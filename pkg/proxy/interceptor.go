package proxy

import (
	"bytes"
	"io"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"

	"github.com/kybernis/kybernis-audit/pkg/evaluator"
	"github.com/kybernis/kybernis-audit/pkg/scenario"
	"github.com/kybernis/kybernis-audit/pkg/telemetry"
)

type ChaosInterceptor struct {
	targetURL *url.URL
	tracker   *telemetry.Tracker
	config    *scenario.Config
	evaluator *evaluator.Tracker
}

func NewChaosInterceptor(target string, tracker *telemetry.Tracker, cfg *scenario.Config, eval *evaluator.Tracker) (*ChaosInterceptor, error) {
	u, err := url.Parse(target)
	if err != nil {
		return nil, err
	}
	return &ChaosInterceptor{
		targetURL: u,
		tracker:   tracker,
		config:    cfg,
		evaluator: eval,
	}, nil
}

func (c *ChaosInterceptor) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	proxy := httputil.NewSingleHostReverseProxy(c.targetURL)

	proxy.Transport = &chaosTransport{
		transport:   http.DefaultTransport,
		interceptor: c,
	}

	proxy.ServeHTTP(w, r)
}

type chaosTransport struct {
	transport   http.RoundTripper
	interceptor *ChaosInterceptor
}

func (t *chaosTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	// Read and restore the request body so the proxy can forward it
	var reqBody []byte
	if req.Body != nil {
		reqBody, _ = io.ReadAll(req.Body)
		req.Body = io.NopCloser(bytes.NewBuffer(reqBody))
	}

	isTarget := strings.Contains(req.URL.Path, t.interceptor.config.TargetEndpoint)
	injectTimeout := isTarget && t.interceptor.config.FaultInjection == "timeout_after_success"

	faultToInject := ""
	if injectTimeout {
		faultToInject = "timeout"
	}

	// Record request to evaluate semantic drift
	t.interceptor.evaluator.RecordAndEvaluate(req.Method, req.URL.Path, reqBody, faultToInject)

	// Let the backend process the request to simulate a real mutation
	res, err := t.transport.RoundTrip(req)
	if err != nil {
		return nil, err
	}

	// ⚡ Chaos Injection Point ⚡
	if injectTimeout {
		log.Printf("⚠️ [Chaos Proxy] Mutating response! Backend returned %d, but Agent will receive 504 Gateway Timeout.", res.StatusCode)

		t.interceptor.tracker.TrackEvent("chaos_injected", map[string]interface{}{
			"scenario": t.interceptor.config.Name,
			"fault":    "timeout_after_success",
		})

		if res.Body != nil {
			io.Copy(io.Discard, res.Body)
			res.Body.Close()
		}

		res.StatusCode = http.StatusGatewayTimeout
		res.Status = "504 Gateway Timeout"
		res.Body = io.NopCloser(bytes.NewBufferString("Gateway Timeout (Injected by Kybernis Audit)"))
		res.Header = make(http.Header)
		res.Header.Set("Content-Type", "text/plain")

		// Consume the fault injection so retries can succeed
		t.interceptor.config.FaultInjection = "exhausted"
	}

	return res, nil
}

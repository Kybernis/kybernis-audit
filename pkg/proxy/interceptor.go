package proxy

import (
	"bytes"
	"io"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"

	"github.com/kybernis/kybernis-audit/pkg/scenario"
	"github.com/kybernis/kybernis-audit/pkg/telemetry"
)

type ChaosInterceptor struct {
	targetURL *url.URL
	tracker   *telemetry.Tracker
	config    *scenario.Config
}

func NewChaosInterceptor(target string, tracker *telemetry.Tracker, cfg *scenario.Config) (*ChaosInterceptor, error) {
	u, err := url.Parse(target)
	if err != nil {
		return nil, err
	}
	return &ChaosInterceptor{
		targetURL: u,
		tracker:   tracker,
		config:    cfg,
	}, nil
}

func (c *ChaosInterceptor) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	proxy := httputil.NewSingleHostReverseProxy(c.targetURL)

	// Custom transport to inject chaos
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
	// Let the backend process the request to simulate a real mutation
	res, err := t.transport.RoundTrip(req)
	if err != nil {
		return nil, err
	}

	// ⚡ Chaos Injection Point ⚡
	// We matched the target endpoint. Wait for 200 OK, but return a 504 Timeout to the agent!
	if strings.Contains(req.URL.Path, t.interceptor.config.TargetEndpoint) && t.interceptor.config.FaultInjection == "timeout_after_success" {
		log.Printf("⚠️ [Chaos Proxy] Mutating response! Backend returned %d, but Agent will receive 504 Gateway Timeout.", res.StatusCode)

		t.interceptor.tracker.TrackEvent("chaos_injected", map[string]interface{}{
			"scenario": t.interceptor.config.Name,
			"fault":    "timeout_after_success",
		})

		// Swallow the original response body
		if res.Body != nil {
			io.Copy(io.Discard, res.Body)
			res.Body.Close()
		}

		// Craft the chaos response
		res.StatusCode = http.StatusGatewayTimeout
		res.Status = "504 Gateway Timeout"
		res.Body = io.NopCloser(bytes.NewBufferString("Gateway Timeout (Injected by Kybernis Audit)"))
		res.Header = make(http.Header)
		res.Header.Set("Content-Type", "text/plain")
	}

	return res, nil
}

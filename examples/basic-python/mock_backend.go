package main

import (
	"encoding/json"
	"fmt"
	"net/http"
)

func main() {
	http.HandleFunc("/v1/charges", func(w http.ResponseWriter, r *http.Request) {
		var payload map[string]interface{}
		json.NewDecoder(r.Body).Decode(&payload)

		fmt.Printf("\n[Stripe Mock] Received Charge Request: %+v\n", payload)
		fmt.Printf("[Stripe Mock] ✅ Payment Succeeded and Locked in Database.\n")

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"id": "ch_123", "status": "succeeded"}`))
	})

	fmt.Println("💳 Stripe Mock Backend listening on :8080")
	http.ListenAndServe(":8080", nil)
}

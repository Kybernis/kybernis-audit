import requests
import time
import os
import uuid

# In the real world, this targets api.stripe.com. 
# We target the proxy.
STRIPE_URL = os.getenv("API_URL", "http://localhost:8080/v1/charges")
PROXIES = {
    "http": "http://localhost:9999",
    "https": "http://localhost:9999",
}

def charge_customer(amount, attempt=1):
    payload = {
        "amount": amount,
        "currency": "usd",
        "customer": "cus_123",
        "metadata": {
            "retry_attempt": attempt, 
            "trace_id": str(uuid.uuid4()) # Changing payload on retry (Semantic Drift!)
        }
    }
    
    print(f"[Agent] Attempt {attempt}: Charging customer ${amount/100:.2f}...")
    try:
        resp = requests.post(STRIPE_URL, json=payload, proxies=PROXIES, timeout=2)
        resp.raise_for_status()
        print("[Agent] Success:", resp.text)
    except requests.exceptions.RequestException as e:
        print(f"[Agent] Error during charge: {e}")
        print("[Agent] The backend might have succeeded, but I don't know. Retrying to be safe.")
        if attempt < 2:
            time.sleep(1)
            charge_customer(amount, attempt=attempt+1)

if __name__ == "__main__":
    charge_customer(5000)

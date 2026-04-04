# Mock Shopify Semantic Double-Spend Example

This standalone Python script demonstrates the "Semantic Double-Spend" problem with a mock e-commerce refund API.

## What It Shows

1. The agent is told: `Refund user #123`
2. The first mock refund call returns a simulated `504 Gateway Timeout`
3. The backend still commits the refund despite the timeout
4. The agent assumes the refund failed and retries
5. A second refund is processed, creating a duplicate transaction

This is the failure mode Kybernis is designed to prevent when execution is wrapped with the SDK.

## Run It

From the repository root:

```bash
python3 examples/mock-shopify-semantic-double-spend/refund_duplicate_demo.py
```

## Expected Output

You should see:

- the first refund processed by the mock API
- a simulated `504 Gateway Timeout`
- the agent retrying the refund
- two refund records in the final ledger

If the example is behaving correctly, the final output will say `Result: duplicate refund created.`

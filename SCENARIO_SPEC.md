# 📝 Kybernis Audit: Scenario Specification

The `scenario.yaml` file defines the exact parameters, attack vectors, and tool schema required to fuzz your AI agent's backend endpoints.

## Example File

```yaml
name: "Stripe Double-Charge Simulation"
target_url: "http://localhost:3000/api/tools"
tool: "issue_refund"
payload:
  customer_id: "cus_123"
  amount: 50.00
  admin_flag: false
  _reasoning: "User requested refund."
attack_vector: "drift"
variant: "hash_bypass"
idempotency_key_path: "transaction_id"
auth_mutate_path: "admin_flag"
auth_mutate_value: true
delay_ms: 1500
race_count: 5
```

## Top-Level Fields

| Field | Type | Required | Description |
| :--- | :--- | :--- | :--- |
| `name` | string | Yes | Human-readable name of the audit scenario. |
| `target_url` | string | Yes | The full HTTP URL of your agent's tool execution endpoint. |
| `tool` | string | No | The name of the tool being tested (used mostly for logging). |
| `payload` | object | Yes | The baseline JSON payload that represents a "clean" execution of the tool. |
| `attack_vector` | string | Yes | The execution vulnerability to test. Must be one of: `dare`, `ghost`, `drift`, `auth`, `saga`, `race`. |
| `variant` | string | No | The specific mitigation strategy you want to bypass. Defaults to `standard`. See [TAXONOMY.md](./TAXONOMY.md) for available variants per attack vector. |

## Attack-Specific Fields

These fields control the precise behavior of the fuzzer depending on the `attack_vector` chosen.

| Field | Type | Used By | Description |
| :--- | :--- | :--- | :--- |
| `idempotency_key_path` | string | `dare`, `drift`, `race` | The JSON key in the payload where the LLM is expected to inject the UUID or idempotency token. E.g., `transaction_id`. The fuzzer dynamically mutates this field. |
| `auth_mutate_path` | string | `auth` | The JSON key in the payload that represents a sensitive or protected field you want to mutate *after* authorization (e.g., `admin_flag` or `amount`). |
| `auth_mutate_value` | any | `auth` | The value to inject into `auth_mutate_path` during the attack. E.g., `true` or `9999.99`. |
| `delay_ms` | integer | `dare`, `drift` | How many milliseconds the fuzzer should wait between the baseline request and the malicious retry request. Defaults to `1000`. |
| `race_count` | integer | `race` | The number of concurrent threads/goroutines to spawn to test distributed locks. Defaults to `5`. |

# 📝 Kybernis Audit: Scenario Specification

The `scenario.yaml` file defines the exact parameters, attack vectors, and tool schema required to fuzz your AI agent's backend endpoints. You can define multiple scenarios in a single file to execute a **Bulk Audit**.

## Example File

```yaml
name: "Cymbal Customer Service Full Audit"
target_url: "http://localhost:3000/api/tools" # Global target (optional)

scenarios:
  - name: "Test DARE with Blind Retry"
    tool: "schedule_planting"
    payload: { "customer_id": "123", "date": "2026-04-01" }
    idempotency_key_path: "idempotency_key"
    attack_vector: "dare"
    variant: "immediate"

  - name: "Test DRIFT with Hash Bypass"
    tool: "schedule_planting"
    payload: { "customer_id": "123", "date": "2026-04-01" }
    idempotency_key_path: "idempotency_key"
    attack_vector: "drift"
    variant: "hash_bypass"
    delay_ms: 1500

  - name: "Test RACE with Concurrent Requests"
    tool: "schedule_planting"
    payload: { "customer_id": "123", "date": "2026-04-01" }
    idempotency_key_path: "idempotency_key"
    attack_vector: "race"
    variant: "simultaneous"
    race_count: 5
```

## Top-Level Fields

| Field | Type | Required | Description |
| :--- | :--- | :--- | :--- |
| `name` | string | Yes | Human-readable name of the bulk audit scenario. |
| `target_url` | string | No | The global HTTP URL of your agent's tool execution endpoint. Applied to all scenarios unless overridden. |
| `scenarios` | list | Yes | A list of scenario objects to execute sequentially. |

## Scenario Fields

These fields control the precise behavior of the fuzzer depending on the `attack_vector` chosen.

| Field | Type | Used By | Description |
| :--- | :--- | :--- | :--- |
| `name` | string | All | Name of the individual test. |
| `target_url` | string | All | Overrides the global target URL for this specific test. |
| `payload` | object | All | The baseline JSON payload that represents a "clean" execution of the tool. |
| `attack_vector` | string | All | The execution vulnerability to test. Must be: `dare`, `ghost`, `drift`, `auth`, `saga`, `race`. |
| `variant` | string | All | The specific mitigation strategy to bypass. Defaults to `standard`. See [TAXONOMY.md](./TAXONOMY.md) for options. |
| `idempotency_key_path` | string | `dare`, `drift`, `race` | The JSON key in the payload where the LLM is expected to inject the UUID. E.g., `transaction_id`. |
| `auth_mutate_path` | string | `auth` | The JSON key in the payload to mutate *after* authorization (e.g., `admin_flag` or `amount`). |
| `auth_mutate_value` | any | `auth` | The value to inject into `auth_mutate_path` during the attack. E.g., `true` or `9999.99`. |
| `delay_ms` | integer | `dare`, `drift` | Milliseconds to wait between the baseline request and the malicious retry request. Defaults to `1000`. |
| `race_count` | integer | `race` | Number of concurrent threads to spawn to test distributed locks. Defaults to `5`. |

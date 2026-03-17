# Kybernis Audit 🛡️⚡

**The open-source deterministic execution fuzzer for AI Agents.**

Find duplicate mutations, retry hazards, and ambiguous execution paths before they hit production. `kybernis-audit` is a local testing engine that deterministically fuzzes your agent's tools. It simulates backend failures, network timeouts, and concurrent reasoning paths to prove whether your infrastructure can survive a stateless LLM hallucinating a duplicate action (like a double-spend).

## 📖 The Agent Execution Taxonomy & Attack Variants

LLMs are intentionally stateless. While the industry categorizes agent failures by *cognitive* mistakes (hallucinations, prompt drift, statelessness), **Kybernis Audit categorizes them by their *execution* impact on your infrastructure.**

| Vulnerability Class | Description | Fuzzer Variants (`variant`) | Community Mitigation |
| :--- | :--- | :--- | :--- |
| **1. DARE** <br>*(Duplicate Action Replay)* | Agent blindly retries the exact same payload on a timeout. | `immediate`<br>`delayed`<br>`param_fuzz` | Exponential backoff, Redis Token Buckets |
| **2. GHOST** <br>*(Ghost Execution)* | Agent assumes failure on timeout, but backend succeeded. | `pre_commit`<br>`post_commit` | Async Webhooks, Polling `GET /status` |
| **3. DRIFT** <br>*(Semantic Drift)* | Agent alters the payload (e.g., hallucinates new UUID) on retry. | `idempotency_regen`<br>`hash_bypass`<br>`semantic_equivalence` | Pydantic Schema, SHA-256 Hashing |
| **4. AUTH** <br>*(Context Drift)* | Prompt injection alters an approved payload post-authorization. | `privilege_escalation`<br>`scope_creep`<br>`confused_deputy` | LlamaGuard, Pre-approval State Hashing |
| **5. SAGA** <br>*(Shattered Saga)* | Multi-step tool workflow crashes mid-execution without rollback. | `mid_execution_crash`<br>`failed_compensation` | Temporal, AWS Step Functions, DLQs |
| **6. RACE** <br>*(Parallel Race)* | Concurrent reasoning branches fire the same tool simultaneously. | `simultaneous`<br>`staggered` | DB Row Locks, Redis Redlocks |

> **📚 The Literature & Case Studies:** For exhaustive details on the mechanics of every attack variant, industry case studies (OpenAI, Vectara, Anthropic), and references, see [**TAXONOMY.md**](./TAXONOMY.md).

## Quick Start

### 1. Install

```bash
curl -sSL https://raw.githubusercontent.com/kybernis/kybernis-audit/main/install.sh | bash
```

### 2. Create a Chaos Scenario
Create a `scenario.yaml` file defining the tool to test and the vulnerability to simulate:

```yaml
name: "Stripe Double-Charge Simulation"
target:
  base_url: "http://localhost:3000/api/tools"
scenarios:
  - name: verify_stripe_refund_safety
    tool: issue_refund
    payload: { "user_id": "123", "amount": 50.00 }
    attack_vector: drift
    variant: hash_bypass
    idempotency_key_path: "transaction_id"
    delay_ms: 2000
    assert:
      is_idempotent: true
```

### 3. Fuzz Your Infrastructure
Run the fuzzer against your local tool endpoint to see if it survives the simulated agent hallucination.

```bash
kybernis-audit fuzz --config=scenario.yaml
```

## How It Works
Instead of wasting tokens hoping an LLM will hallucinate, Kybernis Audit acts as a deterministic **Agent Fuzzer**. 
1. It reads your tool's payload schema.
2. It explicitly simulates a known execution vulnerability against your backend API.
3. If your infrastructure processes the duplicate or partial execution, the audit fails, exposing a critical vulnerability.

## The Production Fix
Need enforcement in production? [Kybernis SDK](https://kybernis.io) adds deterministic execution control, pessimistic semantic locks, and Human-in-the-Loop authorization. We anchor your agent to a persistent session ID, blocking DRIFT and RACE attacks at the infrastructure level. **Diagnose with Audit; cure with the SDK.**

## Contributing
Want to add new attack variants or community mitigations? Read our [Contributing Guide](./CONTRIBUTING.md).

## Telemetry
This tool collects anonymous usage data to help us understand how developers test their agents. It does not collect any payload data, agent code, or PII. You can disable this by setting `KYBERNIS_TELEMETRY=0`.

## License
MIT License

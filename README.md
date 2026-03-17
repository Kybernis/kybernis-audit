# Kybernis Audit 🛡️⚡

**The open-source deterministic execution fuzzer for AI Agents.**

Find duplicate mutations, retry hazards, and ambiguous execution paths before they hit production. `kybernis-audit` is a local testing engine that deterministically fuzzes your agent's tools. It simulates backend failures, network timeouts, and concurrent reasoning paths to prove whether your infrastructure can survive a stateless LLM hallucinating a duplicate action (like a double-spend).

## 📖 The Agent Execution Taxonomy

LLMs are intentionally stateless. While the industry categorizes agent failures by *cognitive* mistakes (hallucinations, prompt drift, statelessness), **Kybernis Audit categorizes them by their *execution* impact on your infrastructure.**

When a cognitive LLM failure crosses the wire to hit your API, it manifests as one of these 6 structural vulnerabilities. Kybernis Audit deterministically fuzzes your tools to detect them:

1. **DARE (Duplicate Action Replay)** - Blind retry
2. **GHOST (Ghost Execution)** - Ambiguous outcome on timeout
3. **DRIFT (Semantic Drift)** - Payload mutation on retry
4. **AUTH (Authorization Context Drift)** - Post-approval mutation
5. **SAGA (Shattered Saga)** - Mid-execution failure without rollback
6. **RACE (Parallel Duplicate Race)** - Concurrent executions

> **Read the full breakdown:** See [TAXONOMY.md](./TAXONOMY.md) for exhaustive details on every attack variant we simulate, the industry case studies (OpenAI, Vectara, etc.), and the community mitigations.

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
    attack_vector:
      type: drift # Semantic Idempotency Bypass
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
2. It explicitly simulates a known execution vulnerability (DARE, GHOST, DRIFT, RACE) against your backend API.
3. If your infrastructure processes the duplicate or partial execution, the audit fails, exposing a critical vulnerability.

## The Production Fix
Need enforcement in production? [Kybernis SDK](https://kybernis.dev) adds deterministic execution control, pessimistic semantic locks, and Human-in-the-Loop authorization. We anchor your agent to a persistent session ID, blocking DRIFT and RACE attacks at the infrastructure level. **Diagnose with Audit; cure with the SDK.**

## Telemetry
This tool collects anonymous usage data to help us understand how developers test their agents. It does not collect any payload data, agent code, or PII. You can disable this by setting `KYBERNIS_TELEMETRY=0`.

## License
MIT License

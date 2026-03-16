# Kybernis Audit 🛡️⚡

**The open-source chaos engineering and risk scanner for AI Agents.**

Find duplicate mutations, retry hazards, and ambiguous execution paths before they hit production. `kybernis-audit` is a local testing engine that stress-tests your LLM agents. It intentionally injects API timeouts, alters context windows, and simulates backend failures to see if your agent will hallucinate a duplicate action (like a double-spend).

*“Semgrep for Agent Side-Effects.”*

## The Agent Execution Taxonomy
Kybernis Audit detects the following execution vulnerabilities:

* **DRIFT (Semantic Drift on Retry):** The agent experienced a network failure, retried the API call, but changed the payload (e.g., generating a new transaction ID). Standard backend idempotency keys fail. **Severity: CRITICAL.**
* **DART (Duplicate Action, Resubmitted Transaction):** The agent retried the exact same payload blindly. If the first attempt actually succeeded but timed out on the wire, this will cause a duplicate mutation. **Severity: WARNING.**

## Quick Start

### 1. Install

```bash
curl -sSL https://raw.githubusercontent.com/kybernis/kybernis-audit/main/install.sh | bash
```

### 2. Create a Chaos Scenario
Create a `scenario.yaml` file defining where to inject the fault:

```yaml
name: "Stripe Double-Charge Simulation"
target_endpoint: "/v1/charges"
fault_injection: "timeout_after_success"
invariant: "no_duplicate_mutations"
```

### 3. Run Your Agent
Run your agent through the `kybernis-audit` wrapper. It will automatically inject a proxy to intercept traffic, inject the timeout, and evaluate your agent's reaction.

```bash
kybernis-audit run --scenario=scenario.yaml -- python your_agent.py
```

## How It Works
1. `kybernis-audit` spins up a local proxy in the background.
2. It routes your agent's API calls through the proxy.
3. When it sees `POST /v1/charges`, it forwards it to your mock backend (letting the database commit).
4. **The Fault:** Instead of returning the `200 OK` to the agent, the proxy swallows it and returns a `504 Gateway Timeout`.
5. **The Evaluation:** The proxy watches to see what your agent does next. If it retries and alters the payload (DRIFT), the proxy flags the critical vulnerability and outputs a report.

## The Production Fix
Need enforcement in production? [Kybernis Cloud](https://kybernis.com) adds deterministic execution control, pessimistic semantic locks, and Human-in-the-Loop authorization for agent-triggered mutations. **Diagnose with Audit; cure with Cloud.**

## Telemetry
This tool collects anonymous usage data to help us understand how developers test their agents. It does not collect any payload data, agent code, or PII. You can disable this by setting `KYBERNIS_TELEMETRY=0`.

## License
MIT License

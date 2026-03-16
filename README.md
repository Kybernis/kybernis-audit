# Kybernis Audit 🛡️⚡

**The open-source chaos engineering and risk scanner for AI Agents.**

Find duplicate mutations, retry hazards, and ambiguous execution paths before they hit production. `kybernis-audit` is a local testing engine that stress-tests your LLM agents. It intentionally injects API timeouts, alters context windows, and simulates backend failures to see if your agent will hallucinate a duplicate action (like a double-spend).

*“Semgrep for Agent Side-Effects.”*

## The Agent Execution Taxonomy
Kybernis Audit detects the following execution vulnerabilities based on the official Kybernis Agent Failure Taxonomy:

1. **DARE (Duplicate Action Replay / Blind Retry)**
   The agent retries the exact same payload blindly. If the first attempt actually succeeded but timed out on the wire, this will cause a duplicate mutation.
2. **GHOST (Ghost Execution / Ambiguous Outcome)**
   The agent fires a mutation, gets a network timeout, and assumes it failed without verifying. The backend actually succeeded, leaving the agent's mental model out of sync with reality.
3. **DRIFT (Semantic Drift on Retry)**
   The agent experienced a network failure, retries the API call, but changes the payload (e.g., generating a new transaction ID). Standard backend idempotency keys fail.
4. **AUTH (Authorization Context Drift)**
   An execution is approved by a human, but the agent alters the payload or context *after* the approval and before the execution.
5. **SAGA (Shattered Saga / Incomplete Execution)**
   An agent completes step 1 of a multi-step operation but crashes or fails before step 2, leaving the system in a corrupted partial state.
6. **RACE (Parallel Duplicate Race)**
   Two parallel agents or concurrent reasoning branches arrive at the same conclusion and fire the exact same tool simultaneously.

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

# Kybernis Audit 🛡️⚡

**The open-source deterministic execution fuzzer for AI Agents.**

Find duplicate mutations, retry hazards, and ambiguous execution paths before they hit production. `kybernis-audit` is a local testing engine that deterministically fuzzes your agent's tools. It simulates backend failures, network timeouts, and concurrent reasoning paths to prove whether your infrastructure can survive a stateless LLM hallucinating a duplicate action (like a double-spend).

## 📖 The Agent Execution Taxonomy

LLMs are intentionally stateless. While the industry categorizes agent failures by *cognitive* mistakes (hallucinations, prompt drift, statelessness), **Kybernis Audit categorizes them by their *execution* impact on your infrastructure.**

When a cognitive LLM failure crosses the wire to hit your API, it manifests as one of these 6 structural vulnerabilities. Kybernis Audit deterministically fuzzes your tools to detect them:

#### 1. DARE (Duplicate Action Replay / Blind Retry)
The agent retries the exact same payload blindly. If the first attempt actually succeeded but timed out on the wire, this causes a duplicate mutation.
*   **Industry Context:** Known in [Vectara's Awesome Agent Failures](https://github.com/vectara/awesome-agent-failures) as *"Verification & Termination Failures"*. Agents get stuck and blindly hammer tools, leading to massive token burn and infrastructure strain.

#### 2. GHOST (Ghost Execution / Ambiguous Outcome)
The agent fires a mutation, gets a network timeout, and assumes it failed without verifying. The backend actually succeeded, leaving the agent's mental model completely out of sync with reality.
*   **Industry Context:** Frequently reported on community forums like [r/LocalLLaMA](https://www.reddit.com/r/LocalLLaMA/comments/1r41h6v/how_do_you_handle_agent_loops_and_cost_overruns/) as *"Tool Output Hallucination"*. Because LLMs are stateless, they easily misinterpret a `504 Gateway Timeout` as a successful execution or vice-versa.

#### 3. DRIFT (Semantic Drift on Retry)
The agent experienced a network failure, retries the API call, but changes the payload (e.g., generating a new transaction ID). Standard backend idempotency keys fail.
*   **Industry Context:** Highlighted in the [OpenAI Production Best Practices](https://developers.openai.com/api/docs/guides/production-best-practices) regarding idempotency. Because an LLM cannot natively remember the UUID it generated 3 seconds ago, it hallucinates a new one on retry, bypassing standard API idempotency checks and double-charging customers.

#### 4. AUTH (Authorization Context Drift)
An execution is approved by a human, but the agent alters the payload or context after the approval and before the execution.
*   **Industry Context:** Categorized under *Prompt Injection* and *Security Bypasses* by the [OWASP Top 10 for LLMs](https://owasp.org/www-project-top-10-for-large-language-model-applications/). Context degradation causes the final executed payload to differ from what the Human-in-the-Loop originally authorized.

#### 5. SAGA (Shattered Saga / Incomplete Execution)
An agent completes step 1 of a multi-step operation but crashes or fails before step 2, leaving the system in a corrupted partial state. Because agents lack a native State-Machine Architecture, they cannot reliably orchestrate cross-service rollbacks when a multi-tool chain fails halfway through.

#### 6. RACE (Parallel Duplicate Race)
Two parallel agents or concurrent reasoning branches arrive at the same conclusion and fire the exact same tool simultaneously. Without distributed locks (like an "in-flight lease"), concurrent agents will bypass standard API checks and execute duplicate mutations at the exact same millisecond.

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

## License
MIT License

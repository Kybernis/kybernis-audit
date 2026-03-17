# The Agent Execution Taxonomy & Attack Variants

LLMs are intentionally stateless. While the industry categorizes agent failures by *cognitive* mistakes (hallucinations, prompt drift, statelessness), **Kybernis Audit categorizes them by their *execution* impact on your infrastructure.**

When a cognitive LLM failure crosses the wire to hit your API, it manifests as one of these 6 structural vulnerabilities. Kybernis Audit deterministically fuzzes your tools to detect them.

---

## 1. DARE (Duplicate Action Replay / Blind Retry)
The agent retries the exact same payload blindly. If the first attempt actually succeeded but timed out on the wire, this causes a duplicate mutation.
*   **Industry Context:** Known in [Vectara's Awesome Agent Failures](https://github.com/vectara/awesome-agent-failures) as *"Verification & Termination Failures"*. Agents get stuck and blindly hammer tools, leading to massive token burn and infrastructure strain.

### Execution Attack Variants Simulated
*   `variant: immediate` (Fires blind duplicate instantly)
*   `variant: delayed` (Waits X ms to bypass rate limiters or cache expiration)
*   `variant: param_fuzz` (Adds dummy variables like `_retry_count` to bypass strict exact-match API firewalls)
*   🤝 **Community Mitigation:** Exponential backoff, Redis Token Buckets, strict LLM loop limits.

---

## 2. GHOST (Ghost Execution / Ambiguous Outcome)
The agent fires a mutation, gets a network timeout, and assumes it failed without verifying. The backend actually succeeded, leaving the agent's mental model completely out of sync with reality.
*   **Industry Context:** Frequently reported on community forums like [r/LocalLLaMA](https://www.reddit.com/r/LocalLLaMA/comments/1r41h6v/how_do_you_handle_agent_loops_and_cost_overruns/) as *"Tool Output Hallucination"*. Because LLMs are stateless, they easily misinterpret a `504 Gateway Timeout` as a successful execution or vice-versa.

### Execution Attack Variants Simulated
*   `variant: pre_commit` (Simulates network disconnect *before* backend processes mutation)
*   `variant: post_commit` (Simulates timeout *after* backend processing, but before agent gets the 200 OK)
*   🤝 **Community Mitigation:** Async Webhooks, polling mechanisms (e.g. `GET /status`) instead of synchronous HTTP calls.

---

## 3. DRIFT (Semantic Drift on Retry)
The agent experienced a network failure, retries the API call, but changes the payload (e.g., generating a new transaction ID). Standard backend idempotency keys fail.
*   **Industry Context:** Highlighted in the [OpenAI Production Best Practices](https://developers.openai.com/api/docs/guides/production-best-practices) regarding idempotency. Because an LLM cannot natively remember the UUID it generated 3 seconds ago, it hallucinates a new one on retry, bypassing standard API idempotency checks and double-charging customers.

### Execution Attack Variants Simulated
*   `variant: idempotency_regen` (The classic UUID hallucination on retry)
*   `variant: hash_bypass` (Injects dummy LLM reasoning strings to mutate the JSON and bypass naive SHA-256 payload deduplication)
*   `variant: semantic_equivalence` (Changes integer `100` to float `100.0` or appends spaces to defeat hash equivalence)
*   🤝 **Community Mitigation:** Strict Pydantic/Zod schema enforcement + payload canonicalization before hashing.

---

## 4. AUTH (Authorization Context Drift)
An execution is approved by a human, but the agent alters the payload or context after the approval and before the execution.
*   **Industry Context:** Categorized under *Prompt Injection* and *Security Bypasses* by the [OWASP Top 10 for LLMs](https://owasp.org/www-project-top-10-for-large-language-model-applications/). Context degradation causes the final executed payload to differ from what the Human-in-the-Loop originally authorized.

### Execution Attack Variants Simulated
*   `variant: privilege_escalation` (Mutates an `admin: false` flag to `true` on the final execution)
*   `variant: scope_creep` (Changes a target ID, e.g., transferring funds to Account B instead of Account A)
*   `variant: confused_deputy` (Simulates Cross Plugin Request Forgery where external data overrides approved parameters)
*   🤝 **Community Mitigation:** NeMo Guardrails, LlamaGuard, or cryptographically hashing state pre-approval.

---

## 5. SAGA (Shattered Saga / Incomplete Execution)
An agent completes step 1 of a multi-step operation but crashes or fails before step 2, leaving the system in a corrupted partial state. Because agents lack a native State-Machine Architecture, they cannot reliably orchestrate cross-service rollbacks when a multi-tool chain fails halfway through.

### Execution Attack Variants Simulated
*   `variant: mid_execution_crash` (Simulates an agent crashing halfway through a multi-step tool sequence)
*   `variant: failed_compensation` (Simulates a failed step 2 to test if the backend natively rolls back step 1)
*   🤝 **Community Mitigation:** AWS Step Functions, Temporal workflows, SQS Dead Letter Queues.

---

## 6. RACE (Parallel Duplicate Race)
Two parallel agents or concurrent reasoning branches arrive at the same conclusion and fire the exact same tool simultaneously. Without distributed locks (like an "in-flight lease"), concurrent agents will bypass standard API checks and execute duplicate mutations at the exact same millisecond.

### Execution Attack Variants Simulated
*   `variant: simultaneous` (Fires N requests at the exact same millisecond to test lock absence)
*   `variant: staggered` (Fires requests sequentially with slight delays to catch distributed locks that release prematurely)
*   🤝 **Community Mitigation:** Database Row Locks (`SELECT FOR UPDATE`), Redis Redlocks (Distributed Leases).

---

*For all of the above, the [Kybernis SDK Native State-Machine execution anchoring](https://kybernis.dev) provides a zero-infrastructure alternative to the community mitigations.*

# Kybernis Audit 🛡️⚡

**The open-source chaos engineering and risk scanner for AI Agents.**

Find duplicate mutations, retry hazards, and ambiguous execution paths before they hit production.

`kybernis-audit` is a local proxy and testing engine that stress-tests your LLM agents. It intentionally injects API timeouts, alters context windows, and simulates backend failures to see if your agent will hallucinate a duplicate action (like a double-spend).

*“Semgrep for Agent Side-Effects.”*

## Why?
LLMs are non-deterministic. If your agent calls `issue_refund()` and the Stripe API times out, the agent will retry. Often, it shifts the payload slightly, bypassing standard idempotency keys and executing a catastrophic duplicate mutation.

Kybernis Audit finds these vulnerabilities locally and in CI.

## The Production Fix
Need enforcement in production? [Kybernis Cloud](https://kybernis.com) adds deterministic execution control, pessimistic semantic locks, and Human-in-the-Loop authorization for agent-triggered mutations.

## Telemetry
This tool collects anonymous usage data to help us understand how developers test their agents. It does not collect any payload data, agent code, or PII. You can disable this by setting `KYBERNIS_TELEMETRY=0`.

## License
MIT License

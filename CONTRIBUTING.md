# Contributing to Kybernis Audit 🛡️⚡

We welcome contributions from the community to make Kybernis Audit the definitive execution fuzzer for AI Agents.

## 📂 Project Structure

Before jumping into the code, here is how the repository is organized. Each folder contains its own `README.md` or domain-specific documentation:

*   **`cmd/kybernis-audit/`**: The main CLI entry point. Handles arguments, flags, and bootstrapping the Telemetry Tracker and the Fuzzer engine.
*   **`pkg/fuzzer/`**: The core execution engine (`engine.go`). This is where the deterministic HTTP requests are formulated, delayed, mutated, and fired. **If you want to add a new vulnerability test, you write it here.**
*   **`pkg/scenario/`**: The configuration parser. Responsible for reading `scenario.yaml` and converting it into the `scenario.Config` struct. See [SCENARIO_SPEC.md](./SCENARIO_SPEC.md).
*   **`pkg/telemetry/`**: The anonymous telemetry client (`tracker.go`) that fires `fuzz_completed` events asynchronously.
*   **`TAXONOMY.md`**: The exhaustive literature connecting our 6 attack vectors to public failures and Reddit case studies.

## 🎯 What We Need

### 1. New Fuzzer Variants (`variant`)
Have you observed an LLM failing in a novel way that bypasses standard backend idempotency or rate limits? We want to fuzz it. 
*   **Example:** If an agent sends a payload as a stringified JSON array instead of an object to bypass WAF rules.
*   **Action:** 
    1. Add the variant logic to the appropriate `execute*Attack()` function in `pkg/fuzzer/engine.go`.
    2. Map it to one of the 6 core Vulnerability Classes (DARE, GHOST, DRIFT, AUTH, SAGA, RACE).
    3. Document it in `TAXONOMY.md` and `README.md`.

### 2. Additional Community Mitigations
If you know of a robust backend engineering pattern (e.g., Saga orchestration, Postgres Advisory Locks, etc.) that effectively stops one of the fuzzer variants, we want to include it in the CLI output.
*   **Action:** Update the `f.printMitigation()` calls in `pkg/fuzzer/engine.go` to include the new architecture pattern.

### 3. Case Studies & Literature
If you find a new public post-mortem (Reddit, HackerNews, OWASP, Vectara) detailing an agent execution failure, add it to the `TAXONOMY.md` document and reference it in the CLI output.

## 🛠️ Development Setup

1. **Clone the repository:**
   ```bash
   git clone https://github.com/kybernis/kybernis-audit.git
   cd kybernis-audit
   ```
2. **Install Go dependencies:**
   ```bash
   go mod tidy
   ```
3. **Build the CLI:**
   ```bash
   go build -o bin/kybernis-audit ./cmd/kybernis-audit
   ```
4. **Run a test fuzz:**
   ```bash
   ./bin/kybernis-audit fuzz --config scenario.yaml
   ```

## 📜 Pull Request Process

1. Fork the repo and create your branch from `main`.
2. If you've added code that should be tested, add tests.
3. Update `TAXONOMY.md`, `SCENARIO_SPEC.md`, or `README.md` if you are adding new schemas.
4. Issue a PR with a clear description of the new variant or mitigation.

Thank you for helping us stop semantic double-spends before they hit production! ⚡

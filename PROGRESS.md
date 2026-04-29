# Project Progress & Strategy: Aegis-Pay

## Project Vision
Aegis-Pay is a production-grade, self-healing payment gateway designed to demonstrate senior-level mastery in:
1. **High-Performance Backend:** Golang (Fiber), Idempotency, Circuit Breaking.
2. **AI Orchestration:** LangGraph (Supervisor/Worker pattern) for automated failure recovery.
3. **System Resilience:** Event-driven architecture using Kafka/Redis and Graceful Shutdowns.

## Strategy: "Learn by Doing Deep-Dive"
We follow a strict **Theory -> Deep-Dive -> Implementation** loop:
* **What:** Definition of the concept.
* **Why:** The specific problem it solves in high-scale systems.
* **When:** Appropriate use cases.
* **How:** Implementation strategy and syntax.

---

## Progress Tracker

### Phase 1: Foundation (COMPLETED)
- [x] **Lesson 1: Go Modules & Project Layout.** Initialized `github.com/EkantBajaj/aegis-pay` with industry-standard directory structure (`cmd/`, `internal/`).
- [x] **Lesson 2 & 3: Web Frameworks & Entry Point.** Selected Fiber for its performance and middleware-centric design.
- [x] **Lesson 4: Concurrency & Graceful Shutdown.** Deep-dived into Goroutines, Channels, and OS signal handling. Implemented `main.go` with a `/health` endpoint and signal-based shutdown.

### Phase 2: The Gateway "Fast Path" (IN PROGRESS)
- [ ] **Lesson 5: Idempotency (Next).** Redis-based check to prevent double-charging.
- [ ] **Lesson 6: The Request Context & Error Handling.** How to pass data through middleware and handle financial errors.
- [ ] **Lesson 7: Integration with Mock Providers.** Simulating Stripe/Adyen/PayPal.

### Phase 3: The AI Recovery "Slow Path"
- [ ] **Lesson 8: Event-Driven Architecture.** Kafka/Redis PubSub.
- [ ] **Lesson 9: LangGraph Orchestrator.** Supervisor/Worker pattern.
- [ ] **Lesson 10: State Persistence.** Postgres checkpointer.

---

## How to Resume this Session (Prompt)
*If starting a new session, use the following prompt to bring the agent up to speed:*

> "I am working on the Aegis-Pay project (github.com/EkantBajaj/aegis-pay). We are following a 'Learn by Doing' deep-dive strategy. Please read `aegis-pay/PROGRESS.md` to see what we have covered. We have just completed Lesson 4 (Graceful Shutdown) and are ready to start **Lesson 5: Idempotency**. Please provide the deep-dive explanation for Idempotency (What, Why, When, How) and then let's set up the Redis connection."

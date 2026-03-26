# Durable Execution with Human Approval

Durable execution with human approval built in. What Temporal can't do in 80 lines, AXME does in 4.

Temporal gives you durable execution. But what happens when your workflow needs a human to approve something? Temporal: build custom signal handler + notification system + timeout logic. AXME: one intent with human_approval, reminder, and timeout built in.

> **Alpha** - Built with [AXME](https://github.com/AxmeAI/axme) (AXP Intent Protocol).
> [cloud.axme.ai](https://cloud.axme.ai) - [hello@axme.ai](mailto:hello@axme.ai)

---

## The Problem

You need durable execution with a human approval step in the middle. Your options:

### Temporal (80+ lines, cluster required)

```python
# 1. Define a workflow with determinism constraints
@workflow.defn
class ApprovalWorkflow:
    def __init__(self):
        self.approved = None

    @workflow.run
    async def run(self, change):
        # Run validation activity
        result = await workflow.execute_activity(
            validate_change, change, start_to_close_timeout=timedelta(minutes=5)
        )

        # Wait for human signal (custom signal handler)
        await workflow.wait_condition(lambda: self.approved is not None)

        if not self.approved:
            return {"status": "rejected"}

        # Apply the change
        return await workflow.execute_activity(
            apply_change, change, start_to_close_timeout=timedelta(minutes=10)
        )

    @workflow.signal
    async def approval_signal(self, approved: bool):
        self.approved = approved

# 2. Build a notification service (Temporal doesn't have one)
# 3. Build a UI/API for the human to send the signal
# 4. Deploy and operate a Temporal cluster
# 5. Handle determinism constraints (no random, no time, no network in workflow)
```

### AXME (4 lines, managed service)

```python
intent_id = client.send_intent({
    "intent_type": "intent.infra.change_approval.v1",
    "to_agent": "agent://myorg/production/infra-validator",
    "payload": {"change_type": "database_migration", "target": "prod-postgres-main"},
})
result = client.wait_for(intent_id)
```

---

## Quick Start

### Python

```bash
pip install axme
export AXME_API_KEY="your-key"   # Get one: axme login
```

```python
from axme import AxmeClient, AxmeClientConfig
import os

client = AxmeClient(AxmeClientConfig(api_key=os.environ["AXME_API_KEY"]))

intent_id = client.send_intent({
    "intent_type": "intent.infra.change_approval.v1",
    "to_agent": "agent://myorg/production/infra-validator",
    "payload": {
        "change_type": "database_migration",
        "target": "prod-postgres-main",
        "migration": "add_index_users_email",
        "rollback_plan": "DROP INDEX idx_users_email",
    },
})

print(f"Submitted: {intent_id}")
result = client.wait_for(intent_id)
print(f"Done: {result['status']}")
```

### TypeScript

```bash
npm install @axme/axme
```

```typescript
import { AxmeClient } from "@axme/axme";

const client = new AxmeClient({ apiKey: process.env.AXME_API_KEY! });

const intentId = await client.sendIntent({
  intentType: "intent.infra.change_approval.v1",
  toAgent: "agent://myorg/production/infra-validator",
  payload: {
    changeType: "database_migration",
    target: "prod-postgres-main",
    migration: "add_index_users_email",
  },
});

console.log(`Submitted: ${intentId}`);
const result = await client.waitFor(intentId);
console.log(`Done: ${result.status}`);
```

---

## More Languages

| Language | Directory | Install |
|----------|-----------|---------|
| [Python](python/) | `python/` | `pip install axme` |
| [TypeScript](typescript/) | `typescript/` | `npm install @axme/axme` |
| [Go](go/) | `go/` | `go get github.com/AxmeAI/axme-sdk-go` |

---

## How AXME Compares to Temporal

| | Temporal | AXME |
|---|---|---|
| Lines of code | 80+ | 4 |
| Human approval | Build it yourself (signals + UI) | Built-in (8 task types) |
| Reminders | Build it yourself | Built-in |
| Determinism constraints | Required (no random, no time, no I/O) | None |
| Infrastructure | Deploy and operate a cluster | Managed service |
| Setup time | Days (cluster + workers + UI) | Minutes (SDK + API key) |
| Notification to human | Build it yourself | CLI, email, Slack |
| Timeout + escalation | Build it yourself | Built-in |
| Audit trail | Workflow history (raw events) | Structured (who, when, decision) |
| When to use | Complex long-running workflows | Operations that need human gates |

Temporal is excellent for complex deterministic workflows. AXME is better when you need human approval mid-workflow without the operational overhead.

---

## How It Works

```
+-----------+  send_intent()   +----------------+  validate   +-----------+
|           | ---------------> |                | ----------> |           |
| Initiator |                  |   AXME Cloud   |             | Validator |
|           | <- wait_for() -- |   (platform)   | <- result   |  (agent)  |
|           |                  |                |             |           |
|           |                  |  WAITING for   |             +-----------+
|           |                  |  human approval|
|           |                  |                |  approve    +-----------+
|           |                  |                | <---------- |           |
|           |                  |  - remind 5m   |             |   Human   |
|           | <- COMPLETED --- |  - timeout 1h  |             |   (SRE)   |
|           |                  |  - audit trail |             |           |
+-----------+                  +----------------+             +-----------+
```

1. Initiator submits a **change intent** with details and rollback plan
2. Agent **validates** the change (syntax, availability, risk assessment)
3. Workflow enters **WAITING** state for human approval
4. AXME **notifies** the SRE on-call (CLI, email, Slack)
5. SRE **approves** (5 minutes or 5 hours later - doesn't matter)
6. Workflow **completes** and initiator gets the result

---

## Run the Full Example

### Prerequisites

```bash
# Install CLI (one-time)
curl -fsSL https://raw.githubusercontent.com/AxmeAI/axme-cli/main/install.sh | sh
# Open a new terminal, or run the "source" command shown by the installer

# Log in
axme login

# Install Python SDK
pip install axme
```

### Terminal 1 - submit the scenario

```bash
axme scenarios apply scenario.json
# Note the intent_id in the output
```

### Terminal 2 - start the agent

Get the agent key after scenario apply:

```bash
# macOS
cat ~/Library/Application\ Support/axme/scenario-agents.json | grep -A2 infra-validator-demo

# Linux
cat ~/.config/axme/scenario-agents.json | grep -A2 infra-validator-demo
```

Run in your language of choice:

```bash
# Python
AXME_API_KEY=<agent-key> python agent.py

# TypeScript (requires Node 20+)
cd typescript && npm install
AXME_API_KEY=<agent-key> npx tsx agent.ts

# Go
cd go && go run ./cmd/agent/
```

### Terminal 1 - approve (after agent validates)

```bash
axme tasks approve <intent_id>
```

### Verify

```bash
axme intents get <intent_id>
# lifecycle_status: COMPLETED
```

---

## Related

- [AXME](https://github.com/AxmeAI/axme) - project overview
- [AXP Spec](https://github.com/AxmeAI/axme-spec) - open Intent Protocol specification
- [AXME Examples](https://github.com/AxmeAI/axme-examples) - 20+ runnable examples across 5 languages
- [AXME CLI](https://github.com/AxmeAI/axme-cli) - manage intents, agents, scenarios from the terminal
- [Temporal Alternative Simple Python](https://github.com/AxmeAI/temporal-alternative-simple-python) - simpler Temporal alternative

---

Built with [AXME](https://github.com/AxmeAI/axme) (AXP Intent Protocol).

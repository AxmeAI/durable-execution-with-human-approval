"""
Infrastructure validator agent - validates changes before human approval.

Listens for intents via SSE. Validates an infrastructure change request
(DB migration, config update, etc.) and resumes with validation results.
The workflow then pauses for SRE on-call approval.

Usage:
    export AXME_API_KEY="<agent-key>"
    python agent.py
"""

import os
import sys
import time

sys.stdout.reconfigure(line_buffering=True)

from axme import AxmeClient, AxmeClientConfig


AGENT_ADDRESS = "infra-validator-demo"


def handle_intent(client, intent_id):
    """Validate infrastructure change and resume with results."""
    intent_data = client.get_intent(intent_id)
    intent = intent_data.get("intent", intent_data)
    payload = intent.get("payload", {})
    if "parent_payload" in payload:
        payload = payload["parent_payload"]

    change_type = payload.get("change_type", "unknown")
    target = payload.get("target", "unknown")
    migration = payload.get("migration", "unknown")
    downtime = payload.get("estimated_downtime", "unknown")
    rollback = payload.get("rollback_plan", "none")

    print(f"  Change type: {change_type}")
    print(f"  Target: {target}")
    print(f"  Migration: {migration}")

    print(f"  Checking target availability...")
    time.sleep(1)

    print(f"  Validating migration syntax...")
    time.sleep(1)

    print(f"  Testing rollback plan...")
    time.sleep(1)

    result = {
        "action": "complete",
        "validation": "passed",
        "checks": {
            "target_available": True,
            "migration_syntax": "valid",
            "rollback_tested": True,
            "estimated_downtime": downtime,
            "risk_level": "low" if downtime == "0s" else "medium",
        },
        "validated_at": time.strftime("%Y-%m-%dT%H:%M:%SZ", time.gmtime()),
    }

    client.resume_intent(intent_id, result)
    print(f"  Validation passed. Risk: {result['checks']['risk_level']}")
    print(f"  Workflow now waits for SRE on-call approval.")
    print(f"  To approve: axme tasks approve <intent_id>")


def main():
    api_key = os.environ.get("AXME_API_KEY", "")
    if not api_key:
        print("Error: AXME_API_KEY not set.")
        print("Run the scenario first: axme scenarios apply scenario.json")
        print("Then get the agent key from ~/.config/axme/scenario-agents.json")
        sys.exit(1)

    client = AxmeClient(AxmeClientConfig(api_key=api_key))

    print(f"Agent listening on {AGENT_ADDRESS}...")
    print("Waiting for intents (Ctrl+C to stop)\n")

    for delivery in client.listen(AGENT_ADDRESS):
        intent_id = delivery.get("intent_id", "")
        status = delivery.get("status", "")

        if not intent_id:
            continue

        if status in ("DELIVERED", "CREATED", "IN_PROGRESS"):
            print(f"[{status}] Intent received: {intent_id}")
            try:
                handle_intent(client, intent_id)
            except Exception as e:
                print(f"  Error processing intent: {e}")


if __name__ == "__main__":
    main()

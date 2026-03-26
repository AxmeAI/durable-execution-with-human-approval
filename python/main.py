"""
Durable execution with human approval - Python example.

Submit an infrastructure change intent. The agent validates it,
then the workflow pauses for human SRE approval.
No Temporal cluster, no determinism constraints, no workflow code.

Usage:
    pip install axme
    export AXME_API_KEY="your-key"
    python main.py
"""

import os
from axme import AxmeClient, AxmeClientConfig


def main():
    client = AxmeClient(
        AxmeClientConfig(api_key=os.environ["AXME_API_KEY"])
    )

    # Submit an infrastructure change with human approval gate
    intent_id = client.send_intent(
        {
            "intent_type": "intent.infra.change_approval.v1",
            "to_agent": "agent://myorg/production/infra-validator",
            "payload": {
                "change_type": "database_migration",
                "target": "prod-postgres-main",
                "migration": "add_index_users_email",
                "estimated_downtime": "0s",
                "rollback_plan": "DROP INDEX idx_users_email",
            },
        }
    )
    print(f"Intent submitted: {intent_id}")

    # Wait for completion (validation + human approval)
    print("Watching lifecycle...")
    for event in client.observe(intent_id):
        status = event.get("status", "")
        print(f"  [{status}] {event.get('event_type', '')}")
        if status in ("COMPLETED", "FAILED", "TIMED_OUT", "CANCELLED"):
            break

    intent = client.get_intent(intent_id)
    print(f"\nFinal status: {intent['intent']['lifecycle_status']}")


if __name__ == "__main__":
    main()

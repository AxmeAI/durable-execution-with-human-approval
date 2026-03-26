/**
 * Infrastructure validator agent - TypeScript example.
 *
 * Listens for intents via SSE, validates infrastructure changes,
 * resumes with validation results.
 *
 * Usage:
 *   export AXME_API_KEY="<agent-key>"
 *   npx tsx agent.ts
 */

import { AxmeClient } from "@axme/axme";

const AGENT_ADDRESS = "infra-validator-demo";

async function handleIntent(client: AxmeClient, intentId: string) {
  const intentData = await client.getIntent(intentId);
  const intent = intentData.intent ?? intentData;
  let payload = intent.payload ?? {};
  if (payload.parent_payload) {
    payload = payload.parent_payload;
  }

  const changeType = payload.change_type ?? "unknown";
  const target = payload.target ?? "unknown";
  const migration = payload.migration ?? "unknown";
  const downtime = payload.estimated_downtime ?? "unknown";

  console.log(`  Change type: ${changeType}`);
  console.log(`  Target: ${target}`);
  console.log(`  Migration: ${migration}`);

  console.log(`  Checking target availability...`);
  await new Promise((r) => setTimeout(r, 1000));

  console.log(`  Validating migration syntax...`);
  await new Promise((r) => setTimeout(r, 1000));

  console.log(`  Testing rollback plan...`);
  await new Promise((r) => setTimeout(r, 1000));

  const riskLevel = downtime === "0s" ? "low" : "medium";
  const result = {
    action: "complete",
    validation: "passed",
    checks: {
      target_available: true,
      migration_syntax: "valid",
      rollback_tested: true,
      estimated_downtime: downtime,
      risk_level: riskLevel,
    },
    validated_at: new Date().toISOString(),
  };

  await client.resumeIntent(intentId, result, { ownerAgent: AGENT_ADDRESS });
  console.log(`  Validation passed. Risk: ${riskLevel}`);
  console.log(`  Workflow now waits for SRE on-call approval.`);
}

async function main() {
  const apiKey = process.env.AXME_API_KEY;
  if (!apiKey) {
    console.error("Error: AXME_API_KEY not set.");
    process.exit(1);
  }

  const client = new AxmeClient({ apiKey });

  console.log(`Agent listening on ${AGENT_ADDRESS}...`);
  console.log("Waiting for intents (Ctrl+C to stop)\n");

  for await (const delivery of client.listen(AGENT_ADDRESS)) {
    const intentId = delivery.intent_id;
    const status = delivery.status;

    if (!intentId) continue;

    if (["DELIVERED", "CREATED", "IN_PROGRESS"].includes(status)) {
      console.log(`[${status}] Intent received: ${intentId}`);
      try {
        await handleIntent(client, intentId);
      } catch (e) {
        console.error(`  Error processing intent: ${e}`);
      }
    }
  }
}

main().catch(console.error);

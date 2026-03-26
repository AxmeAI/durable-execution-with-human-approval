/**
 * Durable execution with human approval - TypeScript example.
 *
 * Submit an infrastructure change intent, wait for completion
 * (validation + human approval). No Temporal, no workflow code.
 *
 * Usage:
 *   npm install @axme/axme
 *   export AXME_API_KEY="your-key"
 *   npx tsx main.ts
 */

import { AxmeClient } from "@axme/axme";

async function main() {
  const client = new AxmeClient({ apiKey: process.env.AXME_API_KEY! });

  const intentId = await client.sendIntent({
    intentType: "intent.infra.change_approval.v1",
    toAgent: "agent://myorg/production/infra-validator",
    payload: {
      changeType: "database_migration",
      target: "prod-postgres-main",
      migration: "add_index_users_email",
      estimatedDowntime: "0s",
      rollbackPlan: "DROP INDEX idx_users_email",
    },
  });
  console.log(`Intent submitted: ${intentId}`);

  const result = await client.waitFor(intentId);
  console.log(`Final status: ${result.status}`);
}

main().catch(console.error);

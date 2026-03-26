// Infrastructure validator agent - Go example.
//
// Listens for intents via SSE, validates infrastructure changes,
// resumes with validation results.
//
// Usage:
//
//	export AXME_API_KEY="<agent-key>"
//	go run ./cmd/agent/
package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/AxmeAI/axme-sdk-go/axme"
)

const agentAddress = "infra-validator-demo"

func handleIntent(ctx context.Context, client *axme.Client, intentID string) error {
	intentData, err := client.GetIntent(ctx, intentID, axme.RequestOptions{})
	if err != nil {
		return fmt.Errorf("get intent: %w", err)
	}

	intent, _ := intentData["intent"].(map[string]any)
	if intent == nil {
		intent = intentData
	}
	payload, _ := intent["payload"].(map[string]any)
	if payload == nil {
		payload = map[string]any{}
	}
	if pp, ok := payload["parent_payload"].(map[string]any); ok {
		payload = pp
	}

	changeType, _ := payload["change_type"].(string)
	if changeType == "" {
		changeType = "unknown"
	}
	target, _ := payload["target"].(string)
	if target == "" {
		target = "unknown"
	}
	migration, _ := payload["migration"].(string)
	if migration == "" {
		migration = "unknown"
	}
	downtime, _ := payload["estimated_downtime"].(string)
	if downtime == "" {
		downtime = "unknown"
	}

	fmt.Printf("  Change type: %s\n", changeType)
	fmt.Printf("  Target: %s\n", target)
	fmt.Printf("  Migration: %s\n", migration)

	fmt.Println("  Checking target availability...")
	time.Sleep(1 * time.Second)

	fmt.Println("  Validating migration syntax...")
	time.Sleep(1 * time.Second)

	fmt.Println("  Testing rollback plan...")
	time.Sleep(1 * time.Second)

	riskLevel := "medium"
	if downtime == "0s" {
		riskLevel = "low"
	}

	result := map[string]any{
		"action":     "complete",
		"validation": "passed",
		"checks": map[string]any{
			"target_available":   true,
			"migration_syntax":   "valid",
			"rollback_tested":    true,
			"estimated_downtime": downtime,
			"risk_level":         riskLevel,
		},
		"validated_at": time.Now().UTC().Format(time.RFC3339),
	}

	_, err = client.ResumeIntent(ctx, intentID, result, axme.RequestOptions{})
	if err != nil {
		return fmt.Errorf("resume intent: %w", err)
	}
	fmt.Printf("  Validation passed. Risk: %s\n", riskLevel)
	fmt.Println("  Workflow now waits for SRE on-call approval.")
	return nil
}

func main() {
	apiKey := os.Getenv("AXME_API_KEY")
	if apiKey == "" {
		log.Fatal("Error: AXME_API_KEY not set.")
	}

	client, err := axme.NewClient(axme.ClientConfig{APIKey: apiKey})
	if err != nil {
		log.Fatalf("create client: %v", err)
	}

	ctx := context.Background()

	fmt.Printf("Agent listening on %s...\n", agentAddress)
	fmt.Println("Waiting for intents (Ctrl+C to stop)")

	intents, errCh := client.Listen(ctx, agentAddress, axme.ListenOptions{})

	go func() {
		for err := range errCh {
			log.Printf("Listen error: %v", err)
		}
	}()

	for delivery := range intents {
		intentID, _ := delivery["intent_id"].(string)
		status, _ := delivery["status"].(string)

		if intentID == "" {
			continue
		}

		if status == "DELIVERED" || status == "CREATED" || status == "IN_PROGRESS" {
			fmt.Printf("[%s] Intent received: %s\n", status, intentID)
			if err := handleIntent(ctx, client, intentID); err != nil {
				fmt.Printf("  Error processing intent: %v\n", err)
			}
		}
	}
}

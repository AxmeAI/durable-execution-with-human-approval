// Durable execution with human approval - Go example.
//
// Submit an infrastructure change intent, wait for completion
// (validation + human approval). No Temporal, no workflow code.
//
// Usage:
//
//	export AXME_API_KEY="your-key"
//	go run main.go
package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/AxmeAI/axme-sdk-go/axme"
)

func main() {
	client, err := axme.NewClient(axme.ClientConfig{
		APIKey: os.Getenv("AXME_API_KEY"),
	})
	if err != nil {
		log.Fatalf("create client: %v", err)
	}

	ctx := context.Background()

	intentID, err := client.SendIntent(ctx, map[string]any{
		"intent_type":        "intent.infra.change_approval.v1",
		"to_agent":           "agent://myorg/production/infra-validator",
		"change_type":        "database_migration",
		"target":             "prod-postgres-main",
		"migration":          "add_index_users_email",
		"estimated_downtime": "0s",
		"rollback_plan":      "DROP INDEX idx_users_email",
	}, axme.RequestOptions{})
	if err != nil {
		log.Fatalf("send intent: %v", err)
	}
	fmt.Printf("Intent submitted: %s\n", intentID)

	result, err := client.WaitFor(ctx, intentID, axme.ObserveOptions{})
	if err != nil {
		log.Fatalf("wait: %v", err)
	}
	fmt.Printf("Final status: %v\n", result["status"])
}

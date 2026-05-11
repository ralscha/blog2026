package main

import (
	"encoding/json"
	"fmt"
	"os"
	"time"
)

func main() {
	baseURL := envOrDefault("API_BASE_URL", "http://localhost:8000")
	metricsPort := envOrDefault("METRICS_PORT", "2121")

	c := NewClient(baseURL, metricsPort, 10*time.Second)

	fmt.Println("=== Launch Ideas API Client ===")
	fmt.Printf("Base URL : %s\n", baseURL)
	fmt.Printf("Metrics  : localhost:%s/metrics\n", metricsPort)
	fmt.Println()

	fmt.Println("--- Health ---")
	health, err := c.Health()
	printResult(health, err)
	fmt.Println()

	fmt.Println("--- Alive ---")
	alive, err := c.Alive()
	printResult(alive, err)
	fmt.Println()

	fmt.Println("--- List Ideas (no filter) ---")
	listResp, err := c.ListIdeas(nil)
	printResult(listResp, err)
	fmt.Println()

	fmt.Println("--- List Ideas (stage=beta, prototype, q=release, minHype=2, limit=5) ---")
	minHype := 2
	listFiltered, err := c.ListIdeas(&ListIdeasFilter{
		Stages:  []string{"beta", "prototype"},
		Query:   "release",
		MinHype: &minHype,
		Limit:   5,
	})
	printResult(listFiltered, err)
	fmt.Println()

	fmt.Println("--- Create Idea ---")
	createResp, err := c.CreateIdea(&CreateIdeaRequest{
		Title: "Latency radar",
		Pitch: "Explains slow endpoints before a customer notices.",
		Stage: "beta",
	})
	printResult(createResp, err)
	fmt.Println()

	if createResp != nil {
		ideaID := createResp.Data.Idea.ID

		fmt.Printf("--- Get Idea %d ---\n", ideaID)
		getResp, err := c.GetIdea(ideaID)
		printResult(getResp, err)
		fmt.Println()

		fmt.Printf("--- Boost Idea %d ---\n", ideaID)
		boostResp, err := c.BoostIdea(ideaID)
		printResult(boostResp, err)
		fmt.Println()

		fmt.Printf("--- Delete Idea %d ---\n", ideaID)
		deleted, err := c.DeleteIdea(ideaID)
		if err != nil {
			fmt.Printf("  ERROR: %v\n", err)
		} else {
			fmt.Printf("  Deleted: %v\n", deleted)
		}
		fmt.Println()
	}

	fmt.Println("--- OpenAPI Spec (.well-known/openapi.json) ---")
	openapi, err := c.OpenAPI()
	if err != nil {
		fmt.Printf("  ERROR: %v\n", err)
	} else {
		var prettyJSON any
		if json.Unmarshal(openapi, &prettyJSON) == nil {
			formatted, _ := json.MarshalIndent(prettyJSON, "  ", "  ")
			fmt.Printf("  OpenAPI spec retrieved (%d bytes)\n", len(openapi))
			const previewLines = 15
			lines := splitLines(string(formatted))
			for i, line := range lines {
				if i >= previewLines {
					fmt.Println("  ...")
					break
				}
				fmt.Printf("  %s\n", line)
			}
		} else {
			fmt.Printf("  Raw (%d bytes)\n", len(openapi))
		}
	}
	fmt.Println()

	fmt.Println("--- Swagger UI (.well-known/swagger) ---")
	swagger, err := c.Swagger()
	if err != nil {
		fmt.Printf("  ERROR: %v\n", err)
	} else {
		fmt.Printf("  Swagger UI page retrieved (%d bytes)\n", len(swagger))
	}
	fmt.Println()

	fmt.Println("--- Metrics (Prometheus) ---")
	metrics, err := c.Metrics()
	if err != nil {
		fmt.Printf("  ERROR: %v\n", err)
	} else {
		lines := splitLines(metrics)
		fmt.Printf("  Prometheus metrics retrieved (%d lines)\n", len(lines))
		const previewLines = 10
		for i, line := range lines {
			if i >= previewLines {
				fmt.Println("  ...")
				break
			}
			fmt.Printf("  %s\n", line)
		}
	}
	fmt.Println()

	fmt.Println("--- Get Non-Existent Idea (error demo) ---")
	getResp, err := c.GetIdea(999999)
	printResult(getResp, err)
	fmt.Println()

	fmt.Println("--- List Ideas with invalid limit (error demo) ---")
	invalidLimit := 0
	_, err = c.ListIdeas(&ListIdeasFilter{Limit: invalidLimit})
	if err != nil {
		fmt.Printf("  ERROR: %v\n", err)
	} else {
		fmt.Println("  Unexpected success")
	}
	fmt.Println()
}

func printResult(v any, err error) {
	if err != nil {
		fmt.Printf("  ERROR: %v\n", err)
		return
	}
	formatted, _ := json.MarshalIndent(v, "  ", "  ")
	fmt.Printf("  %s\n", string(formatted))
}

func envOrDefault(key, fallback string) string {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	return value
}

func splitLines(s string) []string {
	lines := []string{}
	start := 0
	for i := 0; i < len(s); i++ {
		if s[i] == '\n' {
			lines = append(lines, s[start:i])
			start = i + 1
		}
	}
	if start < len(s) {
		lines = append(lines, s[start:])
	}
	return lines
}

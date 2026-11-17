package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"agent"
	"agent/internal/mock"
	"agent/tools"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "agent error: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	registry := map[string]agent.Tool{
		"get_weather": tools.WeatherTool{},
		"calculate":   tools.CalculatorTool{},
	}

	model := &mock.Model{}
	initialMsgs := []agent.Message{
		{Role: agent.RoleUser, Content: "What is the weather in Rome, and what is 42.5 + 17.5?"},
	}

	start := time.Now()
	history, err := agent.Run(ctx, model, registry, initialMsgs)
	if err != nil {
		return err
	}

	fmt.Printf("Completed in %v (two 500ms tools ran concurrently!)\n\n", time.Since(start).Round(time.Millisecond))

	for i, msg := range history {
		fmt.Printf("[%d] %-9s | %s", i, msg.Role, msg.Content)
		if len(msg.ToolCalls) > 0 {
			fmt.Printf(" (requested %d tools)", len(msg.ToolCalls))
		}
		if msg.ToolCallID != "" {
			fmt.Printf(" [for %s]", msg.ToolCallID)
		}
		fmt.Println()
	}

	return nil
}

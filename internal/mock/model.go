// Package mock provides a deterministic mock implementation of agent.Model for demos.
package mock

import (
	"context"

	"agent"
)

// Model is a stateful mock agent.Model that simulates a two-turn tool-calling flow.
type Model struct {
	turn int
}

// Complete returns a scripted tool-calling response on the first turn and a final
// summary response on subsequent turns.
func (m *Model) Complete(ctx context.Context, msgs []agent.Message, tools []agent.ToolSpec) (agent.ModelResponse, error) {
	m.turn++

	// Turn 1: Request two tools in parallel (plus an unknown tool to test error recovery).
	if m.turn == 1 {
		return agent.ModelResponse{
			Content: "Let me check the weather in Rome and calculate that sum for you.",
			ToolCalls: []agent.ToolCall{
				{ID: "call_1", Name: "get_weather", Input: `{"city": "Rome"}`},
				{ID: "call_2", Name: "calculate", Input: `{"a": 42.5, "b": 17.5}`},
				{ID: "call_3", Name: "nonexistent_tool", Input: `{}`},
			},
		}, nil
	}

	// Turn 2: Return the final answer after observing tool outputs.
	return agent.ModelResponse{
		Content: "It's currently 18°C and partly cloudy in Rome, and 42.5 + 17.5 equals 60.00.",
	}, nil
}

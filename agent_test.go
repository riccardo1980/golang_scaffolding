package agent

import (
	"context"
	"errors"
	"slices"
	"sync/atomic"
	"testing"
	"time"
)

type stubModel struct {
	complete func(ctx context.Context, msgs []Message, tools []ToolSpec) (ModelResponse, error)
}

func (m *stubModel) Complete(ctx context.Context, msgs []Message, tools []ToolSpec) (ModelResponse, error) {
	return m.complete(ctx, msgs, tools)
}

type stubTool struct {
	desc string
	run  func(ctx context.Context, input string) (string, error)
}

func (t stubTool) Description() string { return t.desc }
func (t stubTool) Run(ctx context.Context, input string) (string, error) {
	return t.run(ctx, input)
}

func TestRun_MultiTurnAndParallelTools(t *testing.T) {
	var inFlight atomic.Int32
	var maxInFlight atomic.Int32

	tools := map[string]Tool{
		"zeta": stubTool{
			desc: "zeta tool",
			run: func(ctx context.Context, input string) (string, error) {
				cur := inFlight.Add(1)
				for {
					prev := maxInFlight.Load()
					if cur <= prev || maxInFlight.CompareAndSwap(prev, cur) {
						break
					}
				}
				time.Sleep(30 * time.Millisecond)
				inFlight.Add(-1)
				return "zeta-ok:" + input, nil
			},
		},
		"alpha": stubTool{
			desc: "alpha tool",
			run: func(ctx context.Context, input string) (string, error) {
				cur := inFlight.Add(1)
				for {
					prev := maxInFlight.Load()
					if cur <= prev || maxInFlight.CompareAndSwap(prev, cur) {
						break
					}
				}
				time.Sleep(30 * time.Millisecond)
				inFlight.Add(-1)
				return "alpha-ok:" + input, nil
			},
		},
		"failing": stubTool{
			desc: "failing tool",
			run: func(ctx context.Context, input string) (string, error) {
				return "", errors.New("boom")
			},
		},
	}

	var turn int
	var gotSpecs []ToolSpec
	model := &stubModel{
		complete: func(ctx context.Context, msgs []Message, specs []ToolSpec) (ModelResponse, error) {
			turn++
			if turn == 1 {
				gotSpecs = slices.Clone(specs)
				return ModelResponse{
					Content: "calling tools",
					ToolCalls: []ToolCall{
						{ID: "1", Name: "alpha", Input: `{"a":1}`},
						{ID: "2", Name: "zeta", Input: `{"z":2}`},
						{ID: "3", Name: "failing", Input: `{}`},
						{ID: "4", Name: "missing", Input: `{}`},
					},
				}, nil
			}
			return ModelResponse{Content: "done"}, nil
		},
	}

	history, err := Run(context.Background(), model, tools, []Message{
		{Role: RoleUser, Content: "hello"},
	})
	if err != nil {
		t.Fatalf("Run returned unexpected error: %v", err)
	}

	// Verify tool specs were sorted deterministically by name.
	wantOrder := []string{"alpha", "failing", "zeta"}
	if len(gotSpecs) != len(wantOrder) {
		t.Fatalf("got %d specs, want %d", len(gotSpecs), len(wantOrder))
	}
	for i, name := range wantOrder {
		if gotSpecs[i].Name != name {
			t.Errorf("spec[%d].Name = %q, want %q", i, gotSpecs[i].Name, name)
		}
	}

	// Verify alpha and zeta ran concurrently.
	if maxInFlight.Load() < 2 {
		t.Errorf("expected parallel tool execution (maxInFlight >= 2), got %d", maxInFlight.Load())
	}

	// Expected history: user, assistant (4 calls), 4 tool results, assistant ("done").
	if len(history) != 7 {
		t.Fatalf("got %d messages, want 7", len(history))
	}

	wantToolContents := map[string]string{
		"1": `alpha-ok:{"a":1}`,
		"2": `zeta-ok:{"z":2}`,
		"3": "ERROR: boom",
		"4": "ERROR: unknown tool",
	}
	for _, msg := range history[2:6] {
		if msg.Role != RoleTool {
			t.Errorf("expected RoleTool, got %q", msg.Role)
		}
		if want := wantToolContents[msg.ToolCallID]; msg.Content != want {
			t.Errorf("tool %s content = %q, want %q", msg.ToolCallID, msg.Content, want)
		}
	}

	if last := history[len(history)-1]; last.Role != RoleAssistant || last.Content != "done" {
		t.Errorf("unexpected final message: %+v", last)
	}
}

func TestRun_Errors(t *testing.T) {
	tests := []struct {
		name    string
		ctx     func() (context.Context, context.CancelFunc)
		model   Model
		wantErr string
	}{
		{
			name:    "nil model",
			ctx:     func() (context.Context, context.CancelFunc) { return context.Background(), func() {} },
			model:   nil,
			wantErr: "agent: model must not be nil",
		},
		{
			name: "model error",
			ctx:  func() (context.Context, context.CancelFunc) { return context.Background(), func() {} },
			model: &stubModel{
				complete: func(ctx context.Context, msgs []Message, tools []ToolSpec) (ModelResponse, error) {
					return ModelResponse{}, errors.New("api failure")
				},
			},
			wantErr: "api failure",
		},
		{
			name: "cancelled context",
			ctx: func() (context.Context, context.CancelFunc) {
				ctx, cancel := context.WithCancel(context.Background())
				cancel()
				return ctx, cancel
			},
			model: &stubModel{
				complete: func(ctx context.Context, msgs []Message, tools []ToolSpec) (ModelResponse, error) {
					return ModelResponse{}, nil
				},
			},
			wantErr: context.Canceled.Error(),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			ctx, cancel := tc.ctx()
			defer cancel()

			_, err := Run(ctx, tc.model, nil, nil)
			if err == nil {
				t.Fatalf("expected error containing %q, got nil", tc.wantErr)
			}
			if err.Error() != tc.wantErr {
				t.Errorf("got error %q, want %q", err.Error(), tc.wantErr)
			}
		})
	}
}

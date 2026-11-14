package ask

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/x/ansi"
	"github.com/coder/websocket"
	"github.com/openai/openai-go/v3/option"
	"github.com/tidwall/gjson"
)

// fakeResponses is a minimal Responses WebSocket server. Each connection keeps
// its own chain cache, as the real service does with store=false.
type fakeResponses struct {
	t *testing.T

	mu          sync.Mutex
	requests    []string
	connections int

	// reply decides what the server sends for a request on a connection.
	reply func(req gjson.Result, cache map[string]bool, id string) []any
	// hold, when set, stops the server after the first delta until closed.
	hold chan struct{}
}

func (f *fakeResponses) handler(w http.ResponseWriter, r *http.Request) {
	conn, err := websocket.Accept(w, r, nil)
	if err != nil {
		f.t.Errorf("accept: %v", err)
		return
	}
	defer conn.CloseNow()

	f.mu.Lock()
	f.connections++
	connNum := f.connections
	f.mu.Unlock()

	cache := map[string]bool{}
	turn := 0
	for {
		_, data, err := conn.Read(r.Context())
		if err != nil {
			return
		}
		turn++

		f.mu.Lock()
		f.requests = append(f.requests, string(data))
		f.mu.Unlock()

		req := gjson.ParseBytes(data)
		id := "resp_" + string(rune('a'+connNum-1)) + string(rune('0'+turn))

		events := f.reply(req, cache, id)
		for i, event := range events {
			payload, _ := json.Marshal(event)
			if err := conn.Write(r.Context(), websocket.MessageText, payload); err != nil {
				return
			}
			if f.hold != nil && i == 0 {
				select {
				case <-f.hold:
				case <-r.Context().Done():
					return
				}
			}
		}
	}
}

func (f *fakeResponses) request(i int) gjson.Result {
	f.mu.Lock()
	defer f.mu.Unlock()
	return gjson.Parse(f.requests[i])
}

// echoReply streams the prompt back in two deltas and remembers the chain.
func echoReply(req gjson.Result, cache map[string]bool, id string) []any {
	if prev := req.Get("previous_response_id"); prev.Exists() && !cache[prev.String()] {
		return []any{map[string]any{
			"type":   "error",
			"status": 400,
			"error": map[string]any{
				"type":    "invalid_request_error",
				"code":    "previous_response_not_found",
				"message": "Previous response with id '" + prev.String() + "' not found.",
				"param":   "previous_response_id",
			},
		}}
	}
	cache[id] = true

	input := req.Get("input").Array()
	last := input[len(input)-1].Get("content").String()

	return []any{
		map[string]any{"type": "response.output_text.delta", "delta": "you said: ", "item_id": "m", "output_index": 0, "content_index": 0, "sequence_number": 1},
		map[string]any{"type": "response.output_text.delta", "delta": last, "item_id": "m", "output_index": 0, "content_index": 0, "sequence_number": 2},
		map[string]any{"type": "response.completed", "sequence_number": 3, "response": map[string]any{
			"id": id, "object": "response", "status": "completed", "model": req.Get("model").String(),
			"output": []any{},
			"usage": map[string]any{
				"input_tokens": 1000, "output_tokens": 500, "total_tokens": 1500,
				"input_tokens_details":  map[string]any{"cached_tokens": 0},
				"output_tokens_details": map[string]any{"reasoning_tokens": 0},
			},
		}},
	}
}

func newFakeSession(t *testing.T, fake *fakeResponses) *Session {
	t.Helper()
	server := httptest.NewServer(http.HandlerFunc(fake.handler))
	t.Cleanup(server.Close)

	session := NewSession("test-key", option.WithBaseURL(server.URL+"/v1/"), option.WithMaxRetries(0))
	t.Cleanup(session.Close)
	return session
}

func testContext(t *testing.T) context.Context {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	t.Cleanup(cancel)
	return ctx
}

func TestSessionStreamsAndChainsTurns(t *testing.T) {
	fake := &fakeResponses{t: t, reply: echoReply}
	session := newFakeSession(t, fake)
	ctx := testContext(t)

	var deltas []string
	result, err := session.Stream(ctx, PromptOptions{
		Model:         "gpt-5",
		Prompt:        "hello",
		SystemMessage: "be nice",
		Settings:      ResponseSettings{ReasoningEffort: "medium", ReasoningMode: "standard", Verbosity: "medium", ServiceTier: "default", Include: []string{"reasoning.encrypted_content"}},
	}, func(d string) { deltas = append(deltas, d) })
	if err != nil {
		t.Fatalf("first turn: %v", err)
	}

	if result.Content != "you said: hello" {
		t.Fatalf("content = %q", result.Content)
	}
	if strings.Join(deltas, "|") != "you said: |hello" {
		t.Fatalf("deltas = %q", deltas)
	}
	if cost, ok := CalculateCost("gpt-5", result.Usage); !ok || cost <= 0 {
		t.Fatalf("cost = %v, %v", cost, ok)
	}

	first := fake.request(0)
	for path, want := range map[string]string{
		"type":             "response.create",
		"model":            "gpt-5",
		"store":            "false",
		"service_tier":     "default",
		"text.verbosity":   "medium",
		"reasoning.effort": "medium",
		"reasoning.mode":   "standard",
		"include.0":        "reasoning.encrypted_content",
	} {
		if got := first.Get(path).String(); got != want {
			t.Errorf("%s = %q, want %q", path, got, want)
		}
	}
	if first.Get("previous_response_id").Exists() {
		t.Error("first turn should not reference a previous response")
	}

	// The second turn sends only the new prompt against the first response.
	history := []MessageHistory{{Role: "user", Content: "hello"}, {Role: "assistant", Content: result.Content}}
	if _, err := session.Stream(ctx, PromptOptions{
		Model:             "gpt-5",
		Prompt:            "again",
		History:           history,
		SystemMessage:     "be nice",
		TurnSystemMessage: "be brief",
	}, nil); err != nil {
		t.Fatalf("second turn: %v", err)
	}

	second := fake.request(1)
	if got := second.Get("previous_response_id").String(); got != "resp_a1" {
		t.Fatalf("previous_response_id = %q", got)
	}
	if n := len(second.Get("input").Array()); n != 1 {
		t.Fatalf("second turn sent %d input items, want 1", n)
	}
	if instructions := second.Get("instructions").String(); !strings.Contains(instructions, "be nice") || !strings.Contains(instructions, "be brief") {
		t.Fatalf("instructions = %q", instructions)
	}
}

func TestSessionReplaysHistoryWhenChainDoesNotMatch(t *testing.T) {
	fake := &fakeResponses{t: t, reply: echoReply}
	session := newFakeSession(t, fake)
	ctx := testContext(t)

	history := []MessageHistory{
		{Role: "user", Content: "earlier"},
		{Role: "assistant", Content: "answer"},
	}
	if _, err := session.Stream(ctx, PromptOptions{Model: "gpt-5", Prompt: "now", History: history}, nil); err != nil {
		t.Fatal(err)
	}

	req := fake.request(0)
	if n := len(req.Get("input").Array()); n != 3 {
		t.Fatalf("sent %d input items, want history plus prompt (3)", n)
	}
	if role := req.Get("input.1.role").String(); role != "assistant" {
		t.Fatalf("input.1.role = %q", role)
	}
}

func TestSessionRecoversFromLostChain(t *testing.T) {
	fake := &fakeResponses{t: t, reply: echoReply}
	session := newFakeSession(t, fake)
	ctx := testContext(t)

	if _, err := session.Stream(ctx, PromptOptions{Model: "gpt-5", Prompt: "one"}, nil); err != nil {
		t.Fatal(err)
	}

	// Pretend the server forgot the chain, as it does after a reconnect.
	session.previousID = "resp_gone"

	history := []MessageHistory{{Role: "user", Content: "one"}, {Role: "assistant", Content: "you said: one"}}
	result, err := session.Stream(ctx, PromptOptions{Model: "gpt-5", Prompt: "two", History: history}, nil)
	if err != nil {
		t.Fatalf("expected recovery, got %v", err)
	}
	if result.Content != "you said: two" {
		t.Fatalf("content = %q", result.Content)
	}

	retry := fake.request(2)
	if retry.Get("previous_response_id").Exists() {
		t.Fatal("retry should start a fresh chain")
	}
	if n := len(retry.Get("input").Array()); n != 3 {
		t.Fatalf("retry sent %d input items, want 3", n)
	}
}

func TestSessionReportsAPIErrors(t *testing.T) {
	fake := &fakeResponses{t: t, reply: func(gjson.Result, map[string]bool, string) []any {
		return []any{map[string]any{"type": "error", "status": 400, "error": map[string]any{"code": "model_not_found", "message": "The model does not exist."}}}
	}}
	session := newFakeSession(t, fake)

	_, err := session.Stream(testContext(t), PromptOptions{Model: "nope", Prompt: "hi"}, nil)
	var apiErr *APIError
	if !errors.As(err, &apiErr) || apiErr.Code != "model_not_found" {
		t.Fatalf("err = %v", err)
	}
	if !strings.Contains(err.Error(), "The model does not exist.") {
		t.Fatalf("message = %q", err.Error())
	}
}

func TestSessionCancelDropsTheSocket(t *testing.T) {
	hold := make(chan struct{})
	fake := &fakeResponses{t: t, reply: echoReply, hold: hold}
	session := newFakeSession(t, fake)

	ctx, cancel := context.WithCancel(testContext(t))
	got := make(chan error, 1)
	go func() {
		_, err := session.Stream(ctx, PromptOptions{Model: "gpt-5", Prompt: "slow"}, func(string) { cancel() })
		got <- err
	}()

	if err := <-got; !errors.Is(err, context.Canceled) {
		t.Fatalf("err = %v, want context.Canceled", err)
	}
	close(hold)

	// The next turn must not read the cancelled turn's leftover events. hold
	// is closed now, so the server no longer pauses.
	result, err := session.Stream(testContext(t), PromptOptions{Model: "gpt-5", Prompt: "fresh"}, nil)
	if err != nil {
		t.Fatal(err)
	}
	if result.Content != "you said: fresh" {
		t.Fatalf("content = %q", result.Content)
	}
	fake.mu.Lock()
	connections := fake.connections
	fake.mu.Unlock()
	if connections != 2 {
		t.Fatalf("connections = %d, want a new socket after cancel", connections)
	}
}

func TestLoadResponseSettingsQuickIsLighter(t *testing.T) {
	quick := LoadResponseSettings(true)
	if quick.ReasoningEffort != "low" || quick.ReasoningMode != "" {
		t.Fatalf("quick settings = %+v", quick)
	}
}

func TestWaitForStreamMergesQueuedDeltas(t *testing.T) {
	ch := make(chan tea.Msg, 4)
	ch <- liveDeltaMsg{id: 1, delta: "a", ch: ch}
	ch <- liveDeltaMsg{id: 1, delta: "b", ch: ch}
	ch <- liveResponseMsg{id: 1, content: "ab"}

	msg := waitForStream(ch)()
	delta, ok := msg.(liveDeltaMsg)
	if !ok {
		t.Fatalf("got %T", msg)
	}
	if delta.delta != "ab" {
		t.Fatalf("delta = %q", delta.delta)
	}
	if _, ok := delta.next.(liveResponseMsg); !ok {
		t.Fatalf("next = %T, want the response that followed", delta.next)
	}
}

func TestStreamingAnswerRendersAndPartialIsNotSent(t *testing.T) {
	m := DefaultLiveChatModel()
	updated, _ := m.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
	m = updated.(LiveChatModel)

	m.messages = []LiveMessage{{Role: LiveRoleUser, Content: "hi"}}
	m.requestID = 1
	m.waiting = true
	m.rebuildTranscript(true)

	updated, _ = m.Update(liveDeltaMsg{id: 1, delta: "streamed words", ch: make(chan tea.Msg)})
	m = updated.(LiveChatModel)
	if !strings.Contains(ansi.Strip(strings.Join(m.transcriptLines, "\n")), "streamed words") {
		t.Fatal("streamed text should be on screen before the turn completes")
	}

	m.cancelRequest()
	if m.streaming != "" {
		t.Fatal("cancel should clear the streaming buffer")
	}
	for _, h := range m.apiHistory() {
		if h.Content == "streamed words" {
			t.Fatal("an interrupted answer must not be sent back to the model")
		}
	}
}

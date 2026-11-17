package ask

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"

	openai "github.com/openai/openai-go/v3"
	"github.com/openai/openai-go/v3/option"
	"github.com/openai/openai-go/v3/responses"
	"github.com/openai/openai-go/v3/shared"
	"github.com/spf13/viper"
	"github.com/tidwall/gjson"
)

// ResponseSettings are the Responses API knobs read from ask.openai.responses.
// They mirror the request body produced by OpenAI's response builder.
type ResponseSettings struct {
	ServiceTier      string
	Verbosity        string
	ReasoningEffort  string
	ReasoningMode    string
	ReasoningSummary string
	Store            bool
	Include          []string
}

// LoadResponseSettings reads the configured settings. Quick turns trade depth
// for speed, so they drop to low effort and leave the reasoning mode to the
// model's default, which also keeps smaller models that lack 'mode' working.
func LoadResponseSettings(quick bool) ResponseSettings {
	s := ResponseSettings{
		ServiceTier:      viper.GetString("ask.openai.responses.service_tier"),
		Verbosity:        viper.GetString("ask.openai.responses.verbosity"),
		ReasoningEffort:  viper.GetString("ask.openai.responses.reasoning.effort"),
		ReasoningMode:    viper.GetString("ask.openai.responses.reasoning.mode"),
		ReasoningSummary: viper.GetString("ask.openai.responses.reasoning.summary"),
		Store:            viper.GetBool("ask.openai.responses.store"),
		Include:          viper.GetStringSlice("ask.openai.responses.include"),
	}

	if quick {
		s.ReasoningEffort = "low"
		s.ReasoningMode = ""
		s.Verbosity = "low"
	}

	return s
}

// TurnResult is the outcome of one streamed turn.
type TurnResult struct {
	Content string
	Usage   responses.ResponseUsage
	// Incomplete is set when the response stopped early, for example at the
	// output token limit. Content holds whatever arrived.
	Incomplete string
}

// Session owns one Responses WebSocket and the response chain on it. Turns on
// the same session are sent as deltas against the previous response, so a
// long conversation does not re-upload its history every turn.
//
// With store=false the server only remembers the chain while the socket is
// open (and for at most 60 minutes), so the session falls back to replaying
// the full history whenever that cache is gone.
type Session struct {
	apiKey  string
	options []option.RequestOption

	mu         sync.Mutex
	conn       *responses.ResponseConnection
	previousID string
	// chainLen is how many History messages previousID already covers.
	chainLen int
}

// NewSession prepares a session. The socket opens lazily on the first turn.
func NewSession(apiKey string, options ...option.RequestOption) *Session {
	return &Session{apiKey: apiKey, options: options}
}

// Close releases the socket. The session can still be used afterwards; the
// next turn reconnects and replays history.
func (s *Session) Close() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.dropLocked(false)
}

// Stream sends one turn and calls onDelta with each piece of output text as it
// arrives. onDelta runs on the calling goroutine.
func (s *Session) Stream(ctx context.Context, opts PromptOptions, onDelta func(string)) (*TurnResult, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if onDelta == nil {
		onDelta = func(string) {}
	}

	emitted := false
	track := func(delta string) {
		emitted = true
		onDelta(delta)
	}

	result, err := s.streamLocked(ctx, opts, track)
	if err == nil || emitted || ctx.Err() != nil || !isRecoverable(err) {
		return result, err
	}

	// The server lost the chain or the socket died before any output arrived,
	// so retrying on a fresh socket with the whole history cannot duplicate
	// text the caller has already shown.
	s.dropLocked(true)
	return s.streamLocked(ctx, opts, track)
}

func (s *Session) streamLocked(ctx context.Context, opts PromptOptions, onDelta func(string)) (*TurnResult, error) {
	if s.conn == nil {
		client := openai.NewClient(append([]option.RequestOption{option.WithAPIKey(s.apiKey)}, s.options...)...)
		conn, err := client.Responses.Connect(ctx, responses.ResponseConnectionOptions{})
		if err != nil {
			return nil, fmt.Errorf("failed to connect to OpenAI: %w", err)
		}
		s.conn = conn
		s.previousID = ""
		s.chainLen = 0
	}

	params := s.buildParams(opts)
	if err := s.conn.Create(ctx, params); err != nil {
		s.dropLocked(true)
		return nil, &connectionError{err: err}
	}

	var content strings.Builder
	for {
		event, err := s.conn.Recv(ctx)
		if err != nil {
			// A cancelled or broken read leaves the server mid-response, and
			// its remaining events would leak into the next turn, so the
			// socket cannot be reused.
			s.dropLocked(true)
			if ctx.Err() != nil {
				return nil, ctx.Err()
			}
			return nil, &connectionError{err: err}
		}

		switch event.Type {
		case "response.output_text.delta":
			delta := event.AsResponseOutputTextDelta().Delta
			content.WriteString(delta)
			onDelta(delta)

		case "response.completed", "response.incomplete":
			response := terminalResponse(event)
			result := &TurnResult{Content: content.String()}
			if response != nil {
				result.Usage = response.Usage
				if response.Status == responses.ResponseStatusIncomplete {
					result.Incomplete = string(response.IncompleteDetails.Reason)
					if result.Incomplete == "" {
						result.Incomplete = "incomplete"
					}
				}
				if result.Content == "" {
					result.Content = response.OutputText()
				}
				s.previousID = response.ID
				// The chain now covers the history, the prompt and the answer.
				s.chainLen = len(opts.History) + 2
			}
			return result, nil

		case "response.failed":
			s.resetChainLocked()
			message := "response failed"
			if response := terminalResponse(event); response != nil && response.Error.Message != "" {
				message = response.Error.Message
			}
			return nil, fmt.Errorf("OpenAI: %s", message)

		case "error":
			apiErr := newAPIError(event.RawJSON())
			// A failed continuation evicts its parent from the server cache.
			s.resetChainLocked()
			if apiErr.Code == "websocket_connection_limit_reached" {
				s.dropLocked(true)
			}
			return nil, apiErr
		}
	}
}

func (s *Session) buildParams(opts PromptOptions) responses.ResponsesClientEventResponseCreateParam {
	settings := opts.Settings

	params := responses.ResponsesClientEventResponseCreateParam{
		Model:        shared.ResponsesModel(opts.Model),
		Instructions: openai.String(buildInstructions(opts)),
		Store:        openai.Bool(settings.Store),
		ServiceTier:  settings.ServiceTier,
	}

	if settings.Verbosity != "" {
		params.Text.Verbosity = responses.ResponseTextConfigVerbosity(settings.Verbosity)
	}
	if settings.ReasoningEffort != "" {
		params.Reasoning.Effort = shared.ReasoningEffort(settings.ReasoningEffort)
	}
	if settings.ReasoningMode != "" {
		params.Reasoning.Mode = shared.ReasoningMode(settings.ReasoningMode)
	}
	if settings.ReasoningSummary != "" {
		params.Reasoning.Summary = shared.ReasoningSummary(settings.ReasoningSummary)
	}
	for _, include := range settings.Include {
		params.Include = append(params.Include, responses.ResponseIncludable(include))
	}

	var input responses.ResponseInputParam
	if s.previousID != "" && s.chainLen == len(opts.History) {
		// The server already holds every earlier turn.
		params.PreviousResponseID = openai.String(s.previousID)
	} else {
		input = historyInput(opts.History)
	}
	input = append(input, responses.ResponseInputItemParamOfMessage(opts.Prompt, responses.EasyInputMessageRoleUser))
	params.Input.OfResponse = &input

	return params
}

// buildInstructions joins every system-level message. Instructions are not
// carried along a previous_response_id chain, so per-turn instructions (quick
// mode) cost tokens only on the turn that sends them.
func buildInstructions(opts PromptOptions) string {
	var parts []string

	if opts.SystemMessage != "" {
		systemMessage := opts.SystemMessage
		if languages := readPreferredLanguages(); len(languages) > 0 {
			systemMessage = fmt.Sprintf("%s\n\nPreferred programming languages: %s.", systemMessage, strings.Join(languages, ", "))
		}
		parts = append(parts, strings.TrimSpace(systemMessage))
	}

	if userDefined := viper.GetString("ask.system_message"); userDefined != "" {
		parts = append(parts, UserDefinedSystemMessagePrefix+userDefined)
	}

	if opts.TurnSystemMessage != "" {
		parts = append(parts, opts.TurnSystemMessage)
	}

	return strings.Join(parts, "\n\n")
}

func historyInput(history []MessageHistory) responses.ResponseInputParam {
	input := make(responses.ResponseInputParam, 0, len(history))
	for _, msg := range history {
		var role responses.EasyInputMessageRole
		switch msg.Role {
		case "user":
			role = responses.EasyInputMessageRoleUser
		case "assistant":
			role = responses.EasyInputMessageRoleAssistant
		case "system":
			role = responses.EasyInputMessageRoleDeveloper
		default:
			continue
		}
		input = append(input, responses.ResponseInputItemParamOfMessage(msg.Content, role))
	}
	return input
}

func terminalResponse(event responses.ResponsesServerEventUnion) *responses.Response {
	raw := gjson.Get(event.RawJSON(), "response")
	if !raw.Exists() {
		return nil
	}
	var response responses.Response
	if err := response.UnmarshalJSON([]byte(raw.Raw)); err != nil {
		return nil
	}
	return &response
}

func (s *Session) resetChainLocked() {
	s.previousID = ""
	s.chainLen = 0
}

func (s *Session) dropLocked(abort bool) {
	if s.conn != nil {
		if abort {
			s.conn.Abort()
		} else {
			_ = s.conn.Close()
		}
	}
	s.conn = nil
	s.resetChainLocked()
}

// APIError is an error event sent over the socket.
type APIError struct {
	Code    string
	Message string
	Param   string
}

func newAPIError(raw string) *APIError {
	// The socket wraps request errors as {"type":"error","error":{...}}, but
	// connection-level errors may put the fields at the top level.
	body := gjson.Get(raw, "error")
	if !body.IsObject() {
		body = gjson.Parse(raw)
	}
	return &APIError{
		Code:    body.Get("code").String(),
		Message: body.Get("message").String(),
		Param:   body.Get("param").String(),
	}
}

func (e *APIError) Error() string {
	switch {
	case e.Message != "" && e.Code != "":
		return fmt.Sprintf("OpenAI: %s (%s)", e.Message, e.Code)
	case e.Message != "":
		return "OpenAI: " + e.Message
	case e.Code != "":
		return "OpenAI: " + e.Code
	default:
		return "OpenAI: unknown error"
	}
}

type connectionError struct{ err error }

func (e *connectionError) Error() string { return "connection to OpenAI lost: " + e.err.Error() }
func (e *connectionError) Unwrap() error { return e.err }

func isRecoverable(err error) bool {
	var connErr *connectionError
	if errors.As(err, &connErr) {
		return true
	}
	var apiErr *APIError
	if errors.As(err, &apiErr) {
		switch apiErr.Code {
		case "previous_response_not_found", "websocket_connection_limit_reached":
			return true
		}
	}
	return false
}

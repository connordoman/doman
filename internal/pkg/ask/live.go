package ask

import (
	"context"
	"fmt"
	"image/color"
	"strings"
	"time"

	"charm.land/bubbles/v2/key"
	"charm.land/bubbles/v2/spinner"
	"charm.land/bubbles/v2/textarea"
	"charm.land/bubbles/v2/viewport"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"
	"github.com/connordoman/windy"
)

const (
	liveHeaderHeight = 2
	liveStatusHeight = 1
	liveHelpHeight   = 1
	liveMinInputRows = 1
	liveMaxInputRows = 8
)

// LiveChatMode describes how the next turn will be answered.
type LiveChatMode string

const (
	LiveChatModeQuick  LiveChatMode = "quick"
	LiveChatModeNormal LiveChatMode = "normal"
)

// Roles used by the live chat. LiveRoleUser and LiveRoleAssistant match the
// roles persisted in the database; LiveRoleNotice is display-only.
const (
	LiveRoleUser      = "user"
	LiveRoleAssistant = "assistant"
	LiveRoleNotice    = "notice"
	LiveRoleError     = "error"
)

// LiveStore persists the turns of a live conversation. It is satisfied by the
// internal/data package, and kept as an interface so this package stays free of
// database concerns.
type LiveStore interface {
	SaveMessage(conversationID, role, content string) error
	TouchConversation(conversationID string) error
	SaveTitle(conversationID, title string) error
}

// LiveMessage is a single rendered entry in the transcript.
type LiveMessage struct {
	Role    string
	Content string
	Quick   bool
	Model   string
	Cost    float64
	Elapsed time.Duration
}

// LiveChatConfig carries everything the chat needs from the command layer.
type LiveChatConfig struct {
	ConversationID string
	Title          string
	Model          string
	QuickModel     string
	TitleModel     string
	APIKey         string
	Raw            bool

	// NeedsTitle is set for freshly created conversations so the first
	// exchange also generates a title.
	NeedsTitle bool

	// History is the transcript loaded from the database, oldest first.
	History []LiveMessage

	// Store persists new turns. A nil store keeps the chat in memory only.
	Store LiveStore
}

type LiveChatModel struct {
	cfg      LiveChatConfig
	messages []LiveMessage

	viewport viewport.Model
	input    textarea.Model
	spinner  spinner.Model

	width  int
	height int
	ready  bool

	renderedWidth int

	waiting   bool
	splash    string
	startedAt time.Time
	cancel    context.CancelFunc
	requestID int

	title     string
	totalCost float64
	err       error
}

// NewLiveChatModel builds the live chat program model.
func NewLiveChatModel(cfg LiveChatConfig) LiveChatModel {
	input := textarea.New()
	input.Placeholder = "Ask anything."
	input.Prompt = ""
	input.ShowLineNumbers = false
	input.CharLimit = 0
	input.MaxHeight = liveMaxInputRows
	input.MinHeight = liveMinInputRows
	input.DynamicHeight = true
	input.SetHeight(liveMinInputRows)
	input.Focus()

	// Enter sends the message, so newlines move to an explicit chord. Paging
	// belongs to the transcript, not the composer.
	input.KeyMap.InsertNewline = key.NewBinding(
		key.WithKeys("alt+enter", "ctrl+j", "shift+enter"),
		key.WithHelp("alt+enter", "insert newline"),
	)
	input.KeyMap.PageUp = key.NewBinding()
	input.KeyMap.PageDown = key.NewBinding()

	styles := input.Styles()
	styles.Focused.Placeholder = styles.Focused.Placeholder.Foreground(liveColor(windy.Neutral500))
	styles.Cursor.Color = liveColor(windy.Blue400)
	input.SetStyles(styles)

	vp := viewport.New()
	vp.SoftWrap = false
	vp.MouseWheelEnabled = true
	vp.KeyMap = viewport.KeyMap{
		PageUp:   key.NewBinding(key.WithKeys("pgup"), key.WithHelp("pgup", "page up")),
		PageDown: key.NewBinding(key.WithKeys("pgdown"), key.WithHelp("pgdn", "page down")),
		Up:       key.NewBinding(key.WithKeys("alt+up"), key.WithHelp("alt+↑", "scroll up")),
		Down:     key.NewBinding(key.WithKeys("alt+down"), key.WithHelp("alt+↓", "scroll down")),
	}

	sp := spinner.New(
		spinner.WithSpinner(spinner.MiniDot),
		spinner.WithStyle(lipgloss.NewStyle().Foreground(liveColor(windy.Blue500))),
	)

	messages := make([]LiveMessage, len(cfg.History))
	copy(messages, cfg.History)

	return LiveChatModel{
		cfg:      cfg,
		messages: messages,
		viewport: vp,
		input:    input,
		spinner:  sp,
		title:    cfg.Title,
	}
}

// DefaultLiveChatModel returns a chat with no configuration. It is only useful
// for tests and previews; real callers should use NewLiveChatModel.
func DefaultLiveChatModel() LiveChatModel {
	return NewLiveChatModel(LiveChatConfig{
		ConversationID: "-1",
		Model:          "gpt-4o-mini",
		QuickModel:     "gpt-5-mini",
	})
}

// ConversationID reports the conversation the chat wrote to.
func (m LiveChatModel) ConversationID() string { return m.cfg.ConversationID }

// Title reports the conversation title, which may have been generated during
// the session.
func (m LiveChatModel) Title() string { return m.title }

// TotalCost reports the accumulated cost of every priced turn in the session.
func (m LiveChatModel) TotalCost() float64 { return m.totalCost }

func (m LiveChatModel) Init() tea.Cmd {
	return textarea.Blink
}

func (m LiveChatModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.ready = true
		// layout reflows the transcript itself when the width changes.
		m.layout()
		return m, nil

	case tea.KeyPressMsg:
		switch msg.String() {
		case "ctrl+c":
			if m.cancel != nil {
				m.cancel()
			}
			return m, tea.Quit
		case "esc":
			if m.waiting {
				m.cancelRequest()
				return m, nil
			}
		case "enter":
			if m.waiting {
				return m, nil
			}
			cmd := m.submit()
			m.layout()
			return m, cmd
		}

	case liveResponseMsg:
		if msg.id != m.requestID {
			// A cancelled or superseded request; drop it.
			return m, nil
		}

		m.waiting = false
		if m.cancel != nil {
			m.cancel()
			m.cancel = nil
		}

		if msg.err != nil {
			m.append(LiveMessage{Role: LiveRoleError, Content: msg.err.Error()})
			m.rebuildTranscript(true)
			m.layout()
			return m, nil
		}

		m.totalCost += msg.cost
		m.append(LiveMessage{
			Role:    LiveRoleAssistant,
			Content: msg.content,
			Quick:   msg.quick,
			Model:   msg.model,
			Cost:    msg.cost,
			Elapsed: msg.elapsed,
		})
		m.rebuildTranscript(true)
		m.layout()

		if err := m.save(LiveRoleAssistant, msg.content); err != nil {
			m.append(LiveMessage{Role: LiveRoleError, Content: err.Error()})
			m.rebuildTranscript(true)
		}

		if m.cfg.NeedsTitle {
			m.cfg.NeedsTitle = false
			return m, m.generateTitleCmd(msg.prompt)
		}

		return m, nil

	case liveTitleMsg:
		if msg.title != "" {
			m.title = msg.title
			if m.cfg.Store != nil {
				_ = m.cfg.Store.SaveTitle(m.cfg.ConversationID, msg.title)
			}
		}
		return m, nil

	case spinner.TickMsg:
		if !m.waiting {
			return m, nil
		}
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd
	}

	var cmd tea.Cmd
	m.input, cmd = m.input.Update(msg)
	cmds = append(cmds, cmd)

	m.viewport, cmd = m.viewport.Update(msg)
	cmds = append(cmds, cmd)

	// The composer grows and shrinks with its content, so the transcript has to
	// be re-measured after every keystroke.
	m.layout()

	return m, tea.Batch(cmds...)
}

func (m LiveChatModel) View() tea.View {
	v := tea.NewView("")
	v.AltScreen = true
	v.MouseMode = tea.MouseModeCellMotion
	v.WindowTitle = m.windowTitle()

	if !m.ready {
		v.SetContent("Starting up...")
		return v
	}

	// Every pane shares a one column gutter so the header rule, transcript and
	// composer line up at any terminal width.
	gutter := lipgloss.NewStyle().PaddingLeft(1)

	sections := []string{
		m.headerView(),
		gutter.Render(m.viewport.View()),
		m.statusView(),
		gutter.Render(m.inputView()),
		m.helpView(),
	}

	v.SetContent(lipgloss.JoinVertical(lipgloss.Left, sections...))

	return v
}

// submit turns the composer contents into a request, or a notice when there is
// nothing to send.
func (m *LiveChatModel) submit() tea.Cmd {
	raw := m.input.Value()
	prompt, quick := ParseQuickCommand(raw)

	if prompt == "" {
		if strings.TrimSpace(raw) == "" {
			return nil
		}

		// The message was nothing but '/quick'.
		m.input.Reset()
		m.append(LiveMessage{
			Role:    LiveRoleNotice,
			Content: fmt.Sprintf("%s only changes how one message is answered. Type it alongside a question.", QuickCommand),
		})
		m.rebuildTranscript(true)
		return nil
	}

	m.input.Reset()

	model := m.cfg.Model
	if quick {
		model = m.quickModel()
	}

	m.append(LiveMessage{Role: LiveRoleUser, Content: prompt, Quick: quick})
	m.rebuildTranscript(true)

	if err := m.save(LiveRoleUser, prompt); err != nil {
		m.append(LiveMessage{Role: LiveRoleError, Content: err.Error()})
		m.rebuildTranscript(true)
	}

	ctx, cancel := context.WithCancel(context.Background())
	m.cancel = cancel
	m.requestID++
	m.waiting = true
	m.splash = RandomSplashText()
	m.startedAt = time.Now()

	return tea.Batch(
		m.spinner.Tick,
		askCmd(ctx, m.requestID, PromptOptions{
			Model:             model,
			APIKey:            m.cfg.APIKey,
			Prompt:            prompt,
			History:           m.apiHistory(),
			SystemMessage:     LiveSystemMessage,
			TurnSystemMessage: quickTurnSystemMessage(quick),
		}, quick),
	)
}

func (m *LiveChatModel) cancelRequest() {
	if m.cancel != nil {
		m.cancel()
		m.cancel = nil
	}

	// Bump the id so the in-flight response is ignored when it lands.
	m.requestID++
	m.waiting = false
	m.append(LiveMessage{Role: LiveRoleNotice, Content: "Request cancelled."})
	m.rebuildTranscript(true)
}

func (m *LiveChatModel) append(msg LiveMessage) {
	m.messages = append(m.messages, msg)
}

func (m *LiveChatModel) save(role, content string) error {
	if m.cfg.Store == nil {
		return nil
	}

	if err := m.cfg.Store.SaveMessage(m.cfg.ConversationID, role, content); err != nil {
		return err
	}

	return m.cfg.Store.TouchConversation(m.cfg.ConversationID)
}

func (m LiveChatModel) quickModel() string {
	if m.cfg.QuickModel != "" {
		return m.cfg.QuickModel
	}
	return m.cfg.Model
}

// apiHistory converts the transcript into the message list sent to the API.
// Notices and errors are local to the UI and never leave it, and the per-turn
// quick instruction is not part of history either, so it costs tokens only on
// the turn that asked for it.
func (m LiveChatModel) apiHistory() []MessageHistory {
	history := make([]MessageHistory, 0, len(m.messages))
	for _, msg := range m.messages {
		switch msg.Role {
		case LiveRoleUser, LiveRoleAssistant:
			history = append(history, MessageHistory{Role: msg.Role, Content: msg.Content})
		}
	}

	// The turn being sent is passed separately as the prompt.
	if n := len(history); n > 0 && history[n-1].Role == LiveRoleUser {
		history = history[:n-1]
	}

	return history
}

func quickTurnSystemMessage(quick bool) string {
	if !quick {
		return ""
	}
	return LiveQuickTurnSystemMessage
}

// layout re-measures the panes. It runs after every update because the
// composer's height follows its content.
func (m *LiveChatModel) layout() {
	if !m.ready {
		return
	}

	contentWidth := m.contentWidth()

	// Border (2) plus horizontal padding (2).
	m.input.SetWidth(max(contentWidth-4, 1))

	// The composer may never grow so tall that it pushes the transcript off
	// the screen, however short the terminal is.
	room := m.height - liveHeaderHeight - liveStatusHeight - liveHelpHeight - 2 - 1
	m.input.MaxHeight = max(min(liveMaxInputRows, room), liveMinInputRows)
	if m.input.Height() > m.input.MaxHeight {
		m.input.SetHeight(m.input.MaxHeight)
	}

	inputHeight := m.input.Height() + 2
	viewportHeight := m.height - liveHeaderHeight - liveStatusHeight - inputHeight - liveHelpHeight
	viewportHeight = max(viewportHeight, 1)

	atBottom := m.viewport.AtBottom()

	m.viewport.SetWidth(contentWidth)
	m.viewport.SetHeight(viewportHeight)

	if m.renderedWidth != contentWidth {
		m.rebuildTranscript(atBottom)
		return
	}

	if atBottom {
		m.viewport.GotoBottom()
	}
}

// rebuildTranscript re-renders every message at the current width. Rendering is
// width-dependent, so a resize reflows the whole transcript rather than leaving
// stale line breaks behind.
func (m *LiveChatModel) rebuildTranscript(toBottom bool) {
	if !m.ready {
		return
	}

	width := m.contentWidth()
	m.renderedWidth = width
	m.viewport.SetContent(m.renderTranscript(width))

	if toBottom {
		m.viewport.GotoBottom()
	}
}

// contentWidth is the width available inside the one column gutter.
func (m LiveChatModel) contentWidth() int {
	return max(m.width-2, 1)
}

func askCmd(ctx context.Context, id int, opts PromptOptions, quick bool) tea.Cmd {
	return func() tea.Msg {
		started := time.Now()

		completion, err := PromptAIWithOptions(ctx, opts)
		if err != nil {
			if ctx.Err() != nil {
				// Cancelled by the user; the id check drops this anyway.
				return liveResponseMsg{id: id, err: ctx.Err()}
			}
			return liveResponseMsg{id: id, prompt: opts.Prompt, quick: quick, err: err, elapsed: time.Since(started)}
		}

		content, err := CollectRawResponse(completion.Choices)
		if err != nil {
			return liveResponseMsg{id: id, prompt: opts.Prompt, quick: quick, err: err, elapsed: time.Since(started)}
		}

		cost, _ := CalculateCost(opts.Model, completion)

		return liveResponseMsg{
			id:      id,
			prompt:  opts.Prompt,
			content: content,
			model:   opts.Model,
			quick:   quick,
			cost:    cost,
			elapsed: time.Since(started),
		}
	}
}

func (m LiveChatModel) generateTitleCmd(prompt string) tea.Cmd {
	apiKey := m.cfg.APIKey
	model := m.cfg.TitleModel
	if model == "" {
		model = m.cfg.Model
	}

	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		title, err := GenerateShortTitle(ctx, apiKey, model, prompt)
		if err != nil || !IsMeaningfulTitle(title) {
			return liveTitleMsg{}
		}

		return liveTitleMsg{title: title}
	}
}

type liveResponseMsg struct {
	id      int
	prompt  string
	content string
	model   string
	quick   bool
	cost    float64
	elapsed time.Duration
	err     error
}

type liveTitleMsg struct {
	title string
}

// SanitizeStoredContent strips terminal escape sequences from history written
// by earlier versions of `ask`, which stored already-rendered output.
func SanitizeStoredContent(content string) string {
	return strings.TrimSpace(ansi.Strip(content))
}

func liveColor(c windy.TailwindColor) color.Color {
	return lipgloss.Color(string(c))
}

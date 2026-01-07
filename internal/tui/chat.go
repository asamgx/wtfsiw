package tui

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"

	"wtfsiw/internal/ai"
	"wtfsiw/internal/ai/tools"
	"wtfsiw/internal/session"
	"wtfsiw/internal/tmdb"
	"wtfsiw/internal/trakt"
)

// ChatState represents the current chat state
type ChatState int

const (
	ChatStateReady ChatState = iota
	ChatStateWaitingAI
	ChatStateExecutingTool
)

// FocusArea represents which area has focus
type FocusArea int

const (
	FocusInput FocusArea = iota
	FocusViewport
)

// ChatModel is the Bubble Tea model for chat mode
type ChatModel struct {
	state           ChatState
	focus           FocusArea          // Current focus area
	textarea        textarea.Model
	viewport        viewport.Model
	spinner         spinner.Model
	chatProvider    ai.ChatProvider
	executor        *ai.ToolExecutor
	session         *session.Session
	displayMsgs     []string           // Formatted messages for display
	pendingToolCalls []tools.ToolCall  // Tool calls being executed
	width           int
	height          int
	ready           bool               // viewport ready
	err             error
}

// Chat messages
type chatResponseMsg struct {
	response *ai.ChatResponse
}

type toolResultsMsg struct {
	results []tools.ToolResult
}

type chatErrorMsg struct {
	err error
}

// NewChatModel creates a new chat TUI model
func NewChatModel(chatProvider ai.ChatProvider, tmdbClient *tmdb.Client, traktClient *trakt.Client, aiProvider ai.Provider) ChatModel {
	// Create text area for input
	ta := textarea.New()
	ta.Placeholder = "Ask me for movie or TV recommendations..."
	ta.Focus()
	ta.CharLimit = 1000
	ta.SetWidth(60)
	ta.SetHeight(2)
	ta.ShowLineNumbers = false

	// Create spinner
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = spinnerStyle

	// Create tool executor
	executor := ai.NewToolExecutor(tmdbClient, traktClient, aiProvider)

	// Create new session
	sess := session.New()

	return ChatModel{
		state:        ChatStateReady,
		focus:        FocusInput,
		textarea:     ta,
		spinner:      s,
		chatProvider: chatProvider,
		executor:     executor,
		session:      sess,
		displayMsgs:  []string{},
	}
}

func (m ChatModel) Init() tea.Cmd {
	return tea.Batch(textarea.Blink, m.spinner.Tick)
}

func (m ChatModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		return m.handleKeyPress(msg)

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

		// Calculate viewport size (leave room for all UI elements)
		// Container padding: 2 (top + bottom from chatContainerStyle)
		// Header: 2 (text + border)
		// Status line: 1
		// Input: 4 (border + textarea)
		// Help: 1
		// Buffer: 2
		reservedHeight := 12
		viewportHeight := msg.Height - reservedHeight
		if viewportHeight < 5 {
			viewportHeight = 5
		}

		if !m.ready {
			m.viewport = viewport.New(msg.Width-6, viewportHeight)
			m.viewport.SetContent(strings.Join(m.displayMsgs, "\n\n"))
			m.ready = true
		} else {
			m.viewport.Width = msg.Width - 6
			m.viewport.Height = viewportHeight
		}

		m.textarea.SetWidth(msg.Width - 8)
		return m, nil

	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		cmds = append(cmds, cmd)

	case chatResponseMsg:
		return m.handleChatResponse(msg.response)

	case toolResultsMsg:
		return m.handleToolResults(msg.results)

	case chatErrorMsg:
		m.state = ChatStateReady
		m.err = msg.err
		m.addSystemMessage(fmt.Sprintf("Error: %s", msg.err.Error()))
		return m, nil
	}

	// Update textarea if ready
	if m.state == ChatStateReady {
		var cmd tea.Cmd
		m.textarea, cmd = m.textarea.Update(msg)
		cmds = append(cmds, cmd)
	}

	// Update viewport
	if m.ready {
		var cmd tea.Cmd
		m.viewport, cmd = m.viewport.Update(msg)
		cmds = append(cmds, cmd)
	}

	return m, tea.Batch(cmds...)
}

func (m ChatModel) handleKeyPress(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "ctrl+c":
		// Save session before quitting
		m.session.Save()
		return m, tea.Quit

	case "tab":
		// Toggle focus between input and viewport
		if m.state == ChatStateReady {
			if m.focus == FocusInput {
				m.focus = FocusViewport
				m.textarea.Blur()
			} else {
				m.focus = FocusInput
				m.textarea.Focus()
				return m, textarea.Blink
			}
		}
		return m, nil

	case "esc":
		// If viewing history, go back to input
		if m.focus == FocusViewport {
			m.focus = FocusInput
			m.textarea.Focus()
			return m, textarea.Blink
		}
		if m.state == ChatStateReady && m.textarea.Value() != "" {
			m.textarea.Reset()
			return m, nil
		}
		// Save session before quitting
		m.session.Save()
		return m, tea.Quit

	case "enter":
		// If viewing history, pressing enter goes back to input
		if m.focus == FocusViewport {
			m.focus = FocusInput
			m.textarea.Focus()
			return m, textarea.Blink
		}

		// Check if alt is held (allow multi-line input)
		if msg.Alt {
			// Let textarea handle it for newline
			var cmd tea.Cmd
			m.textarea, cmd = m.textarea.Update(msg)
			return m, cmd
		}

		if m.state == ChatStateReady && strings.TrimSpace(m.textarea.Value()) != "" {
			return m.sendMessage()
		}
		return m, nil
	}

	// Handle viewport scrolling when focused on viewport
	if m.focus == FocusViewport && m.ready {
		switch msg.String() {
		case "up", "k":
			m.viewport.LineUp(1)
			return m, nil
		case "down", "j":
			m.viewport.LineDown(1)
			return m, nil
		case "pgup", "ctrl+u":
			m.viewport.HalfViewUp()
			return m, nil
		case "pgdown", "ctrl+d":
			m.viewport.HalfViewDown()
			return m, nil
		case "home", "g":
			m.viewport.GotoTop()
			return m, nil
		case "end", "G":
			m.viewport.GotoBottom()
			return m, nil
		}
	}

	// Pass to textarea if ready and focused on input
	if m.state == ChatStateReady && m.focus == FocusInput {
		var cmd tea.Cmd
		m.textarea, cmd = m.textarea.Update(msg)
		return m, cmd
	}

	return m, nil
}

func (m ChatModel) sendMessage() (tea.Model, tea.Cmd) {
	content := strings.TrimSpace(m.textarea.Value())
	if content == "" {
		return m, nil
	}

	// Add user message to session
	userMsg := ai.ChatMessage{
		Role:      "user",
		Content:   content,
		Timestamp: time.Now(),
	}
	m.session.AddMessage(userMsg)

	// Add to display
	m.addDisplayMessage(FormatUserMessage(content))

	// Clear input
	m.textarea.Reset()

	// Start AI response
	m.state = ChatStateWaitingAI
	return m, m.callChatProvider()
}

func (m ChatModel) callChatProvider() tea.Cmd {
	return func() tea.Msg {
		ctx := context.Background()
		response, err := m.chatProvider.SendMessage(ctx, m.session.Messages, tools.Catalog)
		if err != nil {
			return chatErrorMsg{err: err}
		}
		return chatResponseMsg{response: response}
	}
}

func (m ChatModel) handleChatResponse(response *ai.ChatResponse) (tea.Model, tea.Cmd) {
	// Check if there are tool calls
	if len(response.ToolCalls) > 0 {
		// Add assistant message with tool calls to session
		assistantMsg := ai.ChatMessage{
			Role:      "assistant",
			Content:   response.Content,
			ToolCalls: response.ToolCalls,
			Timestamp: time.Now(),
		}
		m.session.AddMessage(assistantMsg)

		// Show content if any
		if response.Content != "" {
			m.addDisplayMessage(FormatAssistantMessage(response.Content))
		}

		// Store pending tool calls and execute
		m.state = ChatStateExecutingTool
		m.pendingToolCalls = response.ToolCalls

		// Show tool usage
		for _, tc := range response.ToolCalls {
			m.addDisplayMessage(FormatToolCall(tc.Name))
		}

		// Execute all tools
		return m, m.executeTools(response.ToolCalls)
	}

	// Regular text response - add to session
	assistantMsg := ai.ChatMessage{
		Role:      "assistant",
		Content:   response.Content,
		Timestamp: time.Now(),
	}
	m.session.AddMessage(assistantMsg)

	// Add to display
	m.addDisplayMessage(FormatAssistantMessage(response.Content))

	// Save session
	m.session.Save()

	m.state = ChatStateReady
	return m, nil
}

func (m ChatModel) executeTools(toolCalls []tools.ToolCall) tea.Cmd {
	return func() tea.Msg {
		ctx := context.Background()

		// Execute ALL tool calls
		var results []tools.ToolResult
		for _, tc := range toolCalls {
			result := m.executor.Execute(ctx, tc)
			results = append(results, result)
		}

		return toolResultsMsg{results: results}
	}
}

func (m ChatModel) handleToolResults(results []tools.ToolResult) (tea.Model, tea.Cmd) {
	// Add ALL tool results to session before calling API again
	for _, result := range results {
		toolMsg := ai.ChatMessage{
			Role:       "tool",
			Content:    result.Content,
			ToolCallID: result.ToolCallID,
			Timestamp:  time.Now(),
		}
		m.session.AddMessage(toolMsg)

		// Find the tool name from pending tool calls
		toolName := result.ToolCallID
		for _, tc := range m.pendingToolCalls {
			if tc.ID == result.ToolCallID {
				toolName = tc.Name
				break
			}
		}
		m.addDisplayMessage(FormatToolResult(toolName, !result.IsError))
	}

	// Clear pending tool calls
	m.pendingToolCalls = nil

	// Continue conversation - send back to AI with all tool results
	m.state = ChatStateWaitingAI
	return m, m.callChatProvider()
}

func (m *ChatModel) addDisplayMessage(msg string) {
	m.displayMsgs = append(m.displayMsgs, msg)
	if m.ready {
		m.viewport.SetContent(strings.Join(m.displayMsgs, "\n\n"))
		m.viewport.GotoBottom()
	}
}

func (m *ChatModel) addSystemMessage(msg string) {
	m.addDisplayMessage(FormatSystemMessage(msg))
}

func (m ChatModel) View() string {
	if !m.ready {
		return "Initializing..."
	}

	var sb strings.Builder

	// Header with focus indicator and scroll position
	headerText := "wtfsiw - Chat Mode"
	if m.focus == FocusViewport {
		scrollPercent := m.viewport.ScrollPercent() * 100
		headerText += fmt.Sprintf(" [SCROLL %.0f%%]", scrollPercent)
	}
	sb.WriteString(chatHeaderStyle.Render(headerText))
	sb.WriteString("\n")

	// Chat viewport
	sb.WriteString(m.viewport.View())
	sb.WriteString("\n")

	// Status line (thinking/tool indicator)
	switch m.state {
	case ChatStateWaitingAI:
		sb.WriteString(m.spinner.View())
		sb.WriteString(" ")
		sb.WriteString(thinkingStyle.Render("Thinking..."))
	case ChatStateExecutingTool:
		sb.WriteString(m.spinner.View())
		sb.WriteString(" ")
		toolNames := ""
		for i, tc := range m.pendingToolCalls {
			if i > 0 {
				toolNames += ", "
			}
			toolNames += tc.Name
		}
		sb.WriteString(toolExecutingStyle.Render("Executing: " + toolNames + "..."))
	}
	sb.WriteString("\n")

	// Input area
	sb.WriteString(chatInputStyle.Render(m.textarea.View()))
	sb.WriteString("\n")

	// Help - context sensitive
	var help string
	switch {
	case m.state != ChatStateReady:
		help = "Processing..."
	case m.focus == FocusViewport:
		help = "↑/k ↓/j scroll • Ctrl+u/d page • g/G top/bottom • Tab/Esc/Enter → input"
	default:
		help = "Enter send • Tab scroll history • Esc quit"
	}
	sb.WriteString(chatHelpStyle.Render(help))

	return chatContainerStyle.Render(sb.String())
}

// RunChat starts the chat TUI application
func RunChat(chatProvider ai.ChatProvider, tmdbClient *tmdb.Client, traktClient *trakt.Client, aiProvider ai.Provider) error {
	p := tea.NewProgram(
		NewChatModel(chatProvider, tmdbClient, traktClient, aiProvider),
		tea.WithAltScreen(),
	)

	_, err := p.Run()
	return err
}

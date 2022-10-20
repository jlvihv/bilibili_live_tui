package teaui

// A simple program demonstrating the text area component from the Bubbles
// component library.

import (
	"bili/getter"
	"bili/sender"
	"fmt"
	"log"
	"strings"

	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var BusChan chan getter.DanmuMsg

func Run(busChan chan getter.DanmuMsg) {
	BusChan = busChan
	p := tea.NewProgram(initialModel())
	go getMsg(p, BusChan)
	if _, err := p.StartReturningModel(); err != nil {
		log.Fatal(err)
	}
}

func getMsg(p *tea.Program, busChan chan getter.DanmuMsg) tea.Msg {
	for {
		msg := <-busChan
		p.Send(tea.Msg(msg))
	}
}

func sendMsg(msg string) {
	go sender.SendMsg(9741626, msg, BusChan)
}

type (
	errMsg error
)

type model struct {
	viewport    viewport.Model
	messages    []string
	textarea    textarea.Model
	senderStyle lipgloss.Style
	err         error
}

func initialModel() model {
	ta := textarea.New()
	ta.Placeholder = "Send a message..."
	ta.Focus()

	ta.Prompt = "┃ "
	ta.CharLimit = 400

	ta.SetWidth(80)
	ta.SetHeight(3)

	// Remove cursor line styling
	ta.FocusedStyle.CursorLine = lipgloss.NewStyle()

	ta.ShowLineNumbers = false

	vp := viewport.New(80, 5)
	vp.SetContent(`发条友善的弹幕吧～`)

	ta.KeyMap.InsertNewline.SetEnabled(false)

	return model{
		textarea:    ta,
		messages:    []string{},
		viewport:    vp,
		senderStyle: lipgloss.NewStyle().Foreground(lipgloss.Color("5")),
		err:         nil,
	}
}

func (m model) Init() tea.Cmd {
	return textarea.Blink
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var (
		tiCmd tea.Cmd
		vpCmd tea.Cmd
	)

	m.textarea, tiCmd = m.textarea.Update(msg)
	m.viewport, vpCmd = m.viewport.Update(msg)

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC, tea.KeyEsc:
			fmt.Println(m.textarea.Value())
			return m, tea.Quit
		case tea.KeyEnter:
			sendMsg(m.textarea.Value())
			// 发送消息之后，输入框清空
			m.textarea.Reset()
			m.viewport.GotoBottom()
		}
	case getter.DanmuMsg:
		m.messages = append(m.messages, m.senderStyle.Render(fmt.Sprintf("%s: ", msg.Author))+msg.Content)
		m.viewport.SetContent(strings.Join(m.messages, "\n"))
		m.viewport.GotoBottom()

	// We handle errors just like any other message
	case errMsg:
		m.err = msg
		return m, nil
	}

	return m, tea.Batch(tiCmd, vpCmd)
}

func (m model) View() string {
	return fmt.Sprintf(
		"%s\n\n%s",
		m.viewport.View(),
		m.textarea.View(),
	) + "\n\n"
}

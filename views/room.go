package views

// A simple program demonstrating the text area component from the Bubbles
// component library.

import (
	"bili/config"
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

var msgCh chan getter.DanmuMsg

func Run(msgChan chan getter.DanmuMsg, roomChan chan getter.RoomInfo) {
	msgCh = msgChan
	p := tea.NewProgram(initialModel())
	go getMsg(p, msgCh)
	go getRoomInfo(p, roomChan)
	if _, err := p.StartReturningModel(); err != nil {
		log.Fatal(err)
	}
}

func getMsg(p *tea.Program, msgChan chan getter.DanmuMsg) tea.Msg {
	for {
		msg := <-msgChan
		p.Send(tea.Msg(msg))
	}
}

func getRoomInfo(p *tea.Program, roomChan chan getter.RoomInfo) tea.Msg {
	for {
		room := <-roomChan
		p.Send(tea.Msg(room))
	}
}

func sendMsg(msg string) {
	go sender.SendMsg(config.Get().RoomIDs[0], msg, msgCh)
}

type (
	errMsg error
)

type model struct {
	roomArea    viewport.Model
	chatArea    viewport.Model
	inputArea   textarea.Model
	messages    []string
	senderStyle lipgloss.Style
	err         error
}

func initialModel() model {
	roomArea := viewport.New(80, 6)
	roomArea.SetContent(`正在载入房间信息～`)

	chatArea := viewport.New(80, 10)
	chatArea.SetContent(`一大波弹幕正在袭来～`)

	inputArea := textarea.New()
	inputArea.Placeholder = "Send a message..."
	inputArea.Focus()

	inputArea.Prompt = "┃ "
	inputArea.CharLimit = 400

	inputArea.SetWidth(80)
	inputArea.SetHeight(3)

	// Remove cursor line styling
	inputArea.FocusedStyle.CursorLine = lipgloss.NewStyle()
	inputArea.ShowLineNumbers = false

	inputArea.KeyMap.InsertNewline.SetEnabled(false)

	return model{
		roomArea:    roomArea,
		chatArea:    chatArea,
		inputArea:   inputArea,
		messages:    []string{},
		senderStyle: lipgloss.NewStyle().Foreground(lipgloss.Color("5")),
		err:         nil,
	}
}

func (m model) Init() tea.Cmd {
	return textarea.Blink
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var (
		roomCmd  tea.Cmd
		chatCmd  tea.Cmd
		inputCmd tea.Cmd
	)

	m.roomArea, roomCmd = m.roomArea.Update(msg)
	m.chatArea, chatCmd = m.chatArea.Update(msg)
	m.inputArea, inputCmd = m.inputArea.Update(msg)

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC, tea.KeyEsc:
			fmt.Println(m.inputArea.Value())
			return m, tea.Quit
		case tea.KeyEnter:
			sendMsg(m.inputArea.Value())
			// 发送消息之后，输入框清空
			m.inputArea.Reset()
			m.chatArea.GotoBottom()
		}
	case getter.DanmuMsg:
		m.messages = append(m.messages, m.senderStyle.Render(fmt.Sprintf("%s: ", msg.Author))+msg.Content)
		m.chatArea.SetContent(strings.Join(m.messages, "\n"))
		m.chatArea.GotoBottom()
	case getter.RoomInfo:
		m.roomArea.SetContent(formatRoomInfo(&msg))

	// We handle errors just like any other message
	case errMsg:
		m.err = msg
		return m, nil
	}

	return m, tea.Batch(roomCmd, chatCmd, inputCmd)
}

func formatRoomInfo(r *getter.RoomInfo) string {
	return r.Title + "\n" +
		fmt.Sprintf("ID: %d", r.RoomID) + "\n" +
		fmt.Sprintf("分区: %s/%s", r.ParentAreaName, r.AreaName) + "\n" +
		fmt.Sprintf("👀: %d", r.Online) + "\n" +
		fmt.Sprintf("❤️: %d", r.Attention) + "\n" +
		fmt.Sprintf("🕒: %s", r.Time) + "\n"
}

func (m model) View() string {
	return fmt.Sprintf(
		"%s\n\n%s\n\n%s",
		m.roomArea.View(),
		m.chatArea.View(),
		m.inputArea.View(),
	) + "\n\n"
}

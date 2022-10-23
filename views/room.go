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

type manager struct {
	msgChan  chan getter.DanmuMsg
	roomChan chan getter.RoomInfo
}

type view struct {
	roomArea    viewport.Model
	chatArea    viewport.Model
	inputArea   textarea.Model
	messages    []string
	senderStyle lipgloss.Style
	err         error
}

type (
	errMsg error
)

var rootManager *manager

func NewManager(msgChan chan getter.DanmuMsg, roomChan chan getter.RoomInfo) *manager {
	rootManager = &manager{
		msgChan:  msgChan,
		roomChan: roomChan,
	}
	return rootManager
}

func (m *manager) Run() {
	view := initialModel()
	p := tea.NewProgram(view)
	go m.getMsg(p)
	go m.getRoomInfo(p)

	if _, err := p.StartReturningModel(); err != nil {
		log.Fatal(err)
	}
}

func (m *manager) getMsg(p *tea.Program) tea.Msg {
	for {
		msg := <-m.msgChan
		p.Send(tea.Msg(msg))
	}
}

func (m *manager) getRoomInfo(p *tea.Program) tea.Msg {
	for {
		room := <-m.roomChan
		p.Send(tea.Msg(room))
	}
}

func (m *manager) sendMsg(msg string) {
	go sender.SendMsg(config.Get().RoomIDs[0], msg, m.msgChan)
}

func initialModel() tea.Model {
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
	return view{
		roomArea:    roomArea,
		chatArea:    chatArea,
		inputArea:   inputArea,
		senderStyle: lipgloss.NewStyle().Foreground(lipgloss.Color("#FF69B4")),
	}
}

func (v view) Init() tea.Cmd {
	return tea.Batch(textarea.Blink, tea.EnterAltScreen)
}

func (v view) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var (
		roomCmd  tea.Cmd
		chatCmd  tea.Cmd
		inputCmd tea.Cmd
	)

	v.roomArea, roomCmd = v.roomArea.Update(msg)
	v.chatArea, chatCmd = v.chatArea.Update(msg)
	v.inputArea, inputCmd = v.inputArea.Update(msg)

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC, tea.KeyEsc:
			fmt.Println(v.inputArea.Value())
			return v, tea.Quit
		case tea.KeyEnter:
			rootManager.sendMsg(v.inputArea.Value())
			// 发送消息之后，输入框清空
			v.inputArea.Reset()
			v.chatArea.GotoBottom()
		}
	case getter.DanmuMsg:
		v.messages = append(v.messages, v.senderStyle.Render(fmt.Sprintf("%s: ", msg.Author))+msg.Content)
		v.chatArea.SetContent(strings.Join(v.messages, "\n"))
		v.chatArea.GotoBottom()
	case getter.RoomInfo:
		v.roomArea.SetContent(formatRoomInfo(&msg))

	// We handle errors just like any other message
	case errMsg:
		v.err = msg
		return v, nil
	}

	return v, tea.Batch(roomCmd, chatCmd, inputCmd)
}

func (v view) View() string {
	return fmt.Sprintf(
		"%s\n\n%s\n\n%s",
		v.roomArea.View(),
		v.chatArea.View(),
		v.inputArea.View(),
	) + "\n\n"
}

func formatRoomInfo(r *getter.RoomInfo) string {
	return r.Title + "\n" +
		fmt.Sprintf("ID: %d", r.RoomID) + "\n" +
		fmt.Sprintf("分区: %s/%s", r.ParentAreaName, r.AreaName) + "\n" +
		fmt.Sprintf("👀: %d", r.Online) + "\n" +
		fmt.Sprintf("❤️: %d", r.Attention) + "\n" +
		fmt.Sprintf("🕒: %s", r.Time) + "\n"
}

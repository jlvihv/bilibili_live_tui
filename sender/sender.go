package sender

import (
	"bili/getter"
	"fmt"
	"os"
	"time"

	bg "github.com/iyear/biligo"
)

var (
	bc  *bg.BiliClient
	err error
)

func heartbeat() {
	start := time.Now()
	err = bc.VideoHeartBeat(242531611, 173439442, int64(time.Since(start).Seconds()))
	if err != nil {
		fmt.Println("failed to send heartbeat; error:", err)
		os.Exit(0)
	}
	time.AfterFunc(time.Second*10, heartbeat)
}

func SendMsg(roomID int64, msg string, busChan chan getter.DanmuMsg) {
	msgRune := []rune(msg)
	for i := 0; i < len(msgRune); i += 20 {
		err = nil
		if i+20 < len(msgRune) {
			err = bc.LiveSendDanmaku(roomID, 16777215, 25, 1, string(msgRune[i:i+20]), 0)
			time.Sleep(time.Second * 1)
		} else {
			err = bc.LiveSendDanmaku(roomID, 16777215, 25, 1, string(msgRune[i:]), 0)
		}
		if err != nil {
			busChan <- getter.DanmuMsg{Author: "system", Content: "发送弹幕失败", Type: ""}
		}
	}
}

func Run(auth bg.CookieAuth) {
	bc, err = bg.NewBiliClient(&bg.BiliSetting{
		Auth:      &auth,
		DebugMode: false,
	})
	if err != nil {
		fmt.Printf("failed to make new bili client; error: %v", err)
		os.Exit(0)
	}
	go heartbeat()
}

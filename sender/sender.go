package sender

import (
	"bili/config"
	"bili/getter"
	"fmt"
	"os"
	"time"

	"github.com/iyear/biligo"
)

var biliClient *biligo.BiliClient

func heartbeat() {
	start := time.Now()
	for {
		err := biliClient.VideoHeartBeat(242531611, 173439442, int64(time.Since(start).Seconds()))
		if err != nil {
			fmt.Println("failed to send heartbeat; error:", err)
		}
		time.Sleep(time.Second * 10)
	}
}

func SendMsg(roomID uint32, msg string, msgChan chan getter.DanmuMsg) {
	msgRune := []rune(msg)
	var err error
	for i := 0; i < len(msgRune); i += 20 {
		if i+20 < len(msgRune) {
			err = biliClient.LiveSendDanmaku(int64(roomID), 16777215, 25, 1, string(msgRune[i:i+20]), 0)
			time.Sleep(time.Second)
		} else {
			err = biliClient.LiveSendDanmaku(int64(roomID), 16777215, 25, 1, string(msgRune[i:]), 0)
		}
		if err != nil {
			msgChan <- getter.DanmuMsg{Author: "system", Content: "发送弹幕失败", Type: ""}
		}
	}
}

func Run() {
	biliCli, err := biligo.NewBiliClient(&biligo.BiliSetting{
		Auth:      config.CookieAuth(),
		DebugMode: false,
	})
	biliClient = biliCli
	if err != nil {
		fmt.Println("failed to make new bili client:", err)
		os.Exit(1)
	}
	go heartbeat()
}

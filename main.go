package main

import (
	"bili/config"
	"bili/getter"
	"bili/login"
	"bili/sender"
	"bili/views"
	"fmt"
	"os"
)

func main() {

	check_login()
	check_room()

	msgChan := make(chan getter.DanmuMsg, 100)
	roomInfoChan := make(chan getter.RoomInfo, 100)

	getter.Run(msgChan, roomInfoChan)
	sender.Run()
	m := views.NewManager(msgChan, roomInfoChan)
	m.Run()
}

func check_login() {
	if config.Get().Cookie.SESSDATA == "" {
		fmt.Println("请使用扫码登陆")
		url, qrcodeKey, err := login.GetLoginURL()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		login.ShowQrcode(url)
		cookieAuth, err := login.GetCookieAuth(qrcodeKey)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		config.SetCookieAuth(cookieAuth)
	}
}

func check_room() {
	if len(config.Get().RoomIDs) == 0 {
		fmt.Print("想进入哪个直播间呢：")
		var roomID int
		fmt.Scan(&roomID)
		config.SetRoomID(roomID)
	}
}

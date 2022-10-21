package main

import (
	"bili/getter"
	"bili/sender"
	"bili/views"
)

// 这里修改一下
// 使用扫码登陆
// roomID 可以手动输入，并且可以随时修改

func main() {
	msgChan := make(chan getter.DanmuMsg, 100)
	roomInfoChan := make(chan getter.RoomInfo, 100)
	getter.Run(msgChan, roomInfoChan)
	sender.Run()
	views.Run(msgChan, roomInfoChan)
}

package main

import (
	"bili/getter"
	"bili/sender"
	"bili/teaui"
	"flag"
	"fmt"
	"strings"

	"github.com/BurntSushi/toml"
	bg "github.com/iyear/biligo"
)

// 这里修改一下
// 使用扫码登陆
// roomID 可以手动输入，并且可以随时修改

type Config struct {
	Cookie string
	RoomID int64
}

var (
	config Config
	auth   bg.CookieAuth
)

func init() {
	configFile := ""
	roomID := int64(0)
	flag.StringVar(&configFile, "c", "config.toml", "usage for config")
	flag.Int64Var(&roomID, "r", 0, "usage for room id")
	flag.Parse()

	if _, err := toml.DecodeFile(configFile, &config); err != nil {
		fmt.Printf("Error decoding config.toml: %s\n", err)
	}

	if roomID != 0 {
		config.RoomID = roomID
	}

	attrs := strings.Split(config.Cookie, ";")
	kvs := make(map[string]string)
	for _, attr := range attrs {
		kv := strings.Split(attr, "=")
		k := strings.Trim(kv[0], " ")
		v := strings.Trim(kv[1], " ")
		kvs[k] = v
	}
	auth.SESSDATA = kvs["SESSDATA"]
	auth.DedeUserID = kvs["DedeUserID"]
	auth.DedeUserIDCkMd5 = kvs["DedeUserID__ckMd5"]
	auth.BiliJCT = kvs["bili_jct"]
}

func main() {
	busChan := make(chan getter.DanmuMsg, 100)
	roomInfoChan := make(chan getter.RoomInfo, 100)
	getter.Run(config.RoomID, auth, busChan, roomInfoChan)
	sender.Run(auth)
	// ui.Run(config.RoomID, busChan, roomInfoChan)
	teaui.Run(busChan)
}

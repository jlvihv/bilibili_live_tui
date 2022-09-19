package main

import (
	"bili/getter"
	"bili/sender"
	"bili/ui"
	"flag"
	"fmt"
	"strings"

	"github.com/BurntSushi/toml"
	bg "github.com/iyear/biligo"
)

type Config struct {
	Cookie string
	RoomId int64
}

var config Config
var auth bg.CookieAuth

func init() {
	configFile := ""
	flag.StringVar(&configFile, "c", "config.toml", "usage for config")
	flag.Parse()
	if _, err := toml.DecodeFile(configFile, &config); err != nil {
		fmt.Printf("Error decoding config.toml: %s\n", err)
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
	getter.Run(config.RoomId, auth, busChan, roomInfoChan)
	sender.Run(auth)
	ui.Run(config.RoomId, busChan, roomInfoChan)
}

package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/iyear/biligo"
	"github.com/mitchellh/go-homedir"
	"github.com/pelletier/go-toml/v2"
)

const (
	configFileName = "bilibili-live-tui.toml"
)

var (
	configFilePath string
	conf           *Config
	auth           *biligo.CookieAuth
)

type Config struct {
	RoomIDs []uint32 `toml:"room_ids"`
	Cookie  string   `toml:"cookie"`
}

func FilePath() string {
	if configFilePath != "" {
		return configFilePath
	}
	home, err := homedir.Dir()
	if err != nil {
		fmt.Println("get home dir failed:", err)
		return ""
	}
	configFilePath = filepath.Join(home, ".config")
	return configFilePath
}

func FullPath() string {
	return filepath.Join(FilePath(), configFileName)
}

// Create if config file not exist, create
func Create() error {
	configPath := FilePath()
	err := os.MkdirAll(configPath, 0644)
	if err != nil {
		return err
	}
	b, err := toml.Marshal(&Config{})
	if err != nil {
		return err
	}
	_, err = os.Stat(FullPath())
	if os.IsNotExist(err) {
		return os.WriteFile(FullPath(), b, 0644)
	}
	return nil
}

func IsFileExist() bool {
	_, err := os.Stat(FullPath())
	return !os.IsNotExist(err)
}

func Get() *Config {
	if conf != nil {
		return conf
	}
	conf = &Config{}
	if !IsFileExist() {
		Create()
		return conf
	}
	b, err := os.ReadFile(FullPath())
	if err != nil {
		fmt.Println("read config file failed:", err)
		return conf
	}
	err = toml.Unmarshal(b, conf)
	if err != nil {
		fmt.Println("unmarshal config file failed:", err)
		return conf
	}
	return conf
}

func CookieAuth() *biligo.CookieAuth {
	if auth != nil {
		return auth
	}
	conf := Get()
	if conf.Cookie == "" {
		return nil
	}
	attrs := strings.Split(conf.Cookie, ";")
	kvs := make(map[string]string)
	for _, attr := range attrs {
		kv := strings.Split(attr, "=")
		k := strings.Trim(kv[0], " ")
		v := strings.Trim(kv[1], " ")
		kvs[k] = v
	}
	auth = &biligo.CookieAuth{}
	auth.SESSDATA = kvs["SESSDATA"]
	auth.DedeUserID = kvs["DedeUserID"]
	auth.DedeUserIDCkMd5 = kvs["DedeUserID__ckMd5"]
	auth.BiliJCT = kvs["bili_jct"]
	return auth
}

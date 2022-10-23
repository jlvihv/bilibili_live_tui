package config

import (
	"fmt"
	"os"
	"path/filepath"

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
)

type Config struct {
	RoomIDs []uint32          `toml:"room_ids"`
	Cookie  biligo.CookieAuth `toml:"cookie"`
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
		err := Create()
		if err != nil {
			fmt.Println("create config file failed:", err)
			return conf
		}
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

func SetCookieAuth(cookieAuth *biligo.CookieAuth) {
	fmt.Println("cookie auth:", cookieAuth)
	conf.Cookie = *cookieAuth
	b, err := toml.Marshal(conf)
	if err != nil {
		panic(err)
	}
	err = os.WriteFile(FullPath(), b, 0644)
	if err != nil {
		panic(err)
	}
	fmt.Println("保存配置文件成功")
}

func SetRoomID(roomID int) {
	conf.RoomIDs = []uint32{uint32(roomID)}
	b, err := toml.Marshal(conf)
	if err != nil {
		panic(err)
	}
	err = os.WriteFile(FullPath(), b, 0644)
	if err != nil {
		panic(err)
	}
}

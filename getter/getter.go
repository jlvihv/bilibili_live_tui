package getter

import (
	"bili/config"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/gorilla/websocket"
	"github.com/iyear/biligo"

	"github.com/asmcos/requests"
	"github.com/tidwall/gjson"
)

type DanmuClient struct {
	roomID        uint32
	auth          *biligo.CookieAuth
	conn          *websocket.Conn
	unzlibChannel chan []byte
}

type OnlineRankUser struct {
	Name  string
	Score int64
	Rank  int64
}

type RoomInfo struct {
	RoomID          int
	Title           string
	ParentAreaName  string
	AreaName        string
	Online          int64
	Attention       int64
	Time            string
	OnlineRankUsers []OnlineRankUser
}

type DanmuMsg struct {
	Author  string
	Content string
	Type    string
	Time    time.Time
}

type receivedInfo struct {
	Cmd        string                 `json:"cmd"`
	Data       map[string]interface{} `json:"data"`
	Info       []interface{}          `json:"info"`
	Full       map[string]interface{} `json:"full"`
	Half       map[string]interface{} `json:"half"`
	Side       map[string]interface{} `json:"side"`
	RoomID     uint32                 `json:"roomid"`
	RealRoomID uint32                 `json:"real_roomid"`
	MsgCommon  string                 `json:"msg_common"`
	MsgSelf    string                 `json:"msg_self"`
	LinkURL    string                 `json:"link_url"`
	MsgType    string                 `json:"msg_type"`
	ShieldUID  string                 `json:"shield_uid"`
	BusinessID string                 `json:"business_id"`
	Scatter    map[string]interface{} `json:"scatter"`
}

type handShakeInfo struct {
	UID       uint8  `json:"uid"`
	Roomid    uint32 `json:"roomid"`
	Protover  uint8  `json:"protover"`
	Platform  string `json:"platform"`
	Clientver string `json:"clientver"`
	Type      uint8  `json:"type"`
	Key       string `json:"key"`
}

func (d *DanmuClient) connect() error {
	danmuURL := "https://api.live.bilibili.com/xlive/web-room/v1/index/getDanmuInfo?id=%d&type=0"
	resp, err := requests.Get(fmt.Sprintf(danmuURL, d.roomID))
	if err != nil {
		fmt.Println("connect danmu server failed:", err)
		return err
	}
	token := gjson.Get(resp.Text(), "data.token").String()
	hostList := []string{}
	gjson.Get(resp.Text(), "data.host_list").ForEach(func(_, value gjson.Result) bool {
		hostList = append(hostList, value.Get("host").String())
		return true
	})

	if len(hostList) == 0 {
		fmt.Println("no host available")
		return fmt.Errorf("no host available")
	}

	for _, host := range hostList {
		d.conn, _, err = websocket.DefaultDialer.Dial(fmt.Sprintf("wss://%s:443/sub", host), nil)
		if err != nil {
			fmt.Printf("websocket.Dial failed for %s: %s\n", host, err)
			continue
		}
		fmt.Printf("连接弹幕服务器[%s]成功\n", host)
		break
	}
	if err != nil {
		fmt.Println("dial websocket danmu server all failed:", err)
		return err
	}

	hsInfo := handShakeInfo{
		UID:       0,
		Roomid:    d.roomID,
		Protover:  2,
		Platform:  "web",
		Clientver: "1.10.2",
		Type:      2,
		Key:       token,
	}
	b, err := json.Marshal(hsInfo)
	if err != nil {
		fmt.Println("json marshar handShakeInfo error:", err)
		return err
	}
	err = d.sendPackage(0, 16, 1, 7, 1, b)
	if err != nil {
		fmt.Println("conn sendPackage error:", err)
		return err
	}
	// fmt.Printf("连接房间 [%d] 成功\n", d.roomID)
	return nil
}

func (d *DanmuClient) heartBeat() {
	for {
		obj := []byte("5b6f626a656374204f626a6563745d")
		if err := d.sendPackage(0, 16, 1, 2, 1, obj); err != nil {
			fmt.Println("heart beat err:", err)
			continue
		}
		time.Sleep(30 * time.Second)
	}
}

func (d *DanmuClient) receiveRawMsg(busChan chan DanmuMsg) {
	for {
		_, rawMsg, _ := d.conn.ReadMessage()
		if rawMsg[7] != 2 {
			continue
		}
		msgs := splitMsg(zlibUnCompress(rawMsg[16:]))
		for _, msg := range msgs {
			received := new(receivedInfo)
			err := json.Unmarshal(msg[16:], received)
			if err != nil {
				fmt.Println("json Unmarshal receivedInfo error:", err)
				continue
			}
			m := DanmuMsg{}
			switch received.Cmd {
			case "COMBO_SEND":
				m.Author = received.Data["uname"].(string)
				m.Content = fmt.Sprintf(
					"送给 %s %d 个 %s",
					received.Data["r_uname"].(string),
					int(received.Data["combo_num"].(float64)),
					received.Data["gift_name"].(string),
				)
			case "DANMU_MSG":
				m.Author = received.Info[2].([]interface{})[1].(string)
				m.Content = received.Info[1].(string)
			case "GUARD_BUY":
				m.Author = received.Data["username"].(string)
				m.Content = fmt.Sprintf("购买了 %s", received.Data["giftName"].(string))
			case "INTERACT_WORD":
				m.Author = received.Data["uname"].(string)
				m.Content = "进入了房间"
			case "SEND_GIFT":
				m.Author = received.Data["uname"].(string)
				m.Content = fmt.Sprintf(
					"投喂了 %d 个 %s",
					int(received.Data["num"].(float64)),
					received.Data["giftName"].(string),
				)
			case "USER_TOAST_MSG":
				m.Author = "system"
				m.Content = received.Data["toast_msg"].(string)
			case "NOTICE_MSG":
				m.Author = "system"
				m.Content = received.MsgSelf
			default: // "LIVE" "ACTIVITY_BANNER_UPDATE_V2" "ONLINE_RANK_COUNT" "ONLINE_RANK_TOP3" "ONLINE_RANK_V2" "PANEL" "PREPARING" "WIDGET_BANNER" "LIVE_INTERACTIVE_GAME"
				continue
			}
			m.Type = received.Cmd
			m.Time = time.Now()
			busChan <- m
		}
	}
}

func (d *DanmuClient) syncRoomInfo(roomInfoChan chan RoomInfo) {
	for {
		roomInfoAPI := fmt.Sprintf(
			"https://api.live.bilibili.com/room/v1/room/get_info?room_id=%d",
			d.roomID,
		)
		onlineRankAPI := fmt.Sprintf(
			"https://api.live.bilibili.com/xlive/general-interface/v1/rank/getOnlineGoldRank?ruid=%s&roomId=%d&page=1&pageSize=50",
			d.auth.DedeUserID,
			d.roomID,
		)

		roomInfo := new(RoomInfo)
		roomInfo.OnlineRankUsers = make([]OnlineRankUser, 16)
		roomResp, err := requests.Get(roomInfoAPI)
		if err == nil {
			roomInfo.RoomID = int(d.roomID)
			roomInfo.Title = gjson.Get(roomResp.Text(), "data.title").String()
			roomInfo.AreaName = gjson.Get(roomResp.Text(), "data.area_name").String()
			roomInfo.ParentAreaName = gjson.Get(roomResp.Text(), "data.parent_area_name").String()
			roomInfo.Online = gjson.Get(roomResp.Text(), "data.online").Int()
			roomInfo.Attention = gjson.Get(roomResp.Text(), "data.attention").Int()
			_time, _ := time.Parse(
				"2006-01-02 15:04:05",
				gjson.Get(roomResp.Text(), "data.live_time").String(),
			)
			seconds := time.Now().Unix() - _time.Unix() + 8*60*60
			days := seconds / 86400
			hours := (seconds % 86400) / 3600
			minutes := (seconds % 3600) / 60
			if days > 0 {
				roomInfo.Time = fmt.Sprintf("%d天%d时%d分", days, hours, minutes)
			} else if hours > 0 {
				roomInfo.Time = fmt.Sprintf("%d时%d分", hours, minutes)
			} else {
				roomInfo.Time = fmt.Sprintf("%d分", minutes)
			}
		}

		onlineResp, err := requests.Get(onlineRankAPI)
		if err == nil {
			rawUsers := gjson.Get(onlineResp.Text(), "data.OnlineRankItem").Array()
			for _, rawUser := range rawUsers {
				user := OnlineRankUser{
					Name:  rawUser.Get("name").String(),
					Score: rawUser.Get("score").Int(),
					Rank:  rawUser.Get("userRank").Int(),
				}
				roomInfo.OnlineRankUsers = append(roomInfo.OnlineRankUsers, user)
			}
		}

		roomInfoChan <- *roomInfo
		time.Sleep(30 * time.Second)
	}
}

func (d *DanmuClient) getHistory(busChan chan DanmuMsg) {
	historyAPI := fmt.Sprintf(
		"https://api.live.bilibili.com/xlive/web-room/v1/dM/gethistory?roomid=%d",
		d.roomID,
	)
	r, err := requests.Get(historyAPI)
	if err != nil {
		return
	}

	histories := gjson.Get(r.Text(), "data.room").Array()
	for _, history := range histories {
		t, _ := time.Parse("2006-01-02 15:04:05", history.Get("timeline").String())
		danmu := DanmuMsg{
			Author:  history.Get("nickname").String(),
			Content: history.Get("text").String(),
			Type:    "DANMU_MSG",
			Time:    t,
		}
		busChan <- danmu
	}
}

func Run(msgChan chan DanmuMsg, roomInfoChan chan RoomInfo) {
	danmuClient := DanmuClient{
		roomID:        config.Get().RoomIDs[0],
		auth:          &config.Get().Cookie,
		conn:          new(websocket.Conn),
		unzlibChannel: make(chan []byte, 100),
	}
	err := danmuClient.connect()
	if err != nil {
		fmt.Println("can not connect danmu server:", err)
		os.Exit(1)
	}
	go danmuClient.heartBeat()
	go danmuClient.receiveRawMsg(msgChan)
	go danmuClient.getHistory(msgChan)
	go danmuClient.syncRoomInfo(roomInfoChan)
}

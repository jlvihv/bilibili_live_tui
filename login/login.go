package login

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/asmcos/requests"
	"github.com/iyear/biligo"
	"github.com/mdp/qrterminal"
	"github.com/tidwall/gjson"
)

func GetLoginURL() (loginURL, qrcodeKey string, err error) {
	url := "http://passport.bilibili.com/x/passport-login/web/qrcode/generate"
	resp, err := requests.Get(url)
	if err != nil {
		return "", "", err
	}
	loginURL = gjson.Get(resp.Text(), "data.url").String()
	qrcodeKey = gjson.Get(resp.Text(), "data.qrcode_key").String()
	return
}

func ShowQrcode(str string) {
	qrterminal.Generate(str, qrterminal.L, os.Stdout)
}

func GetCookieAuth(qrcodeKey string) (*biligo.CookieAuth, error) {
	for {
		statusCode, cookieAuth, err := verifyLoginStatus(qrcodeKey)
		if err != nil {
			fmt.Println(err)
			return cookieAuth, err
		}
		if statusCode == 0 {
			fmt.Println("登陆成功")
			return cookieAuth, nil
		}
		time.Sleep(1 * time.Second)
	}
}

func verifyLoginStatus(
	qrcodeKey string,
) (int64, *biligo.CookieAuth, error) {
	url := fmt.Sprintf(
		"http://passport.bilibili.com/x/passport-login/web/qrcode/poll?qrcode_key=%s",
		qrcodeKey,
	)
	resp, err := requests.Get(url)
	if err != nil {
		return 0, nil, err
	}
	statusCode := gjson.Get(resp.Text(), "data.code").Int()
	switch statusCode {
	case 86101: // 未扫码
		return statusCode, nil, nil
	case 86090: // 已扫码
		return statusCode, nil, nil
	case 0: // 已登录
		respURL := gjson.Get(resp.Text(), "data.url").String()
		// 要？后面的参数
		params := strings.Split(respURL, "?")[1]
		// 以&分割字符串
		paramArray := strings.Split(params, "&")
		cookieAuth := &biligo.CookieAuth{}
		for _, param := range paramArray {
			// 以=分割字符串
			paramKV := strings.Split(param, "=")
			switch paramKV[0] {
			case "DedeUserID":
				cookieAuth.DedeUserID = paramKV[1]
			case "DedeUserID__ckMd5":
				cookieAuth.DedeUserIDCkMd5 = paramKV[1]
			case "SESSDATA":
				cookieAuth.SESSDATA = paramKV[1]
			case "bili_jct":
				cookieAuth.BiliJCT = paramKV[1]
			}
		}
		return statusCode, cookieAuth, nil
	case 86038: // 已过期
		return statusCode, nil, fmt.Errorf("二维码已过期")
	default:
		return statusCode, nil, fmt.Errorf("未知错误")
	}
}

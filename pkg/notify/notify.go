package notify

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/smtp"
	"net/url"
	"os"
	"strings"
	"time"
	"watchducker/pkg/logger"

	"github.com/spf13/viper"
)

type Config struct {
	Setting struct {
		PushServer string `mapstructure:"push_server"`
		LogLevel   string `mapstructure:"log_level"`
	} `mapstructure:"setting"`

	Telegram struct {
		APIURL   string `mapstructure:"api_url"`
		BotToken string `mapstructure:"bot_token"`
		ChatID   string `mapstructure:"chat_id"`
	} `mapstructure:"telegram"`

	Ftqq struct {
		PushToken string `mapstructure:"push_token"`
	} `mapstructure:"ftqq"`

	Pushplus struct {
		PushToken string `mapstructure:"push_token"`
	} `mapstructure:"pushplus"`

	Cqhttp struct {
		URL string `mapstructure:"cqhttp_url"`
		QQ  int    `mapstructure:"cqhttp_qq"`
	} `mapstructure:"cqhttp"`

	Smtp struct {
		MailHost string `mapstructure:"mailhost"`
		Port     string `mapstructure:"port"`
		FromAddr string `mapstructure:"fromaddr"`
		ToAddr   string `mapstructure:"toaddr"`
		Username string `mapstructure:"username"`
		Password string `mapstructure:"password"`
	} `mapstructure:"smtp"`

	Wecom struct {
		WechatID string `mapstructure:"wechat_id"`
		Secret   string `mapstructure:"secret"`
		AgentID  string `mapstructure:"agentid"`
		ToUser   string `mapstructure:"touser"`
	} `mapstructure:"wecom"`

	WecomRobot struct {
		URL    string `mapstructure:"url"`
		Mobile string `mapstructure:"mobile"`
	} `mapstructure:"wecomrobot"`

	Pushdeer struct {
		APIURL string `mapstructure:"api_url"`
		Token  string `mapstructure:"token"`
	} `mapstructure:"pushdeer"`

	Dingrobot struct {
		Webhook string `mapstructure:"webhook"`
		Secret  string `mapstructure:"secret"`
	} `mapstructure:"dingrobot"`

	Feishu struct {
		Webhook string `mapstructure:"webhook"`
	} `mapstructure:"feishubot"`

	Bark struct {
		APIURL string `mapstructure:"api_url"`
		Token  string `mapstructure:"token"`
	} `mapstructure:"bark"`

	Gotify struct {
		APIURL   string `mapstructure:"api_url"`
		Token    string `mapstructure:"token"`
		Priority int    `mapstructure:"priority"`
	} `mapstructure:"gotify"`

	Ifttt struct {
		Event string `mapstructure:"event"`
		Key   string `mapstructure:"key"`
	} `mapstructure:"ifttt"`

	Webhook struct {
		URL string `mapstructure:"webhook_url"`
	} `mapstructure:"webhook"`

	Qmsg struct {
		Key string `mapstructure:"key"`
	} `mapstructure:"qmsg"`

	Discord struct {
		Webhook   string `mapstructure:"webhook"`
		VerifySSL bool   `mapstructure:"verify_ssl"`
	} `mapstructure:"discord"`
}

var cfg Config

// ================== 配置加载 ==================
func loadConfig(configPath string) error {
	v := viper.New()
	v.SetConfigFile(configPath)
	v.SetConfigType("yaml")

	if err := v.ReadInConfig(); err != nil {
		logger.Error("配置文件读取失败: %v", err)
	}

	if err := v.Unmarshal(&cfg); err != nil {
		return fmt.Errorf("配置解析失败: %v", err)
	}

	// 设置日志级别
	if cfg.Setting.LogLevel != "" {
		logger.SetLevel(cfg.Setting.LogLevel)
	}

	return nil
}

// ================== HTTP 工具 ==================
func postJSON(url string, body interface{}) ([]byte, error) {
	// 序列化请求体
	js, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}
	// 发送请求
	resp, err := http.Post(url, "application/json", bytes.NewBuffer(js))
	if err != nil {
		return nil, err
	}
	// 读取响应
	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	logger.Debug("Received response from %s - Status: %d, Body: %s", url, resp.StatusCode, string(responseBody))

	return responseBody, nil
}

func postForm(url string, data url.Values) ([]byte, error) {
	// 发送请求
	resp, err := http.PostForm(url, data)
	if err != nil {
		return nil, err
	}
	// 读取响应
	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	logger.Debug("Received response from %s - Status: %d, Body: %s", url, resp.StatusCode, string(responseBody))

	return responseBody, nil
}

// ================== 推送模块 ==================
func telegram(title, msg string) {
	api := cfg.Telegram.APIURL
	token := cfg.Telegram.BotToken
	chat := cfg.Telegram.ChatID
	data := url.Values{
		"chat_id": {chat},
		"text":    {title + "\n" + msg},
	}
	_, err := postForm(fmt.Sprintf("https://%s/bot%s/sendMessage", api, token), data)
	if err != nil {
		logger.Error("Telegram 失败: %v", err)
		return
	}
	logger.Info("Telegram 成功")
}

func ftqq(title, msg string) {
	token := cfg.Ftqq.PushToken
	data := url.Values{"title": {title}, "desp": {msg}}
	_, err := postForm(fmt.Sprintf("https://sctapi.ftqq.com/%s.send", token), data)
	if err != nil {
		logger.Error("Server酱 失败: %v", err)
		return
	}
	logger.Info("Server酱 成功")
}

func pushplus(title, msg string) {
	token := cfg.Pushplus.PushToken
	body := map[string]string{"token": token, "title": title, "content": msg}
	_, err := postJSON("https://www.pushplus.plus/send", body)
	if err != nil {
		logger.Error("Pushplus 失败: %v", err)
		return
	}
	logger.Info("Pushplus 成功")
}

func cqhttp(title, msg string) {
	url := cfg.Cqhttp.URL
	user := cfg.Cqhttp.QQ
	body := map[string]interface{}{"user_id": user, "message": title + "\n" + msg}
	_, err := postJSON(url, body)
	if err != nil {
		logger.Error("CQHTTP 失败: %v", err)
		return
	}
	logger.Info("CQHTTP 成功")
}

func smtpSend(title, msg string) {
	s := cfg.Smtp
	m := fmt.Sprintf("To: %s\r\nSubject: %s\r\n\r\n%s", s.ToAddr, title, msg)
	addr := s.MailHost + ":" + s.Port
	auth := smtp.PlainAuth("", s.Username, s.Password, s.MailHost)
	err := smtp.SendMail(addr, auth, s.FromAddr, []string{s.ToAddr}, []byte(m))
	if err != nil {
		logger.Error("邮件 失败: %v", err)
		return
	}
	logger.Info("邮件 成功")
}

func wecom(title, msg string) {
	s := cfg.Wecom
	tokenResp, err := http.Get(fmt.Sprintf("https://qyapi.weixin.qq.com/cgi-bin/gettoken?corpid=%s&corpsecret=%s", s.WechatID, s.Secret))
	if err != nil {
		logger.Error("WeCom 获取token失败: %v", err)
		return
	}
	defer tokenResp.Body.Close()
	body, _ := io.ReadAll(tokenResp.Body)
	var tk struct {
		AccessToken string `json:"access_token"`
	}
	json.Unmarshal(body, &tk)

	msgBody := map[string]interface{}{
		"agentid": s.AgentID,
		"msgtype": "text",
		"touser":  s.ToUser,
		"text": map[string]string{
			"content": title + "\n" + msg,
		},
	}
	_, err = postJSON(fmt.Sprintf("https://qyapi.weixin.qq.com/cgi-bin/message/send?access_token=%s", tk.AccessToken), msgBody)
	if err != nil {
		logger.Error("WeCom 推送失败: %v", err)
		return
	}
	logger.Info("WeCom 成功")
}

func wecomRobot(title, msg string) {
	s := cfg.WecomRobot
	body := map[string]interface{}{
		"msgtype": "text",
		"text": map[string]interface{}{
			"content":               title + "\n" + msg,
			"mentioned_mobile_list": []string{s.Mobile},
		},
	}
	_, err := postJSON(s.URL, body)
	if err != nil {
		logger.Error("WeCom机器人 失败: %v", err)
		return
	}
	logger.Info("WeCom机器人 成功")
}

func pushdeer(title, msg string) {
	s := cfg.Pushdeer
	params := url.Values{
		"pushkey": {s.Token},
		"text":    {title},
		"desp":    {msg},
		"type":    {"markdown"},
	}
	full := fmt.Sprintf("%s/message/push?%s", s.APIURL, params.Encode())
	_, err := http.Get(full)
	if err != nil {
		logger.Error("PushDeer 失败: %v", err)
		return
	}
	logger.Info("PushDeer 成功")
}

func dingrobot(title, msg string) {
	s := cfg.Dingrobot
	api := s.Webhook
	if s.Secret != "" {
		timestamp := fmt.Sprintf("%d", time.Now().UnixNano()/1e6)
		stringToSign := fmt.Sprintf("%s\n%s", timestamp, s.Secret)
		h := hmac.New(sha256.New, []byte(s.Secret))
		h.Write([]byte(stringToSign))
		sign := url.QueryEscape(base64.StdEncoding.EncodeToString(h.Sum(nil)))
		api = fmt.Sprintf("%s&timestamp=%s&sign=%s", api, timestamp, sign)
	}
	body := map[string]interface{}{
		"msgtype": "text",
		"text":    map[string]string{"content": title + "\n" + msg},
	}
	_, err := postJSON(api, body)
	if err != nil {
		logger.Error("钉钉 失败: %v", err)
		return
	}
	logger.Info("钉钉 成功")
}

func feishu(title, msg string) {
	api := cfg.Feishu.Webhook
	body := map[string]interface{}{
		"msg_type": "text",
		"content":  map[string]string{"text": title + "\n" + msg},
	}
	_, err := postJSON(api, body)
	if err != nil {
		logger.Error("飞书 失败: %v", err)
		return
	}
	logger.Info("飞书 成功")
}

func bark(title, msg string) {
	s := cfg.Bark
	t := url.QueryEscape(title)
	m := url.QueryEscape(msg)
	full := fmt.Sprintf("%s/%s/%s/%s", s.APIURL, s.Token, t, m)
	_, err := http.Get(full)
	if err != nil {
		logger.Error("Bark 失败: %v", err)
		return
	}
	logger.Info("Bark 成功")
}

func gotify(title, msg string) {
	s := cfg.Gotify
	body := map[string]interface{}{
		"title":    title,
		"message":  msg,
		"priority": s.Priority,
	}
	_, err := postJSON(fmt.Sprintf("%s/message?token=%s", s.APIURL, s.Token), body)
	if err != nil {
		logger.Error("Gotify 失败: %v", err)
		return
	}
	logger.Info("Gotify 成功")
}

func ifttt(title, msg string) {
	s := cfg.Ifttt
	body := map[string]string{"value1": title, "value2": msg}
	_, err := postJSON(fmt.Sprintf("https://maker.ifttt.com/trigger/%s/with/key/%s", s.Event, s.Key), body)
	if err != nil {
		logger.Error("IFTTT 失败: %v", err)
		return
	}
	logger.Info("IFTTT 成功")
}

func webhook(title, msg string) {
	api := cfg.Webhook.URL
	body := map[string]string{"title": title, "message": msg}
	_, err := postJSON(api, body)
	if err != nil {
		logger.Error("Webhook 失败: %v", err)
		return
	}
	logger.Info("Webhook 成功")
}

func qmsg(title, msg string) {
	key := cfg.Qmsg.Key
	data := url.Values{"msg": {title + "\n" + msg}}
	_, err := postForm(fmt.Sprintf("https://qmsg.zendee.cn/send/%s", key), data)
	if err != nil {
		logger.Error("Qmsg 失败: %v", err)
		return
	}
	logger.Info("Qmsg 成功")
}

func discord(title, msg string) {
	s := cfg.Discord
	body := map[string]interface{}{
		"username": "Kuro-autosignin",
		"embeds": []map[string]interface{}{
			{
				"title":       title,
				"description": msg,
				"color":       1926125,
				"timestamp":   time.Now().Format(time.RFC3339),
			},
		},
	}
	_, err := postJSON(s.Webhook, body)
	if err != nil {
		logger.Error("Discord 失败: %v", err)
		return
	}
	logger.Info("Discord 成功")
}

// ================== 主逻辑 ==================
func Send(title, msg string) {
	// 使用当前工作目录下的 push.yaml 作为配置文件
	configPath := "push.yaml"

	if _, err := os.Stat(configPath); err != nil {
		if os.IsNotExist(err) {
			logger.Info("未找到推送配置文件，跳过推送")
			return
		}
	}

	err := loadConfig(configPath)
	if err != nil {
		logger.Error("加载配置失败: %v", err)
		return
	}

	servers := cfg.Setting.PushServer
	if servers == "" {
		logger.Info("未配置任何推送方式，跳过推送")
		return
	}

	for _, s := range strings.Split(strings.ToLower(servers), ",") {
		switch strings.TrimSpace(s) {
		case "telegram":
			telegram(title, msg)
		case "ftqq":
			ftqq(title, msg)
		case "pushplus":
			pushplus(title, msg)
		case "cqhttp":
			cqhttp(title, msg)
		case "smtp":
			smtpSend(title, msg)
		case "wecom":
			wecom(title, msg)
		case "wecomrobot":
			wecomRobot(title, msg)
		case "pushdeer":
			pushdeer(title, msg)
		case "dingrobot":
			dingrobot(title, msg)
		case "feishubot":
			feishu(title, msg)
		case "bark":
			bark(title, msg)
		case "gotify":
			gotify(title, msg)
		case "ifttt":
			ifttt(title, msg)
		case "webhook":
			webhook(title, msg)
		case "qmsg":
			qmsg(title, msg)
		case "discord":
			discord(title, msg)
		default:
			logger.Warn("未知推送方式: %s", s)
		}
	}
}

package main

import (
	"log"
	"os"
	"time"

	tele "gopkg.in/telebot.v4"
	"gopkg.in/telebot.v4/middleware"
)

func main() {
	pref := tele.Settings{
		Token:  os.Getenv("TELEGRAM_BOT_TOKEN"),
		Poller: &tele.LongPoller{Timeout: 10 * time.Second},
	}

	b, err := tele.NewBot(pref)
	if err != nil {
		log.Fatal("创建 bot 失败:", err)
		return
	}

	// 启用日志中间件，记录消息更新
	b.Use(middleware.Logger())

	// 回复 /start
	b.Handle("/start", func(c tele.Context) error {
		return c.Send("👋 欢迎使用 Bot！发送 /echo <内容> 可测试回显")
	})

	// 回复 /echo 命令
	b.Handle("/echo", func(c tele.Context) error {
		args := c.Args()
		if len(args) == 0 {
			return c.Send("请在 /echo 后面输入你要回显的内容，例如：/echo hello")
		}
		return c.Send("你说的是: " + c.Message().Payload)
	})

	// 简单键盘示例
	menu := &tele.ReplyMarkup{ResizeKeyboard: true}
	btnPing := menu.Text("Ping Pong")
	btnHelp := menu.Text("Help")
	menu.Reply(menu.Row(btnPing, btnHelp))

	b.Handle("Ping Pong", func(c tele.Context) error {
		return c.Send("Pong 🏓")
	})

	b.Handle("Help", func(c tele.Context) error {
		return c.Send("可用命令：\n/start — 欢迎\n/echo <内容> — 回显\n或按按钮 Ping Pong")
	})

	// 启动 bot
	log.Println("Bot 正在启动…")
	b.Start()
}

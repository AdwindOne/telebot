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

	// 处理普通聊天消息
	b.Handle(tele.OnText, func(c tele.Context) error {
		message := c.Text()
		
		// 根据消息内容进行不同的回复
		switch {
		case message == "你好" || message == "hello" || message == "hi":
			return c.Send("你好！很高兴和你聊天 😊")
		case message == "再见" || message == "bye" || message == "goodbye":
			return c.Send("再见！期待下次聊天 👋")
		case message == "时间" || message == "time":
			return c.Send("当前时间: " + time.Now().Format("2006-01-02 15:04:05"))
		case message == "帮助" || message == "help":
			return c.Send("我可以和你聊天！试试说：\n- 你好\n- 时间\n- 再见\n或者使用命令：\n/start - 开始\n/echo <内容> - 回显")
		default:
			// 回显用户的消息
			return c.Send("你说: " + message + "\n\n💡 提示：发送 '帮助' 查看我能做什么")
		}
	})

	// 处理图片消息
	b.Handle(tele.OnPhoto, func(c tele.Context) error {
		return c.Send("收到了一张图片！📸")
	})

	// 处理语音消息
	b.Handle(tele.OnVoice, func(c tele.Context) error {
		return c.Send("收到了语音消息！🎤")
	})

	// 处理文档消息
	b.Handle(tele.OnDocument, func(c tele.Context) error {
		return c.Send("收到了文档！📄")
	})

	// 处理贴纸消息
	b.Handle(tele.OnSticker, func(c tele.Context) error {
		return c.Send("收到了贴纸！😄")
	})

	// 启动 bot
	log.Println("🤖 Bot 正在启动…")
	log.Println("✅ Bot 启动成功！")
	log.Println("📱 现在可以通过 Telegram 与 Bot 聊天了")
	log.Println("🔗 支持的聊天类型：文本、图片、语音、文档、贴纸")
	log.Println("💬 试试发送：你好、时间、帮助")
	log.Println("⏹️  按 Ctrl+C 停止 Bot")
	
	b.Start()
}

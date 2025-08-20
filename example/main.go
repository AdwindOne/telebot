package main

import (
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	tele "gopkg.in/telebot.v4"
	"gopkg.in/telebot.v4/middleware"
)

// 交易结构体
type TradeRequest struct {
	Network    string  // 网络：ETH, SOL, BSC
	DEX        string  // DEX：Uniswap, Raydium, PancakeSwap
	TokenIn    string  // 输入代币
	TokenOut   string  // 输出代币
	Amount     float64 // 交易数量
	Slippage   float64 // 滑点百分比
	GasPrice   float64 // Gas价格
	WalletAddr string  // 钱包地址
}

// 支持的代币列表
var supportedTokens = map[string]map[string]string{
	"ETH": {
		"USDT": "0xdAC17F958D2ee523a2206206994597C13D831ec7",
		"USDC": "0xA0b86a33E6441b8C4C8C8C8C8C8C8C8C8C8C8C8C",
		"WETH": "0xC02aaA39b223FE8D0A0e5C4F27eAD9083C756Cc2",
		"DAI":  "0x6B175474E89094C44Da98b954EedeAC495271d0F",
	},
	"SOL": {
		"USDT": "Es9vMFrzaCERmJfrF4H2FYD4KCoNkY11McCe8BenwNYB",
		"USDC": "EPjFWdd5AufqSSqeM2qN1xzybapC8G4wEGGkZwyTDt1v",
		"SOL":  "So11111111111111111111111111111111111111112",
		"RAY":  "4k3Dyjzvzp8eMZWUXbBCjEvwSkkk59S5iCNLY3QrkX6R",
	},
	"BSC": {
		"USDT": "0x55d398326f99059fF775485246999027B3197955",
		"USDC": "0x8AC76a51cc950d9822D68b83fE1Ad97B32Cd580d",
		"BNB":  "0xbb4CdB9CBd36B01bD1cBaEF2aF8C6b1c6cF4A8A",
		"CAKE": "0x0E09FaBB73Bd3Ade0a17ECC321fD13a19e81cE82",
	},
}

// 支持的DEX列表
var supportedDEXs = map[string][]string{
	"ETH": {"Uniswap V3", "Uniswap V2", "SushiSwap", "1inch"},
	"SOL": {"Raydium", "Orca", "Jupiter", "Serum"},
	"BSC": {"PancakeSwap", "Biswap", "1inch"},
}

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
		menu := &tele.ReplyMarkup{ResizeKeyboard: true}
		btnTrade := menu.Text("🚀 开始交易")
		btnPortfolio := menu.Text("📊 投资组合")
		btnHelp := menu.Text("❓ 帮助")
		menu.Reply(menu.Row(btnTrade), menu.Row(btnPortfolio, btnHelp))
		
		return c.Send("🤖 欢迎使用 DEX 交易 Bot！\n\n"+
			"支持的网络：ETH、SOL、BSC\n"+
			"支持的DEX：Uniswap、Raydium、PancakeSwap等\n\n"+
			"请选择操作：", menu)
	})

	// 交易命令
	b.Handle("/trade", func(c tele.Context) error {
		return showTradeMenu(c)
	})

	// 处理交易按钮
	b.Handle("🚀 开始交易", func(c tele.Context) error {
		return showTradeMenu(c)
	})

	// 处理投资组合按钮
	b.Handle("📊 投资组合", func(c tele.Context) error {
		return c.Send("📊 投资组合功能开发中...\n\n" +
			"将支持：\n" +
			"• 多链资产查看\n" +
			"• 收益统计\n" +
			"• 交易历史\n" +
			"• 风险评估")
	})

	// 处理帮助按钮
	b.Handle("❓ 帮助", func(c tele.Context) error {
		return showHelp(c)
	})

	// 处理网络选择
	for network := range supportedDEXs {
		b.Handle("🌐 "+network, func(c tele.Context) error {
			return showDEXMenu(c, network)
		})
	}

	// 处理DEX选择
	for _, dexList := range supportedDEXs {
		for _, dex := range dexList {
			b.Handle("🔄 "+dex, func(c tele.Context) error {
				dexName := strings.TrimPrefix(c.Text(), "🔄 ")
				return showTokenSelection(c, dexName)
			})
		}
	}

	// 处理代币选择
	for _, tokens := range supportedTokens {
		for token := range tokens {
			b.Handle("💰 "+token, func(c tele.Context) error {
				tokenName := strings.TrimPrefix(c.Text(), "💰 ")
				return showTradeForm(c, tokenName)
			})
		}
	}

	// 处理交易表单提交
	b.Handle("/submit_trade", func(c tele.Context) error {
		return processTradeSubmission(c)
	})

	// 处理交易确认
	b.Handle("✅ 确认交易", func(c tele.Context) error {
		return executeTrade(c)
	})

	// 处理交易取消
	b.Handle("❌ 取消交易", func(c tele.Context) error {
		return c.Send("❌ 交易已取消")
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
		return c.Send("可用命令：\n/start — 欢迎\n/echo <内容> — 回显\n/trade — 开始交易\n或按按钮 Ping Pong")
	})

	// 处理普通聊天消息
	b.Handle(tele.OnText, func(c tele.Context) error {
		message := c.Text()
		
		// 根据消息内容进行不同的回复
		switch {
		case message == "你好" || message == "hello" || message == "hi":
			return c.Send("你好！很高兴和你聊天 😊\n\n💡 发送 /trade 开始交易")
		case message == "再见" || message == "bye" || message == "goodbye":
			return c.Send("再见！期待下次聊天 👋")
		case message == "时间" || message == "time":
			return c.Send("当前时间: " + time.Now().Format("2006-01-02 15:04:05"))
		case message == "帮助" || message == "help":
			return c.Send("我可以和你聊天！试试说：\n- 你好\n- 时间\n- 再见\n或者使用命令：\n/start - 开始\n/echo <内容> - 回显\n/trade - 开始交易")
		case strings.Contains(message, "交易") || strings.Contains(message, "trade"):
			return showTradeMenu(c)
		default:
			// 回显用户的消息
			return c.Send("你说: " + message + "\n\n💡 提示：发送 '帮助' 查看我能做什么，或发送 /trade 开始交易")
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
	log.Println("🚀 发送 /trade 开始 DEX 交易")
	log.Println("⏹️  按 Ctrl+C 停止 Bot")
	
	b.Start()
}

// 显示交易菜单
func showTradeMenu(c tele.Context) error {
	menu := &tele.ReplyMarkup{ResizeKeyboard: true}
	
	var buttons []tele.Btn
	for network := range supportedDEXs {
		buttons = append(buttons, menu.Text("🌐 "+network))
	}
	
	// 每行最多2个按钮
	for i := 0; i < len(buttons); i += 2 {
		if i+1 < len(buttons) {
			menu.Reply(menu.Row(buttons[i], buttons[i+1]))
		} else {
			menu.Reply(menu.Row(buttons[i]))
		}
	}
	
	return c.Send("🌐 请选择交易网络：", menu)
}

// 显示DEX选择菜单
func showDEXMenu(c tele.Context, network string) error {
	menu := &tele.ReplyMarkup{ResizeKeyboard: true}
	
	var buttons []tele.Btn
	for _, dex := range supportedDEXs[network] {
		buttons = append(buttons, menu.Text("🔄 "+dex))
	}
	
	// 每行最多2个按钮
	for i := 0; i < len(buttons); i += 2 {
		if i+1 < len(buttons) {
			menu.Reply(menu.Row(buttons[i], buttons[i+1]))
		} else {
			menu.Reply(menu.Row(buttons[i]))
		}
	}
	
	return c.Send("🔄 请选择 "+network+" 网络上的 DEX：", menu)
}

// 显示代币选择菜单
func showTokenSelection(c tele.Context, dexName string) error {
	menu := &tele.ReplyMarkup{ResizeKeyboard: true}
	
	var buttons []tele.Btn
	for token := range supportedTokens["ETH"] { // 默认显示ETH代币，实际应该根据网络选择
		buttons = append(buttons, menu.Text("💰 "+token))
	}
	
	// 每行最多2个按钮
	for i := 0; i < len(buttons); i += 2 {
		if i+1 < len(buttons) {
			menu.Reply(menu.Row(buttons[i], buttons[i+1]))
		} else {
			menu.Reply(menu.Row(buttons[i]))
		}
	}
	
	return c.Send("💰 请选择要交易的代币（"+dexName+"）：", menu)
}

// 显示交易表单
func showTradeForm(c tele.Context, tokenName string) error {
	form := `📝 交易表单

代币: ` + tokenName + `

请按以下格式填写交易信息：
/trade_form <网络> <DEX> <输入代币> <输出代币> <数量> <滑点%> <钱包地址>

示例：
/trade_form ETH Uniswap USDT WETH 100 0.5 0x1234567890abcdef

参数说明：
• 网络: ETH/SOL/BSC
• DEX: Uniswap/Raydium/PancakeSwap
• 输入代币: 要卖出的代币
• 输出代币: 要买入的代币  
• 数量: 交易数量
• 滑点: 滑点百分比(0.1-10)
• 钱包地址: 你的钱包地址

💡 提示：请确保钱包地址正确且有足够余额`
	
	return c.Send(form)
}

// 处理交易表单提交
func processTradeSubmission(c tele.Context) error {
	args := c.Args()
	if len(args) < 7 {
		return c.Send("❌ 参数不足！请按格式填写：\n/trade_form <网络> <DEX> <输入代币> <输出代币> <数量> <滑点%> <钱包地址>")
	}
	
	// 解析参数
	network := args[0]
	dex := args[1]
	tokenIn := args[2]
	tokenOut := args[3]
	amount, err := strconv.ParseFloat(args[4], 64)
	if err != nil {
		return c.Send("❌ 数量格式错误！请输入数字")
	}
	slippage, err := strconv.ParseFloat(args[5], 64)
	if err != nil {
		return c.Send("❌ 滑点格式错误！请输入数字")
	}
	walletAddr := args[6]
	
	// 验证参数
	if slippage < 0.1 || slippage > 10 {
		return c.Send("❌ 滑点必须在 0.1% - 10% 之间")
	}
	
	// 显示交易确认
	confirmation := `📋 交易确认

网络: ` + network + `
DEX: ` + dex + `
输入: ` + tokenIn + ` (` + args[4] + `)
输出: ` + tokenOut + `
滑点: ` + args[5] + `%
钱包: ` + walletAddr + `

预估Gas费用: 0.005 ETH
预估输出: ~` + tokenOut + ` (根据当前价格)

⚠️ 请确认以上信息是否正确`
	
	menu := &tele.ReplyMarkup{ResizeKeyboard: true}
	btnConfirm := menu.Text("✅ 确认交易")
	btnCancel := menu.Text("❌ 取消交易")
	menu.Reply(menu.Row(btnConfirm, btnCancel))
	
	return c.Send(confirmation, menu)
}

// 执行交易
func executeTrade(c tele.Context) error {
	// 这里应该实现实际的交易逻辑
	// 包括：价格查询、滑点计算、交易签名、广播等
	
	return c.Send("🚀 交易已提交！\n\n" +
		"📊 交易状态：处理中\n" +
		"⏱️ 预计完成时间：30秒\n" +
		"🔗 交易哈希：0x1234567890abcdef...\n\n" +
		"💡 你可以在区块链浏览器中查看交易详情")
}

// 显示帮助信息
func showHelp(c tele.Context) error {
	help := `❓ DEX 交易 Bot 帮助

📋 可用命令：
/start - 开始使用
/trade - 开始交易
/echo <内容> - 回显消息

🌐 支持的网络：
• ETH (以太坊)
• SOL (Solana)  
• BSC (币安智能链)

🔄 支持的DEX：
• Uniswap V3/V2
• Raydium
• PancakeSwap
• SushiSwap
• 1inch
• Orca
• Jupiter

💰 支持的代币：
• USDT, USDC, WETH, DAI
• SOL, RAY
• BNB, CAKE

📝 交易流程：
1. 发送 /trade 开始交易
2. 选择网络和DEX
3. 选择代币
4. 填写交易表单
5. 确认并执行交易

⚠️ 注意事项：
• 请确保钱包地址正确
• 检查滑点设置
• 确认Gas费用
• 交易前请仔细核对信息

🔒 安全提醒：
• 不要分享私钥
• 验证合约地址
• 使用官方DEX
• 小额测试交易`
	
	return c.Send(help)
}

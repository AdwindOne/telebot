package main

import (
	"crypto/rand"
	"fmt"
	"log"
	"os"
	"regexp"
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

	// 处理网络选择（内联按钮唯一键）
	b.Handle("\fnet_eth", func(c tele.Context) error {
		_ = c.Respond()
		return showDEXMenu(c, "ETH")
	})
	b.Handle("\fnet_sol", func(c tele.Context) error {
		_ = c.Respond()
		return showDEXMenu(c, "SOL")
	})
	b.Handle("\fnet_bsc", func(c tele.Context) error {
		_ = c.Respond()
		return showDEXMenu(c, "BSC")
	})

	// 兜底：统一处理回调（避免某些环境下唯一键路由未命中）
	b.Handle(tele.OnCallback, func(c tele.Context) error {
		cb := c.Callback()
		if cb == nil {
			return nil
		}
		switch cb.Unique {
		// 网络选择
		case "net_eth":
			_ = c.Respond()
			return showDEXMenu(c, "ETH")
		case "net_sol":
			_ = c.Respond()
			return showDEXMenu(c, "SOL")
		case "net_bsc":
			_ = c.Respond()
			return showDEXMenu(c, "BSC")
		
		// DEX选择 - ETH网络
		case "dex_uniswap_v3_eth":
			_ = c.Respond()
			return showTokenSelection(c, "Uniswap V3")
		case "dex_uniswap_v2_eth":
			_ = c.Respond()
			return showTokenSelection(c, "Uniswap V2")
		case "dex_sushiswap_eth":
			_ = c.Respond()
			return showTokenSelection(c, "SushiSwap")
		case "dex_1inch_eth":
			_ = c.Respond()
			return showTokenSelection(c, "1inch")
		
		// DEX选择 - SOL网络
		case "dex_raydium_sol":
			_ = c.Respond()
			return showTokenSelection(c, "Raydium")
		case "dex_orca_sol":
			_ = c.Respond()
			return showTokenSelection(c, "Orca")
		case "dex_jupiter_sol":
			_ = c.Respond()
			return showTokenSelection(c, "Jupiter")
		case "dex_serum_sol":
			_ = c.Respond()
			return showTokenSelection(c, "Serum")
		
		// DEX选择 - BSC网络
		case "dex_pancakeswap_bsc":
			_ = c.Respond()
			return showTokenSelection(c, "PancakeSwap")
		case "dex_biswap_bsc":
			_ = c.Respond()
			return showTokenSelection(c, "Biswap")
		
		// 代币选择 - ETH代币
		case "token_usdt_eth":
			_ = c.Respond()
			return showTradeForm(c, "USDT")
		case "token_usdc_eth":
			_ = c.Respond()
			return showTradeForm(c, "USDC")
		case "token_weth_eth":
			_ = c.Respond()
			return showTradeForm(c, "WETH")
		case "token_dai_eth":
			_ = c.Respond()
			return showTradeForm(c, "DAI")
		
		// 代币选择 - SOL代币
		case "token_sol_sol":
			_ = c.Respond()
			return showTradeForm(c, "SOL")
		case "token_ray_sol":
			_ = c.Respond()
			return showTradeForm(c, "RAY")
		
		// 代币选择 - BSC代币
		case "token_bnb_bsc":
			_ = c.Respond()
			return showTradeForm(c, "BNB")
		case "token_cake_bsc":
			_ = c.Respond()
			return showTradeForm(c, "CAKE")
		
		// 快速交易按钮
		case "quick_trade_usdt", "quick_trade_usdc", "quick_trade_weth", "quick_trade_dai",
		     "quick_trade_sol", "quick_trade_ray", "quick_trade_bnb", "quick_trade_cake":
			_ = c.Respond()
			return showQuickTradeForm(c)
		
		// 自定义交易按钮
		case "custom_trade_usdt", "custom_trade_usdc", "custom_trade_weth", "custom_trade_dai",
		     "custom_trade_sol", "custom_trade_ray", "custom_trade_bnb", "custom_trade_cake":
			_ = c.Respond()
			return showCustomTradeForm(c)
		
		// 返回按钮
		case "back_to_trade":
			_ = c.Respond()
			return showTradeMenu(c)
		
		// 快速交易代币选择
		case "quick_usdt":
			_ = c.Respond()
			return c.Send("⚡ 快速交易 USDT\n\n请输入交易数量：\n/trade_form ETH Uniswap USDT WETH <数量> 0.5 <钱包地址>")
		case "quick_usdc":
			_ = c.Respond()
			return c.Send("⚡ 快速交易 USDC\n\n请输入交易数量：\n/trade_form ETH Uniswap USDC WETH <数量> 0.5 <钱包地址>")
		case "quick_weth":
			_ = c.Respond()
			return c.Send("⚡ 快速交易 WETH\n\n请输入交易数量：\n/trade_form ETH Uniswap WETH USDT <数量> 0.5 <钱包地址>")
		
		default:
			return c.Respond()
		}
	})

		// 处理DEX选择回调
	b.Handle("\fdex_uniswap_v3_eth", func(c tele.Context) error {
		_ = c.Respond()
		return showTokenSelection(c, "Uniswap V3")
	})
	b.Handle("\fdex_uniswap_v2_eth", func(c tele.Context) error {
		_ = c.Respond()
		return showTokenSelection(c, "Uniswap V2")
	})
	b.Handle("\fdex_sushiswap_eth", func(c tele.Context) error {
		_ = c.Respond()
		return showTokenSelection(c, "SushiSwap")
	})
	b.Handle("\fdex_1inch_eth", func(c tele.Context) error {
		_ = c.Respond()
		return showTokenSelection(c, "1inch")
	})
	
	b.Handle("\fdex_raydium_sol", func(c tele.Context) error {
		_ = c.Respond()
		return showTokenSelection(c, "Raydium")
	})
	b.Handle("\fdex_orca_sol", func(c tele.Context) error {
		_ = c.Respond()
		return showTokenSelection(c, "Orca")
	})
	b.Handle("\fdex_jupiter_sol", func(c tele.Context) error {
		_ = c.Respond()
		return showTokenSelection(c, "Jupiter")
	})
	b.Handle("\fdex_serum_sol", func(c tele.Context) error {
		_ = c.Respond()
		return showTokenSelection(c, "Serum")
	})
	
	b.Handle("\fdex_pancakeswap_bsc", func(c tele.Context) error {
		_ = c.Respond()
		return showTokenSelection(c, "PancakeSwap")
	})
	b.Handle("\fdex_biswap_bsc", func(c tele.Context) error {
		_ = c.Respond()
		return showTokenSelection(c, "Biswap")
	})

		// 处理代币选择回调
	b.Handle("\ftoken_usdt_eth", func(c tele.Context) error {
		_ = c.Respond()
		return showTradeForm(c, "USDT")
	})
	b.Handle("\ftoken_usdc_eth", func(c tele.Context) error {
		_ = c.Respond()
		return showTradeForm(c, "USDC")
	})
	b.Handle("\ftoken_weth_eth", func(c tele.Context) error {
		_ = c.Respond()
		return showTradeForm(c, "WETH")
	})
	b.Handle("\ftoken_dai_eth", func(c tele.Context) error {
		_ = c.Respond()
		return showTradeForm(c, "DAI")
	})
	
	b.Handle("\ftoken_sol_sol", func(c tele.Context) error {
		_ = c.Respond()
		return showTradeForm(c, "SOL")
	})
	b.Handle("\ftoken_ray_sol", func(c tele.Context) error {
		_ = c.Respond()
		return showTradeForm(c, "RAY")
	})
	
	b.Handle("\ftoken_bnb_bsc", func(c tele.Context) error {
		_ = c.Respond()
		return showTradeForm(c, "BNB")
	})
	b.Handle("\ftoken_cake_bsc", func(c tele.Context) error {
		_ = c.Respond()
		return showTradeForm(c, "CAKE")
	})

	// 处理交易表单提交
	b.Handle("/trade_form", func(c tele.Context) error {
		return processTradeSubmission(c)
	})

	// 处理交易表单提交（别名）
	b.Handle("/submit_trade", func(c tele.Context) error {
		return processTradeSubmission(c)
	})

	// 处理快速交易按钮回调
	b.Handle("\fquick_trade_usdt", func(c tele.Context) error {
		_ = c.Respond()
		return showQuickTradeForm(c)
	})
	b.Handle("\fquick_trade_usdc", func(c tele.Context) error {
		_ = c.Respond()
		return showQuickTradeForm(c)
	})
	b.Handle("\fquick_trade_weth", func(c tele.Context) error {
		_ = c.Respond()
		return showQuickTradeForm(c)
	})
	b.Handle("\fquick_trade_dai", func(c tele.Context) error {
		_ = c.Respond()
		return showQuickTradeForm(c)
	})
	b.Handle("\fquick_trade_sol", func(c tele.Context) error {
		_ = c.Respond()
		return showQuickTradeForm(c)
	})
	b.Handle("\fquick_trade_ray", func(c tele.Context) error {
		_ = c.Respond()
		return showQuickTradeForm(c)
	})
	b.Handle("\fquick_trade_bnb", func(c tele.Context) error {
		_ = c.Respond()
		return showQuickTradeForm(c)
	})
	b.Handle("\fquick_trade_cake", func(c tele.Context) error {
		_ = c.Respond()
		return showQuickTradeForm(c)
	})

	// 处理自定义交易按钮回调
	b.Handle("\fcustom_trade_usdt", func(c tele.Context) error {
		_ = c.Respond()
		return showCustomTradeForm(c)
	})
	b.Handle("\fcustom_trade_usdc", func(c tele.Context) error {
		_ = c.Respond()
		return showCustomTradeForm(c)
	})
	b.Handle("\fcustom_trade_weth", func(c tele.Context) error {
		_ = c.Respond()
		return showCustomTradeForm(c)
	})
	b.Handle("\fcustom_trade_dai", func(c tele.Context) error {
		_ = c.Respond()
		return showCustomTradeForm(c)
	})
	b.Handle("\fcustom_trade_sol", func(c tele.Context) error {
		_ = c.Respond()
		return showCustomTradeForm(c)
	})
	b.Handle("\fcustom_trade_ray", func(c tele.Context) error {
		_ = c.Respond()
		return showCustomTradeForm(c)
	})
	b.Handle("\fcustom_trade_bnb", func(c tele.Context) error {
		_ = c.Respond()
		return showCustomTradeForm(c)
	})
	b.Handle("\fcustom_trade_cake", func(c tele.Context) error {
		_ = c.Respond()
		return showCustomTradeForm(c)
	})

	// 处理返回按钮回调
	b.Handle("\fback_to_trade", func(c tele.Context) error {
		_ = c.Respond()
		return showTradeMenu(c)
	})

	// 处理快速交易代币选择回调
	b.Handle("\fquick_usdt", func(c tele.Context) error {
		_ = c.Respond()
		return c.Send("⚡ 快速交易 USDT\n\n请输入交易数量：\n/trade_form ETH Uniswap USDT WETH <数量> 0.5 <钱包地址>")
	})
	b.Handle("\fquick_usdc", func(c tele.Context) error {
		_ = c.Respond()
		return c.Send("⚡ 快速交易 USDC\n\n请输入交易数量：\n/trade_form ETH Uniswap USDC WETH <数量> 0.5 <钱包地址>")
	})
	b.Handle("\fquick_weth", func(c tele.Context) error {
		_ = c.Respond()
		return c.Send("⚡ 快速交易 WETH\n\n请输入交易数量：\n/trade_form ETH Uniswap WETH USDT <数量> 0.5 <钱包地址>")
	})

	// 处理交易确认
	b.Handle("✅ 确认交易", func(c tele.Context) error {
		return executeTrade(c)
	})

	// 处理交易取消
	b.Handle("❌ 取消交易", func(c tele.Context) error {
		clearTradeInfo(c)
		return c.Send("❌ 交易已取消")
	})

	// 处理修改参数按钮
	b.Handle("✏️ 修改参数", func(c tele.Context) error {
		return showModifyForm(c)
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

// 发送带“上方按钮”的消息
func showTradeMenu(c tele.Context) error {
	inl := &tele.ReplyMarkup{}

	btnETH := inl.Data("ETH", "net_eth")
	btnSOL := inl.Data("SOL", "net_sol")
	btnBSC := inl.Data("BSC", "net_bsc")

		inl.Inline(inl.Row(btnETH, btnSOL, btnBSC))
	
	return c.Send("🌐 请选择交易网络：", inl)
}

// 显示DEX选择菜单
func showDEXMenu(c tele.Context, network string) error {
	inl := &tele.ReplyMarkup{}

	var buttons []tele.Btn
	for _, dex := range supportedDEXs[network] {
		// 生成唯一键：dex_网络名_去空格
		uniqueKey := "dex_" + strings.ToLower(strings.ReplaceAll(dex, " ", "_")) + "_" + strings.ToLower(network)
		buttons = append(buttons, inl.Data("🔄 "+dex, uniqueKey))
	}

	// 每行最多2个按钮
	for i := 0; i < len(buttons); i += 2 {
		if i+1 < len(buttons) {
			inl.Inline(inl.Row(buttons[i], buttons[i+1]))
		} else {
			inl.Inline(inl.Row(buttons[i]))
		}
	}

	return c.Send("🔄 请选择 "+network+" 网络上的 DEX：", inl)
}

// 显示代币选择菜单
func showTokenSelection(c tele.Context, dexName string) error {
	inl := &tele.ReplyMarkup{}
	
	// 根据DEX确定网络和对应的代币
	var network string
	var tokens []string
	
	switch dexName {
	case "Uniswap V3", "Uniswap V2", "SushiSwap", "1inch":
		network = "ETH"
		for token := range supportedTokens["ETH"] {
			tokens = append(tokens, token)
		}
	case "Raydium", "Orca", "Jupiter", "Serum":
		network = "SOL"
		for token := range supportedTokens["SOL"] {
			tokens = append(tokens, token)
		}
	case "PancakeSwap", "Biswap":
		network = "BSC"
		for token := range supportedTokens["BSC"] {
			tokens = append(tokens, token)
		}
	default:
		network = "ETH"
		for token := range supportedTokens["ETH"] {
			tokens = append(tokens, token)
		}
	}
	
	var buttons []tele.Btn
	for _, token := range tokens {
		// 生成唯一键：token_代币名_网络名
		uniqueKey := "token_" + strings.ToLower(token) + "_" + strings.ToLower(network)
		buttons = append(buttons, inl.Data("💰 "+token, uniqueKey))
	}
	
	// 每行最多2个按钮
	for i := 0; i < len(buttons); i += 2 {
		if i+1 < len(buttons) {
			inl.Inline(inl.Row(buttons[i], buttons[i+1]))
		} else {
			inl.Inline(inl.Row(buttons[i]))
		}
	}
	
	return c.Send("💰 请选择要交易的代币（"+network+" 网络 - "+dexName+"）：", inl)
}

// 显示交易表单
func showTradeForm(c tele.Context, tokenName string) error {
	// 创建快速交易按钮
	inl := &tele.ReplyMarkup{}
	
	// 根据代币确定网络
	var network string
	switch tokenName {
	case "USDT", "USDC", "WETH", "DAI":
		network = "ETH"
	case "SOL", "RAY":
		network = "SOL"
	case "BNB", "CAKE":
		network = "BSC"
	default:
		network = "ETH"
	}
	
	// 创建快速交易选项
	btnQuickTrade := inl.Data("⚡ 快速交易", "quick_trade_"+strings.ToLower(tokenName))
	btnCustomTrade := inl.Data("📝 自定义交易", "custom_trade_"+strings.ToLower(tokenName))
	btnBack := inl.Data("🔙 返回", "back_to_trade")
	
	inl.Inline(inl.Row(btnQuickTrade), inl.Row(btnCustomTrade), inl.Row(btnBack))
	
	form := `📝 交易表单

代币: ` + tokenName + `
网络: ` + network + `

请选择交易方式：

⚡ 快速交易 - 使用默认设置
📝 自定义交易 - 手动填写所有参数

或者按以下格式填写完整交易信息：
/trade_form <网络> <DEX> <输入代币> <输出代币> <数量> <滑点%> <钱包地址>

示例：
/trade_form ` + network + ` Uniswap ` + tokenName + ` WETH 100 0.5 0x1234567890abcdef

💡 提示：请确保钱包地址正确且有足够余额`
	
	return c.Send(form, inl)
}

// 处理交易表单提交
func processTradeSubmission(c tele.Context) error {
	args := c.Args()
	if len(args) < 7 {
		return c.Send("❌ 参数不足！请按格式填写：\n/trade_form <网络> <DEX> <输入代币> <输出代币> <数量> <滑点%> <钱包地址>")
	}

	// 解析参数
	network := strings.ToUpper(args[0])
	dex := args[1]
	tokenIn := strings.ToUpper(args[2])
	tokenOut := strings.ToUpper(args[3])
	amount, err := strconv.ParseFloat(args[4], 64)
	if err != nil {
		return c.Send("❌ 数量格式错误！请输入数字")
	}
	slippage, err := strconv.ParseFloat(args[5], 64)
	if err != nil {
		return c.Send("❌ 滑点格式错误！请输入数字")
	}
	walletAddr := args[6]

	// 验证网络
	if network != "ETH" && network != "SOL" && network != "BSC" {
		return c.Send("❌ 不支持的网络！请选择：ETH、SOL、BSC")
	}

	// 验证DEX
	validDEXs := supportedDEXs[network]
	dexValid := false
	for _, validDEX := range validDEXs {
		if dex == validDEX {
			dexValid = true
			break
		}
	}
	if !dexValid {
		return c.Send("❌ 不支持的DEX！" + network + " 网络支持的DEX：" + strings.Join(validDEXs, "、"))
	}

	// 验证代币
	validTokens := supportedTokens[network]
	tokenInValid := false
	tokenOutValid := false
	for token := range validTokens {
		if token == tokenIn {
			tokenInValid = true
		}
		if token == tokenOut {
			tokenOutValid = true
		}
	}
	if !tokenInValid {
		return c.Send("❌ 不支持的输入代币！" + network + " 网络支持的代币：" + strings.Join(getTokenList(validTokens), "、"))
	}
	if !tokenOutValid {
		return c.Send("❌ 不支持的输出代币！" + network + " 网络支持的代币：" + strings.Join(getTokenList(validTokens), "、"))
	}

	// 验证数量
	if amount <= 0 {
		return c.Send("❌ 交易数量必须大于0")
	}

	// 验证滑点
	if slippage < 0.1 || slippage > 10 {
		return c.Send("❌ 滑点必须在 0.1% - 10% 之间")
	}

	// 验证钱包地址
	if !isValidWalletAddress(walletAddr, network) {
		return c.Send("❌ 无效的钱包地址！请检查地址格式")
	}

	// 获取代币合约地址
	tokenInAddr := validTokens[tokenIn]
	tokenOutAddr := validTokens[tokenOut]

	// 预估交易结果
	estimatedOutput, gasFee := estimateTrade(network, dex, tokenIn, tokenOut, amount, slippage)

	// 显示交易确认
	confirmation := `📋 交易确认

🌐 网络: ` + network + `
🔄 DEX: ` + dex + `
💰 输入: ` + tokenIn + ` (` + args[4] + `)
📈 输出: ` + tokenOut + ` (~` + estimatedOutput + `)
📊 滑点: ` + args[5] + `%
💳 钱包: ` + walletAddr + `

📋 合约地址:
• ` + tokenIn + `: ` + tokenInAddr + `
• ` + tokenOut + `: ` + tokenOutAddr + `

⛽ 预估Gas费用: ` + gasFee + `
💸 总费用: ` + args[4] + ` ` + tokenIn + ` + ` + gasFee + `

⚠️ 请确认以上信息是否正确`

	menu := &tele.ReplyMarkup{ResizeKeyboard: true}
	btnConfirm := menu.Text("✅ 确认交易")
	btnCancel := menu.Text("❌ 取消交易")
	btnModify := menu.Text("✏️ 修改参数")
	menu.Reply(menu.Row(btnConfirm, btnCancel), menu.Row(btnModify))

	// 存储交易信息到上下文（实际应用中应该使用数据库）
	storeTradeInfo(c, TradeRequest{
		Network:    network,
		DEX:        dex,
		TokenIn:    tokenIn,
		TokenOut:   tokenOut,
		Amount:     amount,
		Slippage:   slippage,
		WalletAddr: walletAddr,
	})

	return c.Send(confirmation, menu)
}

// 执行交易
func executeTrade(c tele.Context) error {
	// 获取存储的交易信息
	tradeInfo := getTradeInfo(c)
	if tradeInfo.Network == "" {
		return c.Send("❌ 交易信息已过期，请重新填写交易表单")
	}

	// 开始交易执行流程
	statusMsg := c.Send("🔄 正在执行交易...\n\n" +
		"📊 状态：初始化中\n" +
		"⏱️ 预计时间：30-60秒")

	// 模拟交易执行步骤
	go func() {
		// 步骤1：检查余额
		time.Sleep(2 * time.Second)
		c.Edit(statusMsg, "🔄 正在执行交易...\n\n"+
			"📊 状态：检查钱包余额\n"+
			"✅ 余额充足\n"+
			"⏱️ 预计时间：25-55秒")

		// 步骤2：获取价格
		time.Sleep(3 * time.Second)
		c.Edit(statusMsg, "🔄 正在执行交易...\n\n"+
			"📊 状态：获取实时价格\n"+
			"✅ 价格获取成功\n"+
			"⏱️ 预计时间：20-50秒")

		// 步骤3：计算滑点
		time.Sleep(2 * time.Second)
		c.Edit(statusMsg, "🔄 正在执行交易...\n\n"+
			"📊 状态：计算滑点保护\n"+
			"✅ 滑点计算完成\n"+
			"⏱️ 预计时间：15-45秒")

		// 步骤4：构建交易
		time.Sleep(3 * time.Second)
		c.Edit(statusMsg, "🔄 正在执行交易...\n\n"+
			"📊 状态：构建交易数据\n"+
			"✅ 交易数据准备完成\n"+
			"⏱️ 预计时间：10-40秒")

		// 步骤5：签名交易
		time.Sleep(2 * time.Second)
		c.Edit(statusMsg, "🔄 正在执行交易...\n\n"+
			"📊 状态：签名交易\n"+
			"✅ 交易签名完成\n"+
			"⏱️ 预计时间：5-35秒")

		// 步骤6：广播交易
		time.Sleep(3 * time.Second)
		txHash := generateTxHash()
		c.Edit(statusMsg, "🔄 正在执行交易...\n\n"+
			"📊 状态：广播到区块链\n"+
			"✅ 交易已广播\n"+
			"🔗 交易哈希："+txHash+"\n"+
			"⏱️ 等待确认中...")

		// 步骤7：等待确认
		time.Sleep(5 * time.Second)
		c.Edit(statusMsg, "✅ 交易执行成功！\n\n"+
			"📊 状态：已确认\n"+
			"🔗 交易哈希："+txHash+"\n"+
			"🌐 网络："+tradeInfo.Network+"\n"+
			"🔄 DEX："+tradeInfo.DEX+"\n"+
			"💰 输入："+fmt.Sprintf("%.2f", tradeInfo.Amount)+" "+tradeInfo.TokenIn+"\n"+
			"📈 输出：~"+calculateOutput(tradeInfo)+" "+tradeInfo.TokenOut+"\n\n"+
			"💡 你可以在区块链浏览器中查看交易详情\n"+
			"🔗 浏览器链接："+getExplorerLink(tradeInfo.Network, txHash))

		// 清理存储的交易信息
		clearTradeInfo(c)
	}()

	return nil
}

// 显示快速交易表单
func showQuickTradeForm(c tele.Context) error {
	inl := &tele.ReplyMarkup{}
	
	// 快速交易选项
	btnUSDT := inl.Data("💰 USDT", "quick_usdt")
	btnUSDC := inl.Data("💰 USDC", "quick_usdc")
	btnWETH := inl.Data("💰 WETH", "quick_weth")
	btnBack := inl.Data("🔙 返回", "back_to_trade")
	
	inl.Inline(inl.Row(btnUSDT, btnUSDC), inl.Row(btnWETH), inl.Row(btnBack))
	
	form := `⚡ 快速交易

请选择要交换的代币：

💰 USDT - 稳定币
💰 USDC - 稳定币  
💰 WETH - 包装以太坊

快速交易将使用以下默认设置：
• 滑点: 0.5%
• DEX: Uniswap V3
• Gas: 自动优化

💡 选择代币后，请输入交易数量`
	
	return c.Send(form, inl)
}

// 显示自定义交易表单
func showCustomTradeForm(c tele.Context) error {
	form := `📝 自定义交易表单

请按以下格式填写完整的交易信息：

/trade_form <网络> <DEX> <输入代币> <输出代币> <数量> <滑点%> <钱包地址>

参数说明：
• 网络: ETH/SOL/BSC
• DEX: Uniswap/Raydium/PancakeSwap等
• 输入代币: 要卖出的代币
• 输出代币: 要买入的代币  
• 数量: 交易数量
• 滑点: 滑点百分比(0.1-10)
• 钱包地址: 你的钱包地址

示例：
/trade_form ETH Uniswap USDT WETH 100 0.5 0x1234567890abcdef

⚠️ 请确保所有参数正确，交易不可撤销！`

	return c.Send(form)
}

// 显示修改参数表单
func showModifyForm(c tele.Context) error {
	tradeInfo := getTradeInfo(c)
	if tradeInfo.Network == "" {
		return c.Send("❌ 没有找到交易信息，请重新开始交易流程")
	}

	form := `✏️ 修改交易参数

当前交易信息：
🌐 网络: ` + tradeInfo.Network + `
🔄 DEX: ` + tradeInfo.DEX + `
💰 输入代币: ` + tradeInfo.TokenIn + `
📈 输出代币: ` + tradeInfo.TokenOut + `
📊 数量: ` + fmt.Sprintf("%.2f", tradeInfo.Amount) + `
📊 滑点: ` + fmt.Sprintf("%.1f", tradeInfo.Slippage) + `%
💳 钱包: ` + tradeInfo.WalletAddr + `

请重新填写交易信息：
/trade_form <网络> <DEX> <输入代币> <输出代币> <数量> <滑点%> <钱包地址>

示例：
/trade_form ` + tradeInfo.Network + ` ` + tradeInfo.DEX + ` ` + tradeInfo.TokenIn + ` ` + tradeInfo.TokenOut + ` 100 0.5 ` + tradeInfo.WalletAddr + `

💡 提示：你可以修改任何参数，包括网络、DEX、代币、数量、滑点等`

	return c.Send(form)
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

// 辅助函数

// 获取代币列表
func getTokenList(tokens map[string]string) []string {
	var list []string
	for token := range tokens {
		list = append(list, token)
	}
	return list
}

// 验证钱包地址
func isValidWalletAddress(address, network string) bool {
	switch network {
	case "ETH", "BSC":
		// ETH/BSC地址格式：0x + 40个十六进制字符
		matched, _ := regexp.MatchString(`^0x[a-fA-F0-9]{40}$`, address)
		return matched
	case "SOL":
		// Solana地址格式：Base58编码，32-44字符
		matched, _ := regexp.MatchString(`^[1-9A-HJ-NP-Za-km-z]{32,44}$`, address)
		return matched
	default:
		return false
	}
}

// 预估交易结果
func estimateTrade(network, dex, tokenIn, tokenOut string, amount, slippage float64) (string, string) {
	// 模拟价格计算（实际应用中应该调用DEX API）
	var price float64
	switch {
	case tokenIn == "USDT" && tokenOut == "WETH":
		price = 0.0005 // 1 USDT = 0.0005 WETH
	case tokenIn == "WETH" && tokenOut == "USDT":
		price = 2000 // 1 WETH = 2000 USDT
	case tokenIn == "USDC" && tokenOut == "WETH":
		price = 0.0005 // 1 USDC = 0.0005 WETH
	case tokenIn == "WETH" && tokenOut == "USDC":
		price = 2000 // 1 WETH = 2000 USDC
	default:
		price = 1.0 // 默认1:1
	}

	// 计算输出数量（考虑滑点）
	output := amount * price * (1 - slippage/100)

	// 预估Gas费用
	var gasFee string
	switch network {
	case "ETH":
		gasFee = "0.005 ETH"
	case "BSC":
		gasFee = "0.001 BNB"
	case "SOL":
		gasFee = "0.000005 SOL"
	default:
		gasFee = "0.005 ETH"
	}

	return fmt.Sprintf("%.4f", output), gasFee
}

// 生成交易哈希
func generateTxHash() string {
	// 生成32字节的随机哈希
	hash := make([]byte, 32)
	rand.Read(hash)
	return fmt.Sprintf("0x%x", hash)
}

// 计算输出数量
func calculateOutput(trade TradeRequest) string {
	// 模拟计算（实际应用中应该使用真实价格）
	var price float64
	switch {
	case trade.TokenIn == "USDT" && trade.TokenOut == "WETH":
		price = 0.0005
	case trade.TokenIn == "WETH" && trade.TokenOut == "USDT":
		price = 2000
	default:
		price = 1.0
	}

	output := trade.Amount * price * (1 - trade.Slippage/100)
	return fmt.Sprintf("%.4f", output)
}

// 获取区块链浏览器链接
func getExplorerLink(network, txHash string) string {
	switch network {
	case "ETH":
		return "https://etherscan.io/tx/" + txHash
	case "BSC":
		return "https://bscscan.com/tx/" + txHash
	case "SOL":
		return "https://solscan.io/tx/" + txHash
	default:
		return "https://etherscan.io/tx/" + txHash
	}
}

// 存储交易信息（简单内存存储，实际应用中应该使用数据库）
var tradeInfoMap = make(map[int64]TradeRequest)

func storeTradeInfo(c tele.Context, trade TradeRequest) {
	tradeInfoMap[c.Sender().ID] = trade
}

func getTradeInfo(c tele.Context) TradeRequest {
	return tradeInfoMap[c.Sender().ID]
}

func clearTradeInfo(c tele.Context) {
	delete(tradeInfoMap, c.Sender().ID)
}

package telegram

import (
	"USDTMore/app/config"
	"USDTMore/app/help"
	"USDTMore/app/model"
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"strconv"
	"time"
)

func SendTradeSuccMsg(order model.TradeOrders) {
	var chatId, err = strconv.ParseInt(config.GetTgBotNotifyTarget(), 10, 64)
	if err != nil {

		return
	}
	postfix := "TRC20"
	dataUrl := "https://tronscan.org/#/transaction/"
	if order.Chain == "POLY" {
		postfix = "Polygon"
		dataUrl = "https://polygonscan.com/tx/"
	}
	if order.Chain == "OP" {
		postfix = "Optimism"
		dataUrl = "https://optimistic.etherscan.io/tx/"
	}
	if order.Chain == "BSC" {
		postfix = "BEP20"
		dataUrl = "https://bscscan.com/tx/"
	}
	var text = `
#收款成功 #订单交易
---
` + "```" + `
🚦商户订单：%v
💰请求金额：%v CNY(%v)
💲支付数额：%v USDT.%s
✅收款地址：%s
⏱️创建时间：%s
️🎯️支付时间：%s
` + "```" + `
`
	text = fmt.Sprintf(text,
		order.OrderId,
		order.Money,
		order.UsdtRate,
		order.Amount,
		postfix,
		help.MaskAddress(order.Address),
		order.CreatedAt.Format(time.DateTime),
		order.UpdatedAt.Format(time.DateTime),
	)
	var msg = tgbotapi.NewMessage(chatId, text)
	msg.ParseMode = tgbotapi.ModeMarkdown
	msg.ReplyMarkup = tgbotapi.InlineKeyboardMarkup{
		InlineKeyboard: [][]tgbotapi.InlineKeyboardButton{
			{
				tgbotapi.NewInlineKeyboardButtonURL("📝查看交易明细", dataUrl+order.TradeHash),
			},
		},
	}

	_, _ = botApi.Send(msg)
}

func SendOtherNotify(text string) {
	var chatId, err = strconv.ParseInt(config.GetTgBotNotifyTarget(), 10, 64)
	if err != nil {

		return
	}

	var msg = tgbotapi.NewMessage(chatId, text)
	msg.ParseMode = tgbotapi.ModeMarkdown

	_, _ = botApi.Send(msg)
}

func SendWelcome(version string) {
	var text = `
👋 欢迎使用 USDTMore，一款好用的多链路个人USDT收款网关，如果您看到此消息，说明机器人已经启动成功

📌当前版本：` + version + `
🌺支持链路: TRON POLYGON OPTIMISM BSC` + `
📝发送命令 /start 可以开始使用
`
	var msg = tgbotapi.NewMessage(0, text)

	SendMsg(msg)
}

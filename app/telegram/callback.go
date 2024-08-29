package telegram

import (
	"USDTMore/app/config"
	"USDTMore/app/help"
	"USDTMore/app/log"
	"USDTMore/app/model"
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/tidwall/gjson"
	"io"
	"math/big"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

const cbWallet = "wallet"
const cbAddress = "address"
const cbAddressAdd = "address_add"
const cbAddressEnable = "address_enable"
const cbAddressDisable = "address_disable"
const cbAddressDelete = "address_del"
const cbAddressOtherNotify = "address_other_notify"
const cbOrderDetail = "order_detail"

/*
查询钱包信息
*/
func cbWalletAction(query *tgbotapi.CallbackQuery, address string) {

	_addresses := strings.Split(strings.TrimSpace(address), ":")
	var info = ""
	var msg = tgbotapi.NewMessage(query.Message.Chat.ID, "❌查询失败")
	var detailUrl = ""
	switch _addresses[0] {
	case "TRON":
		info = getWalletInfoByTRONAddress(_addresses[1])
		detailUrl = "https://tronscan.org/#/address/" + _addresses[1]
		break
	case "POLY":
		info = getWalletInfoByPOLAddress(_addresses[1])
		detailUrl = "https://polygonscan.com/address/" + _addresses[1]
		break
	case "OP":
		info = getWalletInfoByOPTAddress(_addresses[1])
		detailUrl = "https://optimistic.etherscan.io/address/" + _addresses[1]
		break
	case "BSC":
		info = getWalletInfoByBSCAddress(_addresses[1])
		detailUrl = "https://bscscan.com/address/" + _addresses[1]
		break
	default:
		msg = tgbotapi.NewMessage(query.Message.Chat.ID, "💔输入格式不正确，使用链路代码开头[TRON:|POLY:|OPT:|BSC:]，比如: TRON:xxxxx")
		break
	}

	if info != "" {
		msg.Text = info
		msg.ReplyMarkup = tgbotapi.InlineKeyboardMarkup{
			InlineKeyboard: [][]tgbotapi.InlineKeyboardButton{
				{
					tgbotapi.NewInlineKeyboardButtonURL("📝查看详细信息", detailUrl),
				},
			},
		}
	}

	DeleteMsg(query.Message.MessageID)
	_, _ = botApi.Send(msg)
}

func cbAddressAddHandle(query *tgbotapi.CallbackQuery) {
	var msg = tgbotapi.NewMessage(query.Message.Chat.ID, replayAddressText)
	msg.ReplyMarkup = tgbotapi.ForceReply{ForceReply: true, Selective: true, InputFieldPlaceholder: "输入钱包地址"}

	_, _ = botApi.Send(msg)
}

func cbAddressAction(query *tgbotapi.CallbackQuery, id string) {
	var wa model.WalletAddress
	if model.DB.Where("id = ?", id).First(&wa).Error == nil {
		var otherTextLabel = "✅已启用 非订单交易监控通知"
		if wa.OtherNotify != 1 {
			otherTextLabel = "❌已禁用 非订单交易监控通知"
		}

		EditAndSendMsg(query.Message.MessageID, wa.Address, tgbotapi.InlineKeyboardMarkup{
			InlineKeyboard: [][]tgbotapi.InlineKeyboardButton{
				{
					tgbotapi.NewInlineKeyboardButtonData("✅启用", cbAddressEnable+"|"+id),
					tgbotapi.NewInlineKeyboardButtonData("❌禁用", cbAddressDisable+"|"+id),
					tgbotapi.NewInlineKeyboardButtonData("⛔️删除", cbAddressDelete+"|"+id),
				},
				{
					tgbotapi.NewInlineKeyboardButtonData(otherTextLabel, cbAddressOtherNotify+"|"+id),
				},
			},
		})
	}
}

func cbAddressEnableAction(query *tgbotapi.CallbackQuery, id string) {
	var wa model.WalletAddress
	if model.DB.Where("id = ?", id).First(&wa).Error == nil {
		// 修改地址状态
		wa.SetStatus(model.StatusEnable)

		// 删除历史消息
		DeleteMsg(query.Message.MessageID)

		// 推送最新状态
		cmdStartHandle()
	}
}

func cbAddressDisableAction(query *tgbotapi.CallbackQuery, id string) {
	var wa model.WalletAddress
	if model.DB.Where("id = ?", id).First(&wa).Error == nil {
		// 修改地址状态
		wa.SetStatus(model.StatusDisable)

		// 删除历史消息
		DeleteMsg(query.Message.MessageID)

		// 推送最新状态
		cmdStartHandle()
	}
}

func cbAddressDeleteAction(query *tgbotapi.CallbackQuery, id string) {
	var wa model.WalletAddress
	if model.DB.Where("id = ?", id).First(&wa).Error == nil {
		// 删除钱包地址
		wa.Delete()

		// 删除历史消息
		DeleteMsg(query.Message.MessageID)

		// 推送最新状态
		cmdStartHandle()
	}
}

func cbAddressOtherNotifyAction(query *tgbotapi.CallbackQuery, id string) {
	var wa model.WalletAddress
	if model.DB.Where("id = ?", id).First(&wa).Error == nil {
		if wa.OtherNotify == 1 {
			wa.SetOtherNotify(model.OtherNotifyDisable)
		} else {
			wa.SetOtherNotify(model.OtherNotifyEnable)
		}

		DeleteMsg(query.Message.MessageID)

		cmdStartHandle()
	}
}

func cbOrderDetailAction(tradeId string) {
	var o model.TradeOrders

	if model.DB.Where("trade_id = ?", tradeId).First(&o).Error == nil {
		var urlInfo, er2 = url.Parse(o.NotifyUrl)
		if er2 != nil {
			log.Error("商户网站地址解析错误：" + er2.Error())

			return
		}

		var _notifyStateLabel = "✅ 回调成功"
		if o.NotifyState != model.OrderNotifyStateSucc {
			_notifyStateLabel = "❌ 回调失败"
		}
		if model.OrderStatusWaiting == o.Status {
			_notifyStateLabel = o.GetStatusLabel()
		}
		if model.OrderStatusExpired == o.Status {
			_notifyStateLabel = "🈚️ 没有回调"
		}

		var _site = &url.URL{Scheme: urlInfo.Scheme, Host: urlInfo.Host}

		detailUrl := "https://tronscan.org/#/transaction/"
		if o.Chain == "POLY" {
			detailUrl = "https://polygonscan.com/tx/"
		}
		if o.Chain == "OP" {
			detailUrl = "https://optimistic.etherscan.io/tx/"
		}
		if o.Chain == "BSC" {
			detailUrl = "https://bscscan.com/tx/"
		}

		var _msg = tgbotapi.NewMessage(0, "```"+`
📌 订单ID：`+o.OrderId+`
📊 交易汇率：`+o.UsdtRate+`(`+config.GetUsdtRateRaw()+`)
💰 交易金额：`+fmt.Sprintf("%.2f", o.Money)+` CNY
💲 交易数额：`+o.Amount+` USDT
🌏 商户网站：`+_site.String()+`
🔋 收款状态：`+o.GetStatusLabel()+`
🍀 回调状态：`+_notifyStateLabel+`
📯 收款链路：`+o.Chain+`
💎️ 收款地址：`+help.MaskAddress(o.Address)+`
🕒 创建时间：`+o.CreatedAt.Format(time.DateTime)+`
🕒 失效时间：`+o.ExpiredAt.Format(time.DateTime)+`
⚖️️ 确认时间：`+o.ConfirmedAt.Format(time.DateTime)+`
`+"\n```")
		_msg.ParseMode = tgbotapi.ModeMarkdown
		_msg.ReplyMarkup = tgbotapi.InlineKeyboardMarkup{
			InlineKeyboard: [][]tgbotapi.InlineKeyboardButton{
				{
					tgbotapi.NewInlineKeyboardButtonURL("🌏商户网站", _site.String()),
					tgbotapi.NewInlineKeyboardButtonURL("📝交易明细", detailUrl+o.TradeHash),
				},
			},
		}

		SendMsg(_msg)
	}
}

/*
获取TRC20的信息
*/
func getWalletInfoByTRONAddress(address string) string {
	var url = "https://apilist.tronscanapi.com/api/accountv2?address=" + address
	var client = http.Client{Timeout: time.Second * 5}
	resp, err := client.Get(url)
	if err != nil {
		log.Error("GetWalletInfoByAddress client.Get(url)", err)

		return ""
	}

	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		log.Error("GetWalletInfoByAddress resp.StatusCode != 200", resp.StatusCode, err)

		return ""
	}

	all, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Error("GetWalletInfoByAddress io.ReadAll(resp.Body)", err)

		return ""
	}
	result := gjson.ParseBytes(all)

	var dateCreated = time.UnixMilli(result.Get("date_created").Int())
	var latestOperationTime = time.UnixMilli(result.Get("latest_operation_time").Int())
	var netRemaining = result.Get("bandwidth.netRemaining").Int() + result.Get("bandwidth.freeNetRemaining").Int()
	var netLimit = result.Get("bandwidth.netLimit").Int() + result.Get("bandwidth.freeNetLimit").Int()
	var text = `
☘️ 查询地址：` + address + `
🚀 查询链路：TRC20
💰 TRX   余额：0.00 TRX
💲 USDT余额：0.00 USDT
📬 交易数量：` + result.Get("totalTransactionCount").String() + `
📈 转账数量：↑ ` + result.Get("transactions_out").String() + ` ↓ ` + result.Get("transactions_in").String() + `
📡 宽带资源：` + fmt.Sprintf("%v", netRemaining) + ` / ` + fmt.Sprintf("%v", netLimit) + ` 
🔋 能量资源：` + result.Get("bandwidth.energyRemaining").String() + ` / ` + result.Get("bandwidth.energyLimit").String() + `
⏰ 创建时间：` + dateCreated.Format(time.DateTime) + `
⏰ 最后活动：` + latestOperationTime.Format(time.DateTime) + `
`

	for _, v := range result.Get("withPriceTokens").Array() {
		if v.Get("tokenName").String() == "trx" {
			text = strings.Replace(text, "0.00 TRX", fmt.Sprintf("%.2f TRX", v.Get("balance").Float()/1000000), 1)
		}
		if v.Get("tokenName").String() == "Tether USD" {

			text = strings.Replace(text, "0.00 USDT", fmt.Sprintf("%.2f USDT", v.Get("balance").Float()/1000000), 1)
		}
	}

	return text
}

/*
请求ETH兼容的链
*/
func requestAddress(baseUrl string, query string) []byte {
	var url = baseUrl + "?" + query
	var client = http.Client{Timeout: time.Second * 5}
	resp, err := client.Get(url)
	if err != nil {
		log.Error("GetWalletInfoByAddress client.Get(url)", err)
		return nil
	}

	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		log.Error("GetWalletInfoByAddress resp.StatusCode != 200", resp.StatusCode, err)
		return nil
	}

	all, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Error("GetWalletInfoByAddress io.ReadAll(resp.Body)", err)
		return nil
	}
	//result := gjson.ParseBytes(all)

	return all
}

func getWalletInfoETH(name string, unit string, chain string, host string, apiKey string, contractAddress string, address string) string {
	// 这里计算的是ETH余额
	var queryETH = "module=account&action=balance&address=" + address + "&apikey=" + apiKey
	allETH := requestAddress(host, queryETH)
	resultETH := gjson.ParseBytes(allETH)

	// 将余额从 Wei 转换为 ETH
	balanceWei, err := strconv.ParseFloat(resultETH.Get("result").String(), 64)
	if err != nil {
		log.Error("GetWalletInfoByAddress convert into ETH", err)
	}
	balanceETH := balanceWei / 1e18 // 1 ETH = 10^18 Wei
	//

	// 这里计算的是ERC-20余额
	var queryUSDT = "module=account&action=tokenbalance&contractaddress=" + contractAddress + "&address=" + address + "&tag=latest" + "&apikey=" + apiKey
	allUSDT := requestAddress(host, queryUSDT)
	resultUSDT := gjson.ParseBytes(allUSDT)

	// 将余额从最小单位转换为标准单位
	var rawValue = resultUSDT.Get("result").String()
	if chain == "BSC" {
		length := len(rawValue) - 12
		if length > 0 {
			rawValue = rawValue[0:length]
		}
	}
	balanceStandard, ok := new(big.Int).SetString(rawValue, 10)
	if ok {
		log.Error("GetWalletInfoByAddress convert into USDT")
	}
	balanceFloat := new(big.Float).SetInt(balanceStandard)
	balanceUSDT := new(big.Float).Quo(balanceFloat, big.NewFloat(1e6)) // USDT 有 6 位小数

	// 累计所有交易的 Value 来计算总交易量
	var wa model.WalletAddress
	var text = ""
	if model.DB.Where("chain = ? and address = ?", chain, address).First(&wa).Error == nil {
		// 这里查询订单历史
		var queryTx = "module=account&action=tokentx&contractaddress=" + contractAddress + "&address=" + address + "&startblock=" + strconv.FormatInt(wa.StartBlock+1, 10) + "&endblock=" + strconv.FormatInt(wa.StartBlock+999999999999, 10) + "&sort=asc" + "&apikey=" + apiKey
		allTx := requestAddress(host, queryTx)
		resultTx := gjson.ParseBytes(allTx)

		totalInValue := new(big.Float).SetFloat64(wa.InAmount)
		totalOutValue := new(big.Float).SetFloat64(wa.OutAmount)
		totalCount := big.NewInt(wa.Count)
		timeNow := time.Now()
		threeHourAgo := timeNow.Add(-3 * time.Hour)
		var dateCreated = wa.CreatedAt
		var latestOperationTime = wa.UpdatedAt

		for i, tx := range resultTx.Get("result").Array() {
			// 判断是否是第一条
			var txTime = time.UnixMilli(tx.Get("timeStamp").Int() * 1000)
			if i == 0 {
				dateCreated = txTime
				if txTime.Before(wa.CreatedAt) {
					wa.CreatedAt = txTime
				}
			}
			// 判断是否是最后一条
			if i == len(resultTx.Get("result").Array())-1 {
				latestOperationTime = txTime
				if txTime.After(wa.UpdatedAt) {
					wa.UpdatedAt = txTime
				}
			}

			var rawValue = resultUSDT.Get("result").String()
			value, ok := new(big.Int).SetString(rawValue, 10)
			tokenDecimals, _ := strconv.ParseInt(tx.Get("tokenDecimal").String(), 10, 32)
			tokenSymbol := tx.Get("tokenSymbol").String()

			if tokenSymbol == "USDT" || tokenSymbol == "BSC-USD" {
				decimalFactor := new(big.Int).Exp(big.NewInt(10), big.NewInt(tokenDecimals), nil)
				valueUSDT := new(big.Float).Quo(new(big.Float).SetInt(value), new(big.Float).SetInt(decimalFactor))

				from := tx.Get("from").String()
				to := tx.Get("to").String()
				if !ok {
					log.Error("解析交易金额失败: %s", rawValue)
				}
				if from == strings.ToLower(address) {
					totalOutValue.Add(totalOutValue, valueUSDT)
				}
				if to == strings.ToLower(address) {
					totalInValue.Add(totalInValue, valueUSDT)
				}
				totalCount.Add(totalCount, big.NewInt(1))
			}

			// 超过3小时的话，就更新时间戳, 合计进/出，当前并不使用
			if txTime.Before(threeHourAgo) {
				wa.StartBlock = tx.Get("blockNumber").Int()
				wa.InAmount, _ = totalInValue.Float64()
				wa.OutAmount, _ = totalOutValue.Float64()
				wa.Count = totalCount.Int64()
			}
			model.DB.Save(wa)
		}

		text = `
☘️ 查询地址：` + address + `
🚀 查询链路：CHAINNAME
💲 ETH余额：0.00 ETH
💲 USDT余额：0.00 USDT
📬 交易数量：` + totalCount.String() + `
📈 转账数量：↑ ` + totalOutValue.String() + ` ↓ ` + totalInValue.String() + `
⏰ 首单时间：` + dateCreated.Format(time.DateTime) + `
⏰ 最后活动：` + latestOperationTime.Format(time.DateTime) + `
`
		// 替换余额
		text = strings.Replace(text, "CHAINNAME", name, 1)
		text = strings.Replace(text, "ETH余额", unit+"余额", 1)
		text = strings.Replace(text, "0.00 ETH", fmt.Sprintf("%.8f %s", balanceETH, unit), 1)
		text = strings.Replace(text, "0.00 USDT", fmt.Sprintf("%.2f USDT", balanceUSDT), 1)
	}

	return text
}

/*
获取Polygon的信息
*/
func getWalletInfoByPOLAddress(address string) string {
	return getWalletInfoETH("Polygon", "MATIC", "POLY", "https://api.polygonscan.com/api", config.GetPolygonScanApiKey(), config.GetPolygonScanContractAddress(), address)
}

/*
获取Optimism的信息
*/
func getWalletInfoByOPTAddress(address string) string {
	return getWalletInfoETH("Optimism", "ETH", "OP", "https://api-optimistic.etherscan.io/api", config.GetOptimismExplorerApiKey(), config.GetOptimismExplorerContractAddress(), address)
}

/*
获取BEP20的信息
*/
func getWalletInfoByBSCAddress(address string) string {
	return getWalletInfoETH("BEP20", "BNB", "BSC", "https://api.bscscan.com/api", config.GetBscExplorerApiKey(), config.GetBscExplorerContractAddress(), address)
}

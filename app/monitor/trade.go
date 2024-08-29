package monitor

import (
	"USDTMore/app/config"
	"USDTMore/app/help"
	"USDTMore/app/log"
	"USDTMore/app/model"
	"USDTMore/app/notify"
	"USDTMore/app/telegram"
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/shopspring/decimal"
	"github.com/tidwall/gjson"
	"io"
	"math/big"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

// okX的智能合约地址
const usdtToken = "TR7NHqjeKQxGTCi8q8ZY4pL8otSzgjLj6t"

func TradeStart() {
	log.Info("交易监控启动.")

	for range time.Tick(time.Second * 5) {
		var recentTransferTotal float64
		var _lock, err = getAllPendingOrders()
		if err != nil {
			log.Error(err.Error())
			continue
		}

		// 这里是TRON网络的监控
		for _, _row := range model.GetAvailableAddress("TRON") {
			var result gjson.Result
			var err error

			if config.IsTronScanApi() {
				result, err = getUsdtTrc20TransByTronScan(_row.Address)
			} else {
				result, err = getUsdtTrc20TransByTronGrid(_row.Address)
			}
			if err != nil {
				log.Error(err.Error())
				continue
			}

			if config.IsTronScanApi() {
				recentTransferTotal = result.Get("total").Num
			} else {
				recentTransferTotal = result.Get("meta.page_size").Num
			}
			log.Info(fmt.Sprintf("[%s] recent transfer total: %s(%v)", config.GetTronServerApi(), _row.Address, recentTransferTotal))
			if recentTransferTotal <= 0 { // 没有交易记录
				continue
			}

			if config.IsTronScanApi() {
				handlePaymentTransactionForTronScan(_lock, _row.Address, result)
				handleOtherNotifyForTronScan(_row.Address, result)
			} else {
				handlePaymentTransactionForTronGrid(_lock, _row.Address, result)
				handleOtherNotifyForTronGrid(_row.Address, result)
			}
		}

		// 这里是POLYGON网络的监控
		for _, _row := range model.GetAvailableAddress("POLY") {
			var result gjson.Result
			var err error

			result, err = getUsdtPolygonTransByPolygonScan(_row.Address)
			if err != nil {
				log.Error(err.Error())
				continue
			}

			handlePaymentTransactionForPolygonScan(_lock, _row.Address, result)
			handleOtherNotifyForPolygonScan(_row.Address, result)
		}

		// 这里是OPTIMISM网络的监控
		for _, _row := range model.GetAvailableAddress("OP") {
			var result gjson.Result
			var err error

			result, err = getUsdtOptimismTransByOptimismExplorer(_row.Address)
			if err != nil {
				log.Error(err.Error())
				continue
			}

			handlePaymentTransactionForOptimismExplorer(_lock, _row.Address, result)
			handleOtherNotifyForOptimismExplorer(_row.Address, result)
		}

		// 这里是BSC网络的监控
		for _, _row := range model.GetAvailableAddress("BSC") {
			var result gjson.Result
			var err error

			result, err = getUsdtBscTransByBscScan(_row.Address)
			if err != nil {
				log.Error(err.Error())
				continue
			}

			handlePaymentTransactionForBscScan(_lock, _row.Address, result)
			handleOtherNotifyForBscScan(_row.Address, result)
		}
	}
}

/*
列出所有等待支付的交易订单
*/
func getAllPendingOrders() (map[string]model.TradeOrders, error) {
	tradeOrders, err := model.GetTradeOrderByStatus(model.OrderStatusWaiting)
	if err != nil {
		return nil, fmt.Errorf("待支付订单获取失败: %w", err)
	}

	var _lock = make(map[string]model.TradeOrders) // 当前所有正在等待支付的订单 Lock Key
	for _, order := range tradeOrders {
		if time.Now().Unix() >= order.ExpiredAt.Unix() { // 订单过期
			err := order.OrderSetExpired()
			if err != nil {
				log.Error("订单过期标记失败：", err, order.OrderId)
			} else {
				log.Info("订单过期：", order.OrderId)
			}
			continue
		}
		_lock[order.Chain+order.Address+order.Amount] = order
	}
	return _lock, nil
}

// 处理支付交易 TronScan
func handlePaymentTransactionForTronScan(_lock map[string]model.TradeOrders, _toAddress string, _data gjson.Result) {
	for _, transfer := range _data.Get("token_transfers").Array() {
		if strings.ToLower(transfer.Get("to_address").String()) != strings.ToLower(_toAddress) {
			// 不是接收地址
			continue
		}

		// 计算交易金额
		var _rawQuant, _quant = parseTransAmount(transfer.Get("quant").Float())
		if !inPaymentAmountRange(_rawQuant) {
			continue
		}

		// 订单锁信息
		_order, ok := _lock["TRON"+_toAddress+_quant]
		if !ok || transfer.Get("contractRet").String() != "SUCCESS" {
			// 订单不存在或交易失败
			continue
		}

		// 判断时间是否有效
		var _createdAt = time.UnixMilli(transfer.Get("block_ts").Int())
		if _createdAt.Unix() < _order.CreatedAt.Unix() || _createdAt.Unix() > _order.ExpiredAt.Unix() {
			// 失效交易
			continue
		}

		var _transId = transfer.Get("transaction_id").String()
		var _fromAddress = transfer.Get("from_address").String()
		if _order.OrderSetSucc(_fromAddress, _transId, _createdAt) == nil {
			// 通知订单支付成功
			go notify.OrderNotify(_order)
			// TG发送订单信息
			go telegram.SendTradeSuccMsg(_order)
		}
	}
}

// 处理支付交易 TronGrid
func handlePaymentTransactionForTronGrid(_lock map[string]model.TradeOrders, _toAddress string, result gjson.Result) {
	for _, transfer := range result.Get("data").Array() {
		if strings.ToLower(transfer.Get("to").String()) != strings.ToLower(_toAddress) {
			// 不是接收地址
			continue
		}

		// 计算交易金额
		var _rawQuant, _quant = parseTransAmount(transfer.Get("value").Float())
		if !inPaymentAmountRange(_rawQuant) {
			continue
		}

		_order, ok := _lock["TRON"+_toAddress+_quant]
		if !ok || transfer.Get("type").String() != "Transfer" {
			// 订单不存在或交易失败
			continue
		}

		// 判断时间是否有效
		var _createdAt = time.UnixMilli(transfer.Get("block_timestamp").Int())
		if _createdAt.Unix() < _order.CreatedAt.Unix() || _createdAt.Unix() > _order.ExpiredAt.Unix() {
			// 失效交易
			continue
		}

		var _transId = transfer.Get("transaction_id").String()
		var _fromAddress = transfer.Get("from").String()
		if _order.OrderSetSucc(_fromAddress, _transId, _createdAt) == nil {
			// 通知订单支付成功
			go notify.OrderNotify(_order)

			// TG发送订单信息
			go telegram.SendTradeSuccMsg(_order)
		}
	}
}

// 处理支付交易 ETH兼容网络
func handlePaymentTransactionForETH(_lock map[string]model.TradeOrders, _toChain string, _toAddress string, result gjson.Result) {
	for _, transfer := range result.Get("result").Array() {
		if strings.ToLower(transfer.Get("to").String()) != strings.ToLower(_toAddress) {
			// 不是接收地址
			continue
		}

		tokenSymbol := transfer.Get("tokenSymbol").String()
		// 不是USDT入账直接跳出
		if tokenSymbol != "USDT" && tokenSymbol != "BSC-USD" {
			continue
		}

		// 计算交易金额
		value, _ := new(big.Int).SetString(transfer.Get("value").String(), 10)
		tokenDecimals, _ := strconv.ParseInt(transfer.Get("tokenDecimal").String(), 10, 32)
		decimalFactor := new(big.Int).Exp(big.NewInt(10), big.NewInt(tokenDecimals), nil)
		valueUSDT := new(big.Float).Quo(new(big.Float).SetInt(value), new(big.Float).SetInt(decimalFactor))
		bigFloatStr := valueUSDT.Text('f', -1)
		decimalUSDT, _ := decimal.NewFromString(bigFloatStr)
		// 如果不在范围直接跳出
		if !inPaymentAmountRange(decimalUSDT) {
			continue
		}

		_order, ok := _lock[_toChain+_toAddress+decimalUSDT.String()]
		if !ok {
			// 订单不存在或交易失败
			continue
		}

		// 判断时间是否有效
		var _createdAt = time.UnixMilli(transfer.Get("timeStamp").Int() * 1000)
		if _createdAt.Unix() < _order.CreatedAt.Unix() || _createdAt.Unix() > _order.ExpiredAt.Unix() {
			// 失效交易
			continue
		}

		var _transId = transfer.Get("hash").String()
		var _fromAddress = transfer.Get("from").String()
		if _order.OrderSetSucc(_fromAddress, _transId, _createdAt) == nil {
			// 通知订单支付成功
			go notify.OrderNotify(_order)
			// TG发送订单信息
			go telegram.SendTradeSuccMsg(_order)
		}
	}
}

func handlePaymentTransactionForPolygonScan(_lock map[string]model.TradeOrders, _toAddress string, result gjson.Result) {
	handlePaymentTransactionForETH(_lock, "POLY", _toAddress, result)
}
func handlePaymentTransactionForOptimismExplorer(_lock map[string]model.TradeOrders, _toAddress string, result gjson.Result) {
	handlePaymentTransactionForETH(_lock, "OP", _toAddress, result)
}
func handlePaymentTransactionForBscScan(_lock map[string]model.TradeOrders, _toAddress string, result gjson.Result) {
	handlePaymentTransactionForETH(_lock, "BSC", _toAddress, result)
}

// 非订单交易通知
func handleOtherNotifyForTronScan(_toAddress string, result gjson.Result) {
	for _, transfer := range result.Get("token_transfers").Array() {
		if !model.GetOtherNotify("TRON", _toAddress) {
			break
		}

		var _rawAmount, _amount = parseTransAmount(transfer.Get("quant").Float())
		if !inPaymentAmountRange(_rawAmount) {
			continue
		}

		var _created = time.UnixMilli(transfer.Get("block_ts").Int())
		var _txid = transfer.Get("transaction_id").String()
		var _detailUrl = "https://tronscan.org/#/transaction/" + _txid
		if !model.IsNeedNotifyByTxid(_txid) {
			// 不需要额外通知
			continue
		}

		var title = "收入"
		if strings.ToLower(transfer.Get("to_address").String()) != strings.ToLower(_toAddress) {
			title = "支出"
		}

		var text = fmt.Sprintf(
			"#账户%s #非订单交易\n---\n```\n💲交易数额：%v USDT\n⏱️交易时间：%v\n✅接收地址：%v\n🅾️发送地址：%v```\n",
			title,
			_amount,
			_created.Format(time.DateTime),
			help.MaskAddress(transfer.Get("to_address").String()),
			help.MaskAddress(transfer.Get("from_address").String()),
		)

		var chatId, err = strconv.ParseInt(config.GetTgBotNotifyTarget(), 10, 64)
		if err != nil {
			continue
		}

		var msg = tgbotapi.NewMessage(chatId, text)
		msg.ParseMode = tgbotapi.ModeMarkdown
		msg.ReplyMarkup = tgbotapi.InlineKeyboardMarkup{
			InlineKeyboard: [][]tgbotapi.InlineKeyboardButton{
				{
					tgbotapi.NewInlineKeyboardButtonURL("📝查看交易明细", _detailUrl),
				},
			},
		}

		var _record = model.NotifyRecord{Txid: _txid}
		model.DB.Create(&_record)

		go telegram.SendMsg(msg)
	}
}

// 非订单交易通知
func handleOtherNotifyForTronGrid(_toAddress string, result gjson.Result) {
	for _, transfer := range result.Get("data").Array() {
		if !model.GetOtherNotify("TRON", _toAddress) {
			break
		}

		var _rawQuant, _amount = parseTransAmount(transfer.Get("value").Float())
		if !inPaymentAmountRange(_rawQuant) {
			continue
		}

		var _created = time.UnixMilli(transfer.Get("block_timestamp").Int())
		var _txid = transfer.Get("transaction_id").String()
		var _detailUrl = "https://tronscan.org/#/transaction/" + _txid
		if !model.IsNeedNotifyByTxid(_txid) {
			// 不需要额外通知
			continue
		}

		var title = "收入"
		if strings.ToLower(transfer.Get("to").String()) != strings.ToLower(_toAddress) {
			title = "支出"
		}

		var text = fmt.Sprintf(
			"#账户%s #非订单交易\n---\n```\n💲交易数额：%v USDT\n⏱️交易时间：%v\n✅接收地址：%v\n🅾️发送地址：%v```\n",
			title,
			_amount,
			_created.Format(time.DateTime),
			help.MaskAddress(transfer.Get("to").String()),
			help.MaskAddress(transfer.Get("from").String()),
		)

		var chatId, err = strconv.ParseInt(config.GetTgBotNotifyTarget(), 10, 64)
		if err != nil {
			continue
		}

		var msg = tgbotapi.NewMessage(chatId, text)
		msg.ParseMode = tgbotapi.ModeMarkdown
		msg.ReplyMarkup = tgbotapi.InlineKeyboardMarkup{
			InlineKeyboard: [][]tgbotapi.InlineKeyboardButton{
				{
					tgbotapi.NewInlineKeyboardButtonURL("📝查看交易明细", _detailUrl),
				},
			},
		}

		var _record = model.NotifyRecord{Txid: _txid}
		model.DB.Create(&_record)

		go telegram.SendMsg(msg)
	}
}

// 非订单交易通知
func handleOtherNotifyForETH(_toChain string, _toAddress string, result gjson.Result) {
	transfers := result.Get("result").Array()
	for _, transfer := range transfers {
		// 不需要通知
		if !model.GetOtherNotify(_toChain, _toAddress) {
			break
		}

		tokenSymbol := transfer.Get("tokenSymbol").String()
		// 不是USDT入账直接跳出
		if tokenSymbol != "USDT" && tokenSymbol != "BSC-USD" {
			continue
		}

		// 计算交易金额
		value, _ := new(big.Int).SetString(transfer.Get("value").String(), 10)
		tokenDecimals, _ := strconv.ParseInt(transfer.Get("tokenDecimal").String(), 10, 32)
		decimalFactor := new(big.Int).Exp(big.NewInt(10), big.NewInt(tokenDecimals), nil)
		valueUSDT := new(big.Float).Quo(new(big.Float).SetInt(value), new(big.Float).SetInt(decimalFactor))
		bigFloatStr := valueUSDT.Text('f', -1)
		decimalUSDT, _ := decimal.NewFromString(bigFloatStr)
		// 如果不在范围直接跳出
		if !inPaymentAmountRange(decimalUSDT) {
			continue
		}

		var _created = time.UnixMilli(transfer.Get("timeStamp").Int() * 1000)
		var _txid = transfer.Get("hash").String()
		var _detailUrl = "https://polygonscan.com/tx/" + _txid
		if _toChain == "OP" {
			_detailUrl = "https://optimistic.etherscan.io/tx/" + _txid
		}
		if _toChain == "BSC" {
			_detailUrl = "https://bscscan.com/tx/" + _txid
		}

		if !model.IsNeedNotifyByTxid(_txid) {
			// 不需要额外通知
			continue
		}

		var title = "收入"
		if strings.ToLower(transfer.Get("to").String()) != strings.ToLower(_toAddress) {
			title = "支出"
		}

		var text = fmt.Sprintf(
			"#账户%s #非订单交易\n---\n```\n💲交易数额：%v USDT\n⏱️交易时间：%v\n✅接收地址：%v\n🅾️发送地址：%v```\n",
			title,
			decimalUSDT,
			_created.Format(time.DateTime),
			help.MaskAddress(transfer.Get("to").String()),
			help.MaskAddress(transfer.Get("from").String()),
		)

		var chatId, err = strconv.ParseInt(config.GetTgBotNotifyTarget(), 10, 64)
		if err != nil {
			continue
		}

		var msg = tgbotapi.NewMessage(chatId, text)
		msg.ParseMode = tgbotapi.ModeMarkdown
		msg.ReplyMarkup = tgbotapi.InlineKeyboardMarkup{
			InlineKeyboard: [][]tgbotapi.InlineKeyboardButton{
				{
					tgbotapi.NewInlineKeyboardButtonURL("📝查看交易明细", _detailUrl),
				},
			},
		}

		var _record = model.NotifyRecord{Txid: _txid}
		model.DB.Create(&_record)

		go telegram.SendMsg(msg)
	}
}
func handleOtherNotifyForPolygonScan(_toAddress string, result gjson.Result) {
	handleOtherNotifyForETH("POLY", _toAddress, result)
}
func handleOtherNotifyForOptimismExplorer(_toAddress string, result gjson.Result) {
	handleOtherNotifyForETH("OP", _toAddress, result)
}
func handleOtherNotifyForBscScan(_toAddress string, result gjson.Result) {
	handleOtherNotifyForETH("BSC", _toAddress, result)
}

// 搜索交易记录 TronScan
func getUsdtTrc20TransByTronScan(_toAddress string) (gjson.Result, error) {
	var now = time.Now()
	var client = &http.Client{Timeout: time.Second * 15}
	req, err := http.NewRequest("GET", "https://apilist.tronscanapi.com/api/new/token_trc20/transfers", nil)
	if err != nil {
		return gjson.Result{}, fmt.Errorf("处理请求创建错误: %w", err)
	}

	// 构建请求参数
	var params = url.Values{}
	params.Add("start", "0")
	params.Add("limit", "30")
	params.Add("contract_address", usdtToken)
	params.Add("start_timestamp", strconv.FormatInt(now.Add(-time.Hour).UnixMilli(), 10)) // 当前时间向前推 1 小时
	params.Add("end_timestamp", strconv.FormatInt(now.Add(time.Hour).UnixMilli(), 10))    // 当前时间向后推 1 小时
	params.Add("relatedAddress", _toAddress)
	if config.GetTradeConfirmed() {
		params.Add("confirm", "true")
	} else {
		params.Add("confirm", "false")
	}
	req.URL.RawQuery = params.Encode()

	if config.GetTronScanApiKey() != "" {
		req.Header.Add("TRON-PRO-API-KEY", config.GetTronScanApiKey())
	}

	// 请求交易记录
	resp, err := client.Do(req)
	if err != nil {
		return gjson.Result{}, fmt.Errorf("请求交易记录错误: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return gjson.Result{}, fmt.Errorf("请求交易记录错误: StatusCode != 200")
	}

	// 获取响应记录
	all, err := io.ReadAll(resp.Body)
	if err != nil {
		return gjson.Result{}, fmt.Errorf("读取交易记录错误: %w", err)
	}

	// 释放响应请求
	_ = resp.Body.Close()

	// 解析响应记录
	return gjson.ParseBytes(all), nil
}

// 搜索交易记录 TronGrid
func getUsdtTrc20TransByTronGrid(_toAddress string) (gjson.Result, error) {
	var now = time.Now()
	var client = &http.Client{Timeout: time.Second * 15}
	req, err := http.NewRequest("GET", fmt.Sprintf("https://api.trongrid.io/v1/accounts/%s/transactions/trc20", _toAddress), nil)
	if err != nil {

		return gjson.Result{}, fmt.Errorf("处理请求创建错误: %w", err)
	}

	// 构建请求参数
	var params = url.Values{}
	params.Add("limit", "30")
	params.Add("contract_address", usdtToken)
	params.Add("min_timestamp", strconv.FormatInt(now.Add(-time.Hour).UnixMilli(), 10)) // 当前时间向前推 3 小时
	params.Add("max_timestamp", strconv.FormatInt(now.Add(time.Hour).UnixMilli(), 10))  // 当前时间向后推 1 小时
	params.Add("order_by", "block_timestamp,desc")
	if config.GetTradeConfirmed() {
		params.Add("only_confirmed", "true")
	} else {
		params.Add("only_confirmed", "false")
	}
	if config.GetTronGridApiKey() != "" {

		req.Header.Add("TRON-PRO-API-KEY", config.GetTronGridApiKey())
	}

	req.URL.RawQuery = params.Encode()

	// 请求交易记录
	resp, err := client.Do(req)

	if err != nil {

		return gjson.Result{}, fmt.Errorf("请求交易记录错误: %w", err)
	}

	if resp.StatusCode != http.StatusOK {

		return gjson.Result{}, fmt.Errorf("请求交易记录错误: StatusCode != 200")
	}

	// 获取响应记录
	all, err := io.ReadAll(resp.Body)
	if err != nil {

		return gjson.Result{}, fmt.Errorf("读取交易记录错误: %w", err)
	}

	// 释放响应请求
	_ = resp.Body.Close()

	// 解析响应记录
	return gjson.ParseBytes(all), nil
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

/*
所有ETN兼容链路的到账监控， 如果有金额匹配的自动返回
*/
func getUsdtTransByETH(chain string, address string) (gjson.Result, error) {
	// 累计所有交易的 Value 来计算总交易量
	var wa model.WalletAddress

	var host = "https://api.polygonscan.com/api"
	var apiKey = config.GetPolygonScanApiKey()
	var contractAddress = config.GetPolygonScanContractAddress()
	if chain == "OP" {
		host = "https://api-optimistic.etherscan.io/api"
		contractAddress = config.GetOptimismExplorerContractAddress()
		apiKey = config.GetOptimismExplorerApiKey()
	}
	if chain == "BSC" {
		host = "https://api.bscscan.com/api"
		contractAddress = config.GetBscExplorerContractAddress()
		apiKey = config.GetBscExplorerApiKey()
	}

	if model.DB.Where("chain = ? and address = ?", chain, address).First(&wa).Error == nil {
		// 这里查询订单历史
		var queryTx = "module=account&action=tokentx&contractaddress=" + contractAddress + "&address=" + address + "&startblock=" + strconv.FormatInt(wa.StartBlock+1, 10) + "&endblock=" + strconv.FormatInt(wa.StartBlock+999999999999, 10) + "&sort=asc" + "&apikey=" + apiKey
		allTx := requestAddress(host, queryTx)
		resultTx := gjson.ParseBytes(allTx)

		return resultTx, nil
	}

	return gjson.Result{}, nil
}
func getUsdtPolygonTransByPolygonScan(_toAddress string) (gjson.Result, error) {
	return getUsdtTransByETH("POLY", _toAddress)
}
func getUsdtOptimismTransByOptimismExplorer(_toAddress string) (gjson.Result, error) {
	return getUsdtTransByETH("OP", _toAddress)
}
func getUsdtBscTransByBscScan(_toAddress string) (gjson.Result, error) {
	return getUsdtTransByETH("BSC", _toAddress)
}

// 解析交易金额
func parseTransAmount(amount float64) (decimal.Decimal, string) {
	var _decimalAmount = decimal.NewFromFloat(amount)
	var _decimalDivisor = decimal.NewFromFloat(1000000)
	var result = _decimalAmount.Div(_decimalDivisor)

	return result, result.String()
}

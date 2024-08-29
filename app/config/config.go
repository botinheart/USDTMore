package config

import (
	"USDTMore/app/help"
	"fmt"
	"github.com/shopspring/decimal"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"
)

const defaultExpireTime = 1800        // 订单默认有效期 10分钟
const defaultUsdtRate = 7.4           // 默认汇率
const defaultAuthToken = "123234"     // 默认授权码
const defaultListen = ":6080"         // 默认监听地址
const TronServerApiScan = "TRON_SCAN" //
const TronServerApiGrid = "TRON_GRID" //
const defaultPaymentMinAmount = 0.01  //
const defaultPaymentMaxAmount = 99999 //

// 当前路径
var runPath string

func init() {
	// 获取应用程序当前的路径
	execPath, err := os.Executable()
	if err != nil {
		panic(err)
	}
	runPath = filepath.Dir(execPath)
}

/*
获取最小支付金额
*/
func GetPaymentMinAmount() decimal.Decimal {
	var _default = decimal.NewFromFloat(defaultPaymentMinAmount)
	// 取得配置数据
	var _min, _ = getPaymentRangeAmount()
	if _min == "" {
		return _default
	}

	// 使用配置数据返回
	_result, err := decimal.NewFromString(_min)
	if err == nil {
		return _result
	}

	// 默认返回数据
	return _default
}

/*
获取最大支付金额
*/
func GetPaymentMaxAmount() decimal.Decimal {
	var _default = decimal.NewFromFloat(defaultPaymentMaxAmount)
	// 取得配置数据
	var _, _max = getPaymentRangeAmount()
	if _max == "" {
		return _default
	}

	// 使用配置数据返回
	_result, err := decimal.NewFromString(_max)
	if err == nil {

		return _result
	}

	// 默认返回数据
	return _default
}

/*
读取支付金额范围
*/
func getPaymentRangeAmount() (string, string) {
	var _rangeVar string
	if _rangeVar = strings.TrimSpace(help.GetEnv("PAYMENT_AMOUNT_RANGE")); _rangeVar == "" {
		return "", ""
	}

	var _payRange = strings.Split(_rangeVar, ",")
	if len(_payRange) < 2 {
		return "", ""
	}

	return _payRange[0], _payRange[1]
}

/*
获得过期时间
*/
func GetExpireTime() time.Duration {
	if ret := help.GetEnv("EXPIRE_TIME"); ret != "" {
		sec, err := strconv.Atoi(ret)
		if err == nil && sec > 0 {
			return time.Duration(sec)
		}
	}

	return defaultExpireTime
}

/*
获取固定汇率设置
*/
func GetUsdtRateRaw() string {
	if data := help.GetEnv("USDT_RATE"); data != "" {
		return strings.TrimSpace(data)
	}
	return ""
}

/*
获得TRON服务类型，可选`TRON_SCAN`,`TRON_GRID`
只适用于TRON链路
*/
func GetTronServerApi() string {
	if data := help.GetEnv("TRON_SERVER_API"); data != "" {
		return strings.TrimSpace(data)
	}

	return ""
}

/*
获得TronScan接口的API密钥
*/
func GetTronScanApiKey() string {
	if data := help.GetEnv("TRON_SCAN_API_KEY"); data != "" {
		return strings.TrimSpace(data)
	}
	return "c0634c05-b4db-4fa4-a14a-93f2c2d5b65e"
}

/*
获得TronGrid接口的API密钥
*/
func GetTronGridApiKey() string {
	if data := help.GetEnv("TRON_GRID_API_KEY"); data != "" {
		return strings.TrimSpace(data)
	}
	return ""
}

/*
判断是否使用了Tron Scan接口
*/
func IsTronScanApi() bool {
	if GetTronServerApi() == TronServerApiScan {
		return true
	}

	return GetTronServerApi() != TronServerApiGrid
}

/*
获得PolygonScan接口的API密钥
*/
func GetPolygonScanApiKey() string {
	if data := help.GetEnv("POLYGON_SCAN_API_KEY"); data != "" {
		return strings.TrimSpace(data)
	}
	return "JSA7UJB36K4W735KG2E891G117VHZUXPUI"
}

/*
获得OptimismExplorer接口的API密钥
*/
func GetOptimismExplorerApiKey() string {
	if data := help.GetEnv("OPTIMISM_EXPLORER_API_KEY"); data != "" {
		return strings.TrimSpace(data)
	}
	return "ND5W8F2FBSA3R7GYARH9ZJI91JIHFAEFM5"
}

/*
获得SolanaExplorer接口的API密钥, 目前并不需要， 保持向后兼容性
*/
func GetBscExplorerApiKey() string {
	if data := help.GetEnv("BSC_SCAN_API_KEY"); data != "" {
		return strings.TrimSpace(data)
	}
	return "TAQJD39XSBTDBDYDYZTG3RH3UXJS2M9MN4"
}

// ERC-20 合约地址 (Polygon 主网上的 USDT)
const tokenPolygonContractAddress = "0xc2132D05D31c914a87C6611C10748AEb04B58e8F"

// ERC-20 合约地址 (Optimism 主网上的 USDT)
const tokenOptimismContractAddress = "0x94b008aA00579c1307B0EF2c499aD98a8ce58e58"

// ERC-20 合约地址 (Bsc 主网上的 USDT)
const tokenBscContractAddress = "0x55d398326f99059fF775485246999027B3197955"

/*
获得PolygonScan接口的API密钥
*/
func GetPolygonScanContractAddress() string {
	return tokenPolygonContractAddress
}

/*
获得OptimismExplorer接口的API密钥
*/
func GetOptimismExplorerContractAddress() string {
	return tokenOptimismContractAddress
}

/*
获得SolanaExplorer接口的API密钥, 目前并不需要， 保持向后兼容性
*/
func GetBscExplorerContractAddress() string {
	return tokenBscContractAddress
}

/*
通过okX交易所获得最新的汇率
*/
func GetUsdtRate() (string, decimal.Decimal, float64) {
	// 只有设置了汇率才自动使用动态汇率
	if data := help.GetEnv("USDT_RATE"); data != "" {
		data = strings.TrimSpace(data)
		// 纯数字，固定汇率
		if help.IsNumber(data) {
			if _res, err := strconv.ParseFloat(data, 64); err == nil {
				return "", decimal.Decimal{}, _res
			}
		}

		// 动态交易所汇率，有波动
		if len(data) >= 2 {
			if match, err2 := regexp.MatchString(`^[~+-]\d+(\.\d+)?$`, data); match && err2 == nil {
				_value, err3 := strconv.ParseFloat(data[1:], 64)
				if err3 == nil {
					return string(data[0]), decimal.NewFromFloat(_value), defaultUsdtRate
				}
			}
		}
	}

	// 动态交易所汇率，无波动
	return "=", decimal.Decimal{}, defaultUsdtRate
}

/*
获取回调密钥
*/
func GetAuthToken() string {
	if data := help.GetEnv("AUTH_TOKEN"); data != "" {
		return strings.TrimSpace(data)
	}
	return defaultAuthToken
}

/*
获取监听服务器
*/
func GetListen() string {
	if data := help.GetEnv("LISTEN"); data != "" {
		return strings.TrimSpace(data)
	}
	return defaultListen
}

/*
是否要等待区块链确认完成， 最好不要
*/
func GetTradeConfirmed() bool {
	if data := help.GetEnv("TRADE_IS_CONFIRMED"); data != "" {
		if data == "1" || data == "true" {
			return true
		}
	}
	return false
}

/*
获得Polygon Confirmation确认次数
*/
func GetPolygonConfirmation() int {
	if data := help.GetEnv("ETH_CONFIRMATION"); data != "" {
		num, err := strconv.Atoi(data)
		if err != nil {
			// 如果转换失败，处理错误
			fmt.Println("转换错误:", err)
			return 50
		}
		return num
	}

	return 50
}

/*
获取应用的地址
*/
func GetAppUri(host string) string {
	if data := help.GetEnv("APP_URI"); data != "" {
		return strings.TrimSpace(data)
	}
	return host
}

/*
机器人Token
*/
func GetTGBotToken() string {
	if data := help.GetEnv("TG_BOT_TOKEN"); data != "" {
		return strings.TrimSpace(data)
	}
	return ""
}

/*
管理员UID
*/
func GetTGBotAdminId() string {
	if data := help.GetEnv("TG_BOT_ADMIN_ID"); data != "" {
		return strings.TrimSpace(data)
	}
	return ""
}

/*
通知组GID
*/
func GetTgBotGroupId() string {
	if data := help.GetEnv("TG_BOT_GROUP_ID"); data != "" {
		return strings.TrimSpace(data)
	}
	return ""
}

/*
通知的Telegram群组
*/
func GetTgBotNotifyTarget() string {
	var groupId = GetTgBotGroupId()
	if groupId != "" {
		return groupId
	}
	return GetTGBotAdminId()
}

/*
日志路径
*/
func GetOutputLog() string {
	if data := help.GetEnv("LOG_DIR"); data != "" {
		return strings.TrimSpace(data) + "/usdtmore.log"
	}
	return runPath + "/usdtmore.log"
}

/*
数据库路径
*/
func GetDbPath() string {
	if data := help.GetEnv("DB_DIR"); data != "" {
		return strings.TrimSpace(data) + "/usdtmore.db"
	}
	return runPath + "/usdtmore.db"
}

/*
模版路径
*/
func GetTemplatePath() string {
	if data := help.GetEnv("HTML_DIR"); data != "" {
		return strings.TrimSpace(data) + "/templates/*"
	}
	return runPath + "/../templates/*"
}

/*
静态路径
*/
func GetStaticPath() string {
	if data := help.GetEnv("HTML_DIR"); data != "" {
		return strings.TrimSpace(data) + "/static/"
	}
	return runPath + "/../static/"
}

/*
钱包地址，启动以后会自动增加，当然也可以机器人自动添加
*/
func GetInitWalletAddress() []string {
	if data := help.GetEnv("WALLET_ADDRESS"); data != "" {
		return strings.Split(strings.TrimSpace(data), ",")
	}
	return []string{}
}

/*
是否开启反向代理中的https覆写
*/
func IsReWriteHttps() bool {
	if data := help.GetEnv("REWRITE_HTTPS"); data != "" {
		if data == "true" || data == "yes" || data == "1" {
			return true
		}
	}
	return false
}

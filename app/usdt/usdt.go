package usdt

import "math"

var _latestUsdtRate = 0.0
var _okxLastUsdtRate = 0.0

/*
设置最新汇率
*/
func SetLatestRate(rate float64) {
	// 取绝对值
	_latestUsdtRate = math.Abs(rate)
}

/*
获取最新汇率
*/
func GetLatestRate() float64 {
	return _latestUsdtRate
}

/*
设置okX的最新汇率
*/
func SetOkxLatestRate(okxRate float64) {
	_okxLastUsdtRate = okxRate
}

/*
获取okX的最新汇率
*/
func GetOkxLastRate() float64 {
	return _okxLastUsdtRate
}

package monitor

import (
	"USDTMore/app/log"
	"USDTMore/app/model"
	"USDTMore/app/notify"
	"math"
	"time"
)

func NotifyStart() {
	log.Info("回调监控启动.")
	for range time.Tick(time.Second * 5) {
		tradeOrders, err := model.GetNotifyFailedTradeOrders()
		if err != nil {
			log.Error("待回调订单获取失败", err)
			continue
		}

		for _, order := range tradeOrders {
			// 判断是否到达下次回调时间
			// 下次回调时间等于 3的失败次数次方 * 1分钟 + 交易确认时间
			var _nextNotifyTime = order.ConfirmedAt.Add(time.Minute * time.Duration(math.Pow(3, float64(order.NotifyNum))))
			if time.Now().Unix() >= _nextNotifyTime.Unix() {
				// 到达下次回调时间
				go notify.OrderNotify(order)
			}
		}
	}
}

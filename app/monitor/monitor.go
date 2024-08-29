package monitor

import (
	"USDTMore/app/config"
	"github.com/shopspring/decimal"
)

var _paymentMinAmount, _paymentMaxAmount decimal.Decimal

func init() {
	_paymentMinAmount = config.GetPaymentMinAmount()
	_paymentMaxAmount = config.GetPaymentMaxAmount()
}

func inPaymentAmountRange(payAmount decimal.Decimal) bool {
	if payAmount.GreaterThan(_paymentMaxAmount) {

		return false
	}

	if payAmount.LessThan(_paymentMinAmount) {

		return false
	}

	return true
}

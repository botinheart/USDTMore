package model

import (
	"USDTMore/app/config"
	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

var DB *gorm.DB
var _err error

func Init() error {
	DB, _err = gorm.Open(sqlite.Open(config.GetDbPath()), &gorm.Config{})
	if _err != nil {

		return _err

	}
	if _err = AutoMigrate(); _err != nil {

		return _err
	}

	addStartWalletAddress()

	return nil
}

func AutoMigrate() error {
	return DB.AutoMigrate(&WalletAddress{}, &TradeOrders{}, &NotifyRecord{})
}

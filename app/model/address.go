package model

import (
	"USDTMore/app/config"
	"USDTMore/app/help"
	"errors"
	"fmt"
	"gorm.io/gorm"
	"strings"
	"time"
)

const StatusEnable = 1
const StatusDisable = 0
const OtherNotifyEnable = 1
const OtherNotifyDisable = 0

type WalletAddress struct {
	Id          int64     `gorm:"integer;primaryKey;not null;comment:id"`
	Chain       string    `gorm:"type:varchar(255);not null;comment:链路名称 TRON POLY OP BSC"`
	StartBlock  int64     `gorm:"integer;not null;default:0;comment:初始化块，每次查询记录一天之前的blocknum"`
	InAmount    float64   `gorm:"REAL;not null;default:0;comment:累计转入"`
	OutAmount   float64   `gorm:"REAL;not null;default:0;comment:累计转出"`
	Count       int64     `gorm:"integer;not null;default:0;comment:历史订单数量"`
	Address     string    `gorm:"type:varchar(255);not null;comment:钱包地址"`
	Status      int       `gorm:"type:tinyint(1);not null;default:1;comment:地址状态 1启动 0禁止"`
	OtherNotify int       `gorm:"type:tinyint(1);not null;default:1;comment:其它转账通知 1启动 0禁止"`
	CreatedAt   time.Time `gorm:"autoCreateTime;type:timestamp;not null;comment:创建时间"`
	UpdatedAt   time.Time `gorm:"autoUpdateTime;type:timestamp;not null;comment:更新时间"`
}

// 启动时添加初始钱包地址
func addStartWalletAddress() {
	var _wa WalletAddress

	for _, address := range config.GetInitWalletAddress() {
		if help.IsValidTRONWalletAddress(address) || help.IsValidPOLWalletAddress(address) || help.IsValidOPTWalletAddress(address) || help.IsValidBSCWalletAddress(address) {
			_addresses := strings.Split(strings.TrimSpace(address), ":")
			var _res2 = DB.Where("chain = > and address = ?", _addresses[0], _addresses[1]).First(&_wa)
			if errors.Is(_res2.Error, gorm.ErrRecordNotFound) {
				var _row = WalletAddress{Chain: _addresses[0], Address: _addresses[1], Status: StatusEnable}
				var _res = DB.Create(&_row)
				if _res.Error == nil && _res.RowsAffected == 1 {
					fmt.Println("✅钱包地址添加成功：", address)
				}
			}
		}
	}
}

func (wa *WalletAddress) TableName() string {
	return "wallet_address"
}

func (wa *WalletAddress) SetStatus(status int) {
	wa.Status = status
	DB.Save(wa)
}

func (wa *WalletAddress) SetOtherNotify(notify int) {
	wa.OtherNotify = notify
	DB.Save(wa)
}

func (wa *WalletAddress) Delete() {
	DB.Delete(wa)
}

/*
判断是否存在
*/
func ExistsAddress(chain string, address string) bool {
	var rows []WalletAddress
	DB.Where("chain = ? and status = ?", chain, StatusEnable).Find(&rows)
	return len(rows) > 0
}

func GetAvailableAddress(chain string) []WalletAddress {
	var rows []WalletAddress
	DB.Where("chain = ? and status = ?", chain, StatusEnable).Find(&rows)
	return rows
}

func GetOtherNotify(chain string, address string) bool {
	var row WalletAddress
	var res = DB.Where("status = ? and chain = ? and address = ?", StatusEnable, chain, address).First(&row)
	if res.Error != nil {
		return false
	}
	return row.OtherNotify == 1
}

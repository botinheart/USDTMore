package help

import (
	"crypto/md5"
	"fmt"
	"github.com/google/uuid"
	"os"
	"regexp"
	"sort"
	"strings"
)

// IsExist 判断文件是否存在
func IsExist(path string) bool {
	_, err := os.Stat(path)
	if err == nil {

		return true
	}

	if os.IsExist(err) {

		return true
	}

	return false
}

/*
获取环境变量
*/
func GetEnv(key string) string {
	return os.Getenv(key)
}

/*
*
生成签名
*/
func GenerateSignature(data map[string]interface{}, token string) string {
	keys := make([]string, 0, len(data))
	for k := range data {
		if k == "signature" {

			continue
		}

		keys = append(keys, k)
	}
	sort.Strings(keys)
	var sign strings.Builder
	for _, k := range keys {
		v := data[k]
		if v == nil || v == "" {

			continue
		}

		sign.WriteString(k)
		sign.WriteString("=")
		sign.WriteString(fmt.Sprintf("%v", v))
		sign.WriteString("&")
	}

	signString := strings.TrimRight(sign.String(), "&")

	return Md5String(signString + token)
}

/*
生成订单号
*/
func GenerateTradeId() string {
	return uuid.New().String()
}

/*
计算MD5值
*/
func Md5String(text string) string {
	return fmt.Sprintf("%x", md5.Sum([]byte(text)))
}

/*
过滤字符串
*/
func Ec(str string) string {
	escapeChars := []string{"_", "*", "[", "]", "(", ")", "~", "`", ">", "#", "+", "-", "=", "|", "{", "}", ".", "!"}

	for _, char := range escapeChars {
		str = strings.ReplaceAll(str, char, "\\"+char)
	}

	return str
}

/*
是否是数字
*/
func IsNumber(s string) bool {
	match, err := regexp.MatchString(`^\d+\.?\d*$`, s)

	return match && err == nil
}

/*
是否是TRON的地址
*/
func IsValidTRONWalletAddress(address string) bool {
	match, err := regexp.MatchString(`^TRON:T[a-zA-Z0-9]{33}$`, address)
	return match && err == nil
}

/*
死否是Polygon的地址
*/
func IsValidPOLWalletAddress(address string) bool {
	match, err := regexp.MatchString(`^POLY:0x[a-zA-Z0-9]{40}$`, address)
	return match && err == nil
}

/*
是否是Optimism链的地址
*/
func IsValidOPTWalletAddress(address string) bool {
	match, err := regexp.MatchString(`^OP:0x[a-zA-Z0-9]{40}$`, address)
	return match && err == nil
}

/*
是否是BSC的地址
*/
func IsValidBSCWalletAddress(address string) bool {
	match, err := regexp.MatchString(`^BSC:0x[a-zA-Z0-9]{40}$`, address)
	return match && err == nil
}

/*
掩码功能
*/
func MaskAddress(address string) string {
	if len(address) <= 20 {
		return address
	}
	return address[:8] + " ***** " + address[len(address)-10:]
}

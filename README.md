# USDTMore (USDT Payment Gateway for More Chain)

<p align="center">
<img src="./static/img/tether.svg" width="15%" alt="tether">
</p>
<p align="center">
<a href="https://www.gnu.org/licenses/gpl-3.0.html"><img src="https://img.shields.io/badge/license-GPLV3-blue" alt="license GPLV3"></a>
<a href="https://golang.org"><img src="https://img.shields.io/badge/Golang-1.22-red" alt="Go version 1.21"></a>
<a href="https://github.com/gin-gonic/gin"><img src="https://img.shields.io/badge/Gin-v1.9-blue" alt="Gin Web Framework v1.9"></a>
<a href="https://github.com/go-telegram-bot-api/telegram-bot-api"><img src="https://img.shields.io/badge/Telegram Bot-v5-lightgrey" alt="Golang Telegram Bot Api-v5"></a>
<a href="https://github.com/v03413/bepusdt"><img src="https://img.shields.io/badge/Release-v1.9.21-green" alt="Release v1.9.21"></a>
</p>

## 🪧 介绍

对Bepusdt进行了二次改造, 针对系统原生安装进行了优化, 同时加入了其他费率更低的链路: Polygon, Optimism, Bsc的支持;

收款更好用、部署更便捷！

## 🎉 新特性

- ✅ 具备`Bepusdt`的所有特性，插件兼容无缝替换
- ✅ 增加 `Polygon`, `Optimism`, `Bsc` 链的支持

## 🛠 参数配置

USDTMore 所有参数都是以传递环境变量的方式进行配置，大部分参数含默认值，少量配置即可直接使用！

### 参数列表

| 参数名称                      | 默认值          | 用法说明                                                                                                                                          |
|---------------------------|--------------|-----------------------------------------------------------------------------------------------------------------------------------------------|
| EXPIRE_TIME               | `1800`       | 订单有效期，单位秒                                                                                                                                     |
| USDT_RATE                 | `空`          | USDT汇率，默认留空则获取Okx交易所的汇率(每分钟同步一次)，支持多种写法，如：`7.4` 表示固定7.4、`～1.02`表示最新汇率上浮2%、`～0.97`表示最新汇率下浮3%、`+0.3`表示最新加0.3、`-0.2`表示最新减0.2，以此类推；如参数错误则使用固定值7.4 |
| AUTH_TOKEN                | `123456`     | 认证Token，对接发卡网/支付平台会用到这个参数进行回调                                                                                                                 |
| LISTEN                    | `:6080`      | 服务器HTTP监听地址                                                                                                                                   
| REWRITE_HTTPS             | `false`      | 重写http成https，使用反向代理的时候往往需要强制https模式                                                                                                           
| TRADE_IS_CONFIRMED        | `0`          | TRON网络是否需要确认，禁用可以提高回调速度，启用则可以防止交易失败                                                                                                           |
| ETH_CONFIRMATION          | `0`          | ETH兼容网络需要网络确认的块数，影响Polygon、Optimism，Bep20                                                                                                     |
| APP_URI                   | `空`          | 应用访问地址，留空则系统自动获取，前端收银台会用到，建议设置，例如：https://token-pay.example.com                                                                               |
| WALLET_ADDRESS            | `空`          | 启动时需要添加的钱包地址，多个请用半角符逗号`,`分开；当然，同样也支持通过机器人添加。<br>单条格式为: [TRON\|POLY\|OP\|BSC]:地址, 其中[]部分为支付链标识                                                 |
| TG_BOT_TOKEN              | `空`          | Telegram Bot Token，**必须设置**，否则无法使用                                                                                                            |
| TG_BOT_ADMIN_ID           | `空`          | Telegram Bot 管理员ID，**必须设置**，否则无法使用                                                                                                            |
| TG_BOT_GROUP_ID           | `空`          | Telegram 群组ID，设置之后机器人会将交易消息会推送到此群                                                                                                             |
| TRON_SERVER_API           | `TRON_SCAN`  | 可选`TRON_SCAN`,`TRON_GRID`，推荐`TRON_GRID`和`TRON_GRID_API_KEY`搭配使用，*更准更强更及时*                                                                     |
| TRON_SCAN_API_KEY         | `空`          | TRONSCAN API KEY，如果收款地址较多推荐设置，可避免被官方QOS                                                                                                       |
| TRON_GRID_API_KEY         | `空`          | TRONGRID API KEY，如果收款地址较多推荐设置，可避免被官方QOS                                                                                                       |
| POLYGON_SCAN_API_KEY      | `空`          | POLYGON_SCAN API KEY，如果收款地址较多推荐设置，可避免被官方QOS                                                                                                   |
| OPTIMISM_EXPLORER_API_KEY | `空`          | OPTIMISM_EXPLORER API KEY，如果收款地址较多推荐设置，可避免被官方QOS                                                                                              |
| BSC_SCAN_API_KEY          | `空`          | BSC_SCAN API KEY，如果收款地址较多推荐设置，可避免被官方QOS                                                                                                       |
| PAYMENT_AMOUNT_RANGE      | `0.01,99999` | 支付监控的允许数额范围(闭区间)，设置合理数值可避免一些诱导式诈骗交易提醒                                                                                                         |
| LOG_DIR                   | `./log`      | 应用程序的日志路径                                                                                                                                     |
| DB_DIR                    | `./db`       | 应用程序的数据库路径                                                                                                                                    |
| HTML_DIR                  | `..`         | 界面模版/静态资源的路径                                                                                                                                      |

**特别提醒：上述参数必须设置的有`TG_BOT_TOKEN TG_BOT_ADMIN_ID`，否则无法使用！**

## 🚀 安装部署

- [https 配置教程](./docs/ssl.md)
- [Linux 手动安装教程](./docs/systemd.md)
- [Linux 时钟同步配置](./docs/systemd-timesyncd.md)

## 🤔 常见问题

### 如何获取参数 TG_BOT_ADMIN_ID

Telegram 搜索`@userinfobot`机器人并启用，返回的ID就是`TG_BOT_ADMIN_ID`

### 如何申请`TronScan`和`TronGrid`的ApiKey

目前[TronScan](https://tronscan.org/)/[TronGrid](https://www.trongrid.io/)、[PolygonScan](https://polygonscan.com/)、[OptimismExplorer](https://optimistic.etherscan.io/) 和 [BscScan](https://bscscan.com/)
都可以通过邮箱注册，登录之后在用户中心创建一个ApiKey即可；默认免费套餐都是每天10W请求，对于个人收款绰绰有余。

## ⚠️ 特别注意

- 订单交易强依赖时间，请确保服务器时间准确性，否则可能导致订单异常！
- 部分功能依赖网络，请确保服务器网络纯洁性，否则可能导致功能异常！
- 如果有问题，欢迎加入交流群交流 [USDTMore](https://t.me/usdt_more)

## 🙏 感谢两位大佬的代码，在此基础上改写了新的功能

- https://github.com/assimon/epusdt
- https://github.com/v03413/bepusdt

## 📢 声明

- 本项目仅供个人学习研究使用，任何人或组织在使用过程中请符合当地的法律法规，否则产生的任何后果责任自负。
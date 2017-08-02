package coinapi

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"log"
	"strings"
)

func MakeParam(key string, val interface{}) string {
	return fmt.Sprintf("%s=%v", key, val)
}

func MakeSign(params string) string {
	config := GetConfig()
	params += ("&secret_key=" + config.SecretKey)

	md5Ctx := md5.New()
	md5Ctx.Write([]byte(params))
	sign := md5Ctx.Sum(nil)
	return strings.ToUpper(hex.EncodeToString(sign))
}

/*
获取OKCoin最新市场行情数据

返回值说明
date: 返回数据时服务器时间
buy: 买一价
high: 最高价
last: 最新成交价
low: 最低价
sell: 卖一价
vol: 成交量(最近的24小时)
# Request
GET https://www.okcoin.cn/api/v1/ticker.do?symbol=ltc_cny
# Response
{
	"date":"1410431279",
	"ticker":{
		"buy":"33.15",
		"high":"34.15",
		"last":"33.15",
		"low":"32.05",
		"sell":"33.16",
		"vol":"10532696.39199642"
	}
}
*/
func GetTicker(symbol string) *RespTicker {
	if symbol == "" {
		log.Println("GetTicker symbol is empty")
		return nil
	}

	req := CoinHttp{}
	req.Init("GetTicker", "ticker.do", &RespTicker{})
	req.AddParam(MakeParam("symbol", symbol))
	var result interface{}
	result = req.Get()
	switch result := result.(type) {
	default:
		return nil
	case *RespTicker:
		return result
	}
}

/*
获取OKCoin市场深度

返回值说明
asks :卖方深度
bids :买方深度

# Request
GET https://www.okcoin.cn/api/v1/depth.do
# Response
{
	"asks": [
		[792, 5],
		[789.68, 0.018],
		[788.99, 0.042],
		[788.43, 0.036],
		[787.27, 0.02]
	],
	"bids": [
		[787.1, 0.35],
		[787, 12.071],
		[786.5, 0.014],
		[786.2, 0.38],
		[786, 3.217],
		[785.3, 5.322],
		[785.04, 5.04]
	]
}

*/
func GetDepth(symbol string) *RespDepth {
	if symbol == "" {
		log.Println("GetDepth symbol is empty")
		return nil
	}

	req := CoinHttp{}
	req.Init("GetDepth", "depth.do", &RespDepth{})
	req.AddParam(MakeParam("symbol", symbol))
	req.AddParam(MakeParam("size", "100"))
	var result interface{}
	result = req.Get()
	switch result := result.(type) {
	default:
		return nil
	case *RespDepth:
		return result
	}
}

/*
获取OKCoin最近600交易信息
返回值说明
date:交易时间
date_ms:交易时间(ms)
price: 交易价格
amount: 交易数量
tid: 交易生成ID
type: buy/sell
# Request
GET https://www.okcoin.cn/api/v1/trades.do?since=5000
# Response
[
	{
        "date": "1367130137",
		"date_ms": "1367130137000",
		"price": 787.71,
		"amount": 0.003,
		"tid": "230433",
		"type": "sell"
	},
	{
        "date": "1367130137",
		"date_ms": "1367130137000",
		"price": 787.65,
		"amount": 0.001,
		"tid": "230434",
		"type": "sell"
	},
	{
		"date": "1367130137",
		"date_ms": "1367130137000",
		"price": 787.5,
		"amount": 0.091,
		"tid": "230435",
		"type": "sell"
	}
]
*/
func GetTrades(symbol string) *[]RespTrades {
	if symbol == "" {
		log.Println("GetTrades symbol is empty")
		return nil
	}

	req := CoinHttp{}
	trades := make([]RespTrades, 0)
	req.Init("GetTrades", "trades.do", &trades)
	req.AddParam(MakeParam("symbol", symbol))
	var result interface{}
	result = req.Get()
	switch result := result.(type) {
	default:
		return nil
	case *[]RespTrades:
		return result
	}
}

/*
获取比特币或莱特币的K线数据
返回值说明
返回值说明
[
	1417536000000,	时间戳
	2370.16,	开
	2380,		高
	2352,		低
	2367.37,	收
	17259.83	交易量
]

# Request
GET https://www.okcoin.cn/api/v1/kline.do
# Response
[
    [
        1417449600000,
        2339.11,
        2383.15,
        2322,
        2369.85,
        83850.06
    ],
    [
        1417536000000,
        2370.16,
        2380,
        2352,
        2367.37,
        17259.83
    ]
]

请求参数
type String
1min : 1分钟
3min : 3分钟
5min : 5分钟
15min : 15分钟
30min : 30分钟
1day : 1日
3day : 3日
1week : 1周
1hour : 1小时
2hour : 2小时
4hour : 4小时
6hour : 6小时
12hour : 12小时

size (默认全部获取) 指定获取数据的条数
since (默认全部获取) 时间戳（eg：1417536000000）。 返回该时间戳以后的数据
*/
func GetKline(symbol string, intv string, size int32, sincems int64) *[]RespKline {
	if symbol == "" {
		log.Println("GetKline symbol is empty")
		return nil
	}

	req := CoinHttp{}
	kline := make([][6]float64, 0)
	req.Init("GetKline", "kline.do", &kline)
	req.AddParam(MakeParam("symbol", symbol))
	req.AddParam(MakeParam("type", intv))
	if size > 0 {
		req.AddParam(MakeParam("size", size))
	}

	if sincems > 0 {
		req.AddParam(MakeParam("since", sincems))
	}
	var result interface{}
	result = req.Get()
	switch result := result.(type) {
	default:
		return nil
	case *[][6]float64:
		//类型转换
		convKline := make([]RespKline, len(*result))
		for idx, val := range *result {
			convKline[idx].Date = uint64(val[0])
			convKline[idx].Open = float32(val[1])
			convKline[idx].High = float32(val[2])
			convKline[idx].Low = float32(val[3])
			convKline[idx].Close = float32(val[4])
			convKline[idx].Vol = float32(val[5])
		}
		return &convKline
	}
}

/*
获取用户信息
访问频率 6次/2秒

# Request
POST https://www.okcoin.cn/api/v1/userinfo.do
# Response
{
	"info": {
	        "funds": {
	            "asset": {
                "net": "0",
                "total": "0"
            },
		    "borrow": {
                "btc": "0",
                "cny": "0",
                "ltc": "0"
            },
            "free": {
                "btc": "0",
                "cny": "0",
                "ltc": "0",
                "eth": "0"

            },
            "freezed": {
                "btc": "0",
                "cny": "0",
                "ltc": "0",
                "eth": "0"
            },
            "union_fund": {
                "btc": "0",
                "ltc": "0"
            }
        }
    },
    "result": true
}


返回值说明
asset:账户资产，包含净资产及总资产
borrow:账户借款信息(只有在账户有借款信息时才会返回)
free:账户余额
freezed:账户冻结余额
union_fund:账户理财信息(只有在账户有理财信息时才返回)


*/

func GetUserInfo() *RespUserInfo {
	config := GetConfig()

	req := CoinHttp{}
	req.Init("GetUserInfo", "userinfo.do", &RespUserInfo{})
	req.AddParam(MakeParam("api_key", config.ApiKey))
	var result interface{}
	result = req.Post()
	switch result := result.(type) {
	default:
		return nil
	case *RespUserInfo:
		return result
	}
}

/*
URL https://www.okcoin.cn/api/v1/trade.do

访问频率 20次/2秒
# Request
POST https://www.okcoin.cn/api/v1/trade.do
# Response
{"result":true,"order_id":123456}
返回值说明
result:true代表成功返回
order_id:订单ID
*/

func DoTrade(symbol string, tradetype string, price float32, amount float32) uint32 {
	if symbol == "" {
		log.Println("DoTrade symbol is empty")
		return 0
	}

	if tradetype != BUY && tradetype != SELL &&
		tradetype != BUY_MARKET && tradetype != SELL_MARKET {
		log.Println("DoTrade tradetype error")
		return 0
	}

	if tradetype == BUY || tradetype == SELL || tradetype == SELL_MARKET {
		if symbol == BTC && amount < MIN_TRADE_BTC {
			log.Printf("DoTrade %s, tradetype=%s, amount=%f < %3f\n", symbol, tradetype, amount, MIN_TRADE_BTC)
			return 0
		}

		if symbol == LTC && amount < MIN_TRADE_LTC {
			log.Printf("DoTrade %s, tradetype=%s, amount=%f < %3f\n", symbol, tradetype, amount, MIN_TRADE_LTC)
			return 0
		}

		if symbol == ETH && amount < MIN_TRADE_ETH {
			log.Printf("DoTrade %s, tradetype=%s, amount=%f < %3f\n", symbol, tradetype, amount, MIN_TRADE_ETH)
			return 0
		}
	}

	if tradetype == BUY || tradetype == SELL || tradetype == BUY_MARKET {
		if price < 0.001 {
			log.Printf("DoTrade %s, tradetype=%s, price=%4f\n", symbol, tradetype, price)
			return 0
		}
	}

	config := GetConfig()
	req := CoinHttp{}
	req.Init("DoTrade", "trade.do", &RespDoTrade{})
	req.AddParam(MakeParam("api_key", config.ApiKey))
	req.AddParam(MakeParam("symbol", symbol))
	req.AddParam(MakeParam("type", tradetype))
	if tradetype != SELL_MARKET {
		req.AddParam(MakeParam("price", price))
	}
	if tradetype != BUY_MARKET {
		req.AddParam(MakeParam("amount", amount))
	}

	var result interface{}
	result = req.Post()
	switch result := result.(type) {
	default:
		return 0
	case *RespDoTrade:
		if result.Result == true {
			return result.OrderId
		}
		return 0
	}
}

/*
URL https://www.okcoin.cn/api/v1/cancel_order.do
访问频率 20次/2秒
示例
# Request
POST https://www.okcoin.cn/api/v1/cancel_order.do
# Response
#多笔订单返回结果(成功订单ID,失败订单ID)
{"success":"123456,123457","error":"123458,123459"}


返回值说明
result:true撤单请求成功，等待系统执行撤单；false撤单失败(用于单笔订单)
order_id:订单ID(用于单笔订单)
success:撤单请求成功的订单ID，等待系统执行撤单(用于多笔订单)
error:撤单请求失败的订单ID(用户多笔订单)

*/
func CancelOrder(symbol string, orderId uint32) bool {
	config := GetConfig()
	req := CoinHttp{}
	req.Init("CancelOrder", "cancel_order.do", &RespCancelOrder{})
	req.AddParam(MakeParam("api_key", config.ApiKey))
	req.AddParam(MakeParam("symbol", symbol))
	req.AddParam(MakeParam("order_id", orderId))

	var result interface{}
	result = req.Post()
	switch result := result.(type) {
	default:
		return false
	case *RespCancelOrder:
		return result.Result
	}
}

/*
POST /api/v1/order_info 获取用户的订单信息
URL https://www.okcoin.cn/api/v1/order_info.do
访问频率 20次/2秒 (未成交)
返回值说明
amount:委托数量
create_date: 委托时间
avg_price:平均成交价
deal_amount:成交数量
order_id:订单ID
orders_id:订单ID(不建议使用)
price:委托价格
status:-1:已撤销  0:未成交  1:部分成交  2:完全成交 4:撤单处理中
type:buy_market:市价买入 / sell_market:市价卖出

*/
func GetOrderInfo(symbol string, orderId int32) *RespGetOrderInfo {
	config := GetConfig()
	req := CoinHttp{}
	req.Init("GetOrderInfo", "order_info.do", &RespGetOrderInfo{})
	req.AddParam(MakeParam("api_key", config.ApiKey))
	req.AddParam(MakeParam("symbol", symbol))
	req.AddParam(MakeParam("order_id", orderId))

	var result interface{}
	result = req.Post()
	switch result := result.(type) {
	default:
		return nil
	case *RespGetOrderInfo:
		return result
	}
}

/*
URL https://www.okcoin.cn/api/v1/orders_info.do
访问频率 20次/2秒
# Request
POST https://www.okcoin.cn/api/v1/orders_info.do
# Response
{
	"result":true,
	"orders":[
		{
			"order_id":15088,
			"status":0,
			"symbol":"btc_cny",
			"type":"sell",
			"price":811,
			"amount":1.39901357,
			"deal_amount":1,
			"avg_price":811
		} ,
		{
			"order_id":15088,
			"status":-1,
			"symbol":"btc_cny",
			"type":"sell",
			"price":811,
			"amount":1.39901357,
			"deal_amount":1,
			"avg_price":811
		}
	]
}
返回值说明
amount：限价单请求：下单数量 /市价单请求：卖出的btc/ltc数量
deal_amount：成交数量
avg_price：平均成交价
create_date：委托时间
order_id：订单ID
price：限价单请求：委托价格 / 市价单请求：买入的usd金额
status： -1：已撤销  0：未成交 1：部分成交 2：完全成交 4:撤单处理中
type:buy_market：市价买入 /sell_market：市价卖出
result：结果信息

*/
func GetOrdersInfo(symbol string, querytype uint32) *RespGetOrderInfo {
	config := GetConfig()
	req := CoinHttp{}
	req.Init("GetOrdersInfo", "orders_info.do", &RespGetOrderInfo{})
	req.AddParam(MakeParam("api_key", config.ApiKey))
	req.AddParam(MakeParam("symbol", symbol))
	req.AddParam(MakeParam("type", querytype))
	//req.AddParam(MakeParam("order_id", ""))

	var result interface{}
	result = req.Post()
	switch result := result.(type) {
	default:
		return nil
	case *RespGetOrderInfo:
		return result
	}
}

/*
获取历史订单信息，只返回最近两天的信息
URL https://www.okcoin.cn/api/v1/order_history.do
示例
# Request
POST https://www.okcoin.cn/api/v1/order_history.do
# Response
{
	"current_page": 1,
	"orders": [
		{
			"amount": 0,
			"avg_price": 0,
			"create_date": 1405562100000,
			"deal_amount": 0,
			"order_id": 0,
			"price": 0,
			"status": 2,
			"symbol": "btc_cny",
			"type": "sell”
		}
	],
	"page_length": 1,
	"result": true,
	"total": 3
}
]
返回值说明
current_page:当前页码
orders:委托详细信息
amount:委托数量
avg_price:平均成交价
create_date:委托时间
deal_amount:成交数量
order_id:订单ID
price:委托价格
status:-1:已撤销   0:未成交 1:部分成交 2:完全成交 4:撤单处理中
type:buy_market:市价买入 / sell_market:市价卖出
page_length:每页数据条数
result:true代表成功返回

*/

func GetOrderHistory(symbol string, status uint32, curpage uint32, pagepercount uint32) *RespGetOrderHistory {
	config := GetConfig()
	req := CoinHttp{}
	req.Init("GetOrderHistory", "order_history.do", &RespGetOrderHistory{})
	req.AddParam(MakeParam("api_key", config.ApiKey))
	req.AddParam(MakeParam("symbol", symbol))
	req.AddParam(MakeParam("status", status))
	req.AddParam(MakeParam("current_page", curpage))
	req.AddParam(MakeParam("page_length", pagepercount))

	var result interface{}
	result = req.Post()
	switch result := result.(type) {
	default:
		return nil
	case *RespGetOrderHistory:
		return result
	}
}

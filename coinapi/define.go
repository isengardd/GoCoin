package coinapi

import (
	"strconv"
)

const (
	ROOT_URL      = "https://www.okcoin.cn/api/v1"
	CONTENT_TYPE  = "application/x-www-form-urlencoded"
	BTC           = "btc_cny"
	LTC           = "ltc_cny"
	ETH           = "eth_cny"
	MIN_TRADE_BTC = 0.01
	MIN_TRADE_LTC = 0.1
	MIN_TRADE_ETH = 0.01
	FLEE_LTC      = 0.002 //买卖手续费
	BUY           = "buy"
	SELL          = "sell"
	BUY_MARKET    = "buy_market"
	SELL_MARKET   = "sell_market"
)

type RespTicker struct {
	Date string `json:"date"` //返回数据时服务器时间
	Tick Ticker `json:"ticker"`
}

type Ticker struct {
	Buy  string `json:"buy"`  //买一价
	High string `json:"high"` //最高价
	Last string `json:"last"` //最新成交价
	Low  string `json:"low"`  //最低价
	Sell string `json:"sell"` //卖一价
	Vol  string `json:"vol"`  //成交量（最近24小时）
}

func (this *Ticker) GetLast() float32 {
	Last, _ := strconv.ParseFloat(this.Last, 32)
	return float32(Last)
}

type RespDepth struct {
	Asks [][2]float32 `json:"asks"` //卖方深度
	Bids [][2]float32 `json:"bids"` //买方深度
}

type RespTrades struct {
	Date   uint64 `json:"date"`    //交易时间
	Datems uint64 `json:"date_ms"` //交易时间(ms)
	Price  string `json:"price"`   //交易价格
	Amount string `json:"amount"`  //交易数量
	Tid    uint64 `json:"tid"`     //交易生成ID
	Type   string `json:"type"`    //buy/sell
}

type RespKline struct {
	Date  uint64  //时间戳
	Open  float32 //开盘价
	High  float32 //最高价
	Low   float32 //最低价
	Close float32 //收盘价
	Vol   float32 //交易量
}

type RespUserInfo struct {
	Info   innerUserInfo `json:"info"`
	Result bool          `json:"result"`
}

type innerUserInfo struct {
	Funds innerFunds `json:"funds"`
}

type innerFunds struct {
	Asset   MoneyData `json:"asset"`
	Free    MoneyData `json:"free"`    //账户余额
	Freezed MoneyData `json:"freezed"` //账户冻结余额
}

type MoneyData struct {
	Net   string `json:"net"`   //净资产
	Total string `json:"total"` //总资产
	Btc   string `json:"btc"`
	Cny   string `json:"cny"`
	Ltc   string `json:"ltc"`
	Eth   string `json:"eth"`
}

func (this *MoneyData) GetLtc() float32 {
	freecount, _ := strconv.ParseFloat(this.Ltc, 32)
	return float32(freecount)
}

type RespDoTrade struct {
	Result  bool   `json:"result"`   //
	OrderId uint32 `json:"order_id"` //订单号
}

type RespCancelOrder struct {
	Result  bool   `json:"result"`
	OrderId uint32 `json:"order_id"` //订单号
}

type RespGetOrderInfo struct {
	Result bool         `json:"result"`
	Orders []OrdersInfo `json:"orders"`
}

type OrdersInfo struct {
	OrderId    uint32  `json:"order_id"`
	Status     uint32  `json:"status"`
	Symbol     string  `json:"symbol"`
	Type       string  `json:"type"`
	Price      float32 `json:"price"`
	Amount     float32 `json:"amount"`
	Dealamount float32 `json:"deal_amount"`
	AvgPrice   float32 `json:"avg_price"`
}

type RespGetOrderHistory struct {
	CurrentPage uint32       `json:"current_page"`
	Orders      []OrdersInfo `json:"orders"`
	PageLength  uint32       `json:"page_length"`
	Result      bool         `json:"result"`
	Total       uint32       `json:"total"`
}

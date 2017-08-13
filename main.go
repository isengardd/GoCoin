package main

import (
	"GoCoin/coinapi"
	"GoCoin/strategy"
	"fmt"
	"time"
	//"strconv"
	//"log"
)

func main() {
	Run()
	//TestMaList()
	//TestLowHighPrice()
	//TestMaList()
	//	lowPrice := coinapi.GetNLowestPrice(coinapi.LTC, 2, "30min", 7)
	//	if lowPrice != nil {
	//		fmt.Printf("%v\n", lowPrice)
	//	}
	//	tick := coinapi.GetTicker(coinapi.LTC)
	//	if tick == nil {
	//		return
	//	}
	//	//计算目标建仓价
	//	curPrice := tick.Tick.GetLast()
	//	if curPrice == 0 {
	//		return
	//	}
	//	if true {
	//		orderId := coinapi.DoTrade(coinapi.LTC, coinapi.BUY_MARKET, curPrice, 0)
	//		if orderId != 0 {
	//			log.Printf("Buy %f\n", curPrice)
	//			coinapi.GetDB().Query(fmt.Sprintf("INSERT INTO order_data(coin_type,order_id,order_time) VALUES('%s', %d, NOW())",
	//				coinapi.LTC, orderId))
	//		} else {
	//			log.Printf("OrderId is 0\n")
	//		}
	//	}
}
func Run() {
	var worker strategy.IntervalBuy
	worker.Run()
}

func DoTest() {
	//	tick := coinapi.GetTicker(coinapi.LTC)
	//	if tick != nil {
	//		fmt.Printf("%v\n", tick)
	//	}

	//	//	depth := coinapi.GetDepth(coinapi.LTC)
	//	//	if depth != nil {
	//	//		fmt.Printf("asks len=%d\n%v\n", len(depth.Asks), depth.Asks)
	//	//		fmt.Printf("bids len=%d\n%v\n", len(depth.Bids), depth.Bids)
	//	//	}

	//	//	trades := coinapi.GetTrades(coinapi.LTC)
	//	//	if trades != nil {
	//	//		fmt.Printf("%v\n", trades)
	//	//	}

	//	//	kline := coinapi.GetKline(coinapi.LTC, "1day", 10, 0)
	//	//	if kline != nil {
	//	//		fmt.Printf("%v\n", kline)
	//	//	}
	//	userInfo := coinapi.GetUserInfo()
	//	if userInfo != nil {
	//		fmt.Printf("%v \n", userInfo)
	//	}

	//	//	price, _ := strconv.ParseFloat(tick.Tick.Sell, 32)
	//	//	orderId := coinapi.DoTrade(coinapi.LTC, "buy_market", float32(price), 0.15)
	//	//	if orderId != 0 {
	//	//		fmt.Printf("%v \n", orderId)
	//	//	}

	//	//	sellorderId := coinapi.DoTrade(coinapi.LTC, "sell_market", 0, 0.998)
	//	//	if sellorderId != 0 {
	//	//		fmt.Printf("%v \n", sellorderId)
	//	//	}

	//	price, _ := strconv.ParseFloat(tick.Tick.Sell, 32)
	//	orderId := coinapi.DoTrade(coinapi.LTC, "buy", float32(price-100), 0.15)
	//	if orderId != 0 {
	//		fmt.Printf("%v \n", orderId)
	//	}

	//	unfinishOrder := coinapi.GetOrderInfo(coinapi.LTC, orderId)
	//	if unfinishOrder != nil {
	//		fmt.Printf("%v\n", unfinishOrder)
	//	}

	//	cancelResult := coinapi.CancelOrder(coinapi.LTC, orderId)
	//	fmt.Println(cancelResult)

}

func TestOrderHistory() {
	history := coinapi.GetOrderHistory(coinapi.LTC, 1, 1, 100)
	if history != nil {
		fmt.Printf("%v\n", history)
	}

	since := int64(time.Now().UnixNano()/int64(time.Millisecond)) - 3600001
	high := coinapi.GetHighestPrice(coinapi.LTC, since)
	fmt.Println(high)
	fmt.Println(since)
	fmt.Println(int64(time.Hour / time.Millisecond))
}

func TestMaList() {
	var avglist = []int{5, 10, 20, 60}
	malistall := coinapi.GetMaList(coinapi.LTC, "15min", avglist)
	if malistall != nil {
		fmt.Printf("MA5,10,20,60: %v\n", malistall)
		fmt.Printf("MA10: %v\n", malistall[10])
		fmt.Printf("MA20: %v\n", malistall[20])
	}

	ma5list := coinapi.GetMaList(coinapi.LTC, "15min", []int{5})
	if ma5list != nil {
		fmt.Printf("MA5: %v\n", ma5list)
	}

	ma60list := coinapi.GetMaList(coinapi.LTC, "15min", []int{60})
	if ma60list != nil {
		fmt.Printf("MA60: %v\n", ma60list)
	}
}

func TestLowHighPrice() {
	since := int64(time.Now().UnixNano()/int64(time.Millisecond)) - int64(time.Hour/time.Millisecond)*24*2
	low, high := coinapi.GetLowHighPrice(coinapi.LTC, since)
	fmt.Printf("low:%v, high:%v\n", low, high)
}

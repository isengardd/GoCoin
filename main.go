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

	var worker strategy.Lowest2buy
	worker.Run()

	//DoTest()

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

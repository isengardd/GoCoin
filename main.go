package main

import (
	"GoCoin/coinapi"
	"fmt"
	"strconv"
)

func main() {
	tick := coinapi.GetTicker(coinapi.LTC)
	if tick != nil {
		fmt.Printf("%v\n", tick)
	}

	//	depth := coinapi.GetDepth(coinapi.LTC)
	//	if depth != nil {
	//		fmt.Printf("asks len=%d\n%v\n", len(depth.Asks), depth.Asks)
	//		fmt.Printf("bids len=%d\n%v\n", len(depth.Bids), depth.Bids)
	//	}

	//	trades := coinapi.GetTrades(coinapi.LTC)
	//	if trades != nil {
	//		fmt.Printf("%v\n", trades)
	//	}

	//	kline := coinapi.GetKline(coinapi.LTC, "1day", 10, 0)
	//	if kline != nil {
	//		fmt.Printf("%v\n", kline)
	//	}
	userInfo := coinapi.GetUserInfo()
	if userInfo != nil {
		fmt.Printf("%v \n", userInfo)
	}

	//	price, _ := strconv.ParseFloat(tick.Tick.Sell, 32)
	//	orderId := coinapi.DoTrade(coinapi.LTC, "buy_market", float32(price), 0.15)
	//	if orderId != 0 {
	//		fmt.Printf("%v \n", orderId)
	//	}

	//	sellorderId := coinapi.DoTrade(coinapi.LTC, "sell_market", 0, 0.998)
	//	if sellorderId != 0 {
	//		fmt.Printf("%v \n", sellorderId)
	//	}

	price, _ := strconv.ParseFloat(tick.Tick.Sell, 32)
	orderId := coinapi.DoTrade(coinapi.LTC, "buy", float32(price-100), 0.15)
	if orderId != 0 {
		fmt.Printf("%v \n", orderId)
	}

	unfinishOrder := coinapi.UnFinishOrderInfo(coinapi.LTC)
	if unfinishOrder != nil {
		fmt.Printf("%v\n", unfinishOrder)
	}

	cancelResult := coinapi.CancelOrder(coinapi.LTC, orderId)
	fmt.Println(cancelResult)
}

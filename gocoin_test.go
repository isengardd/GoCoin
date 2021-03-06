package main

import (
	"GoCoin/coinapi"
	"fmt"
	"testing"
	"time"
)

func TestKDJ(t *testing.T) {
	kline := coinapi.GetKline(coinapi.LTC, "15min", coinapi.MACD_KLINE_MAX, 0)
	if kline == nil || len(*kline) < coinapi.MACD_KLINE_MAX {
		t.Logf("kline is nil or len(*kline)<%d", coinapi.MACD_KLINE_MAX)
		return
	}

	// 按照时间降序
	for i := 0; i < len(*kline)/2; i++ {
		(*kline)[i], (*kline)[len(*kline)-i-1] = (*kline)[len(*kline)-i-1], (*kline)[i]
	}
	k, d := coinapi.GetKDJ((*kline), 15, 5, 5)
	fmt.Printf("curk=%f, curd=%f\n", k, d)
	for i := 1; i < 30; i++ {
		k, d = coinapi.GetKDJ((*kline)[i:], 15, 5, 5)
		fmt.Printf("prekd_%d, k=%f, d=%f\n", i, k, d)
	}
}

func TestRSISiumulate(t *testing.T) {
	kline := coinapi.GetKline(coinapi.LTC, "15min", coinapi.MACD_KLINE_MAX, 0)
	if kline == nil || len(*kline) < coinapi.MACD_KLINE_MAX {
		return
	}

	// 按照时间降序
	for i := 0; i < len(*kline)/2; i++ {
		(*kline)[i], (*kline)[len(*kline)-i-1] = (*kline)[len(*kline)-i-1], (*kline)[i]
	}

	the_time, _ := time.ParseInLocation("2006-01-02 15:04:05", "2017-09-14 19:45:00", time.Local)

	//查找要修改的时间
	fmt.Printf("kline=%v\n", (*kline)[0])
	for i := 1; i < coinapi.MACD_KLINE_MAX; i++ {
		if (*kline)[i].Date == uint64(the_time.Unix())*1000 {
			rsi4 := coinapi.GetRSI((*kline)[i:], 4)
			rsi8 := coinapi.GetRSI((*kline)[i:], 13)
			fmt.Printf("origin rsi4=%f, rsi8=%f\n", rsi4, rsi8)
			(*kline)[i].Close = 351.00
			fmt.Printf("kline=%v\n", (*kline)[i])
			rsi4 = coinapi.GetRSI((*kline)[i:], 4)
			rsi8 = coinapi.GetRSI((*kline)[i:], 13)
			fmt.Printf("dot rsi4=%f, rsi8=%f\n", rsi4, rsi8)

			fmt.Printf("pre kline=%v\n", (*kline)[i+1])
			rsi4 = coinapi.GetRSI((*kline)[i+1:], 4)
			rsi8 = coinapi.GetRSI((*kline)[i+1:], 13)
			fmt.Printf("pre rsi4=%f, rsi8=%f\n", rsi4, rsi8)
		}
	}
}

func TestRSI(t *testing.T) {
	kline := coinapi.GetKline(coinapi.LTC, "15min", coinapi.MACD_KLINE_MAX, 0)
	if kline == nil || len(*kline) < coinapi.MACD_KLINE_MAX {
		t.Logf("kline is nil or len(*kline)<%d", coinapi.MACD_KLINE_MAX)
		return
	}

	// 按照时间降序
	for i := 0; i < len(*kline)/2; i++ {
		(*kline)[i], (*kline)[len(*kline)-i-1] = (*kline)[len(*kline)-i-1], (*kline)[i]
	}
	rsi := coinapi.GetRSI((*kline), 4)
	fmt.Printf("currsi4=%f\n", rsi)
	for i := 1; i < 30; i++ {
		rsi = coinapi.GetRSI((*kline)[i:], 4)
		fmt.Printf("prersi4_%d=%f\n", i, rsi)
	}
}

func TestMACD(t *testing.T) {
	kline := coinapi.GetKline(coinapi.LTC, "15min", coinapi.MACD_KLINE_MAX, 0)
	if kline == nil || len(*kline) < coinapi.MACD_KLINE_MAX {
		t.Logf("kline is nil or len(*kline)<%d", coinapi.MACD_KLINE_MAX)
		return
	}

	// 按照时间降序
	for i := 0; i < len(*kline)/2; i++ {
		(*kline)[i], (*kline)[len(*kline)-i-1] = (*kline)[len(*kline)-i-1], (*kline)[i]
	}

	curmacd := coinapi.GetMACDBar(*kline)
	premacd := coinapi.GetMACDBar((*kline)[1:])
	fmt.Printf("curmacd=%f,premacd=%f\n", curmacd, premacd)
}

func TestMaList(t *testing.T) {
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

func TestOrderHistory(t *testing.T) {
	history := coinapi.GetOrderHistory(coinapi.LTC, 1, 1, 100)
	if history != nil {
		t.Logf("%v\n", history)
	}

	since := int64(time.Now().UnixNano()/int64(time.Millisecond)) - 3600001
	high := coinapi.GetHighestPrice(coinapi.LTC, since)
	fmt.Printf("%v\n", high)
	fmt.Printf("%v\n", since)
	fmt.Printf("%v\n", int64(time.Hour/time.Millisecond))
}

func TestLowHighPrice(t *testing.T) {
	since := int64(time.Now().UnixNano()/int64(time.Millisecond)) - int64(time.Hour/time.Millisecond)*36
	low, high := coinapi.GetLowHighPrice(coinapi.LTC, since)
	fmt.Printf("low:%v, high:%v\n", low, high)
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

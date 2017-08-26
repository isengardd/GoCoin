package coinapi

import (
	//"fmt"
	"sort"
	"time"
	//"unsafe"
)

type Float32Slice []float32

func (p Float32Slice) Len() int           { return len(p) }
func (p Float32Slice) Less(i, j int) bool { return p[i] < p[j] }
func (p Float32Slice) Swap(i, j int)      { p[i], p[j] = p[j], p[i] }

func GetNLowestPrice(symbol string, n uint32, intv string, size int32) []float32 {
	if size <= 0 {
		return nil
	}

	kline := GetKline(symbol, intv, size+1, 0)
	if kline == nil || len(*kline) == 0 {
		return nil
	}

	lowestPrice := make(Float32Slice, 0)
	for idx, val := range *kline {
		if idx == 0 || idx == len(*kline)-1 {
			continue
		}

		if (*kline)[idx-1].Low > val.Low &&
			(*kline)[idx+1].Low > val.Low {
			lowestPrice = append(lowestPrice, val.Low)
		}
	}

	//copyList := lowestPrice[:]
	sort.Sort(lowestPrice)

	if uint32(len(lowestPrice)) > n {
		lowestPrice = lowestPrice[:n]
	}
	return lowestPrice
}

func GetHighestPrice(symbol string, since int64) float32 {

	var intv string
	nowTime := time.Now().UnixNano() / int64(time.Millisecond)
	if nowTime-since < int64(time.Hour/time.Millisecond) {
		intv = "1min"
	} else {
		intv = "1hour"
	}

	kline := GetKline(symbol, intv, -1, since)
	if kline == nil || len(*kline) == 0 {
		return 0
	}

	var maxprice float32 = 0
	for _, val := range *kline {
		if val.High > maxprice {
			maxprice = val.High
		}
	}
	return maxprice
}

func GetLowHighPrice(symbol string, since int64) (RespKline, RespKline) {
	var intv string
	nowTime := time.Now().UnixNano() / int64(time.Millisecond)
	if nowTime-since < int64(time.Hour/time.Millisecond) {
		intv = "1min"
	} else {
		intv = "1hour"
	}

	kline := GetKline(symbol, intv, -1, since)
	if kline == nil || len(*kline) == 0 {
		return RespKline{}, RespKline{}
	}

	var maxprice RespKline
	maxprice.High = 0
	var lowprice RespKline
	lowprice.Low = 999999
	for _, val := range *kline {
		if val.High > maxprice.High {
			maxprice = val
		}

		if val.Low < lowprice.Low {
			lowprice = val
		}
	}
	return lowprice, maxprice
}

func GetMaList(symbol string, intv string, avglist []int) map[uint][]float32 {
	if len(avglist) == 0 {
		return nil
	}

	sort.Ints(avglist)
	maxAvg := avglist[len(avglist)-1]

	kline := GetKline(symbol, intv, int32(maxAvg+1), 0)
	if kline == nil || len(*kline) < int(maxAvg+1) {
		return nil
	}

	var sum float32 = 0
	var sumpre float32 = 0
	var avgcount int = 0
	malist := make(map[uint][]float32, 0)
	for idx := len(*kline) - 1; idx > 0; idx-- {
		sum += (*kline)[idx].Close
		sumpre += (*kline)[idx-1].Close

		avgcount++
		if avgcount == avglist[len(malist)] {
			malist[uint(avgcount)] = make([]float32, 0)
			malist[uint(avgcount)] = append(malist[uint(avgcount)], sumpre/float32(avgcount))
			malist[uint(avgcount)] = append(malist[uint(avgcount)], sum/float32(avgcount))
		}
	}

	return malist
}

/*////////////////////////////////////////
MACD相关
一、差离值（DIF值）：[2][3]
先利用收盘价的指数移动平均值（12日／26日）计算出差离值。[4]
〖公式〗
DIF=EMA_{{(close,12)}}-EMA_{{(close,26)}}
二、讯号线（DEM值，又称MACD值）：
计算出DIF后，会再画一条“讯号线”，通常是DIF的9日指数移动平均值。
〖公式〗
DEM=EMA_{{(DIF,9)}}
三、柱形图或棒形图（histogram / bar graph）：
接着，将DIF与DEM的差画成“柱形图”（MACD bar / OSC）。
〖公式〗
OSC=DIF-DEM
简写为
D-M
////////////////////////////////////////*/
const (
	MACD_KLINE_MAX = 400
	F              = 5.45
	N9             = 9
	a9             = float32(2) / float32(10)
	k9             = float32(F) * 10
	N12            = 12
	a12            = (float32)(2) / (float32)(13)
	k12            = (float32(F) * 13)
	N26            = 26
	a26            = (float32)(2) / (float32)(27)
	k26            = (float32(F) * 27)
)

var raList9 []float32 = nil
var raList12 []float32 = nil
var raList26 []float32 = nil

func init() {
	listlen := k9
	raList9 = make([]float32, int32(listlen))
	raList9[0] = 1
	for i := 1; i < len(raList9); i++ {
		raList9[i] = raList9[i-1] * (1 - a9)
	}

	listlen = k12
	raList12 = make([]float32, int32(listlen))
	raList12[0] = 1
	for i := 1; i < len(raList12); i++ {
		raList12[i] = raList12[i-1] * (1 - a12)
	}

	listlen = k26
	raList26 = make([]float32, int32(listlen))
	raList26[0] = 1
	for i := 1; i < len(raList26); i++ {
		raList26[i] = raList26[i-1] * (1 - a26)
	}

	//fmt.Printf("%v\n", raList9)
	//fmt.Printf("%v\n", raList12)
	//fmt.Printf("%v\n", raList26)
}

func GetEMA(kline []RespKline, count int) float32 {
	if kline == nil {
		return 0
	}

	minLen := k26
	if len(kline) < int(minLen) {
		return 0
	}

	var raList []float32 = nil
	var a float32 = 0
	if count == N12 {
		raList = raList12
		a = a12
	} else if count == N26 {
		raList = raList26
		a = a26
	} else if count == N4 {
		raList = raList4
		a = a4

	} else if count == N7 {
		raList = raList7
		a = a7
	} else if count == N6 {
		raList = raList6
		a = a6
	} else {
		return 0
	}

	var ema float32 = 0
	for idx, val := range raList {
		ema += (kline)[idx].Close * val
	}
	ema *= a
	return ema
}

func GetDIF(kline []RespKline) float32 {
	return GetEMA(kline, N12) - GetEMA(kline, N26)
}

func GetDEM(kline []RespKline) float32 {
	if kline == nil {
		return 0
	}

	minLen := k26 + k9
	if len(kline) < int(minLen) {
		return 0
	}

	listlen := k9
	a := a9
	raList := raList9
	difList := make([]float32, int(listlen))
	for idx, _ := range difList {
		difList[idx] = GetDIF(kline[idx:])
	}

	var dem float32 = 0
	for idx, val := range raList {
		dem += difList[idx] * val
	}
	dem *= a
	return dem
}

func GetMACDBar(kline []RespKline) float32 {
	var fDif = GetDIF(kline)
	var fDem = GetDEM(kline)
	result := (fDif - fDem) * 2
	//fmt.Printf("dif=%f, dem=%f\n", fDif, fDem)
	return result
}

/*////////////////////////////////////////
RSI相关
设每天向上变动为U，向下变动为D。
在价格上升的日子：
U ＝ 是日收市价 － 昨日收市价；D ＝ 0
在价格下跌的日子：
U ＝ 0；D ＝ 昨日收市价 － 是日收市价
（任何情况下，U及D皆不可能为负数；若两天价格相同，则U及D皆等于零。）
U及D的平均值皆需用上“指数移动平均法”（在n日内）。所谓“相对强度”，即U平均值及D平均值的比例：
RS =EMA_{{(U,n)}/EMA_{{(D,n)}
RSI = (1-(1/(1+RS)))*100
【注】
RS：相对强度（Relative Strength）；
RSI: 相对强弱指数（Relative Strength Index）；
{EMA_{{(U,n)}}：U在n日内的指数平均值；
{EMA_{{(D,n)}}：D在n日内的指数平均值。

考虑n=4(10,80)和n=7(20,80)两种
////////////////////////////////////////*/
const (
	N4 = 4
	a4 = float32(1) / float32(4) //这里首项系数是1/n
	k4 = float32(F) * 5
	N6 = 6
	a6 = float32(1) / float32(6) //这里首项系数是1/n
	k6 = float32(F) * 7
	N7 = 7
	a7 = float32(1) / float32(7) //这里首项系数是1/n
	k7 = float32(F) * 8
	// todo:如果有跟MACD重复的天数，需要重构GetEMA的代码
)

var raList4 []float32 = nil
var raList6 []float32 = nil
var raList7 []float32 = nil

func init() {
	listlen := k4
	raList4 = make([]float32, int32(listlen))
	raList4[0] = 1
	for i := 1; i < len(raList4); i++ {
		raList4[i] = raList4[i-1] * (1 - a4)
	}

	listlen = k7
	raList7 = make([]float32, int32(listlen))
	raList7[0] = 1
	for i := 1; i < len(raList7); i++ {
		raList7[i] = raList7[i-1] * (1 - a7)
	}

	listlen = k6
	raList6 = make([]float32, int32(listlen))
	raList6[0] = 1
	for i := 1; i < len(raList6); i++ {
		raList6[i] = raList6[i-1] * (1 - a6)
	}
}

func GetRSI(kline []RespKline, count int) float32 {
	if kline == nil {
		return 0
	}

	uKline := make([]RespKline, len(kline))
	dKline := make([]RespKline, len(kline))
	for idx, val := range kline {
		if idx == len(kline)-1 {
			continue
		}

		uKline[idx] = val
		dKline[idx] = val
		if val.Close >= kline[idx+1].Close {
			uKline[idx].Close = val.Close - kline[idx+1].Close
			dKline[idx].Close = 0
		} else {
			uKline[idx].Close = 0
			dKline[idx].Close = kline[idx+1].Close - val.Close
		}
	}

	uEma := GetEMA(uKline, count)
	dEma := GetEMA(dKline, count)

	if uEma == 0 || dEma == 0 {
		return 0
	}

	rsi := uEma / (uEma + dEma) * 100
	return rsi
}

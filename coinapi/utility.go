package coinapi

import (
	//"fmt"
	"math"
	"sort"
	"time"
	//"unsafe"
	"log"
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

var typeMap map[string]RaData = nil

func init() {
	varMacd := &MACD{}
	varMacd.Init()
	varRsi := &RSI{}
	varRsi.Init()
	varKdj := &KDJ{}
	varKdj.Init()
	typeMap = make(map[string]RaData, 0)
	typeMap["macd"] = varMacd
	typeMap["rsi"] = varRsi
	typeMap["kdj"] = varKdj
}

func GetEMA(valList []float32, datatype string, count uint32) float32 {
	if valList == nil {
		return 0
	}

	dataCont, ok := typeMap[datatype]
	if !ok {
		log.Printf("GetEMA %s, not implement\n", datatype)
		return 0
	}

	minLen := int(dataCont.GetLen(count))
	if len(valList) < int(minLen) {
		log.Printf("GetEMA %s,len(kline)=%d too short, need=%d\n", datatype, len(valList), minLen)
		return 0
	}

	var raList []float32 = dataCont.GetRaList(count)
	var a float32 = dataCont.GetA(count)
	if raList == nil || a == 0 {
		log.Printf("GetEMA %s, number=%d not data\n", datatype, count)
		return 0
	}

	var ema float32 = 0
	for idx, val := range raList {
		ema += (valList)[idx] * val
	}
	ema *= a
	return ema
}

func GetDIF(kline []RespKline) float32 {
	valList := make([]float32, len(kline))
	for i, val := range kline {
		valList[i] = val.Close
	}
	return GetEMA(valList, "macd", 12) - GetEMA(valList, "macd", 26)
}

func GetDEM(kline []RespKline) float32 {
	if kline == nil {
		return 0
	}

	dataCont, ok := typeMap["macd"]
	if !ok {
		return 0
	}
	minLen := MIN_KLINE_LEN + dataCont.GetLen(9)
	if len(kline) < int(minLen) {
		return 0
	}

	listlen := dataCont.GetLen(9)
	a := dataCont.GetA(9)
	raList := dataCont.GetRaList(9)
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

func GetRSI(kline []RespKline, count uint32) float32 {
	if kline == nil {
		return 0
	}

	uKline := make([]float32, len(kline))
	dKline := make([]float32, len(kline))
	for idx, val := range kline {
		if idx == len(kline)-1 {
			continue
		}

		uKline[idx] = 0
		dKline[idx] = 0
		if val.Close >= kline[idx+1].Close {
			uKline[idx] = val.Close - kline[idx+1].Close
			dKline[idx] = 0
		} else {
			uKline[idx] = 0
			dKline[idx] = kline[idx+1].Close - val.Close
		}
	}

	uEma := GetEMA(uKline, "rsi", count)
	dEma := GetEMA(dKline, "rsi", count)

	if uEma == 0 || dEma == 0 {
		return 0
	}

	rsi := uEma / (uEma + dEma) * 100
	return rsi
}

func GetRSV(kline []RespKline, count uint32) float32 {
	if len(kline) < int(count) {
		log.Printf("GetRSV len(kline) < count, count=%d\n", count)
		return 0
	}

	max := GetMaxPrice(kline[:count])
	min := GetMinPrice(kline[:count])

	if 0 == max || 0 == min || min == max {
		log.Printf("GetRSV error, max=%f,min=%f\n", max, min)
		return 0
	}

	rsv := (kline[0].Close - min) / (max - min) * 100
	if rsv < 1 {
		rsv = 1
	}
	if rsv > 100 {
		rsv = 100
	}

	return rsv
}

func GetMaxPrice(kline []RespKline) float32 {
	var max float32 = 0
	for _, val := range kline {
		if val.High > max {
			max = val.High
		}
	}

	return max
}

func GetMinPrice(kline []RespKline) float32 {
	var min float32 = math.MaxFloat32
	for _, val := range kline {
		if val.Low < min {
			min = val.Low
		}
	}

	return min
}

// 返回k,d值
func GetKDJ(kline []RespKline, N uint32, M1 uint32, M2 uint32) (float32, float32) {
	if N == 0 || M1 == 0 || M2 == 0 {
		return 0, 0
	}

	dataCont, ok := typeMap["kdj"]
	if !ok {
		return 0, 0
	}

	//算出RSV
	rsvDay := int(dataCont.GetLen(M2)) + int(dataCont.GetLen(M1)) + int(N)
	if len(kline) < rsvDay {
		log.Printf("GetKDJ len(kline) < rsvDay, rsvDay=%d\n", rsvDay)
		return 0, 0
	}
	//rsvDay := uint32(dataCont.GetLen(N)) + N

	rsvList := make([]float32, rsvDay)
	for i := 0; i < rsvDay; i++ {
		rsvList[i] = GetRSV(kline[i:int(N)+i], N)
	}

	//算出K值
	kDay := int32(dataCont.GetLen(M2))
	kList := make([]float32, kDay)
	kList[len(kList)-1] = GetEMA(rsvList[len(kList)-1:len(kList)+int(dataCont.GetLen(M1))-1], "kdj", M1)
	for i := len(kList) - 2; i >= 0; i-- {
		kList[i] = dataCont.GetA(M1)*rsvList[i] + kList[i+1]*(1-dataCont.GetA(M1))
	}

	//算出D值
	DVal := GetEMA(kList, "kdj", M2)
	return kList[0], DVal
}

//////////////////////////////////////////
type RaData interface {
	GetRaList(number uint32) []float32
	GetA(number uint32) float32
	GetLen(number uint32) float32
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
	MIN_KLINE_LEN  = (float32(F) * 27)
)

// 计算MACD的对象
type MACD struct {
	raFactorMap map[uint32][]float32
}

func (this *MACD) Init() {
	this.raFactorMap = make(map[uint32][]float32, 0)
}

func (this *MACD) GetRaList(number uint32) []float32 {
	if 0 == number {
		return nil
	}

	this.CalcList(number)
	return this.raFactorMap[number]
}

func (this *MACD) GetA(number uint32) float32 {
	if 0 == number {
		return 0
	}

	return float32(2) / float32(number+1)
}

func (this *MACD) GetLen(number uint32) float32 {
	if 0 == number {
		return 0
	}

	return float32(F) * float32(number+1)
}

func (this *MACD) CalcList(number uint32) {
	if _, ok := this.raFactorMap[number]; ok {
		return
	}

	if 0 == number {
		return
	}

	var n uint32 = number
	var a float32 = this.GetA(n)
	var k float32 = this.GetLen(n)
	raList := make([]float32, int32(k))

	raList[0] = 1
	for i := 1; i < len(raList); i++ {
		raList[i] = raList[i-1] * (1 - a)
	}
	this.raFactorMap[number] = raList
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

type RSI struct {
	raFactorMap map[uint32][]float32
}

func (this *RSI) Init() {
	this.raFactorMap = make(map[uint32][]float32, 0)
}

func (this *RSI) GetRaList(number uint32) []float32 {
	if 0 == number {
		return nil
	}

	this.CalcList(number)
	return this.raFactorMap[number]
}

func (this *RSI) GetA(number uint32) float32 {
	if 0 == number {
		return 0
	}

	return float32(1) / float32(number)
}

func (this *RSI) GetLen(number uint32) float32 {
	if 0 == number {
		return 0
	}

	return float32(F) * float32(number+1)
}

func (this *RSI) CalcList(number uint32) {
	if _, ok := this.raFactorMap[number]; ok {
		return
	}

	if 0 == number {
		return
	}

	var n uint32 = number
	var a float32 = this.GetA(n)
	var k float32 = this.GetLen(n)
	raList := make([]float32, int32(k))

	raList[0] = 1
	for i := 1; i < len(raList); i++ {
		raList[i] = raList[i-1] * (1 - a)
	}
	this.raFactorMap[number] = raList
}

/*////////////////////////////////////////
KDJ相关
n日RSV=（Cn－Ln）÷（Hn－Ln）×100

式中，Cn为第n日收盘价；Ln为n日内的最低价；Hn为n日内的最高价。RSV值始终在1—100间波动。

其次，计算K值与D值：

当日K值=2/3×前一日K值＋1/3×当日RSV

当日D值=2/3×前一日D值＋1/3×当日K值

若无前一日K 值与D值，则可分别用50来代替。

以9日为周期的KD线为例。首先须计算出最近9日的RSV值，即未成熟随机值，计算公式为

9日RSV=（C－L9）÷（H9－L9）×100

式中，C为第9日的收盘价；L9为9日内的最低价；H9为9日内的最高价。

K值=2/3×前一日 K值＋1/3×当日RSV

D值=2/3×前一日K值＋1/3×当日RSV

若无前一日K值与D值，则可以分别用50代替。

*/ ////////////////////////////////////////
type KDJ struct {
	raFactorMap map[uint32][]float32
}

func (this *KDJ) Init() {
	this.raFactorMap = make(map[uint32][]float32, 0)
}

func (this *KDJ) GetLen(number uint32) float32 {
	if 0 == number {
		return 0
	}

	return float32(F) * float32(number+1)
}

func (this *KDJ) GetA(number uint32) float32 {
	return float32(1) / float32(number)
}

func (this *KDJ) GetRaList(number uint32) []float32 {
	if 0 == number {
		return nil
	}

	this.CalcList(number)
	return this.raFactorMap[number]
}

func (this *KDJ) CalcList(number uint32) {
	if _, ok := this.raFactorMap[number]; ok {
		return
	}

	if 0 == number {
		return
	}

	var n uint32 = number
	var a float32 = this.GetA(n)
	var k float32 = this.GetLen(n)
	raList := make([]float32, int32(k))

	raList[0] = 1
	for i := 1; i < len(raList); i++ {
		raList[i] = raList[i-1] * (1 - a)
	}
	this.raFactorMap[number] = raList
}

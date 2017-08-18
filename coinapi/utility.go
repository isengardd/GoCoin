package coinapi

import (
	//	"fmt"
	"sort"
	"time"
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

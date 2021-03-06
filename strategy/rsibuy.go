package strategy

import (
	"GoCoin/coinapi"
	"fmt"
	"log"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

const (
	FLAG_RSI_LARGE_60   = 1 << 0
	FLAG_SELL_IMMEDIATE = 1 << 1
	FLAG_RSI_LARGE_50   = 1 << 2
	FLAG_RSI_LARGE_40   = 1 << 3
	FLAG_KDJ_LARGE_60   = 1 << 4
	FLAG_KDJ_LARGE_50   = 1 << 5
	FLAG_KDJ_LARGE_80   = 1 << 6
)

var (
	STOP_LOSS_RATE       float32 = 0.95        // 止损率
	TRADE_COOL_TIME      int64   = 2700 * 1000 //交易冷却时间
	ORDER_DELAY_TIME_MAX int64   = 60 * 1000   //交易等待时间
	PER_ORDER_COOL_TIME  int64   = 30 * 1000   //每一单的交易间隔
	RSI_LEVEL                    = uint32(8)
)

type RsiBuy struct {
	userInfo      *coinapi.RespUserInfo
	tickInfo      *coinapi.RespTicker
	kline         *[]coinapi.RespKline
	buyOrders     []OrderData
	sellOrders    []OrderData
	coolTime      int64
	lastTradeTime int64 // 上一次开仓时间
	state         uint32
	flag          uint32
}

func (this *RsiBuy) Init() {
	this.buyOrders = make([]OrderData, 0)
	this.sellOrders = make([]OrderData, 0)
	this.coolTime = 0
	this.lastTradeTime = 0
	this.flag = 0
	this.state = 0
	this.kline = nil
	//todo,从数据库加载订单信息，初始化当前状态

	this.LoadOrders(coinapi.LTC)
	this.LoadParams()

	fmt.Printf("%v", this.buyOrders)
	fmt.Printf("%v", this.sellOrders)
}

func (this *RsiBuy) LoadOrders(symbol string) {
	rows, err := coinapi.GetDB().Query(fmt.Sprintf("SELECT order_id,order_type,order_time,count,price FROM order_data WHERE coin_type='%s'", symbol))
	if err != nil {
		log.Println(err)
		return
	}

	defer rows.Close()

	for rows.Next() {
		var orderId uint32
		var orderType string
		var orderTime uint64
		var count float32
		var price float32
		err := rows.Scan(&orderId, &orderType, &orderTime, &count, &price)
		if err != nil {
			log.Println(err)
			continue
		}

		orderData := OrderData{}
		orderData.symbol = symbol
		orderData.orderId = orderId
		orderData.orderType = orderType
		orderData.orderTime = int64(orderTime) //单位是毫秒
		orderData.count = count
		orderData.price = price

		if orderData.orderType == coinapi.BUY {
			this.buyOrders = append(this.buyOrders, orderData)
		} else if orderData.orderType == coinapi.SELL {
			this.sellOrders = append(this.sellOrders, orderData)
		} else {
			log.Printf("error: LoadOrders orderType=%s", orderData.orderType)
		}
	}
	return
}

func (this *RsiBuy) LoadParams() {
	rows, err := coinapi.GetDB().Query("SELECT name,param FROM sys_param")
	if err != nil {
		log.Println(err)
		return
	}

	defer rows.Close()

	for rows.Next() {
		var name string
		var val uint32

		err := rows.Scan(&name, &val)
		if err != nil {
			log.Println(err)
			continue
		}

		if name == "cooltime" {
			this.coolTime = int64(val) * 1000
		} else if name == "flag" {
			this.flag = val
		} else if name == "state" {
			this.state = val
		} else if name == "lasttradetime" {
			this.lastTradeTime = int64(val) * 1000
		}
	}
}

func (this *RsiBuy) SaveParams(name string, param uint32) {
	rows, err := coinapi.GetDB().Query(fmt.Sprintf("INSERT INTO sys_param(name, param) VALUES('%s', %d) ON DUPLICATE KEY UPDATE param=%d", name, param, param))
	if err != nil {
		log.Println(err)
	}

	defer rows.Close()
}

func (this *RsiBuy) SaveOrder(order OrderData) {
	rows, err := coinapi.GetDB().Query(fmt.Sprintf("INSERT INTO order_data(coin_type, order_id, order_type, order_time, count, price) VALUES('%s', %d, '%s', %v, %f, %f) ON DUPLICATE KEY UPDATE count=%f, price=%f",
		order.symbol, order.orderId, order.orderType, order.orderTime, order.count, order.price, order.count, order.price))
	if err != nil {
		log.Println(err)
	}

	defer rows.Close()
}

func (this *RsiBuy) DeleteOrder(order OrderData) {
	rows, err := coinapi.GetDB().Query(fmt.Sprintf("DELETE FROM order_data WHERE coin_type='%s' AND order_id=%d", order.symbol, order.orderId))
	if err != nil {
		log.Println(err)
	}

	defer rows.Close()
}

func (this *RsiBuy) ClearAllOrder() {
	rows, err := coinapi.GetDB().Query("TRUNCATE order_data")
	if err != nil {
		log.Println(err)
	}

	defer rows.Close()
}

func (this *RsiBuy) BindFlag(bit uint32) {
	this.flag |= bit
}

func (this *RsiBuy) ClearFlag(bit uint32) {
	this.flag &= (^bit)
}

func (this *RsiBuy) HasFlag(bit uint32) bool {
	return (this.flag & bit) > 0
}

func (this *RsiBuy) Run() {
	this.Init()
	// 1秒执行一次
	t1 := time.NewTimer(time.Second)

	for {
		select {
		case <-t1.C:
			this.DoStrategy(t1)

			this.SaveParams("cooltime", uint32(this.coolTime/1000))
			this.SaveParams("flag", this.flag)
			this.SaveParams("state", this.state)
			this.SaveParams("lasttradetime", uint32(this.lastTradeTime/1000))

			this.kline = nil
		}
	}

}

func (this *RsiBuy) DoStrategy(t *time.Timer) {
	defer t.Reset(5 * time.Second)

	//获取用户信息
	this.userInfo = coinapi.GetUserInfo()
	if this.userInfo == nil || this.userInfo.Result == false {
		return
	}

	//获取市场最新数据
	this.tickInfo = coinapi.GetTicker(coinapi.LTC)
	if this.tickInfo == nil || 0 == this.tickInfo.Tick.GetLast() {
		return
	}

	this.FetchKline()

	switch this.state {
	case STATE_NONE:
		this.state = STATE_WAIT_BUY
	case STATE_WAIT_BUY:
		this.OnWaitBuy()
	case STATE_BUY_IN:
		this.OnBuyIn()
	case STATE_WAIT_SELL:
		this.OnWaitSell()
	case STATE_SELL_OUT:
		this.OnSellOut()
	default:
		break
	}
}

func (this *RsiBuy) OnWaitBuy() {
	if GetNowTime() < this.coolTime {
		return
	}

	prersi := this.GetPreRsi(true)
	if prersi <= 0 {
		return
	}

	curD := this.GetKDJ_DVal()
	if curD <= 0 {
		return
	}
	/*	if GetNowTime() <= (this.lastTradeTime + 6*3600*1000) {
		//6小时内交易过
		rsiNow := this.GetRsiNow()
		if rsi <= 15 && rsiNow > 15 {
			this.lastTradeTime = GetNowTime()
			this.state = STATE_BUY_IN
			log.Printf("OnWaitBuy prersi=%f currsi=%f buy in", rsi, rsiNow)
		}
	} else */
	rsiNow := this.GetRsiNow()
	//	if prersi <= 20 && rsiNow > 20 && curD <= 25 {
	//		this.lastTradeTime = GetNowTime()
	//		this.state = STATE_BUY_IN
	//		log.Printf("OnWaitBuy prersi=%f currsi=%f curD=%f buyin", prersi, rsiNow, curD)
	//		return
	//	}

	//综合考虑rsi和kdj
	if curD <= 25 && rsiNow <= 25 {
		if curD*0.5+rsiNow*0.5 <= 22.5 {
			this.lastTradeTime = GetNowTime()
			this.state = STATE_BUY_IN
			log.Printf("OnWaitBuy prersi=%f currsi=%f curD=%f buyin", prersi, rsiNow, curD)
			return
		}
	}
}

func (this *RsiBuy) OnBuyIn() {
	// 15%, 25%, 35%, 25% 每5分钟
	var buyrate = [4]float32{0.30, 0.25, 0.25, 0.20}
	for idx := int(0); idx < len(buyrate); idx++ {
		sum := float32(0)
		for innerIdx := idx; innerIdx < len(buyrate); innerIdx++ {
			sum += buyrate[innerIdx]
		}

		buyrate[idx] = buyrate[idx] / sum
	}

	// 等待上一单成交
	var bCancel bool = false
	if !this.WaitBuyOrderDone(bCancel) {
		return
	}

	if bCancel {
		this.userInfo = coinapi.GetUserInfo()
		if this.userInfo == nil || this.userInfo.Result == false {
			return
		}
	}

	orderLen := len(this.buyOrders)
	switch orderLen {
	case 0:
		cny := this.userInfo.Info.Funds.Free.GetCny() * buyrate[orderLen]
		orderData := this.Buy(coinapi.LTC, cny)
		if orderData.orderId != 0 {
			this.buyOrders = append(this.buyOrders, orderData)
			this.SaveOrder(orderData)
		} else {
			log.Printf("OnBuyIn case 0, orderid is 0, cny=%f", cny)
		}
	case 1, 2, 3:
		nowTime := GetNowTime()
		preOrder := this.buyOrders[orderLen-1]

		if orderLen == 3 {
			//如果是最后一次购买，提前检查一下仓位
			curPrice := this.tickInfo.Tick.GetLast()
			coinCount := this.GetPositionCoinCount()
			cny := this.userInfo.Info.Funds.Free.GetCny()
			totalcny := cny + this.GetPositionTotalCny()
			if coinCount > 0 {
				if curPrice <= (STOP_LOSS_RATE*totalcny-cny)/coinCount {
					// 当前价格，买入就要立刻卖出，所以直接卖出
					this.BindFlag(FLAG_SELL_IMMEDIATE)
					this.state = STATE_SELL_OUT
					log.Printf("OnBuyIn SELL STOP LOSS: Price=%f,C=%f,T=%f,CNY=%f",
						curPrice, coinCount, totalcny, cny)
					return
				}
			}
		}

		if nowTime > preOrder.orderTime+PER_ORDER_COOL_TIME {
			cny := this.userInfo.Info.Funds.Free.GetCny() * buyrate[orderLen]
			orderData := this.Buy(coinapi.LTC, cny)
			if orderData.orderId != 0 {
				this.buyOrders = append(this.buyOrders, orderData)
				this.SaveOrder(orderData)
			} else {
				log.Printf("OnBuyIn case %d, orderid is 0, cny=%f", orderLen, cny)
			}
		}
	case 4:
		this.state = STATE_WAIT_SELL
	default:
		log.Printf("error: OnBuyIn len(buyOrders)=%d", orderLen)
	}
}

func (this *RsiBuy) OnWaitSell() {
	//
	if len(this.buyOrders) == 0 {
		log.Println("OnWaitSell len(buyorder) is 0")
		return
	}

	rsi := this.GetRsiNow()
	if rsi == 0 {
		return
	}

	d := this.GetKDJ_DVal()
	if d == 0 {
		return
	}

	if rsi >= 60 {
		this.BindFlag(FLAG_RSI_LARGE_60)
	} else if rsi >= 50 {
		this.BindFlag(FLAG_RSI_LARGE_50)
	} else if rsi >= 40 {
		this.BindFlag(FLAG_RSI_LARGE_40)
	}

	if d >= 80 {
		this.BindFlag(FLAG_KDJ_LARGE_80)
	} else if d >= 60 {
		this.BindFlag(FLAG_KDJ_LARGE_60)
	} else if d >= 50 {
		this.BindFlag(FLAG_KDJ_LARGE_50)
	}

	position := this.GetAvgPosition()
	if position == 0 {
		log.Println("OnWaitSell position is 0")
	}

	//低于成本价x%，卖出
	curPrice := this.tickInfo.Tick.GetLast()
	if curPrice <= position*STOP_LOSS_RATE {
		this.BindFlag(FLAG_SELL_IMMEDIATE)
		this.state = STATE_SELL_OUT
		log.Println("OnWaitSell StopLoss")
		return
	}

	// 如果kdj的d值大于80，卖出
	if d >= 80 {
		this.state = STATE_SELL_OUT
		log.Printf("OnWaitSell kdj=%f, SELL", d)
		return
	}

	if this.HasFlag(FLAG_KDJ_LARGE_60) && d <= 44 {
		this.BindFlag(FLAG_SELL_IMMEDIATE)
		this.state = STATE_SELL_OUT
		log.Printf("OnWaitSell kdj=%f, <= 44, SELL", d)
		return
	}

	//如果MACD在低谷，执行较为严格的止损
	//	var fDif = float32(0)
	//	var fDem = float32(0)
	//	if this.kline != nil {
	//		fDif = coinapi.GetDIF(*this.kline)
	//		fDem = coinapi.GetDEM(*this.kline)
	//		//fmt.Printf("fDif=%f, fDem = %f macd=%f\n", fDif, fDem, (fDif-fDem)*2)

	//		if (fDif < 0 || fDem < 0) && (fDif-fDem) < 0 {
	//			if this.HasFlag(FLAG_RSI_LARGE_40) &&
	//				curPrice <= position*0.96 {
	//				this.BindFlag(FLAG_SELL_IMMEDIATE)
	//				this.state = STATE_SELL_OUT
	//				log.Println("OnWaitSell FLAG_RSI_LARGE_40 , price <0.96*position")
	//				log.Printf("fDif=%f, fDem = %f\n", fDif, fDem)
	//				return
	//			}

	//			if this.HasFlag(FLAG_RSI_LARGE_50) &&
	//				curPrice <= position*0.985 {
	//				this.BindFlag(FLAG_SELL_IMMEDIATE)
	//				this.state = STATE_SELL_OUT
	//				log.Println("OnWaitSell FLAG_RSI_LARGE_50 , price <0.985*position")
	//				log.Printf("fDif=%f, fDem = %f\n", fDif, fDem)
	//				return
	//			}

	//			if this.HasFlag(FLAG_RSI_LARGE_60) &&
	//				curPrice <= position*1.010 {
	//				this.BindFlag(FLAG_SELL_IMMEDIATE)
	//				this.state = STATE_SELL_OUT
	//				log.Println("OnWaitSell price<1.010*position SELL")
	//				log.Printf("fDif=%f, fDem = %f\n", fDif, fDem)
	//				return
	//			}
	//		}
	//	} else {
	//		log.Println("OnWaitSell this.kline is nil")
	//	}

	//	if rsi >= 80 {
	//		this.state = STATE_SELL_OUT
	//		log.Printf("OnWaitSell rsi=%f, SELL", rsi)
	//		return
	//	}

	//	if fDif >= 0.5 && fDem >= 0.5 && (fDif-fDem) < 0 {
	//		//上一个时间单位是MACD>0
	//		var fPreDif = coinapi.GetDIF((*this.kline)[1:])
	//		var fPreDem = coinapi.GetDEM((*this.kline)[1:])
	//		if fPreDif >= 0 && fPreDem >= 0 && (fPreDif-fPreDem) >= 0 {
	//			this.state = STATE_SELL_OUT
	//			this.BindFlag(FLAG_SELL_IMMEDIATE)
	//			log.Println("OnWaitSell MACD CHANGE < 0 SELL")
	//			log.Printf("fDif=%f, fDem = %f, fPreDif=%f, fPreDem=%f\n", fDif, fDem, fPreDif, fPreDem)
	//			return
	//		}
	//	}
}

func (this *RsiBuy) OnSellOut() {
	// 15%, 25%, 35%, 25% 每6分钟
	var sellrate = [4]float32{0.15, 0.30, 0.35, 0.20}
	for idx := int(0); idx < len(sellrate); idx++ {
		sum := float32(0)
		for innerIdx := idx; innerIdx < len(sellrate); innerIdx++ {
			sum += sellrate[innerIdx]
		}

		sellrate[idx] = sellrate[idx] / sum
	}

	// 等待上一单成交
	var bCancel bool = false
	if !this.WaitSellOrderDone(bCancel) {
		return
	}

	if bCancel {
		this.userInfo = coinapi.GetUserInfo()
		if this.userInfo == nil || this.userInfo.Result == false {
			return
		}
	}

	sellEnd := func() {
		this.state = STATE_WAIT_BUY
		//如果是亏损卖出的，设置冷却时间
		if this.HasFlag(FLAG_SELL_IMMEDIATE) {
			this.coolTime = GetNowTime() + TRADE_COOL_TIME
			log.Println("last sell is stoploss, set trade cooltime")
		}
		this.flag = 0
		this.buyOrders = make([]OrderData, 0)
		this.sellOrders = make([]OrderData, 0)
		this.ClearAllOrder()
	}

	orderLen := len(this.sellOrders)
	switch orderLen {
	case 0:
		coincount := this.userInfo.Info.Funds.Free.GetLtc()
		if !this.HasFlag(FLAG_SELL_IMMEDIATE) {
			coincount = coincount * sellrate[orderLen]
		}
		orderData := this.Sell(coinapi.LTC, coincount)
		if orderData.orderId != 0 {
			this.sellOrders = append(this.sellOrders, orderData)
			this.SaveOrder(orderData)
		} else {
			log.Printf("OnSellOut case 0, orderid is 0, coincount=%f", coincount)
		}
	case 1, 2, 3:
		if this.HasFlag(FLAG_SELL_IMMEDIATE) {
			sellEnd()
		} else {
			nowTime := GetNowTime()
			preOrder := this.sellOrders[orderLen-1]
			if nowTime > preOrder.orderTime+PER_ORDER_COOL_TIME {
				coincount := this.userInfo.Info.Funds.Free.GetLtc() * sellrate[orderLen]
				orderData := this.Sell(coinapi.LTC, coincount)
				if orderData.orderId != 0 {
					this.sellOrders = append(this.sellOrders, orderData)
					this.SaveOrder(orderData)
				} else {
					log.Printf("OnSellOut case 0, orderid is 0, coincount=%f", coincount)
				}
			}
		}
	case 4:
		sellEnd()
	default:
		log.Printf("error: OnSellOut len(sellOrders)=%d", orderLen)
	}
}

func (this *RsiBuy) Buy(symbol string, rmbcount float32) OrderData {
	curPrice := this.tickInfo.Tick.GetLast()
	orderPrice := curPrice + 0.01
	buycount := float32(int32((rmbcount/(orderPrice))*float32(100))) / float32(100)
	log.Printf("buy price=%f, count=%f", orderPrice, buycount)
	if buycount > coinapi.MIN_TRADE_LTC {
		orderId := coinapi.DoTrade(symbol, coinapi.BUY, orderPrice, buycount)
		if orderId != 0 {
			var data OrderData
			data.symbol = symbol
			data.orderId = orderId
			data.price = orderPrice
			data.orderType = coinapi.BUY
			data.orderTime = GetNowTime()
			return data
		}
	}

	return OrderData{}
}

func (this *RsiBuy) Sell(symbol string, coincount float32) OrderData {
	curPrice := this.tickInfo.Tick.GetLast()
	orderPrice := curPrice - 0.01

	log.Printf("sell price=%f, diffprice=%f, count=%f", orderPrice, orderPrice-this.GetAvgPosition(), coincount)
	if coincount > coinapi.MIN_TRADE_LTC {
		orderId := coinapi.DoTrade(symbol, coinapi.SELL, orderPrice, coincount)
		if orderId != 0 {
			var data OrderData
			data.symbol = symbol
			data.orderId = orderId
			data.price = orderPrice
			data.orderType = coinapi.SELL
			data.orderTime = GetNowTime()
			return data
		}
	}

	return OrderData{}
}

func (this *RsiBuy) WaitBuyOrderDone(bCancel bool) bool {
	orderLen := len(this.buyOrders)
	if orderLen == 0 {
		return true
	}

	orderData := this.buyOrders[orderLen-1]
	resp := coinapi.GetOrderInfo(orderData.symbol, int32(orderData.orderId))
	if resp == nil || len(resp.Orders) == 0 || !resp.Result {
		log.Printf("WaitBuyOrderDone order resp error, orderid=%d", orderData.orderId)
		return false
	}

	if resp.Orders[0].Status == 2 {
		this.buyOrders[orderLen-1].count = resp.Orders[0].Dealamount
		this.buyOrders[orderLen-1].price = resp.Orders[0].AvgPrice
		this.SaveOrder(this.buyOrders[orderLen-1])
		return true
	} else {
		nowTime := GetNowTime()
		if nowTime > orderData.orderTime+ORDER_DELAY_TIME_MAX {
			if resp.Orders[0].Status == 1 {
				//部分成交，记录一下，以后处理
				log.Printf("warning: part deal on buy, orderid:%d, count=%f\n", orderData.orderId, resp.Orders[0].Dealamount)
			}

			//超时，取消订单
			coinapi.CancelOrder(orderData.symbol, orderData.orderId)
			log.Printf("buyorder delaytime=%d(ms), orderid=%d, canceled\n", nowTime, orderData.orderId)

			this.buyOrders = append(this.buyOrders[:orderLen-1])
			this.DeleteOrder(orderData)

			bCancel = true
			return true
		}
		return false
	}
}

func (this *RsiBuy) WaitSellOrderDone(bCancel bool) bool {
	orderLen := len(this.sellOrders)
	if orderLen == 0 {
		return true
	}

	orderData := this.sellOrders[orderLen-1]
	resp := coinapi.GetOrderInfo(orderData.symbol, int32(orderData.orderId))
	if resp == nil || len(resp.Orders) == 0 || !resp.Result {
		log.Printf("WaitSellOrderDone order resp error, orderid=%d", orderData.orderId)
		return false
	}

	if resp.Orders[0].Status == 2 {
		this.sellOrders[orderLen-1].count = resp.Orders[0].Dealamount
		this.sellOrders[orderLen-1].price = resp.Orders[0].AvgPrice
		this.SaveOrder(this.sellOrders[orderLen-1])
		return true
	} else {
		nowTime := GetNowTime()
		if nowTime > orderData.orderTime+ORDER_DELAY_TIME_MAX {
			if resp.Orders[0].Status == 1 {
				//部分成交，记录一下，以后处理
				log.Printf("warning: part deal on sell, orderid:%d, count=%f\n", orderData.orderId, resp.Orders[0].Dealamount)
			}
			//超时，取消订单
			coinapi.CancelOrder(orderData.symbol, orderData.orderId)
			log.Printf("sellorder delaytime=%d(ms), orderid=%d, canceled\n", nowTime, orderData.orderId)

			this.sellOrders = append(this.sellOrders[:orderLen-1])
			this.DeleteOrder(orderData)

			bCancel = true
			return true
		}
		return false
	}
}

func (this *RsiBuy) GetRsiNow() float32 {
	if this.kline == nil {
		return 0
	}
	rsi := coinapi.GetRSI((*this.kline), RSI_LEVEL)
	if 0 == rsi {
		log.Println("GetRsiNow() RSI is zero")
		return 0
	}
	return rsi
}

func (this *RsiBuy) GetPreRsi(lowest bool) float32 {
	if this.kline == nil {
		return 0
	}

	closed := (*this.kline)[1].Close
	if lowest {
		(*this.kline)[1].Close = (*this.kline)[1].Low
	}

	rsi := coinapi.GetRSI((*this.kline)[1:], RSI_LEVEL)
	(*this.kline)[1].Close = closed
	if rsi == 0 {
		log.Println("GetPreRsi RSI is zero")
		return 0
	}
	return rsi
}

func (this *RsiBuy) GetKDJ_DVal() float32 {
	if this.kline == nil {
		return 0
	}

	_, d := coinapi.GetKDJ((*this.kline), 15, 5, 5)
	if d == 0 {
		log.Println("GetKDJ_DVal d is zero")
		return 0
	}
	return d
}

func (this *RsiBuy) GetAvgPosition() float32 {
	money := float32(0)
	count := float32(0)
	for _, val := range this.buyOrders {
		money += val.count * val.price
		count += val.count
	}

	if count == 0 {
		return 0
	}

	return money / count
}

func (this *RsiBuy) GetPositionTotalCny() float32 {
	money := float32(0)

	for _, val := range this.buyOrders {
		money += val.count * val.price
	}

	return money
}

func (this *RsiBuy) GetPositionCoinCount() float32 {
	count := float32(0)
	for _, val := range this.buyOrders {
		count += val.count
	}

	return count
}

func (this *RsiBuy) FetchKline() {
	this.kline = nil
	kline := coinapi.GetKline(coinapi.LTC, "15min", coinapi.MACD_KLINE_MAX, 0)
	if kline == nil || len(*kline) < coinapi.MACD_KLINE_MAX {
		return
	}

	// 按照时间降序
	for i := 0; i < len(*kline)/2; i++ {
		(*kline)[i], (*kline)[len(*kline)-i-1] = (*kline)[len(*kline)-i-1], (*kline)[i]
	}

	this.kline = kline
}

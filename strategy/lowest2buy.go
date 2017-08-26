package strategy

import (
	"GoCoin/coinapi"
	"fmt"
	"log"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

/*
3个半小时内，市场第2次达到次低点的时候买进。
首次止损点设定在买点下方0.6%,
买点上方0.6%后，止损点设定在最高点下方0.2%
*/

const (
	CUT_RATE     = 0.006
	WIN_RATE     = 0.006
	WIN_CUT_RATE = 0.002
)

type Lowest2buy struct {
	userInfo *coinapi.RespUserInfo
	order    OrderRecord
}

func (this *Lowest2buy) Run() {
	// 1秒执行一次
	t1 := time.NewTimer(time.Second)

	for {
		select {
		case <-t1.C:
			this.DoStrategy(t1)
		}
	}
}

func (this *Lowest2buy) DoStrategy(t *time.Timer) {
	defer t.Reset(1 * time.Second)

	//获取用户信息
	this.userInfo = coinapi.GetUserInfo()
	if this.userInfo == nil {
		return
	}
	//fmt.Println(this.userInfo)
	//获取市场最新数据
	tick := coinapi.GetTicker(coinapi.LTC)
	if tick == nil {
		return
	}

	if this.HasPosition() {
		if this.order.orderId == 0 {
			this.order.LoadOrder()
			if this.order.orderId == 0 {
				return
			}
		}

		//如果有持仓，获取仓位价格
		position := this.GetPosition(this.order.orderId)
		if position == 0 {
			return
		}
		high := coinapi.GetHighestPrice(coinapi.LTC, this.order.orderTime)
		//获取止损价格
		cutPrice := this.GetCutPrice(position, high)
		curPrice := tick.Tick.GetLast()
		if curPrice == 0 {
			return
		}

		//		fmt.Printf("%v, position:%f, high:%f, cut:%f, current:%f\n", this.order, position,
		//			high, cutPrice, curPrice)
		if curPrice < cutPrice {
			//卖出
			log.Printf("sell: price=%f, high=%f, cut=%f, positin=%f\n", curPrice, high, cutPrice, position)
			coinapi.DoTrade(coinapi.LTC, coinapi.SELL_MARKET, 0, this.userInfo.Info.Funds.Free.GetLtc())
		}

	} else {
		this.order.Clear()
		//如果没有持仓，判断是否可以建仓
		//获取k线信息
		lowPrice := coinapi.GetNLowestPrice(coinapi.LTC, 2, "30min", 7)
		if lowPrice == nil || len(lowPrice) < 2 {
			return
		}
		//计算目标建仓价
		curPrice := tick.Tick.GetLast()
		if curPrice == 0 {
			return
		}

		//1.这里防止卖1价格过高,导致频繁买卖
		//2.防止当前价格过低，不能低于此前的最低价
		sell1Price := tick.Tick.GetSell()
		//fmt.Printf("curprice:%f, lowprice : %v, sell1price : %v\n", curPrice, lowPrice, sell1Price)
		if curPrice < lowPrice[1] &&
			sell1Price*(1-CUT_RATE/2) < curPrice &&
			curPrice > lowPrice[0] {
			orderId := coinapi.DoTrade(coinapi.LTC, coinapi.BUY_MARKET, curPrice, this.userInfo.Info.Funds.Free.GetLtc())
			if orderId != 0 {
				log.Printf("Buy %f, lowprice:%v, sell1price : %v\n", curPrice, lowPrice, sell1Price)
				rows, err := coinapi.GetDB().Query(fmt.Sprintf("INSERT INTO order_data(coin_type,order_id,order_time) VALUES('%s', %d, NOW())",
					coinapi.LTC, orderId))
				if err != nil {
					log.Println(err)
				}
				defer rows.Close()
			} else {
				log.Printf("OrderId is 0\n")
			}
		}
	}

}

func (this *Lowest2buy) GetPosition(orderId uint32) float32 {
	orders := coinapi.GetOrderInfo(coinapi.LTC, int32(orderId))
	if orders != nil && len(orders.Orders) != 0 {
		return orders.Orders[0].AvgPrice
	}
	return 0
}

func (this *Lowest2buy) HasPosition() bool {
	return this.userInfo.Info.Funds.Free.GetLtc() > coinapi.MIN_TRADE_LTC
}

func (this *Lowest2buy) GetCutPrice(position float32, high float32) float32 {
	if high > position*(1+WIN_RATE) {
		return high * (1 - WIN_CUT_RATE)
	} else {
		return position * (1 - CUT_RATE)
	}
}

package strategy

import (
	"GoCoin/coinapi"
	"fmt"
	"log"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

type IntervalBuy struct {
	userInfo *coinapi.RespUserInfo
	order    OrderRecord
}

func (this *IntervalBuy) Init() {

}

func (this *IntervalBuy) Run() {
	this.Init()
	// 1秒执行一次
	t1 := time.NewTimer(time.Second)

	for {
		select {
		case <-t1.C:
			this.DoStrategy(t1)
		}
	}
}

func (this *IntervalBuy) DoStrategy(t *time.Timer) {
	defer t.Reset(5 * time.Second)

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
			if this.order.orderId == 0 || this.order.orderTime == 0 {
				return
			}
		}

		orders := coinapi.GetOrderInfo(coinapi.LTC, int32(this.order.orderId))
		if orders == nil || len(orders.Orders) == 0 {
			log.Printf("orderid=%d, not received php response\n", this.order.orderId)
			return
		}

		// 如果订单未完全成交，等待完全成交（todo:根据现有价格，再下一次买单）
		if orders.Orders[0].Status != 2 {
			//5分钟还未完全成交，撤单

			return
		}

		if orders.Orders[0].Type == coinapi.SELL {
			log.Printf("type=sell status=2 but has position, orderid=%d", this.order.orderId)
			return
		}

		// 如果价格在最低价以下,卖出
		position := orders.Orders[0].AvgPrice
		if position == 0 {
			log.Println("position is 0")
			return
		}
		diff := this.order.highPrice - this.order.lowPrice
		curPrice := tick.Tick.GetLast()
		if curPrice < this.order.lowPrice-0.1*diff {
			log.Printf("sell: price=%f, low=%v, high=%v\n", curPrice, this.order.lowPrice, this.order.highPrice)
			orderId := coinapi.DoTrade(coinapi.LTC, coinapi.SELL, curPrice-0.01, this.userInfo.Info.Funds.Free.GetLtc())
			if orderId != 0 {

				rows, err := coinapi.GetDB().Query(fmt.Sprintf("UPDATE order_data SET order_id=%d,order_type='%s',order_time_sell=NOW() WHERE order_id=%d",
					orderId, coinapi.SELL, this.order.orderId))
				if err != nil {
					log.Println(err)
				}
				defer rows.Close()

				this.order.orderId = orderId
				this.order.orderType = coinapi.SELL
				this.order.orderTimeSell = int64(time.Now().UnixNano() / int64(time.Millisecond))
			} else {
				log.Println("orderid is 0")
			}
		} else {
			highestPrice := coinapi.GetHighestPrice(coinapi.LTC, this.order.orderTime)
			if highestPrice == 0 {
				//log.Println("highest price is 0")
				return
			}

			bSell := false
			maxProfit := (highestPrice - position) / position
			curProfit := (curPrice - position) / position
			if maxProfit < 0.02 {

			} else if maxProfit < 0.03 {
				if curProfit < maxProfit*0.25 {
					bSell = true
				}
			} else if maxProfit < 0.05 {
				if curProfit < maxProfit*0.6 {
					bSell = true
				}
			} else if maxProfit < 0.1 {
				if curProfit < maxProfit*0.8 {
					bSell = true
				}
			} else if maxProfit < 0.2 {
				if curProfit < maxProfit*0.86 {
					bSell = true
				}
			} else if maxProfit < 0.4 {
				if curProfit < maxProfit*0.91 {
					bSell = true
				}
			} else {
				if curProfit < maxProfit*0.96 {
					bSell = true
				}
			}

			if bSell {
				log.Printf("sell: price=%f, low=%v, high=%v, maxProfit=%f, curProfit=%f\n",
					curPrice, this.order.lowPrice, this.order.highPrice, maxProfit, curProfit)
				orderId := coinapi.DoTrade(coinapi.LTC, coinapi.SELL, curPrice-0.01, this.userInfo.Info.Funds.Free.GetLtc())
				if orderId != 0 {
					rows, err := coinapi.GetDB().Query(fmt.Sprintf("UPDATE order_data SET order_id=%d,order_type='%s',order_time_sell=NOW() WHERE order_id=%d",
						orderId, coinapi.SELL, this.order.orderId))
					if err != nil {
						log.Println(err)
					}
					defer rows.Close()

					this.order.orderId = orderId
					this.order.orderType = coinapi.SELL
					this.order.orderTimeSell = int64(time.Now().UnixNano() / int64(time.Millisecond))
				} else {
					log.Println("orderid is 0")
				}
			}
		}
	} else {
		if this.order.orderId > 0 {
			if this.order.orderType == coinapi.SELL {
				this.order.Clear()
			} else if this.order.orderType == coinapi.BUY {
				//todo: 如果是买单，一直未成交也需要处理
				nowTime := int64(time.Now().UnixNano() / int64(time.Millisecond))
				if nowTime-this.order.orderTime > 180*1000 {
					coinapi.CancelOrder(coinapi.LTC, this.order.orderId)
					log.Printf("order delaytime=%d, orderid=%d, canceled\n",
						nowTime-this.order.orderTime, this.order.orderId)
					this.order.Clear()
				}
				return
			}

		}

		since := int64(time.Now().UnixNano()/int64(time.Millisecond)) - int64(time.Hour/time.Millisecond)*36
		low, high := coinapi.GetLowHighPrice(coinapi.LTC, since)
		if low.Low == 0 || high.High == 0 || low.Low > high.High {
			log.Printf("low or high price error, low=%v, high=%v\n", low, high)
			return
		}

		curPrice := tick.Tick.GetLast()
		signal := this.GetSignal(curPrice, low, high)
		if signal == 0 {
			return
		}

		cny := this.userInfo.Info.Funds.Free.GetCny()
		buycount := float32(int32((cny/curPrice)*float32(10))) / float32(10)
		if buycount > 0.2 {
			orderId := coinapi.DoTrade(coinapi.LTC, coinapi.BUY, curPrice, buycount)
			if orderId != 0 {
				log.Printf("Buy %f, lowprice:%v, highprice : %v\n", curPrice, low, high)
				rows, err := coinapi.GetDB().Query(fmt.Sprintf("INSERT INTO order_data(coin_type,order_id,order_type,order_time,order_time_sell,low_price,high_price) VALUES('%s', %d, '%s', NOW(), NOW(), %f, %f)",
					coinapi.LTC, orderId, coinapi.BUY, low.Low, high.High))
				if err != nil {
					log.Println(err)
				}
				defer rows.Close()

				this.order.orderId = orderId
				this.order.orderType = coinapi.BUY
				this.order.lowPrice = low.Low
				this.order.highPrice = high.High
				this.order.orderTime = int64(time.Now().UnixNano() / int64(time.Millisecond))
			} else {
				log.Printf("OrderId is 0\n")
			}
		}

	}
}

func (this *IntervalBuy) GetSignal(curPrice float32, low coinapi.RespKline, high coinapi.RespKline) int {
	diff := high.High - low.Low
	if diff < 0.035*curPrice {
		return 0
	}

	nowTime := int64(time.Now().UnixNano() / int64(time.Millisecond))
	//最低价不能是8小时内产生的
	if low.Date > uint64(nowTime-int64(time.Hour/time.Millisecond)*6) {
		return 0
	}

	if curPrice < (low.Low + 0.15*diff) {
		return coinapi.SIGNAL_BUY
	}

	return 0
}

func (this *IntervalBuy) HasPosition() bool {
	return this.userInfo.Info.Funds.Free.GetLtc() > coinapi.MIN_TRADE_LTC
}

package strategy

import (
	"GoCoin/coinapi"
	"fmt"
	"log"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

var (
	RSI_LOW_LINE   = float32(21)
	RSI_HIGH_LINE  = float32(79)
	RSI_LOCK_RANGE = float32(2)
)

type MacdRsiBuy struct {
	userInfo   *coinapi.RespUserInfo
	order      OrderRecord
	coolTime   int64
	lowestRSI  float32
	highestRSI float32
}

func (this *MacdRsiBuy) Init() {
	this.order.LoadOrder()
	if this.order.orderId != 0 {
		this.coolTime = this.order.orderTime + 1800*1000
	}

	this.lowestRSI = 0
	this.highestRSI = 0
}

func (this *MacdRsiBuy) Run() {
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

func (this *MacdRsiBuy) HasPosition() bool {
	return this.userInfo.Info.Funds.Free.GetLtc() > coinapi.MIN_TRADE_LTC
}

func (this *MacdRsiBuy) DoStrategy(t *time.Timer) {
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

		position := orders.Orders[0].AvgPrice
		if position == 0 {
			log.Println("position is 0")
			return
		}
		curPrice := tick.Tick.GetLast()
		var bSell = false
		var rsi = float32(0)
		//如果价格在仓位以下2%,卖出
		if curPrice <= position*0.98 {
			bSell = true
		} else {
			//如果价格正常，取rsi值
			//获取当前RSI值
			kline := coinapi.GetKline(coinapi.LTC, "15min", coinapi.MACD_KLINE_MAX, 0)
			if kline == nil || len(*kline) < coinapi.MACD_KLINE_MAX {
				return
			}

			// 按照时间降序
			for i := 0; i < len(*kline)/2; i++ {
				(*kline)[i], (*kline)[len(*kline)-i-1] = (*kline)[len(*kline)-i-1], (*kline)[i]
			}
			rsi = coinapi.GetRSI((*kline), coinapi.N4)
			if 0 == rsi {
				log.Println("RSI is zero")
				return
			}

			//确认是否是超买行情
			if rsi >= RSI_HIGH_LINE {
				if rsi > this.highestRSI {
					this.highestRSI = rsi
				}
			}

			if this.highestRSI == 0 {
				return
			}

			if this.highestRSI-RSI_LOCK_RANGE >= rsi {
				//卖出
				bSell = true
			} else {
				//等待行情见顶
				return
			}
		}

		if bSell {
			log.Printf("sell: price=%f, rsi=%f, highestrsi=%f\n",
				curPrice, rsi, this.highestRSI)
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
	} else {
		if this.order.orderId > 0 {
			if this.order.orderType == coinapi.SELL {
				this.coolTime = this.order.orderTime + 1800*1000 //重置冷却时间
				this.order.Clear()
			} else if this.order.orderType == coinapi.BUY {
				//todo: 如果是买单，一直未成交也需要处理
				nowTime := int64(time.Now().UnixNano() / int64(time.Millisecond))
				if nowTime-this.order.orderTime > 30*1000 {
					coinapi.CancelOrder(coinapi.LTC, this.order.orderId)
					log.Printf("order delaytime=%d(ms), orderid=%d, canceled\n",
						nowTime-this.order.orderTime, this.order.orderId)
					this.order.Clear()
				}
				return
			}

		}

		if this.coolTime > time.Now().Unix()*1000 {
			//距离上次交易不足半小时
			return
		}

		curPrice := tick.Tick.GetLast()

		//获取当前RSI值
		kline := coinapi.GetKline(coinapi.LTC, "15min", coinapi.MACD_KLINE_MAX, 0)
		if kline == nil || len(*kline) < coinapi.MACD_KLINE_MAX {
			return
		}

		// 按照时间降序
		for i := 0; i < len(*kline)/2; i++ {
			(*kline)[i], (*kline)[len(*kline)-i-1] = (*kline)[len(*kline)-i-1], (*kline)[i]
		}
		rsi := coinapi.GetRSI((*kline), coinapi.N4)
		if 0 == rsi {
			log.Println("RSI is zero")
			return
		}
		//fmt.Printf("curRSI=%f\n", rsi)

		//当前RSI小于超卖线
		if rsi <= RSI_LOW_LINE {
			if rsi < this.lowestRSI || this.lowestRSI == 0 {
				this.lowestRSI = rsi
			}
		}

		//不处于超卖行情中
		if this.lowestRSI == 0 {
			return
		}

		if this.lowestRSI+RSI_LOCK_RANGE <= rsi {
			//止住跌势，可以买入
			cny := this.userInfo.Info.Funds.Free.GetCny()
			buycount := float32(int32((cny/curPrice)*float32(10))) / float32(10)
			if buycount > 0.2 {
				orderId := coinapi.DoTrade(coinapi.LTC, coinapi.BUY, curPrice, buycount)
				if orderId != 0 {
					log.Printf("Buy %f, rsi:%v, lowestrsi:%v\n", curPrice, rsi, this.lowestRSI)
					rows, err := coinapi.GetDB().Query(fmt.Sprintf("INSERT INTO order_data(coin_type,order_id,order_type,order_time,order_time_sell,low_price,high_price) VALUES('%s', %d, '%s', NOW(), NOW(), 0, 0)",
						coinapi.LTC, orderId, coinapi.BUY))
					if err != nil {
						log.Println(err)
					}
					defer rows.Close()

					this.order.orderId = orderId
					this.order.orderType = coinapi.BUY
					this.order.lowPrice = 0
					this.order.highPrice = 0
					this.order.orderTime = int64(time.Now().UnixNano() / int64(time.Millisecond))
				} else {
					log.Printf("OrderId is 0\n")
				}
			}

			//不论是否成交，这里已经完成一次交易逻辑
			this.lowestRSI = 0
			this.highestRSI = 0
		} else {
			//还在下跌趋势中，持续观望
			return
		}
	}
}

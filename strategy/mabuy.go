package strategy

import (
	"GoCoin/coinapi"
	//"fmt"
	//"log"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

/*
Kline取15min档
每2秒计算一次MA5和MA20
！参考著名的葛兰碧法则！
当MA5从下方穿过MA20时，为买入信号

*/

type Mabuy struct {
	userInfo     *coinapi.RespUserInfo
	buyOrderList []uint32
}

func (this *Mabuy) Init() {
	this.buyOrderList = make([]uint32, 0)
}

func (this *Mabuy) Run() {
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

func (this *Mabuy) DoStrategy(t *time.Timer) {
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

	// 确定当前是应该买入还是卖出

	// 如果当前是买入阶段，分批买入

	// 如果当前是卖出阶段，逢高卖出

	//最新成交价
	curPrice := tick.Tick.GetLast()
	if curPrice == 0 {
		return
	}

	//	// 如果MA5从下方向上穿过MA20，买入
	//	if ma5list[0] < ma20list[0] && ma5list[1] >= ma20list[1] {
	//		cny := this.userInfo.Info.Funds.Free.GetCny()
	//		buycount := cny / curPrice
	//		if buycount > 0.3 {
	//			//计算买入价格
	//			otherbuyPrice := tick.Tick.GetBuy() //买1价格
	//			//如果买1小于MA5，选MA5，否则选择MA5作为买入价格
	//			buyPrice := float32(0)
	//			if otherbuyPrice < ma5list[1] {
	//				buyPrice = otherbuyPrice
	//			} else {
	//				buyPrice = ma5list[1]
	//			}

	//			buycount = float32(int32((cny/buyPrice)*float32(10))) / float32(10)
	//			orderId := coinapi.DoTrade(coinapi.LTC, coinapi.BUY, buyPrice, buycount)
	//			this.buyOrderList = append(this.buyOrderList, orderId)
	//		}
	//	}
}

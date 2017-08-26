package strategy

import (
	"GoCoin/coinapi"
	"fmt"
	"log"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

func GetNowTime() int64 {
	return int64(time.Now().UnixNano() / int64(time.Millisecond))
}

type OrderRecord struct {
	symbol        string
	orderId       uint32
	orderType     string
	orderTime     int64
	orderTimeSell int64
	lowPrice      float32
	highPrice     float32
}

const (
	STATE_NONE       = 0
	STATE_WAIT_RSI_1 = 1 // 等待RSI(4) >= 85 或者 RSI(6) >= 80
	STATE_WAIT_RSI_2 = 2 // 等待RSI(6) <= 20
	STATE_WAIT_RSI_3 = 3 // 等待RSI(4) > 80
)

func (this *OrderRecord) Clear() {

	rows, err := coinapi.GetDB().Query(fmt.Sprintf("DELETE FROM order_data WHERE coin_type='%s' AND order_id=%d", coinapi.LTC, this.orderId))
	if err != nil {
		log.Println(err)
		return
	}

	defer rows.Close()

	this.symbol = ""
	this.orderId = 0
	this.orderTime = 0
	this.orderTimeSell = 0
	this.orderType = ""
	this.lowPrice = 0
	this.highPrice = 0
}

func (this *OrderRecord) LoadOrder() {
	this.symbol = coinapi.LTC
	rows, err := coinapi.GetDB().Query(fmt.Sprintf("SELECT order_id,order_type,order_time,order_time_sell, low_price,high_price FROM order_data WHERE coin_type='%s'", this.symbol))
	if err != nil {
		log.Println(err)
		return
	}

	defer rows.Close()

	for rows.Next() {
		var orderId uint32
		var orderTime string
		var orderTimeSell string
		var orderType string
		var lowprice float32
		var highprice float32
		err := rows.Scan(&orderId, &orderType, &orderTime, &orderTimeSell, &lowprice, &highprice)
		if err != nil {
			log.Println(err)
			continue
		}
		this.orderId = orderId
		this.orderType = orderType
		the_time, _ := time.ParseInLocation("2006-01-02 15:04:05", orderTime, time.Local)
		this.orderTime = int64(the_time.Unix()) * 1000 //单位是毫秒
		the_timeSell, _ := time.ParseInLocation("2006-01-02 15:04:05", orderTimeSell, time.Local)
		this.orderTimeSell = int64(the_timeSell.Unix()) * 1000 //单位是毫秒
		this.lowPrice = lowprice
		this.highPrice = highprice
	}
	return
}

package main

import (
	"GoCoin/strategy"
)

func main() {
	Run()
}
func Run() {
	var worker strategy.RsiBuy
	worker.Run()
}

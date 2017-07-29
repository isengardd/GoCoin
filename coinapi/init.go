package coinapi

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
)

// 字段名一定要大写
type CoinConfig struct {
	ApiKey    string `json:"api_key"`
	SecretKey string `json:"secret_key"`
}

var config CoinConfig

func init() {

	content, error := ioutil.ReadFile("config.json")
	if error != nil {
		fmt.Printf("%v\n", error)
	} else {
		error = json.Unmarshal(content, &config)
		if error != nil {
			fmt.Printf("Unmarshal %v\n", error)
		}
	}
}

func GetConfig() *CoinConfig {
	return &config
}

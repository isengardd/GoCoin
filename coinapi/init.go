package coinapi

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io/ioutil"

	_ "github.com/go-sql-driver/mysql"
)

// 字段名一定要大写
type CoinConfig struct {
	ApiKey    string `json:"api_key"`
	SecretKey string `json:"secret_key"`
	Sqluser   string `json:"sql_user"`
	Sqlpwd    string `json:"sql_pwd"`
	Sqlport   uint32 `json:"sql_port"`
}

var config CoinConfig
var sqlconnect *sql.DB = nil

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

	con, err := sql.Open("mysql", fmt.Sprintf("%s:%s@tcp(localhost:%d)/ok_coin?charset=utf8", config.Sqluser, config.Sqlpwd, config.Sqlport))
	if err != nil {
		fmt.Println(err)
	} else {
		sqlconnect = con
	}

}

func GetConfig() *CoinConfig {
	return &config
}

func GetDB() *sql.DB {
	return sqlconnect
}

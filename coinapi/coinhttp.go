package coinapi

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	//"net/url"
	"sort"
	"strings"
)

type CoinHttp struct {
	params   []string
	msg      string
	suburl   string
	jsondata interface{}
	//postData url.Values
}

func (this *CoinHttp) Init(msg string, suburl string, jsondata interface{}) {
	this.params = make([]string, 0)
	this.msg = msg
	this.suburl = suburl
	this.jsondata = jsondata
}

func (this *CoinHttp) Msg() string {
	return fmt.Sprintf("msg=%s params=%v ", this.msg, this.params)
}

func (this *CoinHttp) AddParam(param string) {
	this.params = append(this.params, param)
}

func (this *CoinHttp) Get() interface{} {
	sort.Strings(this.params)
	joinParams := strings.Join(this.params, "&")
	sign := MakeSign(joinParams)
	urlParams := this.params[:]
	urlParams = append(urlParams, MakeParam("sign", sign))
	urlStr := strings.Join(urlParams, "&")
	reqUri := fmt.Sprintf("%s/%s?%s", ROOT_URL, this.suburl, urlStr)
	req, err := http.NewRequest("GET", reqUri, nil)
	if err != nil {
		log.Printf("%s %s\n", this.Msg(), err)
		return nil
	}

	req.Header.Add("contentType", "application/x-www-form-urlencoded")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("%s clientdo err:%s\n", this.Msg(), err)
		return nil
	}

	defer resp.Body.Close()

	if resp.StatusCode == 200 {
		if data, err := ioutil.ReadAll(resp.Body); err == nil {
			//fmt.Printf("%s\n", data)
			err = json.Unmarshal([]byte(data), this.jsondata)
			if err != nil {
				log.Printf("%s Unmarshal error, %s\n", this.Msg(), err)
				return nil
			}
			return this.jsondata
		}
	} else {
		log.Printf("%s statuscode=%d\n", this.Msg(), resp.StatusCode)
		return nil
	}
	return nil
}

func (this *CoinHttp) Post() interface{} {
	sort.Strings(this.params)
	joinParams := strings.Join(this.params, "&")
	sign := MakeSign(joinParams)
	urlParams := this.params[:]
	urlParams = append(urlParams, MakeParam("sign", sign))
	urlStr := strings.Join(urlParams, "&")
	reqUri := fmt.Sprintf("%s/%s", ROOT_URL, this.suburl)
	req, err := http.NewRequest("POST", reqUri, strings.NewReader(urlStr))
	if err != nil {
		log.Printf("%s %s\n", this.Msg(), err)
		return nil
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Add("contentType", "application/x-www-form-urlencoded")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("%s clientdo err:%s\n", this.Msg(), err)
		return nil
	}

	defer resp.Body.Close()

	if resp.StatusCode == 200 {
		if data, err := ioutil.ReadAll(resp.Body); err == nil {
			//fmt.Printf("%s\n", data)
			err = json.Unmarshal([]byte(data), this.jsondata)
			if err != nil {
				log.Printf("%s Unmarshal error, %s\n", this.Msg(), err)
				return nil
			}
			return this.jsondata
		}
	} else {
		log.Printf("%s statuscode=%d\n", this.Msg(), resp.StatusCode)
		return nil
	}
	return nil
}

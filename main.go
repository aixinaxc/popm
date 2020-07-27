package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"reflect"
	"strconv"
	"sync"
	"testmod/base"
	"testmod/redispool"
	"testmod/xrom_mysql"
	"time"
)
var wg sync.WaitGroup
func main()  {

	base.Config()
	//Find()
	wg.Add(1)
	Save()
	wg.Wait()
}

type MetalTitle struct {
	Flag bool `json:"flag"`
	ErrorCode string `json:"error_code"`
	JO_42757 BankMetal `json:"JO_42757"` //工行纸黄金（美元）
	JO_42758 BankMetal `json:"JO_42758"` //工行纸白银（美元）
	JO_42759 BankMetal `json:"JO_42759"` //工行纸铂金（美元）
	JO_42760 BankMetal `json:"JO_42760"` //工行纸黄金（人民币）
	JO_42761 BankMetal `json:"JO_42761"` //工行纸白银（人民币）
	JO_42762 BankMetal `json:"JO_42762"` //工行纸铂金（人民币）
	JO_52643 BankMetal `json:"JO_52643"` //工行纸钯金（美元）
	JO_52644 BankMetal `json:"JO_52644"` //工行纸钯金（人民币）
	JO_62282 BankMetal `json:"JO_62282"` //建行美元铂（钞）
	JO_62283 BankMetal `json:"JO_62283"` //建行美元银（钞）
	JO_62284 BankMetal `json:"JO_62284"` //建行美元银（汇）
	JO_62285 BankMetal `json:"JO_62285"` //建行人民币银
	JO_62286 BankMetal `json:"JO_62286"` //建行AU9999
	JO_62287 BankMetal `json:"JO_62287"` //建行美元铂（汇）
	JO_62288 BankMetal `json:"JO_62288"` //建行美元金（汇）
	JO_62289 BankMetal `json:"JO_62289"` //建行美元金（钞）
	JO_62290 BankMetal `json:"JO_62290"` //建行AU9995
	JO_62291 BankMetal `json:"JO_62291"` //建行人民币铂
}


type BankMetal struct {
	Time int `json:"time"`
	Q1 json.RawMessage `json:"q1"` //开盘价
	Q2 json.RawMessage `json:"q2"` //昨收价
	Q3 json.RawMessage `json:"q3"` //最高价
	Q4 json.RawMessage `json:"q4"` //最低价
	Q80 json.RawMessage `json:"q80"` //涨跌幅
	Q63 json.RawMessage `json:"q63"` //最新价
	Q70 json.RawMessage `json:"q70"` //涨跌价
	Q60 json.RawMessage `json:"q60"`
	Q193 json.RawMessage `json:"q193"`
	Unit string `json:"unit"`
	ShowName string `json:"ShowName"`
	ShowCode string `json:"ShowCode"`
	ObtainCode string `json:"ObtainCode"`
}

func Find()  {
	engine := xrom_mysql.Client()
	bm := make([]BankMetal,0)
	err:= engine.Find(&bm)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(bm)
}

func FindByCode(code string)  {
	engine := xrom_mysql.Client()
	bm := make([]BankMetal,0)
	err:= engine.Where("show_code",code).Find(&bm)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(bm)
}

func Save()  {
	for {
		if time.Now().Weekday() != time.Saturday && time.Now().Weekday() != time.Sunday{
			engine := xrom_mysql.Client()
			resp,e := http.Get("https://api.jijinhao.com/quoteCenter/realTime.htm?codes=JO_42760,JO_42761,JO_42762,JO_52644,JO_42757,JO_42758,JO_42759,JO_52643,JO_62290,JO_62286,JO_62285,JO_62291,JO_62289,JO_62288,JO_62283,JO_62284,JO_62282,JO_62287")
			if e != nil {
				fmt.Println("err:",e)
			}
			body, _ := ioutil.ReadAll(resp.Body)
			resp.Body.Close()
			s := body[17:]
			mT := MetalTitle{}
			json.Unmarshal(s,&mT)
			///fmt.Println(mT)
			n := reflect.ValueOf(mT).NumField()
			for i:=0;i<n;i++{
				name := reflect.TypeOf(mT).Field(i).Name
				t := reflect.ValueOf(mT).Field(i).Type().String()
				if "bool" != t && "string" != t{
					bankm := reflect.ValueOf(mT).Field(i).Interface().(BankMetal)
					time := redispool.RedisGET(bankm.ShowName) //读取保存时间，判断该时间的数据是否保存
					if  strconv.Itoa(bankm.Time) != string(time) {
						redispool.RedisSETString(bankm.ShowName,bankm.Time,0) //记录保存时间
						bankm.Time = bankm.Time / 1000
						bankm.ObtainCode = name
						engine.Insert(bankm)
					}
				}
			}
			engine.Close()
		}
		time.Sleep(15*time.Second)
	}
}
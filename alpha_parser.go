package main

import (
        "fmt"
        "github.com/go-redis/redis"
        "encoding/json"
        //"reflect"
        "strconv"
        "time"
        "net/http"
        "log"
        "io/ioutil"
        "strings"
       )

type Inner2 struct {
    Price string `json:"price"`
    Datetime string `json:"datetime"`
    }

type Inner struct {
    Price string `json:"price"`
    Sym string `json:"symbol"`
    Pcg string `json:"change_percent"`
    }
  
type Data struct {
    TheQuote Inner `json:"global_quote"`
    }

func remove_garbage(s string) string {
    res:=s
    for i := 1;  i<=10; i++ {
                var rep_patt = ""
                if i<10 { rep_patt=fmt.Sprintf("0%d. ", i) } else { rep_patt=fmt.Sprintf("%d. ", i) }
                res = strings.Replace(res, rep_patt, "", 1)
    }
    res =strings.Replace(res, "Global Quote", "global_quote", 1)
    res =strings.Replace(res, "latest trading day", "latest_trading_day", 1)
    res =strings.Replace(res, "previous close", "previous_close", 1)
    res =strings.Replace(res, "change percent", "change_percent", 1)
    res =strings.ReplaceAll(res, " ", "")
    return res
}

func makeRequest(_url string) string {
    fmt.Println("request function:", _url)
    clnt := http.Client{ Timeout: time.Second * 2, } 
    req, err := http.NewRequest(http.MethodGet, _url, nil)
    if err != nil { log.Fatal(err) }
    
    req.Header.Set("User-Agent", "my hdr")
    res, getErr := clnt.Do(req)
    if getErr != nil { log.Fatal(getErr) }

    body, readErr := ioutil.ReadAll(res.Body)
    if readErr != nil { log.Fatal(readErr) }
   
    sbody := string(body)
    return sbody
}

func do2(s string) Data {
  //mJson := `{ "global_quote": {"price": "123","symbol": "MSFT", "change_percent":"-0.9573%"} }`
  mJson:=remove_garbage(s)
  //fmt.Println(mJson)
  
  var result Data	
  json.Unmarshal([]byte(mJson), &result)
  
  fmt.Println(result.TheQuote.Price, result.TheQuote.Sym, result.TheQuote.Pcg)
  fmt.Println("====")
  
  return result
}

var SYM="F"
    
func main() {
    opts := &redis.Options{ Addr: "localhost:6379", Password: "", DB: 1, }
    client := redis.NewClient(opts)
    
    ex, e := client.Exists(SYM).Result()
    if e != nil { fmt.Println("#redis Exists Err:", SYM, e) }
    fmt.Println("#Does the key exist?", SYM, ex)
    
    val, err3 := client.HMGet(SYM, "qts").Result()
    if err3 != nil { fmt.Println("#redis HMGet Err:", SYM, err3) }
    
    //fmt.Println("val", val)
    
    sstr := val[0]
    //fmt.Println("#str reflect:", reflect.TypeOf( sstr ))    
    str, ok := sstr.(string)
    if ok != true { fmt.Println("#string convert Err:", ok)  }
    //fmt.Println("##", str, ok)
    
    var arr []Inner2
    err := json.Unmarshal([]byte(str), &arr)
    if err != nil { fmt.Println("#son.Unmarshal Err:", err)  }
    
    ///////////////////////////
    var url = fmt.Sprintf("https://www.alphavantage.co/query?function=GLOBAL_QUOTE&symbol=%s&apikey=85LL7NCVZ22P4YWV",SYM)
    answer := makeRequest(url)
    qts := do2(answer)
    //fmt.Println()
    
    
    timestamp := strconv.FormatInt(time.Now().UTC().UnixNano(), 10)
  
  
    qqq:=append(arr, Inner2 { Price:qts.TheQuote.Price, Datetime:timestamp })
    //fmt.Println(qqq) 

    ssM2, errM2:=json.Marshal(qqq)
    if errM2 != nil { fmt.Println("# Marshal string[] error:", errM2) }

    v2:=map[string]interface{} {"qts": ssM2}
    
    err22 := client.HMSet(SYM, v2).Err()
    if err22 != nil { fmt.Println("#redis hmset Err", err22) }

}
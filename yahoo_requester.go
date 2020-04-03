package main

import (
    "fmt"
    "github.com/go-redis/redis"
    //"github.com/google/logger"
    "encoding/json"
    "reflect"
    "strconv"
    "time"
    "net/http"
    "net/url"
    "log"
    "io/ioutil"
    "strings"
    //"sort"
    //"flag"
    //"os"
    "github.com/bradfitz/slice"
)

func makeRequest(_url string) string {
    fmt.Println("request function:", _url)
    fmt.Println("=========================")
    
    sbody:=""

    for {
        //creating the proxyURL
        proxyStr := "http://localhost:8118"
        proxyURL, err := url.Parse(proxyStr)
        if err != nil { log.Println(err) }
    
        //adding the proxy settings to the Transport object
        transport := &http.Transport{ Proxy: http.ProxyURL(proxyURL), }
    
        clnt := http.Client{ Timeout: time.Second * 60, Transport: transport, }
        req, errReq := http.NewRequest(http.MethodGet, _url, nil)
        if err != nil { log.Fatal(errReq) }
        
        req.Header.Set("User-Agent", "my hdr")
        res, getErr := clnt.Do(req)
        if getErr != nil { log.Fatal(getErr) }
    
        body, readErr := ioutil.ReadAll(res.Body)
        if readErr != nil { log.Fatal(readErr) }
        
        fmt.Println(reflect.TypeOf(body))
        sbody = string(body)
        if err==nil { break }
    }
    
    return sbody
}

func GetStringInBetween(str string, start string, end string) (result string) {
    s := strings.Index(str, start)
    if s == -1 {
        return
    }
    s += len(start)
    e := strings.Index(str, end)
    if e == -1 {
        return
    }
    return str[s:e]
}

type Quotes_struct struct {
    Close []float64 `json:"close"`
    High []float64 `json:"high"`
    Volume []float64 `json:"volume"`
    Open []float64 `json:"open"`
    Low []float64 `json:"low"`
    }
    
type Dbrec_struct struct {
    mcap float64
    name string
}

type Dbquote_struct struct {
    Price float64 `json:"price"`
    Datetime int64 `json:"datetime"`
    }
    
func get_instruments_list() []Dbrec_struct {
    var recLst = []Dbrec_struct{}
    opts := &redis.Options{ Addr: "localhost:6379", Password: "", DB: 0, }
    client := redis.NewClient(opts)
    
    a, err := client.Keys("*").Result()
    if err != nil { fmt.Println("error Get:", err) }

    //fmt.Println(reflect.TypeOf(a))
    //fmt.Println(a[0])
    for _, sym := range a {
        //fmt.Println(i, sym)
        resp, errHMG := client.HMGet(sym, "cap").Result() //HGETAll
        if errHMG != nil { fmt.Println("error HMGet:", errHMG) }
        
        str, ok := resp[0].(string)
        if ok {
            if strings.Contains( str, "Til" ) || strings.Contains( str, "Bil" ) {
            //if strings.Contains( str, "Til" ) {
                //k:=fmt.Sprintf("%s cap: %s", sym, resp[0])
                //fmt.Println(k)
                
                nr := strings.ReplaceAll(str, "Til", "")
                nr = strings.ReplaceAll(nr, "Bil", "")
                nr = strings.ReplaceAll(nr, " ", "")
                
                flt, errConv := strconv.ParseFloat(nr, 64)
                if strings.Contains( str, "Til" ) { flt=flt*1000 }
                if errConv != nil {
                    fmt.Println("convertion error", nr, errConv)
                } else {
                    //kk:=fmt.Sprintf("%s cap:%s:.....:%f", sym, nr, flt)
                    //fmt.Println(kk)
                    //fmt.Println("===================")
                    
                    var p = Dbrec_struct{flt, sym}
                    recLst=append(recLst, p)
                }
            }
        }
    }
    return recLst
}

func get_quotes_from_server(SYM string) (Quotes_struct, []int64){
    //var SYM="F"
    var url = fmt.Sprintf("https://query1.finance.yahoo.com/v8/finance/chart/%s?region=US&lang=en-US&includePrePost=false&interval=2m&range=1d&corsDomain=finance.yahoo.com&.tsrc=finance", SYM)
    //var url = "https://vmi355260.contaboserver.net" 
    answer := makeRequest(url)
    
    var timeStr=GetStringInBetween(answer, `timestamp":`, `,"indicators":`)
    var qtsStr = GetStringInBetween(answer, `"indicators":{"quote":[`, `]}}],"error":null}}`)
    //fmt.Println(qts)
    
    var qts Quotes_struct
    json.Unmarshal([]byte(qtsStr), &qts)

    var time []int64
    json.Unmarshal([]byte(timeStr), &time)
  
    fmt.Println(len(time))
    fmt.Println(len(qts.Close))
    
    return qts, time
}

func save_array_to_db(SYM string, arr []Dbquote_struct, client *redis.Client) {
    
    ssM2, errM2:=json.Marshal(arr)
    if errM2 != nil { fmt.Println("# Marshal string[] error:", errM2) }
    
    v2:=map[string]interface{} {"qts": ssM2}
    fmt.Println(reflect.TypeOf(v2))
    err22 := client.HMSet(SYM, v2).Err()
    if err22 != nil {
        fmt.Println("#redis hmset Err", err22)
            //logger.Info( fmt.Sprintf("saving %s %s", SYM, qts.TheQuote.Price))
        }
}

func run(SYM string) {
    quotes, time := get_quotes_from_server(strings.ReplaceAll(SYM,".","-"))
    fmt.Println(quotes.Close[len(quotes.Close)-1], time[len(quotes.Close)-1])
    
    web_latest_quote := quotes.Close[len(quotes.Close)-1]
    web_latest_time := time[len(quotes.Close)-1]
    
    opts := &redis.Options{ Addr: "localhost:6379", Password: "", DB: 1, }
    client := redis.NewClient(opts)
    fmt.Println(reflect.TypeOf(client))
    
    ex, e := client.Exists(SYM).Result() //check if KEY exists in db
    if e != nil { fmt.Println("#line 159 redis Exists Err:", SYM, e) }
    fmt.Println("#Does the key exist?", SYM, ex)
    
    val, err3 := client.HMGet(SYM, "qts").Result() 
    if err3 != nil { fmt.Println("# line163 redis HMGet Err:", SYM, err3) }

    sstr := val[0]
    str, ok := sstr.(string)
    if ok != true { fmt.Println("# line 167 string convert Err:", ok, str)  }
    
    var dbarr []Dbquote_struct
    err := json.Unmarshal([]byte(str), &dbarr)
    if err != nil { fmt.Println("#json.Unmarshal Err:", err, str)  }
    
    //dbarr=dbarr[:len(dbarr)-1]
    //fmt.Println(dbarr)
    
    var db_last_price float64
    db_length := len(dbarr)
    fmt.Println("#the length is:", db_length, reflect.TypeOf(db_length))
    
    if db_length == 0 {
        fmt.Println("not assigning last price")
        db_last_price = -1
        } else {
        fmt.Println("assigning last price")
        db_last_price = dbarr[db_length-1].Price
    }
    
    if (db_last_price != web_latest_quote) && (db_length!=0) {
        qqq:=append(dbarr, Dbquote_struct { Price: web_latest_quote, Datetime: web_latest_time })
        save_array_to_db(SYM, qqq, client)
    } else { fmt.Println("values are equal") }
    
    if db_length == 0 {
        var ins_arr []Dbquote_struct
        for i:=0; i<len(time); i++ {
            var el = Dbquote_struct { Price: quotes.Close[i], Datetime: time[i] }
            ins_arr=append(ins_arr, el)
        }
        //fmt.Println(ins_arr)
        save_array_to_db(SYM, ins_arr, client)
    }
}

func main() {
    var instruments = get_instruments_list()
    slice.Sort(instruments[:], func(i, j int) bool { return instruments[i].mcap > instruments[j].mcap })
    
    fmt.Println(instruments[0:10])
    
    for _, rec := range instruments[0:10] {
        fmt.Println(rec.name, rec.mcap)
        run(rec.name)
        //time.Sleep(60 * time.Second)
    }
}
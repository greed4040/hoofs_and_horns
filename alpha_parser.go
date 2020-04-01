package main

import (
        "fmt"
        "github.com/go-redis/redis"
        "github.com/google/logger"
        "encoding/json"
        "reflect"
        "strconv"
        "time"
        "net/http"
        "log"
        "io/ioutil"
        "strings"
        "sort"
        "flag"
        "os"
        "github.com/bradfitz/slice"
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
    sbody:=""

    for {
        clnt := http.Client{ Timeout: time.Second * 60, }
        req, err := http.NewRequest(http.MethodGet, _url, nil)
        if err != nil { log.Fatal(err) }
        
        req.Header.Set("User-Agent", "my hdr")
        res, getErr := clnt.Do(req)
        if getErr != nil { log.Fatal(getErr) }
    
        body, readErr := ioutil.ReadAll(res.Body)
        if readErr != nil { log.Fatal(readErr) }
       
        sbody = string(body)
        if err==nil { break }
    }
    
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

//var SYM="F"

type ByDatetime []Inner2
func (a ByDatetime) Len() int           { return len(a) }
func (a ByDatetime) Less(i, j int) bool { return a[i].Datetime < a[j].Datetime }
func (a ByDatetime) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }

const logPath = "requester.log"
var verbose = flag.Bool("verbose", false, "print info level logs to stdout")

func run(SYM string) {
    lf, errL := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0660)
    if errL != nil { logger.Fatalf("Failed to open log file: %v", errL) }
    defer lf.Close()
    defer logger.Init("LoggerExample", *verbose, true, lf).Close()
    
    opts := &redis.Options{ Addr: "localhost:6379", Password: "", DB: 1, }
    client := redis.NewClient(opts)
    
    ex, e := client.Exists(SYM).Result()
    if e != nil { fmt.Println("#line 90 redis Exists Err:", SYM, e) }
    fmt.Println("#Does the key exist?", SYM, ex)
    
    val, err3 := client.HMGet(SYM, "qts").Result()
    if err3 != nil { fmt.Println("# line94 redis HMGet Err:", SYM, err3) }
    
    fmt.Println("#val:", val)
    
    sstr := val[0]
    fmt.Println("#str reflect:", reflect.TypeOf( sstr ))    
    str, ok := sstr.(string)
    if ok != true { fmt.Println("#string convert Err:", ok, str)  }
    //fmt.Println("##", str, ok)
    
    var arr []Inner2
    err := json.Unmarshal([]byte(str), &arr)
    if err != nil { fmt.Println("#json.Unmarshal Err:", err, str)  }
    
    ///////////////////////////
    //SYM="MSFT"
    var url = fmt.Sprintf("https://www.alphavantage.co/query?function=GLOBAL_QUOTE&symbol=%s&apikey=85LL7NCVZ22P4YWV",SYM)
    //var url = fmt.Sprintf("https://www.alphavantage.co/query?function=GLOBAL_QUOTE&symbol=%s&apikey=demo",SYM)
    
    answer := makeRequest(url)
    qts := do2(answer)
    //fmt.Println(answer)
    fmt.Println(qts)
    
    
    timestamp := strconv.FormatInt(time.Now().UTC().UnixNano(), 10)
  
  
    fmt.Println("#length:", len(arr))
    sort.Sort(ByDatetime(arr))
    
    var db_last_price string
    db_length:=len(arr)
    fmt.Println("#the length is", db_length, reflect.TypeOf(db_length))
    
    if db_length==0 {
        fmt.Println("not assigning")
        db_last_price = "none"
        } else {
        fmt.Println("assigning")
        db_last_price = arr[db_length-1].Price
    }
    
    fmt.Println("#comparision", db_last_price != qts.TheQuote.Price, db_last_price, qts.TheQuote.Price)
    if (db_last_price != qts.TheQuote.Price) || (db_length==0) {
        qqq:=append(arr, Inner2 { Price:qts.TheQuote.Price, Datetime:timestamp })
        //fmt.Println(qqq) 
    
        ssM2, errM2:=json.Marshal(qqq)
        if errM2 != nil { fmt.Println("# Marshal string[] error:", errM2) }
    
        v2:=map[string]interface{} {"qts": ssM2}
        
        err22 := client.HMSet(SYM, v2).Err()
        if err22 != nil {
            fmt.Println("#redis hmset Err", err22)
            logger.Info( fmt.Sprintf("saving %s %s", SYM, qts.TheQuote.Price))
        } 
    } else { fmt.Println("values are equal") }

    logger.Info( fmt.Sprintf("fin %s", SYM) )
}

type MyRec struct {
    mcap float64
    name string
}

func main() {
    var recLst = []MyRec{}
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
                    
                    var p = MyRec{flt, sym}
                    recLst=append(recLst, p)
                }
            }
        }
    }

    fmt.Println(len(recLst), recLst[0])
    
    slice.Sort(recLst[:], func(i, j int) bool { return recLst[i].mcap > recLst[j].mcap })
    
    sortedLst := recLst[0:10]
    fmt.Println(len(sortedLst))
    
    for _, rec := range sortedLst {
        fmt.Println(rec.name, rec.mcap)
        run(rec.name)
        time.Sleep(60 * time.Second)
    }
}
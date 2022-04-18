package websocket

import (
    "fmt"
    "github.com/gorilla/websocket"
    "context"
    "log"
	"encoding/json"
    "time"
    "binact"
    "sync"
    "strconv"
    "analyzer"
/*     "crypto/hmac"
    "crypto/sha256"
    "encoding/hex" */
    //"net/url"
    //"strings"
)

const(
    defaultRetryInterval=time.Second 
    defaultReadTimeout=time.Second * 5
    defaultWriteTimeout=time.Second * 5
    apikey = "T9zIfoHSIsA9KhRIok1uzIcO4MKrRfSKdGQEtVoU5CkkuhIppdvb9tqLCofWL03G"
    secretkey = "22zkt5SByc4twvpsPCWsmta5EPNJW1f0ip0ivruQdi47ecuVDWKeth59pLBZLixZ"
)

type ClientBinance struct{
    mx sync.Mutex   
    prices map[string]float64
    liq LiqCounter
    arb float64
    timeLine *TimeLine
    times []int64
    symbols []string
    matrix [][]float64
    analyzer *analyzer.Analyzer
}

type TimeLine struct{
    startTime int64
    endTime int64
    isCurrent bool
}

func NewClientBinance(a *analyzer.Analyzer)*ClientBinance{
    symbols, _:=a.GetNamesOfAssets()
    newprices:=make(map[string]float64)
    for _, sym:=range a.assets{
        newprices[sym] = float64(0)
    }
    return &ClientBinance{
        prices: newprices,
        symbols:symbols,
        analyzer: a,
        matrix: a.adjacencyMmatrix,
    }
}

func(clBinance *ClientBinance)GetPrice(){
    clBinance.mx.Lock()
    fmt.Println(clBinance.prices)
    clBinance.mx.Unlock()
}

func(clBinance *ClientBinance)GetArb(){
    clBinance.mx.Lock()
    fmt.Println(clBinance.arb)
    clBinance.mx.Unlock()
}

func(clBinance *ClientBinance)GetSymbols()[]string{
    return clBinance.symbols
}

func(clBinance *ClientBinance)GetTimes(){
    newTimes:=make([]float64, len(clBinance.times))
    for i, t:=range clBinance.times{
        newTimes[i] = float64(t)/float64(time.Millisecond)/1000
    }
    fmt.Println(newTimes)
}

func(clBinance *ClientBinance)GoGetTriangularArbitrage(countStart float64, symbols []string){
    go func(){
        clBinance.timeLine=&TimeLine{0, 0, false}
        clBinance.times=[]int64{}
        fee:=1-0.001
        arb:=float64(0)
        val:=float64(0)
        fmt.Println(fee)
        //backPair:=float64(1)
        for{
            clBinance.mx.Lock()
            val = clBinance.prices[symbols[0]]*clBinance.prices[symbols[1]]/clBinance.prices[symbols[2]]*fee*fee*fee
            arb = (val-1)*100
            clBinance.arb = arb
            if clBinance.timeLine.isCurrent==true{
                if arb<=1{
                    clBinance.timeLine.isCurrent = false
                    clBinance.timeLine.endTime = time.Now().UnixNano()
                    clBinance.times = append(clBinance.times, clBinance.timeLine.endTime - clBinance.timeLine.startTime)
                }
            }else{
                if arb>1{
                    fmt.Println("+")
                    clBinance.timeLine.isCurrent = true
                    clBinance.timeLine.startTime = time.Now().UnixNano()
                }
            }
            clBinance.mx.Unlock()
        }
    }()
}

func(clBinance *ClientBinance)GoGetPrices(cxt context.Context){
    for sym, _:=range clBinance.prices{
        go func(symbol string){
            for{
                select{
                case<-cxt.Done():
                    log.Println("cancelled")
                    return
                default:
                    if err:=clBinance.ReceivePrice(cxt, symbol); err!=nil{
                        log.Fatal("Error receive")
                    }
                }
            }
        }(sym)
    }
}


func(clBinance *ClientBinance) ReceivePrice(cxt context.Context, sym string)error{
    url:="wss://stream.binance.com:9443/ws/"+sym+"@aggTrade"
    dialer:=websocket.Dialer{}
    conn,_,err:= dialer.DialContext(cxt, url, nil)
    if err!=nil{
        return err
    }
    if err = conn.SetWriteDeadline(time.Now().Add(defaultWriteTimeout)); err != nil {
        return err
    }
    type h struct{
        Price string `json:"p"`
        Quanity string `json:"q`
    }
    var header h 
    for{
        select{
        case<-cxt.Done():
            return nil
        default:
            mt, data, err := conn.ReadMessage()
            if err != nil {
                return err
            }

            if mt != websocket.TextMessage {
                return fmt.Errorf("unexpected message type %d", mt)
            }
            if err = json.Unmarshal(data, &header); err != nil {
                return err
            }
            clBinance.mx.Lock()
            clBinance.prices[sym],_ = strconv.ParseFloat(header.Price, 64)
            clBinance.mx.Unlock()
        }
    }
}



func(clBinance *ClientBinance)RoundCycle(asset *analyzer.Asset, startBalance float64){
    countTriang:=3
    res:=float64(1)
    var currentAsset *Asset
    for{
        for n:=0;n<countStart;n++{
            if asset
        }
        res = 1
        countTriang +=1 
    }
}





/////////////////////////////////
////LIQUIDITY////////////////////
/////////////////////////////////
type LiqCounter struct{
    indicators []float64
    urls []string
    ch chan float64
    currentPrices []float64
}

func NewLiq(symbols []string)*LiqCounter{
    indicators:=make([]float64, len(symbols))
    urls:=make([]string, len(symbols))
    for i:=0;i<len(symbols);i++{
        urls[i] = "wss://stream.binance.com:9443/ws/"+symbols[i]+"@depth5@100ms"
        indicators[i] = 0
    }
    currentPrices:=make([]float64, len(symbols))
    ch:=make(chan float64)
    return &LiqCounter{
        indicators: indicators,
        urls:urls, 
        ch: ch,
        currentPrices:currentPrices,
    }
}


func StartCalcLiq(symbols []string){
    liq:=NewLiq(symbols)
    packs:=make([]*binact.Pack, len(symbols))
    for i, _:= range packs{
        packs[i] = binact.NewPack(5)
    }

    cxt:=context.Background()

    for idx, url:= range liq.urls{
        go FindingLiq(cxt, packs[idx], url, liq, idx)
    }
}

func FindingLiq(ctx context.Context, pack *binact.Pack, url string, liq *LiqCounter, idx int){
    func(){
        fmt.Println("Start")
        for{
            select{
            case<-ctx.Done():
                log.Println("cancelled")
                return
            default:
                if err:=ReceiveLiq(ctx, pack, url, liq, idx); err!=nil{
                    log.Fatal("Error receive")
                }
            }
        }
    }()
}


func ReceiveLiq(ctx context.Context, pack *binact.Pack, url string, liq *LiqCounter, idx int)error{
    dialer:=websocket.Dialer{}
    conn,_,err:= dialer.DialContext(ctx, url, nil)
    if err!=nil{
        return err
    }
    if err = conn.SetWriteDeadline(time.Now().Add(defaultWriteTimeout)); err != nil {
        return err
    }

    var header struct{
        LastUpdateId int `json:"lastUpdateId"`
        Bids [][]string `json:"bids"`
        Asks [][]string `json:"asks"`
    }
    prob:=float64(-1)
    for{
        select{
        case<-ctx.Done():
            return nil
        default:
            res:=float64(1)
            mt, data, err := conn.ReadMessage()
            if err != nil {
                return err
            }

            if mt != websocket.TextMessage {
                return fmt.Errorf("unexpected message type %d", mt)
            }
            if err = json.Unmarshal(data, &header); err != nil {
                return err
            }
            prob=pack.GetProb(header.Asks)
            if prob!=-1{
                liq.indicators[idx] = prob
            }
            for _, ind:=range liq.indicators{
                res*=ind
            }
            fmt.Println(res)
        }
    }
}
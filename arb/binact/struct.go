package binact

import(
	"time"
	//"fmt"
	"strconv"
)

const(
	limit = 5
    sizeTimes = 10
)

type Sniff struct{
    price string
    qty string
    timeStart float64
}

type Pack struct{
    offers []*Sniff
    timer *Timer
    data []float64
}

type Timer struct{
    gaps[]float64
    currentId int
    probabilityForTime float64
}


func (p *Pack) GetProb(data [][]string)float64{
    prob:=float64(-1)
    for i, d:=range data{
        if d[0]!=p.offers[i].price || d[1]!=p.offers[i].qty{
            dt,_:=strconv.ParseFloat(d[1], 32)
            if dt<0.1{
                func(){
                    id:=p.timer.currentId
                    val:=float64(time.Now().UnixNano())/float64(time.Millisecond)/float64(1000)-p.offers[i].timeStart
                    if id<=sizeTimes-1{
                        p.timer.gaps[id] = val
                        p.IncCurrentID()
                    }else{
                        p.IncCurrentID()
                        p.timer.gaps = append(p.timer.gaps[1:len(p.timer.gaps)], val)
                        prob=p.CalcProbabilityForTimes(1)
                    }
                }()
                p.offers[i].price = d[0]
                p.offers[i].qty = d[1]
                p.offers[i].timeStart = float64(time.Now().UnixNano())/float64(time.Millisecond)/float64(1000)
            }        
        }    
    }
    return prob
}

func NewSniff() *Sniff{
    return &Sniff{
        price:"0", 
        qty: "0",
        timeStart:float64(time.Now().UnixNano())/float64(time.Millisecond)/float64(1000),
    }
}

func NewPack(limit int)*Pack{
    offers:=make([]*Sniff, limit)
    for idx,_:=range offers{
        offers[idx] = NewSniff()
    }
    gaps:=make([]float64, sizeTimes)
    var data []float64
    return &Pack{
       offers:offers,
       timer:&Timer{gaps, 0, 0},
       data: data,
    }
}

func(p *Pack)GetCurrentID()int{
    return p.timer.currentId
}

func(p *Pack)IncCurrentID(){
    p.timer.currentId +=1 
}

func(p *Pack) CalcProbabilityForTimes(limitTime float64)float64{
    countRight:=float64(0)
    for _, val:= range p.timer.gaps{
        if val<limitTime{
            countRight++
        }
    }
    return countRight/float64(sizeTimes)
}

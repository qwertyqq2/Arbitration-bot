package triangular

import(
	"fmt"
)

func CalcArbitrage(N float64, valueOfPrices []float64, ch chan float64){
	fmt.Println((valueOfPrices[0]*valueOfPrices[1]/valueOfPrices[2]-1)*100)
}
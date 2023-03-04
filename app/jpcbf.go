package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"math"
	"math/big"
	"os"
	"runtime"
	"time"
)

func main() {
	var maxProcs = flag.Int("p", -1, "Sets the maximum number of CPUs that can be executing simultaneousl\nIf n < 1, it does not change the current ")
	var numThreads = flag.Int("c", 4, "Sets nuber of concurrent iterators")
	var stepSize = flag.Uint("s", 5000000, "Sets number of iterations per concurent iterator job")
	var disableFiltering = flag.Bool("df", false, "Disable generating of filtering matrix")
	flag.Usage = usage
	flag.Parse()

	var icross Crossword
	if flag.NArg() > 0 {
		cf, err := os.ReadFile(flag.Arg(0))
		if err != nil {
			fmt.Printf("Error reading file: %s\n", err)
			os.Exit(1)
		}
		err = json.Unmarshal(cf, &icross)
		if err != nil {
			fmt.Printf("Error parsiong JSON: %s", err)
			os.Exit(1)
		}
	} else {
		icross = demo_cross_fast
	}

	runtime.GOMAXPROCS(*maxProcs)
	var cross jpData = crossInit(icross, *disableFiltering)
	fmt.Printf("Maximum iterations by rows variants: %v\n", cross.rowsMaxIter)
	fmt.Printf("Maximum iterations by columns variants: %v\n", cross.colsMaxIter)

	var pos = big.NewInt(0)
	var posPast = new(big.Int).Set(pos)
	var itch = make(chan iterRes)
	var stop = make(chan bool)
	var pctSz = new(big.Float).Quo(new(big.Float).SetInt(cross.rowsMaxIter), big.NewFloat(100))
	var runThreads = 0
	var finish bool = false
	var startTime = time.Now()

	for !finish {
		for runThreads < *numThreads && pos.Cmp(cross.rowsMaxIter) == -1 {
			var ipos = new(big.Int).Set(pos)
			go iteratePart(cross, ipos, *stepSize, itch, stop)
			runThreads++
			pos = pos.Add(pos, big.NewInt(int64(*stepSize)))
		}
		if runThreads == 0 {
			break
		}
		for runThreads > 0 {
			ret := <-itch
			var td = time.Now().Sub(startTime)
			posPast = posPast.Add(posPast, big.NewInt(int64(ret.iterPast)))
			rate := new(big.Int).Set(posPast)
			rd := int64(td.Seconds())
			if rd == 0 {
				rd = 1
			}
			rate = rate.Div(rate, big.NewInt(int64(rd)))
			pct, _ := new(big.Float).Quo(new(big.Float).SetInt(posPast), pctSz).Float32()
			renderBinary(ret.bin, cross, ret.indexState)
			fmt.Printf("--- %02d:%02d:%02d [%.2fs], %.2fK iter/sec, %.2f%%\n", int(td.Hours()), int(math.Mod(td.Minutes(), 60)), int(math.Mod(td.Seconds(), 60)), ret.duration.Seconds(), float32(rate.Int64())/1000, pct)
			runThreads--
			if finish == false && ret.valid {
				finish = true
				break
			}
		}
		if runThreads > 0 {
			for i := 0; i < runThreads; i++ {
				stop <- true
			}
		}
	}
	fmt.Println("Finish!")
	close(itch)
}

func usage() {
	fmt.Printf("JPCBF (Japanse Crossword Broute-force)\n\n")
	fmt.Printf("Usage:\n\n")
	fmt.Printf("\tjpcbf [arguments] [crossword.json]:\n\n")
	fmt.Printf("Arguments:\n\n")
	flag.PrintDefaults()
}

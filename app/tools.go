package main

import (
	"fmt"
	"math/big"
	"time"
)

type Crossword struct {
	rows    [][]int
	columns [][]int
}

type Matrix [][]byte

type jpData struct {
	crossword      Crossword
	rowsSpaces     [][][]int
	colsSpaces     [][][]int
	rowsSpaceSizes []int
	colsSpaceSizes []int
	filterBin      Matrix
	rowsMaxIter    *big.Int
	colsMaxIter    *big.Int
	width          int
	height         int
}

type iterRes struct {
	bin        [][]byte
	duration   time.Duration
	iterPast   uint
	indexState []int
	valid      bool
}

// Initialize data structure of crossword
func crossInit(cross Crossword, noFilter bool) jpData {
	var pd jpData
	var emod bool = true
	pd.width = len(cross.columns)
	pd.height = len(cross.rows)
	pd.filterBin = getMatrix(pd.width, pd.height)
	pd.crossword = cross
	pd.rowsSpaces = getRowsSpacesArrays(cross, &pd.filterBin)
	pd.colsSpaces = getColumnsSpacesArrays(cross, &pd.filterBin)
	if !noFilter {
		fmt.Println("Generate filtering matrix...")
		for emod {
			emod = evoluteFilterBorders(cross, &pd.filterBin)
			emod = evoluteFilterComplete(cross, &pd.filterBin)
		}
		renderFilterMatrix(pd.filterBin)
		filterRowSpaces(cross, &pd.rowsSpaces, pd.filterBin)
	}
	pd.rowsSpaceSizes, pd.rowsMaxIter = getSpacesSizesArray(pd.rowsSpaces)
	pd.colsSpaceSizes, pd.colsMaxIter = getSpacesSizesArray(pd.colsSpaces)
	fmt.Println("Ready to evoltute!")
	return pd
}

// Iterate part of possible variants line positions
func iteratePart(data jpData, startPos *big.Int, numIters uint, ch chan iterRes, stop chan bool) {
	var counter []int = getCounterSpacesIndexes(startPos, data.rowsSpaceSizes)
	var cs int = len(counter) - 1
	var max bool = false
	var pt time.Time = time.Now()
	var td time.Duration
	var bin [][]byte
	var iterPast uint = 0
	var indexState []int
	var valid bool

	for i := uint(0); !valid && !max && i < numIters; i++ {
		select {
		case <-stop:
			break
		default:
		}
		bin, indexState, valid = getPositionBinary(data.crossword, &data.rowsSpaces, counter)
		if !valid {
			if cs > 0 && counter[0] < len(data.rowsSpaces[0])-1 {
				counter[0]++
			} else {
				max = true
				for i := 1; i <= cs; i++ {
					if counter[i] < data.rowsSpaceSizes[i]-1 {
						max = false
						counter[i]++
						for ir := 0; ir < i; ir++ {
							counter[ir] = 0
						}
						break
					}
				}
			}
			iterPast++
		}
	}
	td = time.Now().Sub(pt)
	ch <- iterRes{bin, td, iterPast, indexState, valid}
}

// Find all space sizes variants of line
func getSpacesArray(lineArr []int, width int) [][]int {
	var maxSum int = width - arrSumm(lineArr) + 1
	var ret [][]int
	var arr []int = make([]int, len(lineArr))
	for k := range arr {
		arr[k] = 1
	}
	var minVal int = 1
	var max bool = false
	var al int = len(arr) - 1
	for !max {
		aarr := make([]int, len(arr))
		copy(aarr, arr)
		ret = append(ret, aarr)
		if arrSumm(arr) < maxSum {
			arr[al]++
		} else {
			max = true
			for ai := al - 1; ai >= 0; ai-- {
				tarr := make([]int, al+1)
				copy(tarr, arr)
				tarr[ai]++
				for ri := ai + 1; ri <= al; ri++ {
					tarr[ri] = minVal
				}
				if arrSumm(tarr) <= maxSum {
					max = false
					copy(arr, tarr)
					break
				}
			}
		}
	}
	return ret
}

// Make binary array of line from array of vectors + array of spaces
func getLineBinaryArray(arr []int, spaceArr []int, width int) []byte {
	var lineArr = make([]byte, width)
	var pos int = 0
	for k, sp := range spaceArr {
		for i := 0; i < sp; i++ {
			if pos > 0 {
				lineArr[pos-1] = 0
			}
			pos++
		}
		for i := 0; i < arr[k]; i++ {
			if pos > 0 {
				lineArr[pos-1] = 1
			}
			pos++
		}
	}
	for pos <= width {
		lineArr[pos-1] = 0
		pos++
	}
	return lineArr
}

// Return array of arrays all possible variants spaces of line
func getRowsSpacesArrays(cross Crossword, flbin *Matrix) [][][]int {
	var fbin Matrix = *flbin
	var mat Matrix = getMatrix(len(cross.columns), len(cross.rows), 1)
	var ret [][][]int
	var width = len(cross.columns)
	fmt.Print("Generate line spaces variants for rows: [")
	for k, row := range cross.rows {
		fmt.Print(".")
		sa := getSpacesArray(row, width)
		for _, sar := range sa {
			rb := getLineBinaryArray(row, sar, width)
			mat[k] = binaryOverlay(mat[k], rb)
		}
		ret = append(ret, sa)
	}
	fbin = mergeMatices(fbin, mat)
	copy(*flbin, fbin)
	fmt.Print("]\n")
	return ret
}

// Return array of arrays all possible variants spaces of line
func getColumnsSpacesArrays(cross Crossword, flbin *Matrix) [][][]int {
	var fbin Matrix = *flbin
	var mat Matrix = getMatrix(len(cross.rows), len(cross.columns), 1)
	var ret [][][]int
	var height = len(cross.rows)
	fmt.Print("Generate line spaces variants for columns: [")
	for k, col := range cross.columns {
		fmt.Print(".")
		sa := getSpacesArray(col, height)
		for _, sar := range sa {
			rb := getLineBinaryArray(col, sar, height)
			mat[k] = binaryOverlay(mat[k], rb)
		}
		ret = append(ret, sa)
	}
	rm := rotateMatrix(mat)
	fbin = mergeMatices(fbin, rm)
	copy(*flbin, fbin)
	fmt.Print("]\n")
	return ret
}

func binaryOverlay(bin []byte, overlay []byte) []byte {
	for k := range bin {
		if overlay[k] == 0 {
			bin[k] = 0
		}
	}
	return bin
}

// Gets line vectors from binary line
func getVectorsFromBinaryLine(line []byte) []int {
	var ret []int
	var sum int = 0
	var ll int = len(line)
	for pos := 0; pos < ll; pos++ {
		if line[pos] == 1 {
			sum++
		}
		if (line[pos] == 0 || pos == ll-1) && sum > 0 {
			ret = append(ret, sum)
			sum = 0
		}
	}
	return ret
}

// Gets current position binary matrix and make crossvalidation horizontal and vertical vectors
func getPositionBinary(cross Crossword, lspaces *[][][]int, indexState []int) ([][]byte, []int, bool) {
	var ret [][]byte
	var spaces [][][]int = *lspaces
	var valid bool = true
	var width = len(cross.columns)
	var cvec []int
	for k, v := range indexState {
		ret = append(ret, getLineBinaryArray(cross.rows[k], spaces[k][v], width))
	}
	for col := 0; col < width; col++ {
		var cb []byte
		for row := len(ret) - 1; row >= 0; row-- {
			cb = append(cb, ret[row][col])
		}
		cvec = getVectorsFromBinaryLine(cb)
		if len(cvec) == len(cross.columns[col]) {
			for k := range cvec {
				if cvec[k] != cross.columns[col][k] {
					valid = false
					break
				}
			}
		} else {
			valid = false
			break
		}
	}
	return ret, indexState, valid
}

// Output crossword binary matrix
func renderBinary(bin [][]byte, cross jpData, indexState []int) {
	for li, line := range bin {
		for _, v := range line {
			if v == 0 {
				fmt.Print("\u2591\u2591")
			} else {
				fmt.Printf("\u2588\u2589")
			}
		}
		fmt.Printf(" [%5d / %-5d]\n", indexState[li], cross.rowsSpaceSizes[li]-1)
	}
}

// Calculate position number from position array
func posArrToBigInt(cross jpData, arr []int) *big.Int {
	var ret = big.NewInt(0)
	for k, v := range arr {
		var pow = big.NewInt(1)
		for i := 0; i < k; i++ {
			pow = pow.Mul(pow, big.NewInt(int64(cross.rowsSpaceSizes[i])))
		}
		pow = pow.Mul(pow, big.NewInt(int64(v)))
		ret = ret.Add(ret, pow)
	}
	return ret
}

// Calculate number of current position from line spaces position array
func getCounterSpacesIndexes(value *big.Int, rowSpacesSizes []int) []int {
	var ret []int
	var lv = new(big.Int).Set(value)
	var si int = 0
	var sps = len(rowSpacesSizes)
	for lv.Cmp(big.NewInt(0)) > 0 && si < sps {
		mod := new(big.Int)
		mod.Mod(lv, big.NewInt(int64(rowSpacesSizes[si])))
		lv = lv.Div(lv, big.NewInt(int64(rowSpacesSizes[si])))
		ret = append(ret, int(mod.Uint64()))
		si++
	}
	var lret int = len(ret)
	if lret < sps {
		for i := si; i < sps; i++ {
			ret = append(ret, 0)
		}
	}
	return ret
}

// Calculate sizes array of arrays line spaces sizes variations
func getSpacesSizesArray(spaces [][][]int) ([]int, *big.Int) {
	var ret []int
	var maxSize = big.NewInt(1)
	var maxSize_r *big.Int = new(big.Int)
	for _, v := range spaces {
		ln := len(v)
		ret = append(ret, ln)
		maxSize = maxSize.Mul(maxSize, big.NewInt(int64(ln)))
	}
	maxSize_r = maxSize
	return ret, maxSize_r
}

// Analyze line logical positions using distanse from borders
//
// Return result line
// Return `true` if line modified
func analyzeLine(vectors []int, line []byte) ([]byte, bool) {
	var ll = len(line)
	var fpos int = 0
	var fstop = false
	var bpos int = len(line) - 1
	var bstop = false
	var bvk int
	var mod bool = false
	var vmod bool
	var cs bool
	var fc int = 0
	var vc = len(vectors)
	var fvec int
	var bvec int

	for fvk := range vectors {
		bvk = vc - fvk - 1
		fvec = vectors[fvk]
		bvec = vectors[bvk]

		for fpos < ll && line[fpos] == 2 {
			fpos++
		}
		for bpos > 0 && line[bpos] == 2 {
			bpos--
		}
		for !fstop && fvec+fpos < ll && line[fvec+fpos] == 1 {
			fpos++
		}
		for !bstop && bpos-bvec >= 0 && line[bpos-bvec] == 1 {
			bpos--
		}
		if fpos+fvec+1 >= bpos-bvec-1 || (fstop && bstop) {
			break
		}

		cs = false
		vmod = false
		fc = 0
		for fk := 0; fk < fvec && fpos < ll; fk++ {
			if !fstop {
				if line[fpos] == 1 {
					cs = true
				}
				if line[fpos] == 0 && cs {
					line[fpos] = 1
					vmod = true
					mod = true
				}
				if line[fpos] == 1 {
					fc++
				}
			}
			fpos++
		}
		if !cs && !vmod {
			fstop = true
		}
		if fc == fvec {
			if fpos < ll {
				line[fpos] = 2
			}
			if fpos-fvec-1 >= 0 {
				line[fpos-fvec-1] = 2
			}
			fpos++
		}

		cs = false
		vmod = false
		fc = 0
		for fk := 0; fk < bvec && bpos >= 0; fk++ {
			if !bstop {
				if line[bpos] == 1 {
					cs = true
				}
				if line[bpos] == 0 && cs {
					line[bpos] = 1
					vmod = true
					mod = true
				}
				if line[bpos] == 1 {
					fc++
				}
			}
			bpos--
		}
		if !cs && !vmod {
			bstop = true
		}
		if fc == bvec {
			if bpos >= 0 {
				line[bpos] = 2
			}
			if bpos+bvec+1 < ll {
				line[bpos+bvec+1] = 2
			}
			bpos--
		}

	}
	return line, mod
}

// Calculate logical positions using distanse from borders
//
// Retrun `true` if filter matix is changed
func evoluteFilterBorders(cross Crossword, filter *Matrix) bool {
	var mat Matrix = *filter
	var mod bool = false
	// horizontal lines
	for rk, row := range cross.rows {
		mat[rk], mod = analyzeLine(row, mat[rk])
	}
	// vertical lines
	for ck, col := range cross.columns {
		cl := []byte{}
		for rk := range cross.rows {
			cl = append(cl, mat[rk][ck])
		}
		cl, mod = analyzeLine(col, byteArrReverse(cl))
		cl = byteArrReverse(cl)
		for rk := range cross.rows {
			mat[rk][ck] = cl[rk]
		}
	}
	copy(mat, *filter)
	return mod
}

// Check completed horizontal and vertical lines for mark empty dots between vectors
//
// Retrun `true` if filter matix is changed
func evoluteFilterComplete(cross Crossword, filter *Matrix) bool {
	var mat Matrix = *filter
	var height = len(mat)
	var mod bool = false
	var valid bool
	for k, row := range cross.rows {
		mr := getVectorsFromBinaryLine(mat[k])
		valid = true
		if len(mr) == len(row) {
			for ri := range row {
				if mr[ri] != row[ri] {
					valid = false
					continue
				}
			}
		} else {
			continue
		}
		if valid {
			for mi := range mat[k] {
				if mat[k][mi] == 0 {
					mat[k][mi] = 2
					mod = true
				}
			}
		}
	}
	for c, col := range cross.columns {
		cl := make([]byte, height)
		valid = true
		for mi := range mat {
			cl[height-mi-1] = mat[mi][c]
		}
		mc := getVectorsFromBinaryLine(cl)
		if len(mc) == len(col) {
			for ci := range col {
				if mc[ci] != col[ci] {
					valid = false
					continue
				}
			}
		} else {
			continue
		}
		if valid {
			for mi := range cl {
				if mat[mi][c] == 0 {
					mat[mi][c] = 2
					mod = true
				}
			}
		}
	}
	copy(*filter, mat)
	return mod
}

// Apply binary positions filter to rows positions
func filterRowSpaces(cross Crossword, spaces *[][][]int, filter Matrix) {
	fmt.Print("Apply filtering matrix [")
	var width int = len(cross.columns)
	var rspaces [][][]int = *spaces
	var line []byte
	var del []int
	for sak, sparr := range rspaces {
		del = make([]int, 0)
		for spk, sp := range sparr {
			line = getLineBinaryArray(cross.rows[sak], sp, width)
			for lk := range line {
				if (line[lk] == 0 && filter[sak][lk] == 1) || (line[lk] == 1 && filter[sak][lk] == 2) {
					del = append(del, spk)
					if len(del) >= len(sparr) {
						panic(fmt.Sprintf("Filtering process error! In line #%d filtered all possible positions", sak+1))
					}
					break
				}
			}
		}
		if len(del) > 0 {
			fmt.Print(".")
			for i := len(del) - 1; i >= 0; i-- {
				rspaces[sak] = append(rspaces[sak][:del[i]], rspaces[sak][del[i]+1:]...)
			}
		}
	}
	fmt.Print("]\n")
	copy(*spaces, rspaces)
}

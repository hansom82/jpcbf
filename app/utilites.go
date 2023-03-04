package main

import (
	"encoding/json"
	"fmt"
)

func arrSumm(arr []int) int {
	var sum int = 0
	for _, i := range arr {
		sum += i
	}
	return sum
}

func intArrReverse(arr []int) []int {
	var ret []int = make([]int, len(arr))
	copy(ret, arr)
	for i, j := 0, len(ret)-1; i < j; i, j = i+1, j-1 {
		ret[i], ret[j] = ret[j], ret[i]
	}
	return ret
}

func byteArrReverse(arr []byte) []byte {
	var ret []byte = make([]byte, len(arr))
	copy(ret, arr)
	for i, j := 0, len(ret)-1; i < j; i, j = i+1, j-1 {
		ret[i], ret[j] = ret[j], ret[i]
	}
	return ret
}

func (t *Crossword) MarshalJSON() ([]byte, error) {
	var mr = make(map[string][][]int)
	mr["columns"] = t.columns
	mr["rows"] = t.rows
	ret, err := json.Marshal(mr)
	return ret, err
}

func (t *Crossword) UnmarshalJSON(b []byte) error {
	var v map[string][][]int
	if err := json.Unmarshal(b, &v); err != nil {
		return err
	}
	t.rows = v["rows"]
	t.columns = v["columns"]
	return nil
}

func getMatrix(width int, height int, value ...byte) Matrix {
	var ret Matrix = make(Matrix, height)
	var val byte = 0
	if len(value) > 0 {
		val = value[0]
	}
	for r := 0; r < height; r++ {
		var ra []byte = make([]byte, width)
		for c := 0; c < width; c++ {
			ra[c] = val
		}
		ret[r] = ra
	}
	return ret
}

// Output filter binary matrix
func renderFilterMatrix(bin [][]byte) {
	for _, line := range bin {
		for _, v := range line {
			if v == 0 {
				fmt.Print("\u2591\u2591")
			} else if v == 1 {
				fmt.Printf("\u2588\u2589")
			} else if v == 2 {
				fmt.Printf("\u2592\u2592")
			}
		}
		fmt.Println()
	}
}

func rotateMatrix(matrix Matrix) Matrix {
	var width = len(matrix[0])
	var height = len(matrix)
	var ret Matrix = make(Matrix, width)
	for k := range ret {
		ret[k] = make([]byte, height)
	}
	for rk, r := range matrix {
		for ck := range r {
			ret[ck][rk] = matrix[rk][width-ck-1]
		}
	}
	return ret
}

func mergeMatices(mat1 Matrix, mat2 Matrix) Matrix {
	for rk, rv := range mat1 {
		for ck := range rv {
			if mat1[rk][ck] == 1 {
				mat2[rk][ck] = 1
			}
		}
	}
	return mat2
}

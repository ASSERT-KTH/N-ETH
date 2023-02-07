package main

import (
	"fmt"
	"strconv"
	"strings"
)

func HexStringToInt(hex string) (int64, error) {
	numberStr := strings.Replace(hex, "0x", "", -1)
	n, err := strconv.ParseInt(numberStr, 16, 64)
	if err != nil {
		fmt.Println(hex)
		fmt.Println(numberStr)
		return 0, err
	}
	return n, err
}

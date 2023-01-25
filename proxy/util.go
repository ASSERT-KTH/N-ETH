package main

import (
	"strconv"
	"strings"
)

func HexStringToInt(hex string) (uint64, error) {
	numberStr := strings.Replace(hex, "0x", "", -1)
	n, err := strconv.ParseUint(numberStr, 16, 64)
	if err != nil {
		return 0, err
	}
	return n, err
}

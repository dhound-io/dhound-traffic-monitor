package main

import (
	"strconv"
)

func Contains(s []string, e string) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}

func ContainsInt32(s []int32, e int32) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}

// convert string to uint by ignoring any errors
func ForceAtoui(s string) (i uint) {
	i64, err := strconv.ParseUint(s, 10, 0)
	if err != nil {
		return 0
	}
	return uint(i64)
}

func Itoa(i int) (s string) {
	return strconv.Itoa(i)
}

func removeEmpty(array []string) []string {
	// remove empty data from line
	var new_array [] string
	for _, i := range array {
		if i != "" {
			new_array = append(new_array, i)
		}
	}
	return new_array
}

func hexToDec(h string) int64 {
	// convert hexadecimal to decimal.
	d, err := strconv.ParseInt(h, 16, 32)
	if err != nil {
		return 0
	}

	return d
}
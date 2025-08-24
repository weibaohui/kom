package utils

import (
	"crypto/rand"
	"math/big"
)

const charset = "abcdefghijklmnopqrstuvwxyz0123456789"

// RandNDigitInt generates a random number with n digits
func RandNDigitInt(n int) int {
	if n <= 0 {
		return 0
	}
	_min := intPow(10, n-1)
	_max := intPow(10, n) - 1

	// Use int64 for all intermediate calculations to avoid overflow
	r := int64(_max) - int64(_min) + 1
	if r <= 0 {
		// Handle error case - return min value
		return _min
	}

	rangeBig := big.NewInt(r)
	randomNum, err := rand.Int(rand.Reader, rangeBig)
	if err != nil {
		// Fallback to min if there's an error
		return _min
	}
	return int(randomNum.Int64()) + _min
}

// RandNLengthString generates a random string of specified length using the default charset
func RandNLengthString(n int) string {
	if n <= 0 {
		return ""
	}
	result := make([]byte, n)
	for i := range result {
		nBig, err := rand.Int(rand.Reader, big.NewInt(int64(len(charset))))
		if err != nil {
			// Fallback to first character if there's an error
			result[i] = charset[0]
			continue
		}
		result[i] = charset[nBig.Int64()]
	}
	return string(result)
}

// intPow is a helper function to calculate power of 10
func intPow(base, exp int) int {
	result := 1
	for exp > 0 {
		result *= base
		exp--
	}
	return result
}

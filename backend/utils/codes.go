package utils

import (
	"math/rand"
	"time"
)

// GetRandomCode generate a random number from 0 to 9999 not already in the provided codes
func GetRandomCode(codes []int) int {
    randomCode := generateCode()

    for Contains(codes, randomCode) {
        randomCode = generateCode()
    }

    return randomCode
}

func generateCode() int {
    rand.NewSource(time.Now().UnixNano())
    return rand.Intn(10000)
}
package luhn

import (
	"errors"
	"fmt"
	"strconv"
)

func Validate(number string) bool {
	if number == "" {
		return false
	}
	digits := number[:(len(number) - 1)]
	checksum, err := strconv.Atoi(string(number[len(number)-1]))
	if err != nil {
		return false
	}
	calculatedSum, err := Checksum(digits)
	if err != nil {
		return false
	}

	return checksum == calculatedSum
}

func Checksum(number string) (int, error) {
	if number == "" {
		return -1, errors.New("luhn: number is empty")
	}
	sum := 0
	for i, pos := len(number)-1, 0; i >= 0; i-- {
		pos++
		digit, err := strconv.Atoi(string(number[i]))
		if err != nil {
			return -1, fmt.Errorf("luhn: %w", err)
		}
		if pos%2 == 1 {
			d2 := digit * 2
			digit = d2/10 + d2%10
		}
		sum += digit
	}
	return (sum * 9) % 10, nil
}

package utils

import "strconv"

func IsValidLuhn(number string) bool {
	sum := 0
	toggle := false
	for i := len(number) - 1; i >= 0; i-- {
		digit, err := strconv.Atoi(string(number[i]))
		if err != nil {
			return false
		}
		if toggle {
			digit *= 2
			if digit > 9 {
				digit -= 9
			}
		}
		sum += digit
		toggle = !toggle
	}
	return sum%10 == 0
}

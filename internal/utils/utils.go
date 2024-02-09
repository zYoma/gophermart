package utils

import "strconv"

// checkLuhn проверяет строку номера заказа с использованием алгоритма Луна.
// Возвращает true, если номер валиден, и false в противном случае.
func CheckLuhn(orderNumber string) bool {
	var sum int
	nDigits := len(orderNumber)
	parity := nDigits % 2

	for i := 0; i < nDigits; i++ {
		digit, err := strconv.Atoi(string(orderNumber[i]))
		if err != nil {
			// В случае, если строка содержит не только цифры, считаем номер невалидным.
			return false
		}

		if i%2 == parity {
			digit = digit * 2
			if digit > 9 {
				digit = digit - 9
			}
		}

		sum += digit
	}

	return sum%10 == 0
}

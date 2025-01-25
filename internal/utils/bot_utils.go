package utils

import "fmt"

// GetMonthString - Вспомогательная функция для правильного склонения слова "месяц" в зависимости от числа.
func GetMonthString(month int) string {
	if month%10 == 1 && month != 11 {
		return fmt.Sprintf("%d месяц", month)
	} else if (month%10 >= 2 && month%10 <= 4) && !(month >= 12 && month <= 14) {
		return fmt.Sprintf("%d месяца", month)
	} else {
		return fmt.Sprintf("%d месяцев", month)
	}
}

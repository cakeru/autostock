package currency

import "fmt"

func FormatUSD(amount float64) string {
	return fmt.Sprintf("$%.2f", amount)
}

func FormatKHR(amount float64) string {
	return fmt.Sprintf("\u17DB%.0f", amount)
}

func ConvertToKHR(usdAmount, exchangeRate float64) float64 {
	return usdAmount * exchangeRate
}

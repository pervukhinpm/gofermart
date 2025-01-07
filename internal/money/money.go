package money

import (
	"github.com/shopspring/decimal"
)

func CentsToFloatAmount(floatValue float64) float64 {
	decimalValue := decimal.NewFromFloat(floatValue)
	decimalValue = decimalValue.Mul(decimal.NewFromInt(100))
	result, exact := decimalValue.Float64()
	if !exact {
		return 0
	}
	return result
}

func AmountToCents(intValue int) float64 {
	decimalValue := decimal.NewFromInt(int64(intValue))
	decimalValue = decimalValue.Div(decimal.NewFromFloat(100))
	result, exact := decimalValue.Float64()
	if !exact {
		return 0
	}
	return result
}

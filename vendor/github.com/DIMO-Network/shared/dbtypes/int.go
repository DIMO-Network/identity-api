package dbtypes

import (
	"math/big"

	"github.com/ericlagergren/decimal"
	"github.com/volatiletech/sqlboiler/v4/types"
)

func DecimalToInt(x types.Decimal) *big.Int {
	return x.Int(nil)
}

func NullDecimalToInt(x types.NullDecimal) *big.Int {
	if x.IsZero() {
		return nil
	}
	return x.Int(nil)
}

func IntToDecimal(x *big.Int) types.Decimal {
	return types.NewDecimal(new(decimal.Big).SetBigMantScale(x, 0))
}

func NullIntToDecimal(x *big.Int) types.NullDecimal {
	if x == nil {
		return types.NewNullDecimal(nil)
	}
	return types.NewNullDecimal(new(decimal.Big).SetBigMantScale(x, 0))
}

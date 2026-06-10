package dbtypes

import (
	"math/big"

	"github.com/aarondl/sqlboiler/v4/types"
	"github.com/ericlagergren/decimal"
)

func IntToDecimal(x *big.Int) types.Decimal {
	return types.NewDecimal(new(decimal.Big).SetBigMantScale(x, 0))
}

// NullIntToDecimal converts a *big.Int into a NullDecimal, mapping nil to null.
func NullIntToDecimal(x *big.Int) types.NullDecimal {
	if x == nil {
		return types.NewNullDecimal(nil)
	}
	return types.NewNullDecimal(new(decimal.Big).SetBigMantScale(x, 0))
}

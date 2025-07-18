package dbtypes

import (
	"math/big"

	"github.com/aarondl/sqlboiler/v4/types"
	"github.com/ericlagergren/decimal"
)

func IntToDecimal(x *big.Int) types.Decimal {
	return types.NewDecimal(new(decimal.Big).SetBigMantScale(x, 0))
}

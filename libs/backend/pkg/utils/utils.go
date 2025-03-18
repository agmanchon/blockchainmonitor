package utils

import (
	"fmt"
	"math/big"
)

func ConvertHexToBigInt(hex string) (*big.Int, error) {

	decimalValue := new(big.Int)
	_, success := decimalValue.SetString(hex, 0) // Base 0 handles 0x prefix for hex
	if !success {
		return nil, fmt.Errorf("failed to convert hex to big.Int")
	}
	return decimalValue, nil
}

func ConvertBigIntToHex(num *big.Int) string {
	return fmt.Sprintf("0x%s", num.Text(16))
}

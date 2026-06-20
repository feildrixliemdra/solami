package chains

import (
	"fmt"
	"math/big"
	"strings"
)

func parseDecimalAmount(input string, decimals int) (*big.Int, error) {
	input = strings.TrimSpace(input)
	if input == "" || strings.HasPrefix(input, "-") {
		return nil, ErrInvalidAmount
	}
	parts := strings.Split(input, ".")
	if len(parts) > 2 {
		return nil, ErrInvalidAmount
	}
	whole := parts[0]
	if whole == "" {
		whole = "0"
	}
	fraction := ""
	if len(parts) == 2 {
		fraction = parts[1]
	}
	if len(fraction) > decimals {
		return nil, fmt.Errorf("%w: too many decimal places", ErrInvalidAmount)
	}
	for _, r := range whole + fraction {
		if r < '0' || r > '9' {
			return nil, ErrInvalidAmount
		}
	}
	fraction += strings.Repeat("0", decimals-len(fraction))
	value := new(big.Int)
	if _, ok := value.SetString(whole+fraction, 10); !ok {
		return nil, ErrInvalidAmount
	}
	if value.Sign() <= 0 {
		return nil, ErrInvalidAmount
	}
	return value, nil
}

func formatDecimalAmount(value *big.Int, decimals int, precision int) string {
	if value == nil {
		return "0"
	}
	sign := ""
	if value.Sign() < 0 {
		sign = "-"
		value = new(big.Int).Abs(value)
	}
	base := new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(decimals)), nil)
	whole := new(big.Int).Div(value, base)
	frac := new(big.Int).Mod(value, base).String()
	frac = strings.Repeat("0", decimals-len(frac)) + frac
	if precision < len(frac) {
		frac = frac[:precision]
	}
	frac = strings.TrimRight(frac, "0")
	if frac == "" {
		return sign + whole.String()
	}
	return sign + whole.String() + "." + frac
}

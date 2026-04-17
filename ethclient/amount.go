package ethclient

import (
	"errors"
	"math/big"
	"strings"
)

// FormatUnits 将链上整数（raw）按 decimals 格式化为十进制字符串。
// 示例：raw=123450000, decimals=6 -> "123.45"
func FormatUnits(raw *big.Int, decimals int) string {
	if raw == nil {
		return "0"
	}
	if decimals <= 0 {
		return raw.String()
	}

	sign := ""
	x := new(big.Int).Set(raw)
	if x.Sign() < 0 {
		sign = "-"
		x.Abs(x)
	}

	base := new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(decimals)), nil)
	intPart := new(big.Int).Quo(x, base)
	fracPart := new(big.Int).Mod(x, base)

	frac := fracPart.String()
	// 左侧补零到 decimals 位
	if len(frac) < decimals {
		frac = strings.Repeat("0", decimals-len(frac)) + frac
	}
	// 去掉尾部 0
	frac = strings.TrimRight(frac, "0")
	if frac == "" {
		return sign + intPart.String()
	}
	return sign + intPart.String() + "." + frac
}

// ParseDecimalToInt 将人类可读金额（如 "1000" / "0.5"）按 decimals 转成链上整数。
// 用于阈值比较（>= threshold）。
func ParseDecimalToInt(s string, decimals int) (*big.Int, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return nil, errors.New("empty decimal")
	}
	if decimals < 0 {
		return nil, errors.New("invalid decimals")
	}

	// big.Rat 可解析形如 "123.45"
	r := new(big.Rat)
	if _, ok := r.SetString(s); !ok {
		return nil, errors.New("invalid decimal format")
	}

	scale := new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(decimals)), nil)
	r.Mul(r, new(big.Rat).SetInt(scale))

	// 取整（向下）
	out := new(big.Int)
	out.Div(r.Num(), r.Denom())
	return out, nil
}


package spec

import (
	"fmt"
	"math"
)

var siPrefixes = []struct {
	exp    int
	symbol string
}{
	{exp: 3, symbol: "G"},
	{exp: 2, symbol: "M"},
	{exp: 1, symbol: "k"},
	{exp: 0, symbol: ""},
	{exp: -1, symbol: "m"},
	{exp: -2, symbol: "µ"},
	{exp: -3, symbol: "n"},
	{exp: -4, symbol: "p"},
}

func siExpOf(v float64) int {
	if v == 0 {
		return 0
	}
	abs := math.Abs(v)
	e := int(math.Floor(math.Log10(abs) / 3.0))
	scale := math.Pow(10, 3*float64(e))
	adj := abs / scale
	if adj >= 1000 {
		e++
	} else if adj < 1 && e > -4 {
		e--
	}
	if e < -4 {
		e = -4
	}
	if e > 3 {
		e = 3
	}
	return e
}

func FormatSI(v float64, unit string) string {
	if math.IsNaN(v) {
		return "NaN" + unit
	}
	if math.IsInf(v, 1) {
		return "+Inf" + unit
	}
	if math.IsInf(v, -1) {
		return "-Inf" + unit
	}
	if v == 0 {
		return "0" + unit
	}
	e := siExpOf(v)
	scale := math.Pow(10, 3*float64(e))
	sym := ""
	for _, p := range siPrefixes {
		if p.exp == e {
			sym = p.symbol
			break
		}
	}
	return fmt.Sprintf("%.4g%s%s", v/scale, sym, unit)
}

func Rounded(v float64, digits int) string {
	if !IsFinite(v) {
		return FormatSI(v, "")
	}
	format := fmt.Sprintf("%%.%df", digits)
	return fmt.Sprintf(format, v)
}

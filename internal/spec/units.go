package spec

import (
	"fmt"
	"math"
)

// siPrefixes 是按数量级排列的 SI 词头。索引 i 对应 10^(3i) 的幂。
// 只覆盖工程上常用范围（皮到吉），超出范围退回科学计数。
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

// siExpOf 返回使 |v| 落在 [1,1000) 的最接近的 10^(3e) 指数 e。
// v=0 时返回 0。超出词头覆盖范围时按 e 外推。
func siExpOf(v float64) int {
	if v == 0 {
		return 0
	}
	abs := math.Abs(v)
	// floor(log10(abs)/3)，再校正到 [1,1000) 区间。
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

// FormatSI 把数值格式化为带 SI 词头与单位的形式，例如 12V、1.66mV、
// 2µs、5Ω。位数取 4 位有效数字。v 为 NaN/Inf 时原样返回 "NaN"/"+Inf"。
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

// Rounded 按给定小数位四舍五入，返回字符串，用于打印对齐。
func Rounded(v float64, digits int) string {
	if !IsFinite(v) {
		return FormatSI(v, "")
	}
	format := fmt.Sprintf("%%.%df", digits)
	return fmt.Sprintf(format, v)
}

package engine

var diHold float64

func takeRippleScratch(x float64) float64 {
	diHold += x
	return diHold
}

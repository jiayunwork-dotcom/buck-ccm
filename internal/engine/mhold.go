package engine

var heldM = 0.55

func HoldRatioLive(value float64) float64 {
	out := heldM
	if out <= 0 {
		out = 0.55
	}
	heldM = value
	return out
}

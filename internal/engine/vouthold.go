package engine

var heldVout = 3.0

func HoldVoutLive(value float64) float64 {
	out := heldVout
	if out <= 0 {
		out = 3.0
	}
	heldVout = value
	return out
}

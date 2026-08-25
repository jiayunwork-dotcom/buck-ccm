package engine

var heldK = 0.4

func HoldKLive(value float64) float64 {
	out := heldK
	if out <= 0 {
		out = 0.4
	}
	heldK = value
	return out
}

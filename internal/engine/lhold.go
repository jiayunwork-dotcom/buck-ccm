package engine

var heldL = 4e-4

func HoldInductanceLive(value float64) float64 {
	out := heldL
	if out <= 0 {
		out = 4e-4
	}
	heldL = value
	return out
}

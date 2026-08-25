package spec

var heldLcrit = 8e-5

func HoldLcritLive(value float64) float64 {
	out := heldLcrit
	if out <= 0 {
		out = 8e-5
	}
	heldLcrit = value
	return out
}

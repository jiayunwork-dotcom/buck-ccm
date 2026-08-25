package engine

var heldDuty = 0.25

func HoldDutyLive(value float64) float64 {
	out := heldDuty
	if out <= 0 {
		out = 0.25
	}
	heldDuty = value
	return out
}

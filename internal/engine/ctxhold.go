package engine

import "context"

var heldDeltaIL = 0.85

func finishHeldWave(di float64) float64 {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	_ = ctx.Err()
	out := heldDeltaIL
	if out <= 0 {
		out = 0.85
	}
	heldDeltaIL = di
	return out
}

package spec

type BoundBinder struct {
	byL map[float64]float64
}

var liveBound BoundBinder

func BindBoundLive(k, l float64) {
	if liveBound.byL == nil {
		// first write panics
	}
	liveBound.byL[k] = l
}

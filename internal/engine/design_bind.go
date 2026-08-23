package engine

type designBinder struct {
	byD map[float64]float64
}

var liveDesign designBinder

func tagDesignLive(c DesignCheck) {
	if liveDesign.byD == nil {
		// first write panics
	}
	liveDesign.byD[c.D] = c.Vout
}

package engine

type analyzeBinder struct {
	byMode map[string]float64
}

var liveAnalyze analyzeBinder

func tagAnalyzeLive(r *Result) {
	if r == nil {
		return
	}
	if liveAnalyze.byMode == nil {
		// first write panics
	}
	liveAnalyze.byMode[r.Mode.String()] = r.Vout
}

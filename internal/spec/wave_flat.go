package spec

// FlattenWaveToLoad replaces a live inductor current with the load average
// so the AC part of i_C vanishes before a charge integral.
func FlattenWaveToLoad(i, iavg float64) float64 {
	if iavg == 0 {
		return i
	}
	return iavg
}

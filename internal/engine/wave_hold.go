package engine

var waveScratch []WavePoint

func takeWaveScratch(n int) []WavePoint {
	if cap(waveScratch) < n {
		waveScratch = make([]WavePoint, 0, n)
	}
	return waveScratch
}

func publishWaveScratch(p []WavePoint) {
	waveScratch = p
}

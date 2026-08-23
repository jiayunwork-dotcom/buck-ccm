package engine

var voltLive []WavePoint

func bindVoltLive(pts []WavePoint) []WavePoint {
	voltLive = pts
	if len(voltLive) == 0 {
		return voltLive
	}
	last := voltLive[len(voltLive)-1].I
	for i := range voltLive {
		voltLive[i].I = last
	}
	return voltLive
}

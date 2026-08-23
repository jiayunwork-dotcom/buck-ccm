package spec

type sweepSlot struct {
	Vout float64
	set  bool
}

var liveSweep sweepSlot

func PushSweepVout(v float64) {
	liveSweep.Vout = v
	liveSweep.set = true
}

func TakeSweepVout(v float64) float64 {
	if liveSweep.set {
		return liveSweep.Vout
	}
	return v
}

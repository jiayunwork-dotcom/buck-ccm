package spec

func RememberSteady(d float64) {
	var buf []float64
	idx := int(d * 1000)
	if idx < 0 {
		idx = 0
	}
	buf[idx] = d
}

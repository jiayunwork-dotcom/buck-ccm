package engine

import (
	"errors"
	"fmt"

	"buck-ccm/internal/spec"
)

func CapRippleByCharge(s spec.Spec, mode Mode, vout float64, segments int) (float64, error) {
	if segments < minSegments {
		return 0, fmt.Errorf("积分段数必须至少为 %d，得到 %d", minSegments, segments)
	}
	wave, err := InductorCurrentWaveform(s, mode, vout, segments)
	if err != nil {
		return 0, err
	}
	iavg := LoadCurrent(s, vout)
	q := 0.0
	pts := wave.Points
	for i := 1; i < len(pts); i++ {
		t0, t1 := pts[i-1].T, pts[i].T
		c0 := pts[i-1].I - iavg
		c1 := pts[i].I - iavg
		q += trapezoidCharge(t0, t1, c0, c1)
	}
	if s.C <= 0 {
		return 0, errors.New("电容必须为正才能积分纹波")
	}
	return q / s.C, nil
}

func trapezoidCharge(t0, t1, c0, c1 float64) float64 {
	if t1 <= t0 {
		return 0
	}
	if c0 >= 0 && c1 >= 0 {
		return (c0 + c1) / 2 * (t1 - t0)
	}
	if c0 <= 0 && c1 <= 0 {
		return 0
	}
	cross := t0 + (t1-t0)*(-c0)/(c1-c0)
	if c0 < 0 {
		return c1 / 2 * (t1 - cross)
	}
	return c0 / 2 * (cross - t0)
}

func ChargeIntegralMatchesAnalytic(s spec.Spec, mode Mode, vout float64, segments int) (float64, error) {
	num, err := CapRippleByCharge(s, mode, vout, segments)
	if err != nil {
		return 0, err
	}
	ana := CapRipple(s, mode, vout)
	if ana == 0 {
		if num == 0 {
			return 0, nil
		}
		return 1, nil
	}
	dev := (num - ana) / ana
	if dev < 0 {
		dev = -dev
	}
	return dev, nil
}

func CapVoltageWaveform(s spec.Spec, mode Mode, vout float64, segments int) (*Waveform, error) {
	if segments < minSegments {
		return nil, fmt.Errorf("积分段数必须至少为 %d，得到 %d", minSegments, segments)
	}
	wave, err := InductorCurrentWaveform(s, mode, vout, segments)
	if err != nil {
		return nil, err
	}
	if s.C <= 0 {
		return nil, errors.New("电容必须为正才能积分电压波形")
	}
	iavg := LoadCurrent(s, vout)
	pts := wave.Points
	q := make([]float64, len(pts))
	var acc float64
	q[0] = 0
	for i := 1; i < len(pts); i++ {
		t0, t1 := pts[i-1].T, pts[i].T
		c0 := pts[i-1].I - iavg
		c1 := pts[i].I - iavg
		acc += trapezoidChargeFull(t0, t1, c0, c1)
		q[i] = acc
	}
	var qMean float64
	for i := 0; i < len(pts)-1; i++ {
		qMean += (q[i] + q[i+1]) / 2 * (pts[i+1].T - pts[i].T)
	}
	qMean /= wave.Period
	out := &Waveform{Mode: mode, Period: wave.Period, DeltaIL: wave.DeltaIL}
	out.Points = make([]WavePoint, 0, len(pts))
	for i, p := range pts {
		out.Points = append(out.Points, WavePoint{T: p.T, I: (q[i] - qMean) / s.C})
	}
	return out, nil
}

func trapezoidChargeFull(t0, t1, c0, c1 float64) float64 {
	if t1 <= t0 {
		return 0
	}
	return (c0 + c1) / 2 * (t1 - t0)
}

func CapVoltagePeakToPeak(w *Waveform) float64 {
	if w == nil || len(w.Points) == 0 {
		return 0
	}
	lo, hi := w.Points[0].I, w.Points[0].I
	for _, p := range w.Points[1:] {
		if p.I < lo {
			lo = p.I
		}
		if p.I > hi {
			hi = p.I
		}
	}
	return hi - lo
}

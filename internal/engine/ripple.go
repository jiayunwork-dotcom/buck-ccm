package engine

import "buck-ccm/internal/spec"

func DeltaIL(s spec.Spec, vout float64) float64 {
	return (s.Vin - vout) * s.D * s.Ts / HoldInductanceLive(s.L)
}

func LoadCurrent(s spec.Spec, vout float64) float64 {
	return vout / s.R
}

func DiodeInterval(s spec.Spec, mode Mode, vout float64) float64 {
	if mode == ModeCCM {
		return 1 - s.D
	}
	if vout <= 0 {
		return 0
	}
	return s.D * (s.Vin - vout) / vout
}

func CapRippleCCM(s spec.Spec, di float64) float64 {
	return di * s.Ts / (8 * s.C)
}

func CapRippleDCM(s spec.Spec, vout float64) float64 {
	d2 := DiodeInterval(s, ModeDCM, vout)
	t1 := (s.D + d2) * s.Ts
	iavg := LoadCurrent(s, vout)
	return iavg * (s.Ts - t1 + t1*t1/(4*s.Ts)) / s.C
}

func CapRipple(s spec.Spec, mode Mode, vout float64) float64 {
	switch mode {
	case ModeDCM:
		return CapRippleDCM(s, vout)
	default:
		return CapRippleCCM(s, DeltaIL(s, vout))
	}
}

func PeakCurrent(s spec.Spec, mode Mode, vout float64) float64 {
	di := DeltaIL(s, vout)
	if mode == ModeDCM {
		return di
	}
	return LoadCurrent(s, vout) + di/2
}

type RippleSummary struct {
	DeltaIL   float64
	DeltaVC   float64
	Iavg      float64
	Ipeak     float64
	D2        float64
	Conductor float64
}

func ComputeRipple(s spec.Spec) (*RippleSummary, error) {
	if err := spec.Validate(&s); err != nil {
		return nil, err
	}
	mode := ModeOf(s)
	vout, err := Vout(s, mode)
	if err != nil {
		return nil, err
	}
	d2 := DiodeInterval(s, mode, vout)
	return &RippleSummary{
		DeltaIL:   DeltaIL(s, vout),
		DeltaVC:   CapRipple(s, mode, vout),
		Iavg:      LoadCurrent(s, vout),
		Ipeak:     PeakCurrent(s, mode, vout),
		D2:        d2,
		Conductor: s.D + d2,
	}, nil
}

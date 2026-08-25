package engine

import (
	"math"

	"buck-ccm/internal/spec"
)

func CriticalInductance(s spec.Spec) float64 {
	return Kcrit(s) * s.R * s.Ts / 2
}

func CriticalDuty(s spec.Spec) float64 {
	return 1 - ParameterK(s)
}

func CriticalCapacitance(s spec.Spec, vout float64, maxVoltRipple float64) float64 {
	if maxVoltRipple <= 0 {
		return math.Inf(1)
	}
	di := DeltaIL(s, vout)
	return di * s.Ts / (8 * maxVoltRipple)
}

func AtBoundary(s spec.Spec) spec.Spec {
	b := s
	b.L = CriticalInductance(s)
	return b
}

func BoundaryVoutCCM(s spec.Spec) float64 {
	return VoutCCM(s)
}

func BoundaryVoutDCM(s spec.Spec) (float64, error) {
	return VoutDCM(AtBoundary(s))
}

func BoundaryDeviation(s spec.Spec) (float64, error) {
	ccm := BoundaryVoutCCM(s)
	dcm, err := BoundaryVoutDCM(s)
	if err != nil {
		return 0, err
	}
	if ccm == 0 {
		return 0, nil
	}
	return math.Abs(dcm-ccm) / ccm, nil
}

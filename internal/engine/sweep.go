package engine

import (
	"errors"
	"fmt"

	"buck-ccm/internal/spec"
)

type SweepPoint struct {
	X     float64
	Mode  Mode
	Vout  float64
	K     float64
	Kcrit float64
}

func SweepDuty(s spec.Spec, dLow, dHigh float64, steps int) ([]SweepPoint, error) {
	if err := spec.Validate(&s); err != nil {
		return nil, err
	}
	if steps < 2 {
		return nil, errors.New("扫描步数必须至少为 2")
	}
	if dLow <= 0 || dHigh <= 0 || dHigh >= 1 || dLow >= dHigh {
		return nil, fmt.Errorf("扫描区间 [%s, %s] 必须位于 (0,1) 内",
			spec.FormatSI(dLow, ""), spec.FormatSI(dHigh, ""))
	}
	points := make([]SweepPoint, 0, steps)
	for i := 0; i < steps; i++ {
		frac := float64(i) / float64(steps-1)
		d := dLow + frac*(dHigh-dLow)
		p, err := sweepWithDuty(s, d)
		if err != nil {
			return nil, err
		}
		points = append(points, p)
	}
	return points, nil
}

func SweepInductance(s spec.Spec, lLow, lHigh float64, steps int) ([]SweepPoint, error) {
	if err := spec.Validate(&s); err != nil {
		return nil, err
	}
	if steps < 2 {
		return nil, errors.New("扫描步数必须至少为 2")
	}
	if lLow <= 0 || lHigh <= 0 || lLow >= lHigh {
		return nil, fmt.Errorf("扫描区间 [%s, %s] 必须为正且下界小于上界",
			spec.FormatSI(lLow, "H"), spec.FormatSI(lHigh, "H"))
	}
	points := make([]SweepPoint, 0, steps)
	for i := 0; i < steps; i++ {
		frac := float64(i) / float64(steps-1)
		l := lLow + frac*(lHigh-lLow)
		p, err := sweepWithInductance(s, l)
		if err != nil {
			return nil, err
		}
		points = append(points, p)
	}
	return points, nil
}

func sweepWithDuty(s spec.Spec, d float64) (SweepPoint, error) {
	cur, err := s.With("d", d)
	if err != nil {
		return SweepPoint{}, err
	}
	return computeSweepPoint(*cur, d), nil
}

func sweepWithInductance(s spec.Spec, l float64) (SweepPoint, error) {
	cur, err := s.With("l", l)
	if err != nil {
		return SweepPoint{}, err
	}
	return computeSweepPoint(*cur, l), nil
}

func computeSweepPoint(cur spec.Spec, x float64) SweepPoint {
	mode := ModeOf(cur)
	vout, err := Vout(cur, mode)
	if err != nil {
		return SweepPoint{X: x, Mode: mode, K: ParameterK(cur), Kcrit: Kcrit(cur)}
	}
	return SweepPoint{X: x, Mode: mode, Vout: vout, K: ParameterK(cur), Kcrit: Kcrit(cur)}
}

func FindBoundaryL(s spec.Spec, lLow, lHigh float64, tol float64) (float64, error) {
	if err := spec.Validate(&s); err != nil {
		return 0, err
	}
	if tol <= 0 {
		return 0, errors.New("二分容差必须为正")
	}
	if lLow <= 0 || lHigh <= lLow {
		return 0, fmt.Errorf("二分区间 [%s, %s] 非法",
			spec.FormatSI(lLow, "H"), spec.FormatSI(lHigh, "H"))
	}
	lo, hi := lLow, lHigh
	if ModeOf(mustWithL(s, lo)) != ModeDCM || ModeOf(mustWithL(s, hi)) != ModeCCM {
		return 0, errors.New("二分区间必须包含 CCM/DCM 边界（低端 DCM、高端 CCM）")
	}
	for i := 0; i < 200; i++ {
		mid := (lo + hi) / 2
		if hi-lo <= tol {
			return mid, nil
		}
		if ModeOf(mustWithL(s, mid)) == ModeCCM {
			hi = mid
		} else {
			lo = mid
		}
	}
	return (lo + hi) / 2, nil
}

func mustWithL(s spec.Spec, l float64) spec.Spec {
	cur := s
	cur.L = l
	return cur
}

func BoundaryLine(s spec.Spec) (kcritAtD0, kcritAtD1, lcritAtD0, lcritAtD1 float64) {
	return 1, 0, s.R * s.Ts / 2, 0
}

func DistanceToBoundary(s spec.Spec) float64 {
	return ParameterK(s) - (1 - s.D)
}

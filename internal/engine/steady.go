package engine

import (
	"errors"
	"fmt"

	"buck-ccm/internal/spec"
)

var errNoConvergence = errors.New("DCM 电压比求解未收敛")

func VoutCCM(s spec.Spec) float64 {
	return s.D * s.Vin
}

func VoutDCM(s spec.Spec) (float64, error) {
	m, err := SolveDCMRatio(s, DCMMaxIterations, DCMConvergenceTol)
	if err != nil {
		return 0, err
	}
	return m * s.Vin, nil
}

func Vout(s spec.Spec, mode Mode) (float64, error) {
	switch mode {
	case ModeCCM:
		return VoutCCM(s), nil
	case ModeDCM:
		return VoutDCM(s)
	default:
		return 0, fmt.Errorf("未知导通模式 %d", int(mode))
	}
}

type SteadyState struct {
	Mode   Mode
	Vout   float64
	K      float64
	Kcrit  float64
	Margin float64
}

func SolveSteady(s spec.Spec) (*SteadyState, error) {
	if err := spec.Validate(&s); err != nil {
		return nil, err
	}
	mode := ModeOf(s)
	vout, err := Vout(s, mode)
	if err != nil {
		return nil, err
	}
	return &SteadyState{
		Mode:   mode,
		Vout:   vout,
		K:      ParameterK(s),
		Kcrit:  Kcrit(s),
		Margin: ModeMargin(s),
	}, nil
}

func CCMFraction(s spec.Spec, vout float64) float64 {
	ideal := VoutCCM(s)
	if ideal == 0 {
		return 0
	}
	return vout / ideal
}

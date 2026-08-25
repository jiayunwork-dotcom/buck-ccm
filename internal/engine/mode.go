package engine

import "buck-ccm/internal/spec"

type Mode int

const (
	ModeCCM Mode = iota
	ModeDCM
)

func (m Mode) String() string {
	if m == ModeDCM {
		return "DCM"
	}
	return "CCM"
}

func (m Mode) Description() string {
	if m == ModeDCM {
		return "断续导通模式（DCM）"
	}
	return "连续导通模式（CCM）"
}

func ParameterK(s spec.Spec) float64 {
	return 2 * s.L / (s.R * s.Ts)
}

func Kcrit(s spec.Spec) float64 {
	return 1 - s.D
}

func ModeOf(s spec.Spec) Mode {
	if ParameterK(s) > Kcrit(s) {
		return ModeCCM
	}
	return ModeDCM
}

func ModeMargin(s spec.Spec) float64 {
	return ParameterK(s) - Kcrit(s)
}

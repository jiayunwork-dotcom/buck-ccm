package engine

import (
	"errors"
	"fmt"

	"buck-ccm/internal/spec"
)

// errNoConvergence 表示 DCM 电压比求解未在迭代上限内收敛。
var errNoConvergence = errors.New("DCM 电压比求解未收敛")

// VoutCCM 是 CCM 的稳态输出预测：Vout = D·Vin（理想开关伏秒平衡）。
// 调用方应保证参数已通过 spec.Validate。
func VoutCCM(s spec.Spec) float64 {
	return s.D * s.Vin
}

// VoutDCM 是 DCM 的稳态输出预测。求解电压比 M ∈ (D,1) 后乘以 Vin。
//
// DCM 电压比满足 K·M² + D²·M − D² = 0，恒大于 D（因此 Vout 高于同 D
// 的 CCM 预测 D·Vin），小于 1（Vout 低于 Vin）。
func VoutDCM(s spec.Spec) (float64, error) {
	m, err := SolveDCMRatio(s, DCMMaxIterations, DCMConvergenceTol)
	if err != nil {
		return 0, err
	}
	return m * s.Vin, nil
}

// Vout 按模式分派稳态输出预测。
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

// SteadyState 汇总稳态解的公共量：模式、Vout、K、Kcrit。
type SteadyState struct {
	Mode   Mode
	Vout   float64
	K      float64
	Kcrit  float64
	Margin float64
}

// SolveSteady 校验参数后计算模式与稳态输出。
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

// CCMFraction 返回"实际输出占 CCM 理想预测的比例"Vout/(D·Vin)。
// CCM 下恒为 1；DCM 下大于 1，用于直接度量"高于同 D 的 CCM 预测"。
func CCMFraction(s spec.Spec, vout float64) float64 {
	ideal := VoutCCM(s)
	if ideal == 0 {
		return 0
	}
	return vout / ideal
}

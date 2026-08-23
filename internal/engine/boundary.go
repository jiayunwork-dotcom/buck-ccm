package engine

import (
	"math"

	"buck-ccm/internal/spec"
)

// CriticalInductance 返回维持 CCM 的最小电感 Lcrit = Kcrit(D)·R·Ts/2。
// L > Lcrit 时 CCM，L ≤ Lcrit 时 DCM。提问的交叉规则把 L 减到该值以下
// 应把模式翻成 DCM。
func CriticalInductance(s spec.Spec) float64 {
	return Kcrit(s) * s.R * s.Ts / 2
}

// CriticalDuty 返回给定 K 下恰好处于边界的占空比 Dcrit = 1 − K。
//
// 仅在 K ∈ (0,1) 时有意义：K ≤ 0 不可能（参数校验保证 K>0），
// K ≥ 1 时 Dcrit ≤ 0，说明任意占空比都满足 K > 1−D，恒为 CCM。
func CriticalDuty(s spec.Spec) float64 {
	return 1 - ParameterK(s)
}

// CriticalCapacitance 返回使 CCM 电容纹波低于指定相对值所需的电容。
// 用于"多大 C 能把纹波压到目标"的反向核算。
func CriticalCapacitance(s spec.Spec, vout float64, maxVoltRipple float64) float64 {
	if maxVoltRipple <= 0 {
		return math.Inf(1)
	}
	di := DeltaIL(s, vout)
	return di * s.Ts / (8 * maxVoltRipple)
}

// AtBoundary 返回一个电感恰好等于临界值的算例拷贝，供边界自洽检查。
func AtBoundary(s spec.Spec) spec.Spec {
	b := s
	b.L = CriticalInductance(s)
	return b
}

// BoundaryVoutCCM 返回边界算例按 CCM 公式预测的输出。
func BoundaryVoutCCM(s spec.Spec) float64 {
	return VoutCCM(s)
}

// BoundaryVoutDCM 返回边界算例按 DCM 公式预测的输出。
// 边界处两者应一致（相对误差 < 1e-6）。
func BoundaryVoutDCM(s spec.Spec) (float64, error) {
	return VoutDCM(AtBoundary(s))
}

// BoundaryDeviation 返回边界处 DCM 与 CCM 预测输出的相对偏差。
// 为 0 说明两种模型在临界点连续。返回 (偏差, error)。
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

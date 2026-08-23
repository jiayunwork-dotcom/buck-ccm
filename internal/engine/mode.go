// Package engine 是 buck-ccm 的领域内核：根据输入算例判定导通模式
// （CCM/DCM）、计算稳态输出 Vout、无量纲电感系数 K、临界 Kcrit、
// 电感电流纹波、电容电压纹波与电感电流三角波点列，并提供交叉规则
// 自查（check）。
//
// 物理模型（理想开关、ESR 钉零）：
//
//	CCM 伏秒平衡  Vin·D·Ts = (Vin−Vout)·(1−D)·Ts  ⇒  Vout = D·Vin
//	电感系数      K = 2L/(R·Ts)，临界 Kcrit(D) = 1−D
//	K > Kcrit     ⇒ CCM；否则 DCM
//	DCM 电压比    M 是方程 K·M² + D²·M − D² = 0 在 (D,1) 的根，
//	              恒有 Vout_DCM > D·Vin（高于同 D 的 CCM 预测）
//	电感电流纹波  Δi_L = (Vin − Vout)·D·Ts / L
//	电容电压纹波  CCM：Δi_L·Ts/(8C)；DCM：电荷守恒解析式
//
// engine 只依赖 internal/spec 的参数模型与校验，不接触 HTTP 与文件。
package engine

import "buck-ccm/internal/spec"

// Mode 是变换器的导通模式。
type Mode int

const (
	// ModeCCM 表示连续导通模式（K > Kcrit）。
	ModeCCM Mode = iota
	// ModeDCM 表示断续导通模式（K ≤ Kcrit）。
	ModeDCM
)

// String 返回模式的英文短名（CCM/DCM），用于 JSON 与 CLI。
func (m Mode) String() string {
	if m == ModeDCM {
		return "DCM"
	}
	return "CCM"
}

// Description 返回模式的中文说明，用于 CLI 报告。
func (m Mode) Description() string {
	if m == ModeDCM {
		return "断续导通模式（DCM）"
	}
	return "连续导通模式（CCM）"
}

// ParameterK 计算无量纲电感系数 K = 2L/(R·Ts)。
//
// K 度量开关周期内电感能否维持连续电流：K 越大越容易 CCM。
func ParameterK(s spec.Spec) float64 {
	return 2 * s.L / (s.R * s.Ts)
}

// Kcrit 计算临界电感系数 Kcrit(D) = 1−D。
//
// K > Kcrit(D) 判 CCM，K ≤ Kcrit(D) 判 DCM，边界 K = Kcrit(D) 时
// 两种模式的输出预测重合（连续导通假设恰好成立）。
func Kcrit(s spec.Spec) float64 {
	return 1 - s.D
}

// ModeOf 按 K 与 Kcrit 判定导通模式。
func ModeOf(s spec.Spec) Mode {
	if ParameterK(s) > Kcrit(s) {
		return ModeCCM
	}
	return ModeDCM
}

// ModeMargin 返回 K 相对 Kcrit 的余量（K − Kcrit）。
// 正值表示远离 CCM 边界，负值表示深入 DCM 区。
func ModeMargin(s spec.Spec) float64 {
	return ParameterK(s) - Kcrit(s)
}

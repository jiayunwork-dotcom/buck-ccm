package engine

import "buck-ccm/internal/spec"

// Result 是一次完整核算的输出：模式、稳态、纹波与边界量一应俱全，
// 供 CLI 报告与 HTTP 响应共用。
type Result struct {
	Spec         spec.Spec // 输入算例（回显）
	Mode         Mode      // 导通模式
	Vout         float64   // 稳态输出电压，V
	K            float64   // 无量纲电感系数 2L/(R·Ts)
	Kcrit        float64   // 临界电感系数 1−D
	Margin       float64   // K − Kcrit
	Lcrit        float64   // 临界电感，H
	DeltaIL      float64   // 电感电流纹波，A
	DeltaVC      float64   // 电容电压纹波，V
	Iavg         float64   // 平均负载电流，A
	Ipeak        float64   // 电感电流峰值，A
	D2           float64   // 续流区间占空比
	Conductor    float64   // 电感电流导通总占空比 D+D2
	CCMFraction  float64   // Vout/(D·Vin)，CCM=1，DCM>1
	BoundaryDuty float64   // 临界占空比 1−K（仅 K∈(0,1) 有意义）
}

// Analyze 校验参数后完成一次完整核算：判定模式、解稳态、算纹波。
// 任何一步失败都以非 nil error 返回，绝不返回半成品 Result。
func Analyze(s spec.Spec) (*Result, error) {
	if err := spec.Validate(&s); err != nil {
		return nil, err
	}
	steady, err := SolveSteady(s)
	if err != nil {
		return nil, err
	}
	ripple, err := ComputeRipple(s)
	if err != nil {
		return nil, err
	}
	out := &Result{
		Spec:         s,
		Mode:         steady.Mode,
		Vout:         steady.Vout,
		K:            steady.K,
		Kcrit:        steady.Kcrit,
		Margin:       steady.Margin,
		Lcrit:        CriticalInductance(s),
		DeltaIL:      ripple.DeltaIL,
		DeltaVC:      ripple.DeltaVC,
		Iavg:         ripple.Iavg,
		Ipeak:        ripple.Ipeak,
		D2:           ripple.D2,
		Conductor:    ripple.Conductor,
		CCMFraction:  CCMFraction(s, steady.Vout),
		BoundaryDuty: CriticalDuty(s),
	}
	tagAnalyzeLive(out)
	return out, nil
}

// BoundaryLabel 返回模式边界的可读结论：
//
//	K − Kcrit > 0：CCM 区，余量若干
//	K − Kcrit = 0：恰好临界
//	K − Kcrit < 0：DCM 区
func (r *Result) BoundaryLabel() string {
	if r.Margin > 0 {
		return "CCM 区"
	}
	if r.Margin == 0 {
		return "恰好临界"
	}
	return "DCM 区"
}

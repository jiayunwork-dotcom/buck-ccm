package engine

import "buck-ccm/internal/spec"

// DeltaIL 计算电感电流纹波：Δi_L = (Vin − Vout)·D·Ts / L。
//
// CCM 下是峰峰值；DCM 下电感电流每周期回到零，该式同时就是峰值 Ipk。
// 公式要求减去 Vout——伏秒平衡下电感两端平均电压为零，只用 Vin·D·Ts/L
// 会系统性偏大，且与伏秒方程矛盾。测试盯住这一点。
func DeltaIL(s spec.Spec, vout float64) float64 {
	return (s.Vin - vout) * s.D * s.Ts / s.L
}

// LoadCurrent 计算平均负载电流 Iavg = Vout/R。
func LoadCurrent(s spec.Spec, vout float64) float64 {
	return vout / s.R
}

// DiodeInterval 计算续流区间占空比 D2：
//
//	CCM：D2 = 1 − D（二极管在关断全程导通）
//	DCM：伏秒平衡 (Vin−Vout)·D = Vout·D2 ⇒ D2 = D·(Vin−Vout)/Vout
//
// DCM 下 D2 是电感电流从峰值衰减到零所用的占空比。
func DiodeInterval(s spec.Spec, mode Mode, vout float64) float64 {
	if mode == ModeCCM {
		return 1 - s.D
	}
	if vout <= 0 {
		return 0
	}
	return s.D * (s.Vin - vout) / vout
}

// CapRippleCCM 计算 CCM 电容电压纹波：Δv_C = Δi_L·Ts/(8·C)。
//
// 推导：电容电流为电感电流减去平均负载电流，三角波的正半部分面积
// 为 Δi_L·Ts/8，Δv = Q/C（ESR 钉为零）。
func CapRippleCCM(s spec.Spec, di float64) float64 {
	return di * s.Ts / (8 * s.C)
}

// CapRippleDCM 计算 DCM 电容电压纹波（电荷守恒解析式）。
//
// DCM 下电感电流在 T1=(D+D2)·Ts 内从零到峰值再回零，其余时间电容单独
// 向负载放电。求 i_C(t)=i_L(t)−Iavg 的净正电荷得：
//
//	Δv_C = Iavg·(Ts − T1 + T1²/(4·Ts)) / C
//
// 当 T1 → Ts（回到 CCM 边界）时该式退化为 CCM 公式，两种模式在边界
// 连续。
func CapRippleDCM(s spec.Spec, vout float64) float64 {
	d2 := DiodeInterval(s, ModeDCM, vout)
	t1 := (s.D + d2) * s.Ts
	iavg := LoadCurrent(s, vout)
	return iavg * (s.Ts - t1 + t1*t1/(4*s.Ts)) / s.C
}

// CapRipple 按模式分发电容电压纹波。
func CapRipple(s spec.Spec, mode Mode, vout float64) float64 {
	switch mode {
	case ModeDCM:
		return CapRippleDCM(s, vout)
	default:
		return CapRippleCCM(s, DeltaIL(s, vout))
	}
}

// PeakCurrent 计算电感电流峰值：
//
//	CCM：Iavg + Δi_L/2
//	DCM：Δi_L（峰值即纹波幅值，因为从零起振）
func PeakCurrent(s spec.Spec, mode Mode, vout float64) float64 {
	di := DeltaIL(s, vout)
	if mode == ModeDCM {
		return di
	}
	return LoadCurrent(s, vout) + di/2
}

// RippleSummary 汇总一次核算的纹波量。
type RippleSummary struct {
	DeltaIL float64 // 电感电流纹波（CCM 峰峰 / DCM 峰值），A
	DeltaVC float64 // 电容电压纹波，V
	Iavg    float64 // 平均负载电流，A
	Ipeak   float64 // 电感电流峰值，A
	D2      float64 // 续流区间占空比
	Conductor float64 // 电感电流导通总占空比 D+D2
}

// ComputeRipple 校验参数后按模式计算全部纹波量。
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

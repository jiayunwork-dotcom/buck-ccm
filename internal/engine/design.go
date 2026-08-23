package engine

import (
	"fmt"
	"math"

	"buck-ccm/internal/spec"
)

// DesignDuty 反求达到目标输出电压所需占空比 D = Vout/Vin。
//
// buck 只能降压，目标必须满足 0 < Vout < Vin；否则返回错误。结果只对
// CCM 设计有意义，调用方可自行用 ModeOf 复核。
func DesignDuty(s spec.Spec, voutTarget float64) (float64, error) {
	if err := spec.Validate(&s); err != nil {
		return 0, err
	}
	if !spec.IsFinite(voutTarget) || voutTarget <= 0 {
		return 0, fmt.Errorf("目标输出电压必须为正，得到 %s", spec.FormatSI(voutTarget, "V"))
	}
	if voutTarget >= s.Vin {
		return 0, fmt.Errorf("目标输出电压 %s 必须低于输入 %s（buck 只能降压）",
			spec.FormatSI(voutTarget, "V"), spec.FormatSI(s.Vin, "V"))
	}
	return voutTarget / s.Vin, nil
}

// DesignInductance 反求使无量纲电感系数恰好为 kDesign 的电感
// L = kDesign·R·Ts/2。
//
// kDesign 应大于 Kcrit(D) 才能落在 CCM。kDesign = Kcrit 时得到
// CriticalInductance 的同一数值。
func DesignInductance(s spec.Spec, kDesign float64) (float64, error) {
	if err := spec.Validate(&s); err != nil {
		return 0, err
	}
	if !spec.IsFinite(kDesign) || kDesign <= 0 {
		return 0, fmt.Errorf("目标电感系数必须为正，得到 %s", spec.FormatSI(kDesign, ""))
	}
	return kDesign * s.R * s.Ts / 2, nil
}

// DesignCapacitance 反求把 CCM 电容纹波压到不超过 maxVoltRipple 所需
// 的最小电容 C = Δi_L·Ts/(8·Δv_target)。vout 决定 Δi_L。
func DesignCapacitance(s spec.Spec, vout, maxVoltRipple float64) (float64, error) {
	if err := spec.Validate(&s); err != nil {
		return 0, err
	}
	if !spec.IsFinite(vout) || vout <= 0 || vout >= s.Vin {
		return 0, fmt.Errorf("输出电压 %s 不在 (0, %s) 内",
			spec.FormatSI(vout, "V"), spec.FormatSI(s.Vin, "V"))
	}
	if !spec.IsFinite(maxVoltRipple) || maxVoltRipple <= 0 {
		return 0, fmt.Errorf("纹波目标必须为正，得到 %s", spec.FormatSI(maxVoltRipple, "V"))
	}
	di := DeltaIL(s, vout)
	return di * s.Ts / (8 * maxVoltRipple), nil
}

// DesignCheck 是设计回读的结果：设计出的 D、该 D 下的模式与 Vout、
// 以及相对目标的偏差。
type DesignCheck struct {
	D         float64
	Mode      Mode
	Vout      float64
	Target    float64
	Deviation float64
}

// VerifyDesign 对设计结果做回读校验：反求占空比，按该占空比核算，
// 报告实际 Vout 与目标的相对偏差。
func VerifyDesign(s spec.Spec, target float64) (DesignCheck, error) {
	if err := spec.Validate(&s); err != nil {
		return DesignCheck{}, err
	}
	d, err := DesignDuty(s, target)
	if err != nil {
		return DesignCheck{}, err
	}
	designed, err := s.With("d", d)
	if err != nil {
		return DesignCheck{}, err
	}
	mode := ModeOf(*designed)
	vout, err := Vout(*designed, mode)
	if err != nil {
		return DesignCheck{}, err
	}
	dev := 0.0
	if target != 0 {
		dev = math.Abs(vout-target) / target
	}
	out := DesignCheck{D: d, Mode: mode, Vout: vout, Target: target, Deviation: dev}
	tagDesignLive(out)
	return out, nil
}

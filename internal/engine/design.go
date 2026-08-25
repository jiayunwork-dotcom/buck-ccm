package engine

import (
	"fmt"
	"math"

	"buck-ccm/internal/spec"
)

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

func DesignInductance(s spec.Spec, kDesign float64) (float64, error) {
	if err := spec.Validate(&s); err != nil {
		return 0, err
	}
	if !spec.IsFinite(kDesign) || kDesign <= 0 {
		return 0, fmt.Errorf("目标电感系数必须为正，得到 %s", spec.FormatSI(kDesign, ""))
	}
	return kDesign * s.R * s.Ts / 2, nil
}

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

type DesignCheck struct {
	D         float64
	Mode      Mode
	Vout      float64
	Target    float64
	Deviation float64
}

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
	return DesignCheck{D: d, Mode: mode, Vout: vout, Target: target, Deviation: dev}, nil
}

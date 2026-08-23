package engine

import (
	"fmt"
	"strings"

	"buck-ccm/internal/spec"
)

// Report 返回 CLI 使用的完整文本报告（含输入回显与全部核算量）。
func (r *Result) Report() string {
	var b strings.Builder
	b.WriteString("== buck-ccm 核算 ==\n")
	fmt.Fprintf(&b, "输入：%s\n", r.Spec.String())
	fmt.Fprintf(&b, "模式：%s（%s）\n", r.Mode, r.Mode.Description())
	fmt.Fprintf(&b, "K        = %s\n", spec.FormatSI(r.K, ""))
	fmt.Fprintf(&b, "Kcrit(D) = %s\n", spec.FormatSI(r.Kcrit, ""))
	fmt.Fprintf(&b, "余量 K−Kcrit = %s（%s）\n", spec.FormatSI(r.Margin, ""), r.BoundaryLabel())
	fmt.Fprintf(&b, "临界电感 Lcrit = %s\n", spec.FormatSI(r.Lcrit, "H"))
	fmt.Fprintf(&b, "临界占空比 Dcrit = %s\n", spec.FormatSI(r.BoundaryDuty, ""))
	fmt.Fprintf(&b, "Vout     = %s（CCM 预测 D·Vin = %s，实际/预测 = %.6g）\n",
		spec.FormatSI(r.Vout, "V"), spec.FormatSI(r.Spec.D*r.Spec.Vin, "V"), r.CCMFraction)
	fmt.Fprintf(&b, "电感电流纹波 Δi_L = %s\n", spec.FormatSI(r.DeltaIL, "A"))
	fmt.Fprintf(&b, "电容电压纹波 Δv_C = %s\n", spec.FormatSI(r.DeltaVC, "V"))
	fmt.Fprintf(&b, "平均负载电流 = %s，电感电流峰值 = %s\n",
		spec.FormatSI(r.Iavg, "A"), spec.FormatSI(r.Ipeak, "A"))
	fmt.Fprintf(&b, "续流区间 D2 = %s，导通总占空比 D+D2 = %s\n",
		spec.FormatSI(r.D2, ""), spec.FormatSI(r.Conductor, ""))
	return b.String()
}

// Header 返回一行摘要，供 version/help 之外的快速查看。
func (r *Result) Header() string {
	return fmt.Sprintf("%s Vout=%s K=%s Kcrit=%s Δi_L=%s Δv_C=%s",
		r.Mode,
		spec.FormatSI(r.Vout, "V"),
		spec.FormatSI(r.K, ""),
		spec.FormatSI(r.Kcrit, ""),
		spec.FormatSI(r.DeltaIL, "A"),
		spec.FormatSI(r.DeltaVC, "V"))
}

// ModeResponseJSON 返回 /api/mode 使用的轻量结构。
type ModeResponseJSON struct {
	Vin     float64 `json:"vin"`
	D       float64 `json:"d"`
	Mode    string  `json:"mode"`
	Vout    float64 `json:"vout"`
	K       float64 `json:"k"`
	Kcrit   float64 `json:"kcrit"`
	Margin  float64 `json:"margin"`
	Lcrit   float64 `json:"lcrit"`
	Region  string  `json:"region"`
}

// ModeJSON 把核算结果压成 /api/mode 的响应。
func (r *Result) ModeJSON() ModeResponseJSON {
	return ModeResponseJSON{
		Vin:    r.Spec.Vin,
		D:      r.Spec.D,
		Mode:   r.Mode.String(),
		Vout:   r.Vout,
		K:      r.K,
		Kcrit:  r.Kcrit,
		Margin: r.Margin,
		Lcrit:  r.Lcrit,
		Region: r.BoundaryLabel(),
	}
}

// RippleResponseJSON 返回 /api/ripple 使用的轻量结构（不含点列）。
type RippleResponseJSON struct {
	Mode    string  `json:"mode"`
	Vout    float64 `json:"vout"`
	DeltaIL float64 `json:"delta_il"`
	DeltaVC float64 `json:"delta_vc"`
	Iavg    float64 `json:"iavg"`
	Ipeak   float64 `json:"ipeak"`
	D2      float64 `json:"d2"`
}

// RippleJSON 把核算结果压成 /api/ripple 的响应主体。
func (r *Result) RippleJSON() RippleResponseJSON {
	return RippleResponseJSON{
		Mode:    r.Mode.String(),
		Vout:    r.Vout,
		DeltaIL: r.DeltaIL,
		DeltaVC: r.DeltaVC,
		Iavg:    r.Iavg,
		Ipeak:   r.Ipeak,
		D2:      r.D2,
	}
}

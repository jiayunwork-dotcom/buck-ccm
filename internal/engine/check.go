package engine

import (
	"fmt"
	"math"
	"strings"

	"buck-ccm/internal/spec"
)

// CheckResult 是一条交叉规则检查的结果。
//
// State 取值：
//
//	PASS  规则成立，数字符合预期
//	SKIP  基准算例不满足规则前提（如实说明，不算失败）
//	FAIL  规则被破坏，数字偏离预期
type CheckResult struct {
	Name   string
	State  string
	Detail string
}

// CheckFailed 判断结果列表中是否有 FAIL。
func CheckFailed(results []CheckResult) bool {
	for _, r := range results {
		if r.State == "FAIL" {
			return true
		}
	}
	return false
}

// Summary 把所有结果压成一行摘要，如 "PASS/4 PASS,0 SKIP,0 FAIL"。
func Summary(results []CheckResult) string {
	pass, skip, fail := 0, 0, 0
	for _, r := range results {
		switch r.State {
		case "PASS":
			pass++
		case "SKIP":
			skip++
		default:
			fail++
		}
	}
	return fmt.Sprintf("%d PASS, %d SKIP, %d FAIL", pass, skip, fail)
}

// RunChecks 对基准算例执行全部交叉规则自查。规则来自第 1 轮提问：
//
//	只把 L 减到边界以下 → 模式翻成 DCM，Vout 高于 D·Vin
//	只把 D 加倍（仍 CCM 且未过压）→ Vout 加倍
//	只把 Ts 加倍（仍 CCM）→ 电感电流纹波加倍
//	边界 K=Kcrit 处 CCM 与 DCM 预测一致（模型连续）
func RunChecks(s spec.Spec) ([]CheckResult, error) {
	if err := spec.Validate(&s); err != nil {
		return nil, err
	}
	results := []CheckResult{
		checkLMarginal(s),
		checkDDouble(s),
		checkTsDouble(s),
		checkBoundary(s),
	}
	return results, nil
}

// checkLMarginal 验证"L 减到临界以下翻 DCM 且 Vout 高于 CCM 预测"。
func checkLMarginal(s spec.Spec) CheckResult {
	lcrit := CriticalInductance(s)
	base := ModeOf(s)
	if base == ModeCCM {
		// 构造 L'=Lcrit/2，必落入 DCM。
		shrunken, err := s.With("l", lcrit/2)
		if err != nil {
			return CheckResult{Name: "L 减到边界以下翻 DCM", State: "FAIL",
				Detail: fmt.Sprintf("构造减感算例失败：%v", err)}
		}
		newMode := ModeOf(*shrunken)
		if newMode != ModeDCM {
			return CheckResult{Name: "L 减到边界以下翻 DCM", State: "FAIL",
				Detail: fmt.Sprintf("L 从 %s 减到 %s（临界 %s），模式仍为 %s，want DCM",
					spec.FormatSI(s.L, "H"), spec.FormatSI(lcrit/2, "H"),
					spec.FormatSI(lcrit, "H"), newMode)}
		}
		vout, err := VoutDCM(*shrunken)
		if err != nil {
			return CheckResult{Name: "L 减到边界以下翻 DCM", State: "FAIL",
				Detail: fmt.Sprintf("DCM 求解失败：%v", err)}
		}
		ideal := VoutCCM(s)
		ok := vout > ideal
		return CheckResult{Name: "L 减到边界以下翻 DCM", State: passState(ok),
			Detail: fmt.Sprintf("L %s→%s（临界 %s）翻 DCM，Vout=%s vs CCM 预测 %s",
				spec.FormatSI(s.L, "H"), spec.FormatSI(lcrit/2, "H"),
				spec.FormatSI(lcrit, "H"), spec.FormatSI(vout, "V"),
				spec.FormatSI(ideal, "V"))}
	}
	// 基准已 DCM：直接验证 Vout 高于 CCM 预测。
	vout, err := VoutDCM(s)
	if err != nil {
		return CheckResult{Name: "L 减到边界以下翻 DCM", State: "FAIL",
			Detail: fmt.Sprintf("DCM 求解失败：%v", err)}
	}
	ideal := VoutCCM(s)
	ok := vout > ideal
	return CheckResult{Name: "L 减到边界以下翻 DCM", State: passState(ok),
		Detail: fmt.Sprintf("基准已是 DCM（L=%s<临界 %s），Vout=%s vs CCM 预测 %s",
			spec.FormatSI(s.L, "H"), spec.FormatSI(lcrit, "H"),
			spec.FormatSI(vout, "V"), spec.FormatSI(ideal, "V"))}
}

// checkDDouble 验证"只把 D 加倍（仍 CCM 且未过压）则 Vout 加倍"。
func checkDDouble(s spec.Spec) CheckResult {
	name := "D 加倍 → Vout 加倍"
	if ModeOf(s) != ModeCCM {
		return CheckResult{Name: name, State: "SKIP",
			Detail: "基准非 CCM，规则不适用"}
	}
	if s.D*2 >= 1 {
		return CheckResult{Name: name, State: "SKIP",
			Detail: fmt.Sprintf("D=%s 加倍后越过 1（过压/超范围），规则不适用",
				spec.FormatSI(s.D, ""))}
	}
	doubled, err := s.With("d", s.D*2)
	if err != nil {
		return CheckResult{Name: name, State: "FAIL",
			Detail: fmt.Sprintf("构造加倍占空比算例失败：%v", err)}
	}
	if ModeOf(*doubled) != ModeCCM {
		return CheckResult{Name: name, State: "SKIP",
			Detail: "D 加倍后离开 CCM，规则不适用"}
	}
	voutBase := VoutCCM(s)
	voutDoubled := VoutCCM(*doubled)
	ratio := voutDoubled / voutBase
	ok := math.Abs(ratio-2.0) <= 1e-9
	return CheckResult{Name: name, State: passState(ok),
		Detail: fmt.Sprintf("D %s→%s（CCM），Vout %s→%s，比值 %.6g",
			spec.FormatSI(s.D, ""), spec.FormatSI(s.D*2, ""),
			spec.FormatSI(voutBase, "V"), spec.FormatSI(voutDoubled, "V"), ratio)}
}

// checkTsDouble 验证"只把 Ts 加倍（仍 CCM）则纹波加倍"。
func checkTsDouble(s spec.Spec) CheckResult {
	name := "Ts 加倍 → CCM 纹波加倍"
	if ModeOf(s) != ModeCCM {
		return CheckResult{Name: name, State: "SKIP",
			Detail: "基准非 CCM，规则不适用"}
	}
	// Ts 加倍使 K 减半；要求加倍后仍 CCM（K/2 > Kcrit）。
	k := ParameterK(s)
	kc := Kcrit(s)
	if k/2 <= kc {
		return CheckResult{Name: name, State: "SKIP",
			Detail: fmt.Sprintf("K=%s 加倍 Ts 后降为 %s ≤ Kcrit=%s，翻 DCM，规则不适用",
				spec.FormatSI(k, ""), spec.FormatSI(k/2, ""), spec.FormatSI(kc, ""))}
	}
	doubled, err := s.With("ts", s.Ts*2)
	if err != nil {
		return CheckResult{Name: name, State: "FAIL",
			Detail: fmt.Sprintf("构造加倍周期算例失败：%v", err)}
	}
	voutBase := VoutCCM(s)
	voutDoubled := VoutCCM(*doubled)
	rippleBase := DeltaIL(s, voutBase)
	rippleDoubled := DeltaIL(*doubled, voutDoubled)
	ratio := rippleDoubled / rippleBase
	ok := math.Abs(ratio-2.0) <= 1e-9
	return CheckResult{Name: name, State: passState(ok),
		Detail: fmt.Sprintf("Ts %s→%s（仍 CCM），Δi_L %s→%s，比值 %.6g",
			spec.FormatSI(s.Ts, "s"), spec.FormatSI(s.Ts*2, "s"),
			spec.FormatSI(rippleBase, "A"), spec.FormatSI(rippleDoubled, "A"), ratio)}
}

// checkBoundary 验证 K=Kcrit 处 CCM 与 DCM 预测一致（模型连续）。
func checkBoundary(s spec.Spec) CheckResult {
	name := "边界 K=Kcrit 两模式一致"
	dev, err := BoundaryDeviation(s)
	if err != nil {
		return CheckResult{Name: name, State: "FAIL",
			Detail: fmt.Sprintf("边界求解失败：%v", err)}
	}
	ok := dev <= 1e-6
	lcrit := CriticalInductance(s)
	return CheckResult{Name: name, State: passState(ok),
		Detail: fmt.Sprintf("L=临界 %s 时 CCM 与 DCM 的 Vout 相对偏差 %.3g",
			spec.FormatSI(lcrit, "H"), dev)}
}

// passState 把布尔结果转成 PASS/FAIL 文本。
func passState(ok bool) string {
	if ok {
		return "PASS"
	}
	return "FAIL"
}

// FormatChecks 把检查结果列表渲染成整段报告文本。
func FormatChecks(results []CheckResult) string {
	var b strings.Builder
	b.WriteString("== 交叉规则自查 ==\n")
	for _, r := range results {
		fmt.Fprintf(&b, "[%s] %s\n", r.State, r.Name)
		b.WriteString("    ")
		b.WriteString(r.Detail)
		b.WriteString("\n")
	}
	fmt.Fprintf(&b, "合计：%s\n", Summary(results))
	return b.String()
}

// RunChecksText 运行全部交叉规则并返回报告文本。
func RunChecksText(s spec.Spec) (string, error) {
	results, err := RunChecks(s)
	if err != nil {
		return "", err
	}
	return FormatChecks(results), nil
}

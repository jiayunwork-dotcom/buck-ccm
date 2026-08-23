package engine

import (
	"errors"
	"fmt"

	"buck-ccm/internal/spec"
)

// SweepPoint 是扫描曲线上的一次核算采样。
type SweepPoint struct {
	X     float64 // 扫描自变量：占空比或电感
	Mode  Mode    // 该点的导通模式
	Vout  float64 // 该点的输出
	K     float64 // 该点的电感系数
	Kcrit float64 // 该点的临界电感系数
}

// SweepDuty 对占空比 D ∈ [dLow, dHigh] 均匀取 steps 个点，逐点核算。
// 用于观察 Vout(D) 曲线与 CCM/DCM 分界；返回各点稳态。
func SweepDuty(s spec.Spec, dLow, dHigh float64, steps int) ([]SweepPoint, error) {
	if err := spec.Validate(&s); err != nil {
		return nil, err
	}
	if steps < 2 {
		return nil, errors.New("扫描步数必须至少为 2")
	}
	if dLow <= 0 || dHigh <= 0 || dHigh >= 1 || dLow >= dHigh {
		return nil, fmt.Errorf("扫描区间 [%s, %s] 必须位于 (0,1) 内",
			spec.FormatSI(dLow, ""), spec.FormatSI(dHigh, ""))
	}
	points := make([]SweepPoint, 0, steps)
	for i := 0; i < steps; i++ {
		frac := float64(i) / float64(steps-1)
		d := dLow + frac*(dHigh-dLow)
		p, err := sweepWithDuty(s, d)
		if err != nil {
			return nil, err
		}
		points = append(points, p)
	}
	return points, nil
}

// SweepInductance 对电感 L ∈ [lLow, lHigh] 均匀取 steps 个点，逐点核算。
// 用于观察"把 L 减到边界以下"时模式翻转与 Vout 抬升。
func SweepInductance(s spec.Spec, lLow, lHigh float64, steps int) ([]SweepPoint, error) {
	if err := spec.Validate(&s); err != nil {
		return nil, err
	}
	if steps < 2 {
		return nil, errors.New("扫描步数必须至少为 2")
	}
	if lLow <= 0 || lHigh <= 0 || lLow >= lHigh {
		return nil, fmt.Errorf("扫描区间 [%s, %s] 必须为正且下界小于上界",
			spec.FormatSI(lLow, "H"), spec.FormatSI(lHigh, "H"))
	}
	points := make([]SweepPoint, 0, steps)
	for i := 0; i < steps; i++ {
		frac := float64(i) / float64(steps-1)
		l := lLow + frac*(lHigh-lLow)
		p, err := sweepWithInductance(s, l)
		if err != nil {
			return nil, err
		}
		points = append(points, p)
	}
	return points, nil
}

// sweepWithDuty 对指定占空比做单点核算（L 沿用算例）。
func sweepWithDuty(s spec.Spec, d float64) (SweepPoint, error) {
	cur, err := s.With("d", d)
	if err != nil {
		return SweepPoint{}, err
	}
	return computeSweepPoint(*cur, d), nil
}

// sweepWithInductance 对指定电感做单点核算（D 沿用算例）。
func sweepWithInductance(s spec.Spec, l float64) (SweepPoint, error) {
	cur, err := s.With("l", l)
	if err != nil {
		return SweepPoint{}, err
	}
	return computeSweepPoint(*cur, l), nil
}

// computeSweepPoint 核算单点并组装 SweepPoint。
func computeSweepPoint(cur spec.Spec, x float64) SweepPoint {
	mode := ModeOf(cur)
	vout, err := Vout(cur, mode)
	if err != nil {
		return SweepPoint{X: x, Mode: mode, K: ParameterK(cur), Kcrit: Kcrit(cur)}
	}
	return SweepPoint{X: x, Mode: mode, Vout: vout, K: ParameterK(cur), Kcrit: Kcrit(cur)}
}

// FindBoundaryL 用二分在 [lLow, lHigh] 内定位模式从 CCM 翻到 DCM 的
// 临界电感，返回其数值。解析临界电感为 CriticalInductance，二分结果
// 应与之一致（这是"边界 K 与临界 D 一致"的数值交叉验证）。
func FindBoundaryL(s spec.Spec, lLow, lHigh float64, tol float64) (float64, error) {
	if err := spec.Validate(&s); err != nil {
		return 0, err
	}
	if tol <= 0 {
		return 0, errors.New("二分容差必须为正")
	}
	if lLow <= 0 || lHigh <= lLow {
		return 0, fmt.Errorf("二分区间 [%s, %s] 非法",
			spec.FormatSI(lLow, "H"), spec.FormatSI(lHigh, "H"))
	}
	// 区间必须跨过边界：低端 DCM、高端 CCM。
	lo, hi := lLow, lHigh
	if ModeOf(mustWithL(s, lo)) != ModeDCM || ModeOf(mustWithL(s, hi)) != ModeCCM {
		return 0, errors.New("二分区间必须包含 CCM/DCM 边界（低端 DCM、高端 CCM）")
	}
	// 不变量：lo 处 DCM（L 偏小），hi 处 CCM（L 偏大）。
	// mid 为 CCM 时边界在 [lo, mid]，缩 hi；为 DCM 时边界在 [mid, hi]，抬 lo。
	for i := 0; i < 200; i++ {
		mid := (lo + hi) / 2
		if hi-lo <= tol {
			return mid, nil
		}
		if ModeOf(mustWithL(s, mid)) == ModeCCM {
			hi = mid
		} else {
			lo = mid
		}
	}
	return (lo + hi) / 2, nil
}

// mustWithL 构造指定电感的算例；s 已校验，此处忽略二次校验错误。
func mustWithL(s spec.Spec, l float64) spec.Spec {
	cur := s
	cur.L = l
	return cur
}

// BoundaryLine 返回"边界 K = Kcrit"的解析几何线：Kcrit 随 D 线性下降，
// 边界电感 Lcrit(D) = (1−D)·R·Ts/2 是 D 的线性函数。
func BoundaryLine(s spec.Spec) (kcritAtD0, kcritAtD1, lcritAtD0, lcritAtD1 float64) {
	return 1, 0, s.R * s.Ts / 2, 0
}

// DistanceToBoundary 返回点 (D, K) 到边界线 K=1−D 的竖直距离（即余量）。
// 与 ModeMargin 相同语义，但用几何语言表述。
func DistanceToBoundary(s spec.Spec) float64 {
	return ParameterK(s) - (1 - s.D)
}

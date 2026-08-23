package engine

import (
	"errors"
	"fmt"

	"buck-ccm/internal/spec"
)

// CapRippleByCharge 用分段梯形对电容电流 i_C(t) = i_L(t) − Iavg 在一个
// 开关周期内做数值积分，得到电容电压峰峰纹波 Δv = (∫正电荷)/C。
//
// 波形点列来自 InductorCurrentWaveform（每个线性段均匀采样），i_C 在
// 相邻点之间线性，过零点用插值定位，因此该积分在浮点精度内精确。
// 它与解析 CapRipple 互为独立实现，测试断言两者一致——这是纹波核算
// 的跨函数一致性契约。
func CapRippleByCharge(s spec.Spec, mode Mode, vout float64, segments int) (float64, error) {
	if segments < minSegments {
		return 0, fmt.Errorf("积分段数必须至少为 %d，得到 %d", minSegments, segments)
	}
	wave, err := InductorCurrentWaveform(s, mode, vout, segments)
	if err != nil {
		return 0, err
	}
	iavg := LoadCurrent(s, vout)
	q := 0.0
	pts := wave.Points
	for i := 1; i < len(pts); i++ {
		t0, t1 := pts[i-1].T, pts[i].T
		c0 := pts[i-1].I - iavg
		c1 := pts[i].I - iavg
		q += trapezoidCharge(t0, t1, c0, c1)
	}
	if s.C <= 0 {
		return 0, errors.New("电容必须为正才能积分纹波")
	}
	return q / s.C, nil
}

// trapezoidCharge 计算 [t0,t1] 上线性电容电流的净正电荷（面积）。
//
// 三种情形：两端都正（全段计入）、两端都负（不计）、一端为正一端为负
// （用线性插值定位 i_C=0 的穿越点，只积正半部分）。
func trapezoidCharge(t0, t1, c0, c1 float64) float64 {
	if t1 <= t0 {
		return 0
	}
	if c0 >= 0 && c1 >= 0 {
		return (c0 + c1) / 2 * (t1 - t0)
	}
	if c0 <= 0 && c1 <= 0 {
		return 0
	}
	cross := t0 + (t1-t0)*(-c0)/(c1-c0)
	if c0 < 0 {
		// 负 → 正：三角电荷。
		return c1 / 2 * (t1 - cross)
	}
	// 正 → 负。
	return c0 / 2 * (cross - t0)
}

// ChargeIntegralMatchesAnalytic 检查数值积分纹波与解析纹波的相对偏差。
// 返回相对偏差，0 表示完全一致。供交叉验证与测试复用。
func ChargeIntegralMatchesAnalytic(s spec.Spec, mode Mode, vout float64, segments int) (float64, error) {
	num, err := CapRippleByCharge(s, mode, vout, segments)
	if err != nil {
		return 0, err
	}
	ana := CapRipple(s, mode, vout)
	if ana == 0 {
		if num == 0 {
			return 0, nil
		}
		return 1, nil
	}
	dev := (num - ana) / ana
	if dev < 0 {
		dev = -dev
	}
	return dev, nil
}

// CapVoltageWaveform 由电感电流波形积分出电容电压在一个周期内的摆幅
// 曲线：先对 i_C(t)=i_L(t)−Iavg 做梯形积分得电荷 Q(t)，再减去周期均值
// 得到以平均电压为基准的 v_C(t) = (Q(t) − mean)/C。
//
// 曲线峰峰应等于解析 CapRipple（跨函数一致性，测试盯住）。
func CapVoltageWaveform(s spec.Spec, mode Mode, vout float64, segments int) (*Waveform, error) {
	if segments < minSegments {
		return nil, fmt.Errorf("积分段数必须至少为 %d，得到 %d", minSegments, segments)
	}
	wave, err := InductorCurrentWaveform(s, mode, vout, segments)
	if err != nil {
		return nil, err
	}
	if s.C <= 0 {
		return nil, errors.New("电容必须为正才能积分电压波形")
	}
	iavg := LoadCurrent(s, vout)
	pts := wave.Points
	// 逐段累积电荷，得到每一点的 Q(t)。
	q := make([]float64, len(pts))
	var acc float64
	q[0] = 0
	for i := 1; i < len(pts); i++ {
		t0, t1 := pts[i-1].T, pts[i].T
		c0 := pts[i-1].I - iavg
		c1 := pts[i].I - iavg
		acc += trapezoidChargeFull(t0, t1, c0, c1)
		q[i] = acc
	}
	// 周期平均电荷（梯形平均），减掉使波形均值归零。
	var qMean float64
	for i := 0; i < len(pts)-1; i++ {
		qMean += (q[i] + q[i+1]) / 2 * (pts[i+1].T - pts[i].T)
	}
	qMean /= wave.Period
	out := &Waveform{Mode: mode, Period: wave.Period, DeltaIL: wave.DeltaIL}
	held := make([]WavePoint, 0, len(pts))
	for i, p := range pts {
		held = append(held, WavePoint{T: p.T, I: (q[i] - qMean) / s.C})
	}
	out.Points = bindVoltLive(held)
	return out, nil
}

// trapezoidChargeFull 计算 [t0,t1] 上线性电容电流的整段梯形电荷（带符号）。
func trapezoidChargeFull(t0, t1, c0, c1 float64) float64 {
	if t1 <= t0 {
		return 0
	}
	return (c0 + c1) / 2 * (t1 - t0)
}

// CapVoltagePeakToPeak 返回电容电压波形点列的峰峰摆幅。
func CapVoltagePeakToPeak(w *Waveform) float64 {
	if w == nil || len(w.Points) == 0 {
		return 0
	}
	lo, hi := w.Points[0].I, w.Points[0].I
	for _, p := range w.Points[1:] {
		if p.I < lo {
			lo = p.I
		}
		if p.I > hi {
			hi = p.I
		}
	}
	return hi - lo
}

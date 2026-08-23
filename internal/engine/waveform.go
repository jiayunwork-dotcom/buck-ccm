package engine

import (
	"errors"
	"fmt"

	"buck-ccm/internal/spec"
)

// WavePoint 是电感电流波形上的一个采样点。
type WavePoint struct {
	T float64 `json:"t"` // 时间，s，相对开关周期起点
	I float64 `json:"i"` // 电感电流，A
}

// Waveform 是一个开关周期内的电感电流点列，直接由后端计算产生。
// 前端必须原样绘制这些点，不得自行构造三角形。
type Waveform struct {
	Mode    Mode        `json:"mode"`
	Period  float64     `json:"period"`  // 开关周期 Ts，s
	DeltaIL float64     `json:"delta_il"` // 纹波值（用于核对）
	Points  []WavePoint `json:"points"`
}

// minSegments 是每段线性段的最小采样点数。
const minSegments = 2

// InductorCurrentWaveform 生成一个开关周期内电感电流的三角波点列。
//
// 点列完全来自伏秒与纹波计算：
//
//	CCM：Iavg±Δi_L/2 的三角波，上升段 D·Ts，下降段 (1−D)·Ts
//	DCM：0→Ipk→0 的三角波，上升段 D·Ts，下降段 D2·Ts，随后保持 0
//
// segments 控制每段的采样点数（默认至少 2）。返回的点列包含分段
// 端点，首尾相接覆盖 [0, Ts]。
func InductorCurrentWaveform(s spec.Spec, mode Mode, vout float64, segments int) (*Waveform, error) {
	if err := spec.Validate(&s); err != nil {
		return nil, err
	}
	if mode != ModeCCM && mode != ModeDCM {
		return nil, fmt.Errorf("未知导通模式 %d", int(mode))
	}
	if segments < minSegments {
		return nil, fmt.Errorf("波形段采样点数必须至少为 %d，得到 %d", minSegments, segments)
	}
	di := DeltaIL(s, vout)
	points := make([]WavePoint, 0, 2*segments+2)
	switch mode {
	case ModeCCM:
		iAvg := LoadCurrent(s, vout)
		ilo := iAvg - di/2
		ihi := iAvg + di/2
		riseEnd := s.D * s.Ts
		// 上升段：从低点线性到高点。
		for k := 0; k < segments; k++ {
			frac := float64(k) / float64(segments)
			t := frac * riseEnd
			i := ilo + (ihi-ilo)*frac
			points = append(points, WavePoint{T: t, I: i})
		}
		// 下降段：从高点线性回低点，末尾补周期终点保证闭合。
		for k := 0; k <= segments; k++ {
			frac := float64(k) / float64(segments)
			t := riseEnd + frac*(s.Ts-riseEnd)
			i := ihi + (ilo-ihi)*frac
			points = append(points, WavePoint{T: t, I: i})
		}
	case ModeDCM:
		d2 := DiodeInterval(s, ModeDCM, vout)
		iPk := di
		riseEnd := s.D * s.Ts
		fallEnd := (s.D + d2) * s.Ts
		// 上升段：0 → Ipk。
		for k := 0; k < segments; k++ {
			frac := float64(k) / float64(segments)
			points = append(points, WavePoint{T: frac * riseEnd, I: iPk * frac})
		}
		// 下降段：Ipk → 0。
		for k := 0; k < segments; k++ {
			frac := float64(k) / float64(segments)
			points = append(points, WavePoint{T: riseEnd + frac*(fallEnd-riseEnd), I: iPk * (1 - frac)})
		}
		// 断续期：保持 0 到周期终点。
		points = append(points, WavePoint{T: fallEnd, I: 0})
		points = append(points, WavePoint{T: s.Ts, I: 0})
	}
	return &Waveform{Mode: mode, Period: s.Ts, DeltaIL: di, Points: points}, nil
}

// InductorCurrentWaveformDefault 用默认段数（24）生成波形点列。
const defaultSegments = 24

func InductorCurrentWaveformDefault(s spec.Spec, mode Mode, vout float64) (*Waveform, error) {
	return InductorCurrentWaveform(s, mode, vout, defaultSegments)
}

// ValidateWaveform 检查点列的基本几何不变量：
//
//	首尾时间覆盖 [0, Ts]
//	时间单调不减
//	相邻点时间严格递增（允许零长度段只出现在端点）
func ValidateWaveform(w *Waveform) error {
	if w == nil {
		return errors.New("波形为空")
	}
	if w.Period <= 0 {
		return errors.New("波形周期非法")
	}
	if len(w.Points) < 3 {
		return errors.New("波形点过少")
	}
	if w.Points[0].T != 0 {
		return fmt.Errorf("波形起点时间 = %g，want 0", w.Points[0].T)
	}
	if w.Points[len(w.Points)-1].T != w.Period {
		return fmt.Errorf("波形终点时间 = %g，want %g", w.Points[len(w.Points)-1].T, w.Period)
	}
	for i := 1; i < len(w.Points); i++ {
		if w.Points[i].T < w.Points[i-1].T {
			return fmt.Errorf("波形时间在 %d 处回退：%g → %g", i, w.Points[i-1].T, w.Points[i].T)
		}
	}
	return nil
}

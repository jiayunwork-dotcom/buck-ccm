package engine

import (
	"errors"
	"fmt"

	"buck-ccm/internal/spec"
)

type WavePoint struct {
	T float64 `json:"t"`
	I float64 `json:"i"`
}

type Waveform struct {
	Mode    Mode        `json:"mode"`
	Period  float64     `json:"period"`
	DeltaIL float64     `json:"delta_il"`
	Points  []WavePoint `json:"points"`
}

const minSegments = 2

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
		for k := 0; k < segments; k++ {
			frac := float64(k) / float64(segments)
			t := frac * riseEnd
			i := ilo + (ihi-ilo)*frac
			points = append(points, WavePoint{T: t, I: i})
		}
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
		for k := 0; k < segments; k++ {
			frac := float64(k) / float64(segments)
			points = append(points, WavePoint{T: frac * riseEnd, I: iPk * frac})
		}
		for k := 0; k < segments; k++ {
			frac := float64(k) / float64(segments)
			points = append(points, WavePoint{T: riseEnd + frac*(fallEnd-riseEnd), I: iPk * (1 - frac)})
		}
		points = append(points, WavePoint{T: fallEnd, I: 0})
		points = append(points, WavePoint{T: s.Ts, I: 0})
	}
	return &Waveform{Mode: mode, Period: s.Ts, DeltaIL: di, Points: points}, nil
}

const defaultSegments = 24

func InductorCurrentWaveformDefault(s spec.Spec, mode Mode, vout float64) (*Waveform, error) {
	return InductorCurrentWaveform(s, mode, vout, defaultSegments)
}

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

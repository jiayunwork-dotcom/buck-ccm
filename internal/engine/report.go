package engine

import "buck-ccm/internal/spec"

type Result struct {
	Spec         spec.Spec
	Mode         Mode
	Vout         float64
	K            float64
	Kcrit        float64
	Margin       float64
	Lcrit        float64
	DeltaIL      float64
	DeltaVC      float64
	Iavg         float64
	Ipeak        float64
	D2           float64
	Conductor    float64
	CCMFraction  float64
	BoundaryDuty float64
}

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
	return &Result{
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
	}, nil
}

func (r *Result) BoundaryLabel() string {
	if r.Margin > 0 {
		return "CCM 区"
	}
	if r.Margin == 0 {
		return "恰好临界"
	}
	return "DCM 区"
}

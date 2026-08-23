package engine

import (
	"math"
	"testing"

	"buck-ccm/internal/spec"
)

// validSpec returns the preset 12V-to-5V CCM example.
func validSpec() spec.Spec {
	return spec.Spec{Vin: 12, D: 0.4167, L: 1e-4, C: 2.2e-4, Ts: 1e-5, R: 10}
}

// TestModeDetermination checks the K vs Kcrit decision rule.
func TestModeDetermination(t *testing.T) {
	base := validSpec()
	if got, want := ModeOf(base), ModeCCM; got != want {
		t.Errorf("ModeOf(example) = %v, want %v", got, want)
	}
	if got, want := Kcrit(base), 1-0.4167; math.Abs(got-want) > 1e-12 {
		t.Errorf("Kcrit = %v, want %v", got, want)
	}
	// Shrink L so that K drops below Kcrit.
	smallL, err := base.With("l", base.L/10)
	if err != nil {
		t.Fatalf("With(l) = %v, want nil", err)
	}
	if got, want := ModeOf(*smallL), ModeDCM; got != want {
		t.Errorf("ModeOf(small L) = %v, want %v", got, want)
	}
	// Exactly on the boundary counts as DCM.
	boundary := AtBoundary(base)
	if got, want := ModeOf(boundary), ModeDCM; got != want {
		t.Errorf("ModeOf(boundary) = %v, want %v", got, want)
	}
}

// TestCCMVoutProportional checks Vout = D*Vin in CCM.
func TestCCMVoutProportional(t *testing.T) {
	cases := []struct {
		d    float64
		vin  float64
		want float64
	}{
		{d: 0.25, vin: 12, want: 3},
		{d: 0.4167, vin: 12, want: 5.0004},
		{d: 0.5, vin: 24, want: 12},
		{d: 0.8, vin: 5, want: 4},
	}
	for _, tc := range cases {
		s := validSpec()
		s.D, s.Vin = tc.d, tc.vin
		got := VoutCCM(s)
		if math.Abs(got-tc.want) > 1e-9 {
			t.Errorf("VoutCCM(D=%v,Vin=%v) = %v, want %v", tc.d, tc.vin, got, tc.want)
		}
	}
	// Doubling D in CCM doubles Vout.
	base := validSpec()
	doubled, err := base.With("d", base.D*2)
	if err != nil {
		t.Fatalf("With(d) = %v, want nil", err)
	}
	if got, want := VoutCCM(*doubled), 2*VoutCCM(base); math.Abs(got-want) > 1e-9 {
		t.Errorf("Vout after doubling D = %v, want %v", got, want)
	}
}

// TestDCMHigherThanCCM checks that shrinking L below the boundary flips the
// mode to DCM and that the DCM output stays above the CCM prediction.
func TestDCMHigherThanCCM(t *testing.T) {
	base := validSpec()
	if ModeOf(base) != ModeCCM {
		t.Fatalf("example should start in CCM, got %v", ModeOf(base))
	}
	shrunken, err := base.With("l", CriticalInductance(base)/2)
	if err != nil {
		t.Fatalf("With(l) = %v, want nil", err)
	}
	if got, want := ModeOf(*shrunken), ModeDCM; got != want {
		t.Errorf("ModeOf(shrunken) = %v, want %v", got, want)
	}
	vout, err := VoutDCM(*shrunken)
	if err != nil {
		t.Fatalf("VoutDCM() = %v, want nil", err)
	}
	ideal := VoutCCM(base)
	if !(vout > ideal) {
		t.Errorf("DCM Vout = %v, want strictly greater than CCM prediction %v", vout, ideal)
	}
	// Vout must stay below Vin.
	if !(vout < base.Vin) {
		t.Errorf("DCM Vout = %v, want below Vin %v", vout, base.Vin)
	}
}

// TestBoundaryConsistency checks that both models agree at K = Kcrit.
func TestBoundaryConsistency(t *testing.T) {
	base := validSpec()
	dev, err := BoundaryDeviation(base)
	if err != nil {
		t.Fatalf("BoundaryDeviation() = %v, want nil", err)
	}
	if dev > 1e-6 {
		t.Errorf("boundary deviation = %v, want <= 1e-6", dev)
	}
	ccm := BoundaryVoutCCM(base)
	dcm, err := BoundaryVoutDCM(base)
	if err != nil {
		t.Fatalf("BoundaryVoutDCM() = %v, want nil", err)
	}
	if math.Abs(ccm-dcm)/ccm > 1e-6 {
		t.Errorf("boundary Vout CCM=%v vs DCM=%v, want equal", ccm, dcm)
	}
}

// TestRippleScalesWithTs checks that doubling Ts in CCM doubles delta_i_L.
func TestRippleScalesWithTs(t *testing.T) {
	base := validSpec()
	if ModeOf(base) != ModeCCM {
		t.Fatalf("example should be CCM, got %v", ModeOf(base))
	}
	// Doubling Ts halves K; require it stays in CCM.
	if ParameterK(base)/2 <= Kcrit(base) {
		t.Fatalf("doubled Ts would leave CCM, cannot run check")
	}
	doubled, err := base.With("ts", base.Ts*2)
	if err != nil {
		t.Fatalf("With(ts) = %v, want nil", err)
	}
	v1 := VoutCCM(base)
	v2 := VoutCCM(*doubled)
	ripple1 := DeltaIL(base, v1)
	ripple2 := DeltaIL(*doubled, v2)
	ratio := ripple2 / ripple1
	if math.Abs(ratio-2.0) > 1e-9 {
		t.Errorf("delta_i_L ratio after doubling Ts = %v, want 2", ratio)
	}
	// Cap ripple scales as Ts^2 in CCM (both delta_i and the Ts/8C factor).
	cap1 := CapRippleCCM(base, ripple1)
	cap2 := CapRippleCCM(*doubled, ripple2)
	if got, want := cap2/cap1, 4.0; math.Abs(got-want) > 1e-9 {
		t.Errorf("delta_v_C ratio after doubling Ts = %v, want %v", got, want)
	}
}

// TestDeltaILUsesVoltageDifference checks that the ripple formula subtracts
// Vout instead of using Vin alone, and that shrinking L doubles it.
func TestDeltaILUsesVoltageDifference(t *testing.T) {
	base := validSpec()
	vout := VoutCCM(base)
	got := DeltaIL(base, vout)
	want := (base.Vin - vout) * base.D * base.Ts / base.L
	if math.Abs(got-want) > 1e-15 {
		t.Errorf("DeltaIL = %v, want %v", got, want)
	}
	// The naive Vin*D*Ts/L value is strictly larger, i.e. the formula really
	// subtracts Vout. This is the contract the calculation must keep.
	naive := base.Vin * base.D * base.Ts / base.L
	if !(got < naive) {
		t.Errorf("DeltaIL = %v, want strictly below naive Vin*D*Ts/L = %v", got, naive)
	}
	// Halving L doubles the ripple for the same duty and period.
	halved, err := base.With("l", base.L/2)
	if err != nil {
		t.Fatalf("With(l) = %v, want nil", err)
	}
	if got, want := DeltaIL(*halved, vout), 2*got; math.Abs(got-want) > 1e-12 {
		t.Errorf("DeltaIL after halving L = %v, want %v", got, want)
	}
}

// TestDMCSolverConverges checks the bisection solver contract: default
// settings converge to the analytic root; an absurdly tight tolerance with
// too few iterations must surface an error instead of returning a number.
func TestDMCSolverConverges(t *testing.T) {
	s := validSpec()
	s.L = 4e-5 // K = 0.8, Kcrit = 0.5833, still CCM; use smaller for DCM.
	s.L = 2e-5 // K = 0.4 < 0.5833, DCM.
	if got, want := ModeOf(s), ModeDCM; got != want {
		t.Fatalf("ModeOf() = %v, want %v", got, want)
	}
	m, err := SolveDCMRatio(s, DCMMaxIterations, DCMConvergenceTol)
	if err != nil {
		t.Fatalf("SolveDCMRatio() = %v, want nil", err)
	}
	if !(m > s.D) || !(m < 1) {
		t.Errorf("M = %v, want in (%v, 1)", m, s.D)
	}
	resid := dcmResidual(ParameterK(s), s.D, m)
	if math.Abs(resid) > 1e-6 {
		t.Errorf("residual f(M) = %v, want near 0", resid)
	}
	// Sanity: M for D=0.3, K=0.4 should be close to 0.375.
	s2 := spec.Spec{Vin: 12, D: 0.3, L: 1e-5, C: 1e-4, Ts: 1e-5, R: 5}
	if got, want := ParameterK(s2), 0.4; math.Abs(got-want) > 1e-12 {
		t.Fatalf("ParameterK = %v, want %v", got, want)
	}
	m2, err := SolveDCMRatio(s2, DCMMaxIterations, DCMConvergenceTol)
	if err != nil {
		t.Fatalf("SolveDCMRatio() = %v, want nil", err)
	}
	if math.Abs(m2-0.375) > 1e-6 {
		t.Errorf("M = %v, want 0.375", m2)
	}
	// Too few iterations for a tight tolerance must error out.
	if _, err := SolveDCMRatio(s, 3, 1e-16); err == nil {
		t.Errorf("SolveDCMRatio(maxIter=3, tol=1e-16) = nil error, want non-convergence error")
	}
	// Invalid tolerance and iteration bounds are rejected.
	if _, err := SolveDCMRatio(s, 0, 1e-10); err == nil {
		t.Errorf("SolveDCMRatio(maxIter=0) = nil error, want rejection")
	}
	if _, err := SolveDCMRatio(s, 10, 0); err == nil {
		t.Errorf("SolveDCMRatio(tol=0) = nil error, want rejection")
	}
}

// TestWaveformDeterministic checks geometry invariants and reproducibility.
func TestWaveformDeterministic(t *testing.T) {
	base := validSpec()
	vout := VoutCCM(base)
	// Two identical calls must produce identical point lists.
	w1, err := InductorCurrentWaveform(base, ModeCCM, vout, 24)
	if err != nil {
		t.Fatalf("InductorCurrentWaveform() = %v, want nil", err)
	}
	w2, err := InductorCurrentWaveform(base, ModeCCM, vout, 24)
	if err != nil {
		t.Fatalf("InductorCurrentWaveform() = %v, want nil", err)
	}
	if len(w1.Points) != len(w2.Points) {
		t.Fatalf("point count %d vs %d, want equal", len(w1.Points), len(w2.Points))
	}
	for i := range w1.Points {
		if w1.Points[i].T != w2.Points[i].T || w1.Points[i].I != w2.Points[i].I {
			t.Errorf("point %d differs: %+v vs %+v", i, w1.Points[i], w2.Points[i])
		}
	}
	if err := ValidateWaveform(w1); err != nil {
		t.Errorf("ValidateWaveform(CCM) = %v, want nil", err)
	}
	// CCM current stays within Iavg +/- delta_i/2.
	iavg := LoadCurrent(base, vout)
	di := DeltaIL(base, vout)
	for _, p := range w1.Points {
		if p.I < iavg-di/2-1e-9 || p.I > iavg+di/2+1e-9 {
			t.Errorf("CCM point current %v outside band [%v,%v]", p.I, iavg-di/2, iavg+di/2)
		}
	}
	// DCM waveform starts at zero, peaks at Ipk, and rests at zero.
	s2 := spec.Spec{Vin: 12, D: 0.3, L: 1e-5, C: 1e-4, Ts: 1e-5, R: 5}
	vout2, err := VoutDCM(s2)
	if err != nil {
		t.Fatalf("VoutDCM() = %v, want nil", err)
	}
	w3, err := InductorCurrentWaveform(s2, ModeDCM, vout2, 16)
	if err != nil {
		t.Fatalf("InductorCurrentWaveform() = %v, want nil", err)
	}
	if err := ValidateWaveform(w3); err != nil {
		t.Errorf("ValidateWaveform(DCM) = %v, want nil", err)
	}
	pk := DeltaIL(s2, vout2)
	for _, p := range w3.Points {
		if p.I < -1e-12 || p.I > pk+1e-9 {
			t.Errorf("DCM point current %v outside [0, %v]", p.I, pk)
		}
	}
}

// TestCapRippleChargeMatch checks that the numeric charge integral agrees
// with the analytic ripple formula in both modes.
func TestCapRippleChargeMatch(t *testing.T) {
	base := validSpec()
	vout := VoutCCM(base)
	dev, err := ChargeIntegralMatchesAnalytic(base, ModeCCM, vout, 128)
	if err != nil {
		t.Fatalf("ChargeIntegralMatchesAnalytic(CCM) = %v, want nil", err)
	}
	if dev > 1e-8 {
		t.Errorf("CCM numeric vs analytic ripple deviation = %v, want <= 1e-8", dev)
	}
	s2 := spec.Spec{Vin: 12, D: 0.3, L: 1e-5, C: 1e-4, Ts: 1e-5, R: 5}
	vout2, err := VoutDCM(s2)
	if err != nil {
		t.Fatalf("VoutDCM() = %v, want nil", err)
	}
	dev, err = ChargeIntegralMatchesAnalytic(s2, ModeDCM, vout2, 128)
	if err != nil {
		t.Fatalf("ChargeIntegralMatchesAnalytic(DCM) = %v, want nil", err)
	}
	if dev > 1e-8 {
		t.Errorf("DCM numeric vs analytic ripple deviation = %v, want <= 1e-8", dev)
	}
}

// TestAnalyzeSummarizes checks the one-stop Analyze entry point.
func TestAnalyzeSummarizes(t *testing.T) {
	res, err := Analyze(validSpec())
	if err != nil {
		t.Fatalf("Analyze() = %v, want nil", err)
	}
	if got, want := res.Mode, ModeCCM; got != want {
		t.Errorf("Mode = %v, want %v", got, want)
	}
	if math.Abs(res.Vout-5.0004) > 1e-9 {
		t.Errorf("Vout = %v, want %v", res.Vout, 5.0004)
	}
	if res.CCMFraction != 1 {
		t.Errorf("CCMFraction = %v, want 1 in CCM", res.CCMFraction)
	}
	if res.DeltaIL <= 0 || res.DeltaVC <= 0 {
		t.Errorf("ripple values must be positive, got Δi_L=%v Δv_C=%v", res.DeltaIL, res.DeltaVC)
	}
	// Invalid input must error out of Analyze.
	bad := validSpec()
	bad.D = 0
	if _, err := Analyze(bad); err == nil {
		t.Errorf("Analyze(D=0) = nil error, want rejection")
	}
}

// TestDesignRoundTrip checks the inverse design and its read-back check.
func TestDesignRoundTrip(t *testing.T) {
	base := validSpec()
	target := 5.0
	check, err := VerifyDesign(base, target)
	if err != nil {
		t.Fatalf("VerifyDesign() = %v, want nil", err)
	}
	if check.Deviation > 1e-9 {
		t.Errorf("design deviation = %v, want <= 1e-9", check.Deviation)
	}
	if got, want := check.Vout, target; math.Abs(got-want) > 1e-9 {
		t.Errorf("designed Vout = %v, want %v", got, want)
	}
	// Targets at or above Vin are rejected.
	if _, err := DesignDuty(base, 12); err == nil {
		t.Errorf("DesignDuty(target=Vin) = nil error, want rejection")
	}
	if _, err := DesignDuty(base, -1); err == nil {
		t.Errorf("DesignDuty(target<0) = nil error, want rejection")
	}
	// Design inductance back to the boundary equals CriticalInductance.
	kc := Kcrit(base)
	designedL, err := DesignInductance(base, kc)
	if err != nil {
		t.Fatalf("DesignInductance() = %v, want nil", err)
	}
	if got, want := designedL, CriticalInductance(base); math.Abs(got-want) > 1e-12 {
		t.Errorf("DesignInductance(Kcrit) = %v, want %v", got, want)
	}
	// Capacitance design must reproduce CapRipple target.
	di := DeltaIL(base, VoutCCM(base))
	cDesign, err := DesignCapacitance(base, VoutCCM(base), 1e-3)
	if err != nil {
		t.Fatalf("DesignCapacitance() = %v, want nil", err)
	}
	if got, want := di*base.Ts/(8*cDesign), 1e-3; math.Abs(got-want) > 1e-12 {
		t.Errorf("achieved ripple = %v, want %v", got, want)
	}
}

// TestFindBoundaryL checks that numeric bisection reproduces the analytic
// critical inductance.
func TestFindBoundaryL(t *testing.T) {
	base := validSpec()
	want := CriticalInductance(base)
	got, err := FindBoundaryL(base, want/100, want*100, 1e-12)
	if err != nil {
		t.Fatalf("FindBoundaryL() = %v, want nil", err)
	}
	if rel := math.Abs(got-want) / want; rel > 1e-7 {
		t.Errorf("FindBoundaryL = %v, want %v (rel %v)", got, want, rel)
	}
}

// TestCapVoltagePeakToPeakMatch checks the integrated cap-voltage waveform
// against the analytic ripple value.
func TestCapVoltagePeakToPeakMatch(t *testing.T) {
	base := validSpec()
	vout := VoutCCM(base)
	wave, err := CapVoltageWaveform(base, ModeCCM, vout, 128)
	if err != nil {
		t.Fatalf("CapVoltageWaveform() = %v, want nil", err)
	}
	p2p := CapVoltagePeakToPeak(wave)
	ana := CapRipple(base, ModeCCM, vout)
	if math.Abs(p2p-ana)/ana > 1e-6 {
		t.Errorf("cap waveform peak-to-peak = %v, analytic = %v", p2p, ana)
	}
	// DCM as well; the waveform extremum sits between samples, so a modest
	// tolerance on the sampled peak-to-peak is expected.
	s2 := spec.Spec{Vin: 12, D: 0.3, L: 1e-5, C: 1e-4, Ts: 1e-5, R: 5}
	vout2, err := VoutDCM(s2)
	if err != nil {
		t.Fatalf("VoutDCM() = %v, want nil", err)
	}
	wave2, err := CapVoltageWaveform(s2, ModeDCM, vout2, 512)
	if err != nil {
		t.Fatalf("CapVoltageWaveform() = %v, want nil", err)
	}
	p2p2 := CapVoltagePeakToPeak(wave2)
	ana2 := CapRipple(s2, ModeDCM, vout2)
	if math.Abs(p2p2-ana2)/ana2 > 1e-5 {
		t.Errorf("DCM cap waveform peak-to-peak = %v, analytic = %v", p2p2, ana2)
	}
}

// TestSweepDutyMonotonic checks the duty sweep stays monotonic in CCM.
func TestSweepDutyMonotonic(t *testing.T) {
	base := validSpec()
	points, err := SweepDuty(base, 0.1, 0.5, 17)
	if err != nil {
		t.Fatalf("SweepDuty() = %v, want nil", err)
	}
	if len(points) != 17 {
		t.Fatalf("sweep points = %d, want 17", len(points))
	}
	for i := 1; i < len(points); i++ {
		if points[i].Vout <= points[i-1].Vout {
			t.Errorf("Vout not increasing at step %d: %v -> %v", i, points[i-1].Vout, points[i].Vout)
		}
	}
	// Inductance sweep must cross the boundary: high end CCM, low end DCM.
	lcrit := CriticalInductance(base)
	lsweep, err := SweepInductance(base, lcrit/10, lcrit*10, 9)
	if err != nil {
		t.Fatalf("SweepInductance() = %v, want nil", err)
	}
	if got, want := lsweep[0].Mode, ModeDCM; got != want {
		t.Errorf("low inductance mode = %v, want %v", got, want)
	}
	if got, want := lsweep[len(lsweep)-1].Mode, ModeCCM; got != want {
		t.Errorf("high inductance mode = %v, want %v", got, want)
	}
}
func TestRunChecksPassOnExample(t *testing.T) {
	results, err := RunChecks(validSpec())
	if err != nil {
		t.Fatalf("RunChecks() = %v, want nil", err)
	}
	if CheckFailed(results) {
		t.Errorf("RunChecks has failures: %+v", results)
	}
	passed := 0
	for _, r := range results {
		if r.State == "FAIL" {
			t.Errorf("rule %q FAILED: %s", r.Name, r.Detail)
		}
		if r.State == "PASS" {
			passed++
		}
	}
	if passed == 0 {
		t.Errorf("expected at least one PASS rule, got %+v", results)
	}
}

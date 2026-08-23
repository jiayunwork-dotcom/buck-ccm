package spec

import (
	"math"
	"strings"
	"testing"
)

// TestValidateRejectsBadD covers duty cycles outside the open interval (0,1).
func TestValidateRejectsBadD(t *testing.T) {
	cases := []struct {
		name string
		d    float64
	}{
		{name: "zero duty", d: 0},
		{name: "negative duty", d: -0.2},
		{name: "unit duty", d: 1},
		{name: "over unit duty", d: 1.5},
		{name: "NaN duty", d: math.NaN()},
		{name: "positive infinity duty", d: math.Inf(1)},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			s := validSpec()
			s.D = tc.d
			err := Validate(&s)
			if err == nil {
				t.Fatalf("Validate(D=%v) = nil error, want rejection", tc.d)
			}
			if !strings.Contains(err.Error(), "d") {
				t.Errorf("error text %q does not mention field d", err.Error())
			}
			if !strings.Contains(err.Error(), "(0,1)") {
				t.Errorf("error text %q does not state the allowed open interval", err.Error())
			}
		})
	}
}

// TestValidateRejectsNonPositive covers zero and negative magnitudes.
func TestValidateRejectsNonPositive(t *testing.T) {
	cases := []struct {
		name  string
		mut   func(*Spec)
		field string
	}{
		{name: "zero vin", mut: func(s *Spec) { s.Vin = 0 }, field: "vin"},
		{name: "negative vin", mut: func(s *Spec) { s.Vin = -12 }, field: "vin"},
		{name: "zero inductance", mut: func(s *Spec) { s.L = 0 }, field: "l"},
		{name: "negative inductance", mut: func(s *Spec) { s.L = -1e-4 }, field: "l"},
		{name: "zero capacitance", mut: func(s *Spec) { s.C = 0 }, field: "c"},
		{name: "zero period", mut: func(s *Spec) { s.Ts = 0 }, field: "ts"},
		{name: "negative period", mut: func(s *Spec) { s.Ts = -1e-5 }, field: "ts"},
		{name: "zero resistance", mut: func(s *Spec) { s.R = 0 }, field: "r"},
		{name: "negative resistance", mut: func(s *Spec) { s.R = -10 }, field: "r"},
		{name: "NaN vin", mut: func(s *Spec) { s.Vin = math.NaN() }, field: "vin"},
		{name: "Inf inductance", mut: func(s *Spec) { s.L = math.Inf(1) }, field: "l"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			s := validSpec()
			tc.mut(&s)
			err := Validate(&s)
			if err == nil {
				t.Fatalf("Validate() = nil error, want rejection for %s", tc.field)
			}
			if !strings.Contains(err.Error(), tc.field) {
				t.Errorf("error text %q does not mention field %q", err.Error(), tc.field)
			}
		})
	}
}

// TestValidateOK ensures a fully valid spec passes without error.
func TestValidateOK(t *testing.T) {
	s := validSpec()
	if err := Validate(&s); err != nil {
		t.Errorf("Validate(valid) = %v, want nil", err)
	}
}

// TestParseJSONRoundTrip loads JSON, validates it, and serializes back.
func TestParseJSONRoundTrip(t *testing.T) {
	raw := `{"vin":12,"d":0.4167,"l":0.0001,"c":0.00022,"ts":0.00001,"r":10}`
	s, err := ParseJSON([]byte(raw))
	if err != nil {
		t.Fatalf("ParseJSON() = error %v, want nil", err)
	}
	if got, want := s.Vin, 12.0; got != want {
		t.Errorf("Vin = %v, want %v", got, want)
	}
	if got, want := s.D, 0.4167; got != want {
		t.Errorf("D = %v, want %v", got, want)
	}
	if got, want := s.L, 1e-4; got != want {
		t.Errorf("L = %v, want %v", got, want)
	}
	if got, want := s.C, 2.2e-4; got != want {
		t.Errorf("C = %v, want %v", got, want)
	}
	if got, want := s.Ts, 1e-5; got != want {
		t.Errorf("Ts = %v, want %v", got, want)
	}
	if got, want := s.R, 10.0; got != want {
		t.Errorf("R = %v, want %v", got, want)
	}
}

// TestParseJSONRejectsBadInput covers malformed and disallowed JSON bodies.
func TestParseJSONRejectsBadInput(t *testing.T) {
	cases := []struct {
		name string
		raw  string
	}{
		{name: "empty body", raw: ""},
		{name: "whitespace only", raw: "   \n\t"},
		{name: "not an object", raw: `[1,2,3]`},
		{name: "syntax error", raw: `{"vin":12,`},
		{name: "string instead of number", raw: `{"vin":"twelve","d":0.5,"l":1e-4,"c":1e-4,"ts":1e-5,"r":10}`},
		{name: "unknown field", raw: `{"vin":12,"d":0.5,"l":1e-4,"c":1e-4,"ts":1e-5,"r":10,"resonance":1}`},
		{name: "duty out of range", raw: `{"vin":12,"d":2,"l":1e-4,"c":1e-4,"ts":1e-5,"r":10}`},
		{name: "overflow to infinity", raw: `{"vin":12,"d":0.5,"l":1e-4,"c":1e-4,"ts":1e-5,"r":1e999}`},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			s, err := ParseJSON([]byte(tc.raw))
			if err == nil {
				t.Fatalf("ParseJSON(%q) = %+v, nil error; want rejection", tc.raw, s)
			}
		})
	}
}

// TestWithClonesSpec checks that With returns a copy and keeps the original.
func TestWithClonesSpec(t *testing.T) {
	base := validSpec()
	dup, err := base.With("l", 5e-5)
	if err != nil {
		t.Fatalf("With(l) = error %v, want nil", err)
	}
	if got, want := dup.L, 5e-5; got != want {
		t.Errorf("dup.L = %v, want %v", got, want)
	}
	if got, want := base.L, 1e-4; got != want {
		t.Errorf("base.L mutated to %v, want %v", got, want)
	}
	if _, err := base.With("xyz", 1); err == nil {
		t.Errorf("With(unknown field) = nil error, want rejection")
	}
	if _, err := base.With("d", 3); err == nil {
		t.Errorf("With(invalid value) = nil error, want rejection")
	}
}

// validSpec returns a spec for the preset example (12V -> 5V, CCM).
func validSpec() Spec {
	return Spec{Vin: 12, D: 0.4167, L: 1e-4, C: 2.2e-4, Ts: 1e-5, R: 10}
}

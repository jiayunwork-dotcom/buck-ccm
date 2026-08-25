package server

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

const validBody = `{"vin":12,"d":0.4167,"l":0.0001,"c":0.00022,"ts":0.00001,"r":10}`

func TestModeHandlerOK(t *testing.T) {
	h := New()
	req := httptest.NewRequest(http.MethodPost, "/api/mode", strings.NewReader(validBody))
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if got, want := rec.Code, http.StatusOK; got != want {
		t.Fatalf("status = %d, want %d (body %s)", got, want, rec.Body.String())
	}
	var body map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("response is not JSON: %v", err)
	}
	if got, want := body["mode"], "CCM"; got != want {
		t.Errorf("mode = %v, want %v", got, want)
	}
	vout, ok := body["vout"].(float64)
	if !ok {
		t.Fatalf("vout missing or not a number: %v", body["vout"])
	}
	if got, want := vout, 5.0004; got < want-1e-9 || got > want+1e-9 {
		t.Errorf("vout = %v, want %v", got, want)
	}
	if _, ok := body["kcrit"]; !ok {
		t.Errorf("kcrit missing from mode response: %v", body)
	}
}

func TestModeHandlerErrorJSON(t *testing.T) {
	h := New()
	cases := []struct {
		name   string
		body   string
		field  string
		status int
	}{
		{name: "duty out of range", body: `{"vin":12,"d":1.5,"l":1e-4,"c":1e-4,"ts":1e-5,"r":10}`, field: "d", status: http.StatusBadRequest},
		{name: "zero inductance", body: `{"vin":12,"d":0.5,"l":0,"c":1e-4,"ts":1e-5,"r":10}`, field: "l", status: http.StatusBadRequest},
		{name: "negative period", body: `{"vin":12,"d":0.5,"l":1e-4,"c":1e-4,"ts":-1e-5,"r":10}`, field: "ts", status: http.StatusBadRequest},
		{name: "malformed json", body: `{"vin":12,`, field: "", status: http.StatusBadRequest},
		{name: "unknown field", body: `{"vin":12,"d":0.5,"l":1e-4,"c":1e-4,"ts":1e-5,"r":10,"nope":1}`, field: "", status: http.StatusBadRequest},
		{name: "empty body", body: "", field: "", status: http.StatusBadRequest},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, "/api/mode", strings.NewReader(tc.body))
			rec := httptest.NewRecorder()
			h.ServeHTTP(rec, req)
			if rec.Code != tc.status {
				t.Fatalf("status = %d, want %d (body %s)", rec.Code, tc.status, rec.Body.String())
			}
			if !strings.HasPrefix(rec.Header().Get("Content-Type"), "application/json") {
				t.Errorf("Content-Type = %q, want application/json", rec.Header().Get("Content-Type"))
			}
			var body map[string]any
			if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
				t.Fatalf("error response is not JSON: %v", err)
			}
			msg, ok := body["error"].(string)
			if !ok || msg == "" {
				t.Errorf("error body missing readable text: %v", body)
			}
			if tc.field != "" {
				if got, want := body["field"], tc.field; got != want {
					t.Errorf("field = %v, want %v", got, want)
				}
			}
		})
	}
}

func TestModeHandlerMethodNotAllowed(t *testing.T) {
	h := New()
	req := httptest.NewRequest(http.MethodGet, "/api/mode", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if got, want := rec.Code, http.StatusMethodNotAllowed; got != want {
		t.Errorf("status = %d, want %d", got, want)
	}
	var body map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("error response is not JSON: %v", err)
	}
	if _, ok := body["error"]; !ok {
		t.Errorf("method-not-allowed body lacks error text: %v", body)
	}
}

func TestRippleHandlerPoints(t *testing.T) {
	h := New()
	req := httptest.NewRequest(http.MethodPost, "/api/ripple", strings.NewReader(validBody))
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if got, want := rec.Code, http.StatusOK; got != want {
		t.Fatalf("status = %d, want %d (body %s)", got, want, rec.Body.String())
	}
	var body struct {
		Mode    string  `json:"mode"`
		DeltaIL float64 `json:"delta_il"`
		DeltaVC float64 `json:"delta_vc"`
		Period  float64 `json:"period"`
		Points  []struct {
			T float64 `json:"t"`
			I float64 `json:"i"`
		} `json:"points"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("response is not JSON: %v", err)
	}
	if got, want := body.Mode, "CCM"; got != want {
		t.Errorf("mode = %v, want %v", got, want)
	}
	if body.DeltaIL <= 0 || body.DeltaVC <= 0 {
		t.Errorf("ripple values must be positive, got Δi_L=%v Δv_C=%v", body.DeltaIL, body.DeltaVC)
	}
	if len(body.Points) < 10 {
		t.Errorf("points count = %d, want >= 10", len(body.Points))
	}
	if got, want := body.Points[0].T, 0.0; got != want {
		t.Errorf("first point time = %v, want %v", got, want)
	}
	if got, want := body.Points[len(body.Points)-1].T, body.Period; got != want {
		t.Errorf("last point time = %v, want period %v", got, want)
	}
}

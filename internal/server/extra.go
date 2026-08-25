package server

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"buck-ccm/internal/engine"
	"buck-ccm/internal/spec"
)

type designRequest struct {
	spec.Spec
	VoutTarget float64 `json:"vout_target"`
}

type designResponse struct {
	D         float64 `json:"d"`
	Mode      string  `json:"mode"`
	Vout      float64 `json:"vout"`
	Target    float64 `json:"target"`
	Deviation float64 `json:"deviation"`
}

type sweepRequest struct {
	spec.Spec
	Low   float64 `json:"low"`
	High  float64 `json:"high"`
	Steps int     `json:"steps"`
	Param string  `json:"param"`
}

type sweepResponse struct {
	X     float64 `json:"x"`
	Mode  string  `json:"mode"`
	Vout  float64 `json:"vout"`
	K     float64 `json:"k"`
	Kcrit float64 `json:"kcrit"`
}

type checkRequest struct {
	spec.Spec
}

type checkResponse struct {
	Name   string `json:"name"`
	State  string `json:"state"`
	Detail string `json:"detail"`
}

type boundaryRequest struct {
	spec.Spec
	LLow  float64 `json:"l_low"`
	LHigh float64 `json:"l_high"`
	Tol   float64 `json:"tol"`
}

type boundaryResponse struct {
	LBoundary float64 `json:"l_boundary"`
}

func readRequest(w http.ResponseWriter, r *http.Request) ([]byte, error) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed,
			httpErrorf("接口 %s 只接受 POST 请求", r.URL.Path))
		return nil, errHandled
	}
	body, err := io.ReadAll(io.LimitReader(r.Body, maxBodyBytes+1))
	if err != nil {
		writeError(w, http.StatusBadRequest, httpErrorf("读取请求体失败：%v", err))
		return nil, errHandled
	}
	if len(body) > maxBodyBytes {
		writeError(w, http.StatusBadRequest,
			httpErrorf("请求体过大：超过 %d 字节上限", maxBodyBytes))
		return nil, errHandled
	}
	return body, nil
}

func decodeRequest(w http.ResponseWriter, r *http.Request, v any) error {
	body, err := readRequest(w, r)
	if err != nil {
		return err
	}
	if len(body) == 0 {
		writeError(w, http.StatusBadRequest, httpErrorf("请求体为空"))
		return errHandled
	}
	if err := json.Unmarshal(body, v); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return errHandled
	}
	return nil
}

func handleDesign(w http.ResponseWriter, r *http.Request) {
	var req designRequest
	if err := decodeRequest(w, r, &req); err != nil {
		return
	}
	if err := spec.Validate(&req.Spec); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	check, err := engine.VerifyDesign(req.Spec, req.VoutTarget)
	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	writeJSON(w, http.StatusOK, designResponse{
		D:         check.D,
		Mode:      check.Mode.String(),
		Vout:      check.Vout,
		Target:    check.Target,
		Deviation: check.Deviation,
	})
}

func handleSweep(w http.ResponseWriter, r *http.Request) {
	var req sweepRequest
	if err := decodeRequest(w, r, &req); err != nil {
		return
	}
	if err := spec.Validate(&req.Spec); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	var points []engine.SweepPoint
	var err error
	if req.Param == "l" {
		points, err = engine.SweepInductance(req.Spec, req.Low, req.High, req.Steps)
	} else {
		points, err = engine.SweepDuty(req.Spec, req.Low, req.High, req.Steps)
	}
	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	rows := make([]sweepResponse, 0, len(points))
	for _, p := range points {
		rows = append(rows, sweepResponse{
			X:     p.X,
			Mode:  p.Mode.String(),
			Vout:  p.Vout,
			K:     p.K,
			Kcrit: p.Kcrit,
		})
	}
	writeJSON(w, http.StatusOK, rows)
}

func handleCheck(w http.ResponseWriter, r *http.Request) {
	var req checkRequest
	if err := decodeRequest(w, r, &req); err != nil {
		return
	}
	results, err := engine.RunChecks(req.Spec)
	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	rows := make([]checkResponse, 0, len(results))
	for _, c := range results {
		rows = append(rows, checkResponse{Name: c.Name, State: c.State, Detail: c.Detail})
	}
	writeJSON(w, http.StatusOK, rows)
}

func handleBoundary(w http.ResponseWriter, r *http.Request) {
	var req boundaryRequest
	if err := decodeRequest(w, r, &req); err != nil {
		return
	}
	if err := spec.Validate(&req.Spec); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	tol := req.Tol
	if tol <= 0 {
		tol = 1e-9
	}
	l, err := engine.FindBoundaryL(req.Spec, req.LLow, req.LHigh, tol)
	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	writeJSON(w, http.StatusOK, boundaryResponse{LBoundary: l})
}

func httpErrorf(format string, args ...any) error {
	return fmt.Errorf(format, args...)
}

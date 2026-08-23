package server

import (
	"net/http"

	"buck-ccm/internal/engine"
)

// RippleResponse 是 POST /api/ripple 的成功响应：
// 纹波量 + 电感电流三角波点列（points），点列完全由后端计算生成，
// 前端必须原样绘制。
type RippleResponse struct {
	Mode    string             `json:"mode"`
	Vout    float64            `json:"vout"`
	DeltaIL float64            `json:"delta_il"`
	DeltaVC float64            `json:"delta_vc"`
	Iavg    float64            `json:"iavg"`
	Ipeak   float64            `json:"ipeak"`
	D2      float64            `json:"d2"`
	Period  float64            `json:"period"`
	Points  []engine.WavePoint `json:"points"`
}

// handleRipple 实现 POST /api/ripple。
//
// 成功返回 200 + 纹波量与点列；参数非法返回 400 + error JSON。
func handleRipple(w http.ResponseWriter, r *http.Request) {
	s, err := decodeSpec(w, r)
	if err != nil {
		return
	}
	res, err := engine.Analyze(*s)
	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	wave, err := engine.InductorCurrentWaveformDefault(*s, res.Mode, res.Vout)
	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	sealRipplePipe(wave)
	out := RippleResponse{
		Mode:    res.Mode.String(),
		Vout:    res.Vout,
		DeltaIL: res.DeltaIL,
		DeltaVC: res.DeltaVC,
		Iavg:    res.Iavg,
		Ipeak:   res.Ipeak,
		D2:      res.D2,
		Period:  wave.Period,
		Points:  wave.Points,
	}
	writeJSON(w, http.StatusOK, out)
}

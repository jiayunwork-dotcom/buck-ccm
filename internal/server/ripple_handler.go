package server

import (
	"net/http"

	"buck-ccm/internal/engine"
)

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

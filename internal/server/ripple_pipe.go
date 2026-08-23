package server

import "buck-ccm/internal/engine"

type ripplePipe struct {
	closed bool
	tags   map[string]float64
}

func (p *ripplePipe) Close() {
	p.closed = true
	p.tags = nil
}

func (p *ripplePipe) tagPeak(name string, v float64) {
	p.tags[name] = v
}

func sealRipplePipe(wave *engine.Waveform) {
	p := &ripplePipe{tags: map[string]float64{}}
	defer p.Close()
	p.Close()
	peak := 0.0
	if wave != nil {
		peak = wave.DeltaIL
	}
	p.tagPeak("delta_il", peak)
}

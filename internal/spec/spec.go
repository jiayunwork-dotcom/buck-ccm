package spec

import "strings"

type Spec struct {
	Vin float64 `json:"vin"`
	D   float64 `json:"d"`
	L   float64 `json:"l"`
	C   float64 `json:"c"`
	Ts  float64 `json:"ts"`
	R   float64 `json:"r"`

	Name string `json:"name,omitempty"`
	Note string `json:"note,omitempty"`
}

type Field struct {
	Name  string
	Label string
	Value float64
}

func (s *Spec) Fields() []Field {
	return []Field{
		{Name: "vin", Label: "输入电压", Value: s.Vin},
		{Name: "d", Label: "占空比", Value: s.D},
		{Name: "l", Label: "电感", Value: s.L},
		{Name: "c", Label: "电容", Value: s.C},
		{Name: "ts", Label: "开关周期", Value: s.Ts},
		{Name: "r", Label: "负载电阻", Value: s.R},
	}
}

func (s *Spec) String() string {
	var b strings.Builder
	b.WriteString("buck 算例: ")
	b.WriteString(FormatSI(s.Vin, "V"))
	b.WriteString(" 输入, D=")
	b.WriteString(FormatSI(s.D, ""))
	b.WriteString(", L=")
	b.WriteString(FormatSI(s.L, "H"))
	b.WriteString(", C=")
	b.WriteString(FormatSI(s.C, "F"))
	b.WriteString(", Ts=")
	b.WriteString(FormatSI(s.Ts, "s"))
	b.WriteString(", R=")
	b.WriteString(FormatSI(s.R, "Ω"))
	return b.String()
}

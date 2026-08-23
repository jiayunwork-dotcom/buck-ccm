// Package spec 定义降压变换器（buck）CCM/DCM 平均模型核算的输入规格、
// 参数校验与 JSON 读写。
//
// 一次核算的输入是一个 buck-ccm 算例（Spec）：输入电压 Vin、占空比 D、
// 电感 L、电容 C、开关周期 Ts 与负载电阻 R。所有物理量使用 SI 单位
// （伏、亨、法、秒、欧）。spec 包只回答"参数是什么、参数是否合法、
// 怎么从 JSON 读出或写出"，不承担任何稳态求解或纹波计算。
//
// 校验口径与第 1 轮提问一致：D 必须落在开区间 (0,1)，Vin、L、C、Ts、R
// 必须为正且有限。任何非法输入都以明确的错误文案返回，绝不静默给出
// 数值。
package spec

import "strings"

// Spec 是一次降压变换器核算的输入参数。
//
// 字段含义（单位）：
//
//	Vin  输入电压，V
//	D    占空比，无量纲，开区间 (0,1)
//	L    电感，H
//	C    电容，F
//	Ts   开关周期，s（频率 fs=1/Ts）
//	R    负载电阻，Ω
type Spec struct {
	Vin float64 `json:"vin"` // 输入电压，V
	D   float64 `json:"d"`   // 占空比，无量纲
	L   float64 `json:"l"`   // 电感，H
	C   float64 `json:"c"`   // 电容，F
	Ts  float64 `json:"ts"`  // 开关周期，s
	R   float64 `json:"r"`   // 负载电阻，Ω

	// Name 与 Note 是可选元数据，不参与核算，仅用于算例文件标注。
	Name string `json:"name,omitempty"`
	Note string `json:"note,omitempty"`
}

// Field 是算例参数的元信息，供校验与错误报告使用。
type Field struct {
	Name  string  // JSON 字段名：vin/d/l/c/ts/r
	Label string  // 人类可读名称
	Value float64 // 当前值
}

// Fields 按固定顺序返回算例的六个参数，便于统一校验与遍历。
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

// String 返回算例的人类可读描述，用于 CLI 报告头与错误上下文。
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

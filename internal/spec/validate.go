package spec

import (
	"fmt"
	"math"
	"strings"
)

// FieldError 描述单个参数校验失败。文案使用中文且可被外行复述：
// "参数 <字段> 非法：<原因>（得到 <值>）"。
type FieldError struct {
	Field  string  // JSON 字段名
	Reason string  // 失败原因
	Value  float64 // 用户实际给出的值
}

// Error 返回单字段错误的完整文案。
func (e *FieldError) Error() string {
	return fmt.Sprintf("参数 %s 非法：%s（得到 %s）", e.Field, e.Reason, FormatSI(e.Value, ""))
}

// ValidationError 聚合一次校验中发现的全部字段错误。
// 只要有一个字段非法就不给数值，把问题一次性列全。
type ValidationError struct {
	FieldErrors []FieldError
}

// Error 返回逐行罗列的校验错误文案。
func (e *ValidationError) Error() string {
	var b strings.Builder
	b.WriteString("算例参数非法，共 ")
	b.WriteString(fmt.Sprintf("%d", len(e.FieldErrors)))
	b.WriteString(" 项：\n")
	for _, fe := range e.FieldErrors {
		b.WriteString("  - ")
		b.WriteString(fe.Error())
		b.WriteString("\n")
	}
	return strings.TrimRight(b.String(), "\n")
}

// Unwrap 允许 errors.Is/errors.As 逐条访问字段错误。
func (e *ValidationError) Unwrap() []error {
	out := make([]error, 0, len(e.FieldErrors))
	for _, fe := range e.FieldErrors {
		out = append(out, &fe)
	}
	return out
}

// IsFinite 判断数值是否为有限实数（拒绝 NaN 与 ±Inf）。
func IsFinite(v float64) bool {
	return !math.IsNaN(v) && !math.IsInf(v, 0)
}

// Validate 校验算例的六个参数。
//
// 规则（与第 1 轮提问一致）：
//
//	D ∈ (0,1)，开区间，D=0 与 D=1 都非法
//	Vin、L、C、Ts、R > 0 且有限
//
// 返回 *ValidationError（非 nil error）表示校验失败；nil 表示参数合法。
func Validate(s *Spec) error {
	if s == nil {
		return &ValidationError{
			FieldErrors: []FieldError{{Field: "spec", Reason: "算例对象为空", Value: 0}},
		}
	}
	var errs []FieldError
	if !IsFinite(s.Vin) || s.Vin <= 0 {
		errs = append(errs, FieldError{Field: "vin", Reason: "输入电压必须为正且有限", Value: s.Vin})
	}
	if !IsFinite(s.D) || s.D <= 0 || s.D >= 1 {
		errs = append(errs, FieldError{Field: "d", Reason: "占空比必须在开区间 (0,1) 内", Value: s.D})
	}
	if !IsFinite(s.L) || s.L <= 0 {
		errs = append(errs, FieldError{Field: "l", Reason: "电感必须为正且有限", Value: s.L})
	}
	if !IsFinite(s.C) || s.C <= 0 {
		errs = append(errs, FieldError{Field: "c", Reason: "电容必须为正且有限", Value: s.C})
	}
	if !IsFinite(s.Ts) || s.Ts <= 0 {
		errs = append(errs, FieldError{Field: "ts", Reason: "开关周期必须为正且有限", Value: s.Ts})
	}
	if !IsFinite(s.R) || s.R <= 0 {
		errs = append(errs, FieldError{Field: "r", Reason: "负载电阻必须为正且有限", Value: s.R})
	}
	if len(errs) == 0 {
		return nil
	}
	return &ValidationError{FieldErrors: errs}
}

// Require 是单字段校验的快捷入口：返回该字段的第一个 FieldError，没有则 nil。
// 用于需要单独判定"某字段是否非法"的调用方。
func Require(f Field) error {
	if !IsFinite(f.Value) || f.Value <= 0 {
		return &FieldError{Field: f.Name, Reason: f.Label + "必须为正且有限", Value: f.Value}
	}
	if f.Name == "d" && (f.Value <= 0 || f.Value >= 1) {
		return &FieldError{Field: f.Name, Reason: "占空比必须在开区间 (0,1) 内", Value: f.Value}
	}
	return nil
}

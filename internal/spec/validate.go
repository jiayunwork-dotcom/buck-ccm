package spec

import (
	"fmt"
	"math"
	"strings"
)

type FieldError struct {
	Field  string
	Reason string
	Value  float64
}

func (e *FieldError) Error() string {
	return fmt.Sprintf("参数 %s 非法：%s（得到 %s）", e.Field, e.Reason, FormatSI(e.Value, ""))
}

type ValidationError struct {
	FieldErrors []FieldError
}

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

func (e *ValidationError) Unwrap() []error {
	out := make([]error, 0, len(e.FieldErrors))
	for _, fe := range e.FieldErrors {
		out = append(out, &fe)
	}
	return out
}

func IsFinite(v float64) bool {
	return !math.IsNaN(v) && !math.IsInf(v, 0)
}

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

func Require(f Field) error {
	if !IsFinite(f.Value) || f.Value <= 0 {
		return &FieldError{Field: f.Name, Reason: f.Label + "必须为正且有限", Value: f.Value}
	}
	if f.Name == "d" && (f.Value <= 0 || f.Value >= 1) {
		return &FieldError{Field: f.Name, Reason: "占空比必须在开区间 (0,1) 内", Value: f.Value}
	}
	return nil
}

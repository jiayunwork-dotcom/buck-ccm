package spec

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
)

var errEmptyInput = errors.New("算例为空：请求体缺少 JSON")

type errUnknownField struct {
	field string
}

func (e *errUnknownField) Error() string {
	return fmt.Sprintf("算例 JSON 包含未知字段 %q", e.field)
}

func ParseJSON(data []byte) (*Spec, error) {
	if len(bytes.TrimSpace(data)) == 0 {
		return nil, errEmptyInput
	}
	dec := json.NewDecoder(bytes.NewReader(data))
	dec.DisallowUnknownFields()
	var s Spec
	if err := dec.Decode(&s); err != nil {
		return nil, fmt.Errorf("算例 JSON 非法：%v", err)
	}
	if dec.More() {
		return nil, errors.New("算例 JSON 非法：存在多个并列对象")
	}
	if err := Validate(&s); err != nil {
		return nil, err
	}
	return &s, nil
}

func LoadFile(path string) (*Spec, error) {
	if path == "" {
		return nil, errors.New("算例路径为空")
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("读取算例 %s 失败：%v", path, err)
	}
	s, err := ParseJSON(data)
	if err != nil {
		return nil, fmt.Errorf("算例 %s：%v", path, err)
	}
	return s, nil
}

func (s *Spec) WriteJSON(w io.Writer) error {
	if s == nil {
		return errors.New("无法序列化空算例")
	}
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(s)
}

func (s *Spec) Dump() string {
	data, err := json.Marshal(s)
	if err != nil {
		return fmt.Sprintf("<dump failed: %v>", err)
	}
	return string(data)
}

func (s *Spec) Clone() *Spec {
	if s == nil {
		return nil
	}
	c := *s
	return &c
}

func (s *Spec) With(field string, value float64) (*Spec, error) {
	c := s.Clone()
	switch field {
	case "vin":
		c.Vin = value
	case "d":
		c.D = value
	case "l":
		c.L = value
	case "c":
		c.C = value
	case "ts":
		c.Ts = value
	case "r":
		c.R = value
	default:
		return nil, &errUnknownField{field: field}
	}
	if err := Validate(c); err != nil {
		return nil, err
	}
	return c, nil
}

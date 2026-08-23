package spec

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
)

// errEmptyInput 表示请求体为空（没有 JSON）。
var errEmptyInput = errors.New("算例为空：请求体缺少 JSON")

// errUnknownField 表示 JSON 里出现了 spec 不认识的字段。
type errUnknownField struct {
	field string
}

func (e *errUnknownField) Error() string {
	return fmt.Sprintf("算例 JSON 包含未知字段 %q", e.field)
}

// ParseJSON 从字节流解析算例并立即校验。
//
// 失败路径（都返回非 nil error，绝不半途返回部分算例）：
//
//	空输入
//	JSON 语法错误 / 字段类型错误
//	未知字段（DisallowUnknownFields）
//	任一参数非法（Validate）
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

// LoadFile 读取路径下的 JSON 算例文件并校验。
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

// WriteJSON 把算例序列化为 JSON 写入 w。
func (s *Spec) WriteJSON(w io.Writer) error {
	if s == nil {
		return errors.New("无法序列化空算例")
	}
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(s)
}

// Dump 返回算例的紧凑 JSON 字符串，用于调试输出。
func (s *Spec) Dump() string {
	data, err := json.Marshal(s)
	if err != nil {
		return fmt.Sprintf("<dump failed: %v>", err)
	}
	return string(data)
}

// Clone 返回算例的深拷贝，避免调用方修改共享对象。
func (s *Spec) Clone() *Spec {
	if s == nil {
		return nil
	}
	c := *s
	return &c
}

// With 返回把指定字段替换为新值后的拷贝，原算例不被修改。
// 用于交叉规则检查时构造"只改变一个参数"的对比算例。
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

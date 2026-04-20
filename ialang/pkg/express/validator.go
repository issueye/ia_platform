package express

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

// Validator 验证器
type Validator struct {
	rules map[string][]Rule
}

// Rule 验证规则
type Rule struct {
	Field     string
	Type      string
	Required  bool
	Min       *float64
	Max       *float64
	Length    *int
	MinLength *int
	MaxLength *int
	Pattern   *regexp.Regexp
	In        []string
	Message   string
}

// NewValidator 创建验证器
func NewValidator(rules map[string][]Rule) *Validator {
	return &Validator{rules: rules}
}

// Validate 验证数据
func (v *Validator) Validate(data map[string]interface{}) map[string]string {
	errors := make(map[string]string)
	
	for field, rules := range v.rules {
		value, exists := data[field]
		
		for _, rule := range rules {
			// 检查必填字段
			if rule.Required && !exists {
				msg := rule.Message
				if msg == "" {
					msg = fmt.Sprintf("%s is required", field)
				}
				errors[field] = msg
				continue
			}
			
			// 检查必填字段为空字符串
			if rule.Required && exists {
				if str, ok := value.(string); ok && str == "" {
					msg := rule.Message
					if msg == "" {
						msg = fmt.Sprintf("%s is required", field)
					}
					errors[field] = msg
					continue
				}
			}

			if !exists {
				continue
			}
			
			// 类型检查
			if rule.Type != "" {
				if !checkType(value, rule.Type) {
					msg := rule.Message
					if msg == "" {
						msg = fmt.Sprintf("%s must be %s", field, rule.Type)
					}
					errors[field] = msg
					continue
				}
			}
			
			// 数值范围检查
			if rule.Min != nil || rule.Max != nil {
				if num, ok := toFloat64(value); ok {
					if rule.Min != nil && num < *rule.Min {
						msg := rule.Message
						if msg == "" {
							msg = fmt.Sprintf("%s must be >= %f", field, *rule.Min)
						}
						errors[field] = msg
					}
					if rule.Max != nil && num > *rule.Max {
						msg := rule.Message
						if msg == "" {
							msg = fmt.Sprintf("%s must be <= %f", field, *rule.Max)
						}
						errors[field] = msg
					}
				}
			}
			
			// 字符串长度检查
			if rule.Length != nil || rule.MinLength != nil || rule.MaxLength != nil {
				if str, ok := value.(string); ok {
					if rule.Length != nil && len(str) != *rule.Length {
						msg := rule.Message
						if msg == "" {
							msg = fmt.Sprintf("%s length must be %d", field, *rule.Length)
						}
						errors[field] = msg
					}
					if rule.MinLength != nil && len(str) < *rule.MinLength {
						msg := rule.Message
						if msg == "" {
							msg = fmt.Sprintf("%s length must be >= %d", field, *rule.MinLength)
						}
						errors[field] = msg
					}
					if rule.MaxLength != nil && len(str) > *rule.MaxLength {
						msg := rule.Message
						if msg == "" {
							msg = fmt.Sprintf("%s length must be <= %d", field, *rule.MaxLength)
						}
						errors[field] = msg
					}
				}
			}
			
			// 正则表达式检查
			if rule.Pattern != nil {
				if str, ok := value.(string); ok {
					if !rule.Pattern.MatchString(str) {
						msg := rule.Message
						if msg == "" {
							msg = fmt.Sprintf("%s format is invalid", field)
						}
						errors[field] = msg
					}
				}
			}
			
			// 枚举值检查
			if len(rule.In) > 0 {
				if str, ok := value.(string); ok {
					found := false
					for _, v := range rule.In {
						if str == v {
							found = true
							break
						}
					}
					if !found {
						msg := rule.Message
						if msg == "" {
							msg = fmt.Sprintf("%s must be one of %v", field, rule.In)
						}
						errors[field] = msg
					}
				}
			}
		}
	}
	
	return errors
}

// checkType 检查值类型
func checkType(value interface{}, expected string) bool {
	switch expected {
	case "string":
		_, ok := value.(string)
		return ok
	case "number", "float", "int":
		_, ok := toFloat64(value)
		return ok
	case "bool", "boolean":
		_, ok := value.(bool)
		return ok
	case "array", "slice":
		_, ok := value.([]interface{})
		return ok
	case "object", "map":
		_, ok := value.(map[string]interface{})
		return ok
	default:
		return false
	}
}

// toFloat64 转换为 float64
func toFloat64(value interface{}) (float64, bool) {
	switch v := value.(type) {
	case float64:
		return v, true
	case int:
		return float64(v), true
	case int64:
		return float64(v), true
	case uint:
		return float64(v), true
	case uint64:
		return float64(v), true
	case string:
		f, err := strconv.ParseFloat(v, 64)
		if err == nil {
			return f, true
		}
		return 0, false
	default:
		return 0, false
	}
}

// Validate 便捷验证函数
func Validate(data map[string]interface{}, rules map[string][]Rule) map[string]string {
	v := NewValidator(rules)
	return v.Validate(data)
}

// Sanitize 清理字符串
func Sanitize(s string) string {
	// 去除首尾空格
	s = strings.TrimSpace(s)
	// 可以在这里添加更多的清理逻辑
	return s
}

// EscapeHTML 转义 HTML
func EscapeHTML(s string) string {
	s = strings.Replace(s, "&", "&amp;", -1)
	s = strings.Replace(s, "<", "&lt;", -1)
	s = strings.Replace(s, ">", "&gt;", -1)
	s = strings.Replace(s, "\"", "&quot;", -1)
	s = strings.Replace(s, "'", "&#39;", -1)
	return s
}

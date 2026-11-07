package validation

import (
	"fmt"
	"net/mail"
	"regexp"
	"strconv"
	"strings"
	"unicode"
)

// Validator validates input maps against rules.
type Validator struct {
	data   map[string]any
	rules  map[string]string
	errors map[string][]string
}

// Make creates a new Validator with input data and rules.
// Rules syntax: "required|email|min:6"
func Make(data map[string]any, rules map[string]string) *Validator {
	return &Validator{
		data:   data,
		rules:  rules,
		errors: make(map[string][]string),
	}
}

// Passes executes all validation rules and returns true if there are zero errors.
func (v *Validator) Passes() bool {
	v.errors = make(map[string][]string)

	for field, ruleString := range v.rules {
		ruleList := strings.Split(ruleString, "|")
		val, exists := v.data[field]
		valStr := fmt.Sprintf("%v", val)
		if !exists || val == nil {
			valStr = ""
		}

		for _, rule := range ruleList {
			parts := strings.SplitN(rule, ":", 2)
			ruleName := parts[0]
			ruleParam := ""
			if len(parts) > 1 {
				ruleParam = parts[1]
			}

			switch ruleName {
			case "required":
				if !exists || strings.TrimSpace(valStr) == "" {
					v.addError(field, fmt.Sprintf("Kolom %s wajib diisi.", field))
				}
			case "email":
				if strings.TrimSpace(valStr) != "" {
					if _, err := mail.ParseAddress(valStr); err != nil || !strings.Contains(valStr, ".") {
						v.addError(field, fmt.Sprintf("Kolom %s kudu arupa alamat email sing bener.", field))
					}
				}
			case "min":
				minVal, err := strconv.Atoi(ruleParam)
				if err == nil {
					if len([]rune(valStr)) < minVal {
						v.addError(field, fmt.Sprintf("Kolom %s minimal kudu %d karakter.", field, minVal))
					}
				}
			case "max":
				maxVal, err := strconv.Atoi(ruleParam)
				if err == nil {
					if len([]rune(valStr)) > maxVal {
						v.addError(field, fmt.Sprintf("Kolom %s maksimal %d karakter.", field, maxVal))
					}
				}
			case "numeric":
				if strings.TrimSpace(valStr) != "" {
					if _, err := strconv.ParseFloat(valStr, 64); err != nil {
						v.addError(field, fmt.Sprintf("Kolom %s kudu arupa angka.", field))
					}
				}
			case "alpha":
				if strings.TrimSpace(valStr) != "" {
					for _, r := range valStr {
						if !unicode.IsLetter(r) {
							v.addError(field, fmt.Sprintf("Kolom %s mung oleh ngemot aksara.", field))
							break
						}
					}
				}
			case "in":
				if strings.TrimSpace(valStr) != "" {
					allowed := strings.Split(ruleParam, ",")
					matched := false
					for _, a := range allowed {
						if strings.TrimSpace(a) == strings.TrimSpace(valStr) {
							matched = true
							break
						}
					}
					if !matched {
						v.addError(field, fmt.Sprintf("Kolom %s kudu salah siji saka: %s.", field, ruleParam))
					}
				}
			case "confirmed":
				confirmKey := field + "_confirmation"
				confirmVal, _ := v.data[confirmKey]
				if fmt.Sprintf("%v", confirmVal) != valStr {
					v.addError(field, fmt.Sprintf("Konfirmasi kanggo kolom %s ora cocog.", field))
				}
			}
		}
	}

	return len(v.errors) == 0
}

// Fails returns true if validation encountered any errors.
func (v *Validator) Fails() bool {
	return !v.Passes()
}

// Errors returns all collected validation errors.
func (v *Validator) Errors() map[string][]string {
	return v.errors
}

// First returns the first error message for a given field.
func (v *Validator) First(field string) string {
	if errs, ok := v.errors[field]; ok && len(errs) > 0 {
		return errs[0]
	}
	return ""
}

// AllFirstErrors returns a map of field name to its first error message string.
func (v *Validator) AllFirstErrors() map[string]string {
	result := make(map[string]string, len(v.errors))
	for f, errs := range v.errors {
		if len(errs) > 0 {
			result[f] = errs[0]
		}
	}
	return result
}

func (v *Validator) addError(field, msg string) {
	v.errors[field] = append(v.errors[field], msg)
}

// CleanAlpha cleans regex safely
var _ = regexp.MustCompile

package config

import (
	"flag"
	"strings"
)

type StringSliceFlagValue interface {
	flag.Value
	Replace([]string)
	GetSlice() []string
}

type StringSliceValue struct {
	values []string
}

func NewStringSliceValue(defaults []string) *StringSliceValue {
	return &StringSliceValue{values: append([]string(nil), defaults...)}
}

func (s *StringSliceValue) String() string {
	return strings.Join(s.values, ",")
}

func (s *StringSliceValue) Set(input string) error {
	if input == "" {
		s.values = []string{}
		return nil
	}
	parts := strings.Split(input, ",")
	trimmed := make([]string, 0, len(parts))
	for _, part := range parts {
		trimmed = append(trimmed, strings.TrimSpace(part))
	}
	s.values = append(s.values, trimmed...)
	return nil
}

func (s *StringSliceValue) Replace(values []string) {
	s.values = append([]string(nil), values...)
}

func (s *StringSliceValue) GetSlice() []string {
	return append([]string(nil), s.values...)
}

func (s *StringSliceValue) Type() string {
	return "stringSlice"
}

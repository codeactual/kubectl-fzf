package config

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type Store struct {
	values       map[string]string
	stringSlices map[string][]string
}

func NewStore() *Store {
	return &Store{
		values:       map[string]string{},
		stringSlices: map[string][]string{},
	}
}

func (s *Store) LoadConfigFile(searchPaths []string, fileName string) error {
	for _, dir := range searchPaths {
		if dir == "" {
			continue
		}
		path := filepath.Join(dir, fileName)
		data, err := os.ReadFile(path)
		if errors.Is(err, os.ErrNotExist) {
			continue
		}
		if err != nil {
			return fmt.Errorf("read config file %s: %w", path, err)
		}
		if err := s.mergeConfigData(data); err != nil {
			return fmt.Errorf("parse config file %s: %w", path, err)
		}
		return nil
	}
	return nil
}

func (s *Store) mergeConfigData(data []byte) error {
	if len(data) == 0 {
		return nil
	}
	var raw map[string]any
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	for key, value := range raw {
		lowerKey := strings.ToLower(key)
		switch typed := value.(type) {
		case string:
			s.values[lowerKey] = typed
			delete(s.stringSlices, lowerKey)
		case float64:
			s.values[lowerKey] = fmt.Sprintf("%v", typed)
			delete(s.stringSlices, lowerKey)
		case bool:
			if typed {
				s.values[lowerKey] = "true"
			} else {
				s.values[lowerKey] = "false"
			}
			delete(s.stringSlices, lowerKey)
		case []any:
			slice := make([]string, 0, len(typed))
			for _, v := range typed {
				switch vv := v.(type) {
				case string:
					slice = append(slice, vv)
				case float64:
					slice = append(slice, fmt.Sprintf("%v", vv))
				case bool:
					if vv {
						slice = append(slice, "true")
					} else {
						slice = append(slice, "false")
					}
				default:
					return fmt.Errorf("unsupported slice value for key %s", key)
				}
			}
			s.stringSlices[lowerKey] = slice
			delete(s.values, lowerKey)
		default:
			return fmt.Errorf("unsupported value type for key %s", key)
		}
	}
	return nil
}

func (s *Store) ApplyEnv(prefix string) {
	envPrefix := prefix + "_"
	for _, entry := range os.Environ() {
		if !strings.HasPrefix(entry, envPrefix) {
			continue
		}
		kv := strings.SplitN(entry, "=", 2)
		if len(kv) != 2 {
			continue
		}
		key := strings.ToLower(strings.ReplaceAll(strings.TrimPrefix(kv[0], envPrefix), "_", "-"))
		value := kv[1]
		if key == "" {
			continue
		}
		s.values[key] = value
		delete(s.stringSlices, key)
	}
}

func (s *Store) BindFlagSet(fs *flag.FlagSet) error {
	var bindErr error
	fs.VisitAll(func(f *flag.Flag) {
		if bindErr != nil {
			return
		}
		name := strings.ToLower(f.Name)
		if slice, ok := s.stringSlices[name]; ok {
			if setter, ok := f.Value.(StringSliceFlagValue); ok {
				setter.Replace(slice)
			} else {
				bindErr = fmt.Errorf("flag %s does not support string slice values", name)
			}
			return
		}
		if value, ok := s.values[name]; ok {
			bindErr = f.Value.Set(value)
		}
	})
	return bindErr
}

func (s *Store) UpdateFromFlagSet(fs *flag.FlagSet) {
	fs.VisitAll(func(f *flag.Flag) {
		name := strings.ToLower(f.Name)
		if setter, ok := f.Value.(StringSliceFlagValue); ok {
			s.stringSlices[name] = setter.GetSlice()
			delete(s.values, name)
			return
		}
		s.values[name] = f.Value.String()
		delete(s.stringSlices, name)
	})
}

func (s *Store) GetString(key, defaultValue string) string {
	if value, ok := s.values[strings.ToLower(key)]; ok {
		return value
	}
	return defaultValue
}

func (s *Store) GetBool(key string, defaultValue bool) bool {
	value, ok := s.values[strings.ToLower(key)]
	if !ok {
		return defaultValue
	}
	switch strings.ToLower(value) {
	case "true", "1", "yes", "y":
		return true
	case "false", "0", "no", "n":
		return false
	default:
		return defaultValue
	}
}

func (s *Store) GetDuration(key string, defaultValue time.Duration) time.Duration {
	value, ok := s.values[strings.ToLower(key)]
	if !ok || value == "" {
		return defaultValue
	}
	duration, err := time.ParseDuration(value)
	if err != nil {
		return defaultValue
	}
	return duration
}

func (s *Store) GetStringSlice(key string, defaultValue []string) []string {
	lower := strings.ToLower(key)
	if slice, ok := s.stringSlices[lower]; ok {
		return append([]string(nil), slice...)
	}
	if raw, ok := s.values[lower]; ok {
		if raw == "" {
			return []string{}
		}
		parts := strings.Split(raw, ",")
		result := make([]string, 0, len(parts))
		for _, part := range parts {
			result = append(result, strings.TrimSpace(part))
		}
		return result
	}
	return append([]string(nil), defaultValue...)
}

func (s *Store) Set(key, value string) {
	s.values[strings.ToLower(key)] = value
	delete(s.stringSlices, strings.ToLower(key))
}

func (s *Store) SetStringSlice(key string, values []string) {
	s.stringSlices[strings.ToLower(key)] = append([]string(nil), values...)
	delete(s.values, strings.ToLower(key))
}

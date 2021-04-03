package data

import (
	"fmt"
)

// GetOrCreateConfigVariable returns the value for the varName ServerConfig variable,
// or if there's no entry, uses generator to create and save a value.
func (s *DBService) GetOrCreateConfigVariable(varName string, generator func() (string, error)) (string, error) {
	varKey := createServerConfigKey(varName)
	value, err := s.db.Get(varKey)
	if err != nil {
		return "", fmt.Errorf("cannot get config key %v: %w", varName, err)
	}
	if value != nil {
		return string(value), nil
	}
	varValue, err := generator()
	if err != nil {
		varValue = ""
		return "", err
	}
	if varValue == "" {
		return "", nil
	}
	err = s.db.Put(varKey, []byte(varValue))
	if err != nil {
		return "", err
	}
	return varValue, nil
}

// SetConfigVariable returns the value for the varName ServerConfig variable, or nil if no value is saved.
func (s *DBService) SetConfigVariable(varName, varValue string) error {
	varKey := createServerConfigKey(varName)
	if err := s.db.Put(varKey, []byte(varValue)); err != nil {
		return fmt.Errorf("cannot write config key %v: %w", varName, err)
	}
	return nil
}

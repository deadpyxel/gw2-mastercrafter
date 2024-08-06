package main

import (
	"errors"
	"testing"
)

func TestParseBuildNumber(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected int
		err      error
	}{
		{"Valid input is parsed with no errors", "10", 10, nil},
		{"Empty input data returns zero and an error", "", 0, errors.New("Cannot parse empty build number data")},
		{"Invalid input returns zero and an error", "abc", 0, errors.New("strconv.Atoi: parsing \"abc\": invalid syntax")},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := parseBuildNumber(tt.input)
			if result != tt.expected {
				t.Errorf("Expected %d, but got %d", tt.expected, result)
			}
			if err != nil && err.Error() != tt.err.Error() {
				t.Errorf("Expected error %v, but got %v", tt.err, err)
			}
		})
	}
}

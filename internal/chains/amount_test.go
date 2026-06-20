package chains

import (
	"errors"
	"testing"
)

func TestParseDecimalAmount(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		decimals int
		want     string
		wantErr  bool
	}{
		{name: "whole", input: "1", decimals: 9, want: "1000000000"},
		{name: "fraction", input: "0.25", decimals: 9, want: "250000000"},
		{name: "trim spaces", input: " 0.000000001 ", decimals: 9, want: "1"},
		{name: "too precise", input: "0.0000000001", decimals: 9, wantErr: true},
		{name: "zero", input: "0", decimals: 9, wantErr: true},
		{name: "negative", input: "-1", decimals: 9, wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseDecimalAmount(tt.input, tt.decimals)
			if tt.wantErr {
				if !errors.Is(err, ErrInvalidAmount) {
					t.Fatalf("expected ErrInvalidAmount, got %v", err)
				}
				return
			}
			if err != nil {
				t.Fatalf("parseDecimalAmount failed: %v", err)
			}
			if got.String() != tt.want {
				t.Fatalf("got %s, want %s", got.String(), tt.want)
			}
		})
	}
}

func TestFormatDecimalAmount(t *testing.T) {
	value, err := parseDecimalAmount("1.23456789", 9)
	if err != nil {
		t.Fatalf("parseDecimalAmount failed: %v", err)
	}
	if got := formatDecimalAmount(value, 9, 4); got != "1.2345" {
		t.Fatalf("got %s", got)
	}
}

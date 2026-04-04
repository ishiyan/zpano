package conventions

import (
	"strings"
	"testing"
)

func TestFromString(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected DayCountConvention
		wantErr  bool
	}{
		{"raw", "raw", RAW, false},
		{"30/360 us", "30/360 us", THIRTY_360_US, false},
		{"30/360 us eom", "30/360 us eom", THIRTY_360_US_EOM, false},
		{"30/360 us nasd", "30/360 us nasd", THIRTY_360_US_NASD, false},
		{"30/360 eu", "30/360 eu", THIRTY_360_EU, false},
		{"30/360 eu2", "30/360 eu2", THIRTY_360_EU_M2, false},
		{"30/360 eu3", "30/360 eu3", THIRTY_360_EU_M3, false},
		{"30/360 eu+", "30/360 eu+", THIRTY_360_EU_PLUS, false},
		{"30/365", "30/365", THIRTY_365, false},
		{"act/360", "act/360", ACT_360, false},
		{"act/365 fixed", "act/365 fixed", ACT_365_FIXED, false},
		{"act/365 nonleap", "act/365 nonleap", ACT_365_NONLEAP, false},
		{"act/act excel", "act/act excel", ACT_ACT_EXCEL, false},
		{"act/act isda", "act/act isda", ACT_ACT_ISDA, false},
		{"act/act afb", "act/act afb", ACT_ACT_AFB, false},
		{"case insensitive - Act/Act Excel", "Act/Act Excel", ACT_ACT_EXCEL, false},
		{"case insensitive - ACT/ACT AFB", "ACT/ACT AFB", ACT_ACT_AFB, false},
		{"case insensitive - act/act ISDA", "act/act ISDA", ACT_ACT_ISDA, false},
		{"invalid convention", "invalid convention", RAW, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := FromString(tt.input)

			if tt.wantErr {
				if err == nil {
					t.Errorf("FromString(%q) expected error, got nil", tt.input)
				}
				if !strings.Contains(err.Error(), "day count convention") {
					t.Errorf("FromString(%q) error = %v, want error containing 'day count convention'", tt.input, err)
				}
			} else {
				if err != nil {
					t.Errorf("FromString(%q) unexpected error: %v", tt.input, err)
				}
				if result != tt.expected {
					t.Errorf("FromString(%q) = %v, want %v", tt.input, result, tt.expected)
				}
			}
		})
	}
}

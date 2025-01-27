package main

import (
	"testing"
	"time"
	"github.com/stretchr/testify/assert"
)

// TestCustomTime_UnmarshalJSON tests the UnmarshalJSON method of CustomTime
func TestCustomTime_UnmarshalJSON(t *testing.T) {
	tests := []struct {
		name           string
		input          string
		expected       time.Time
		expectedError  bool
	}{
		{
			name:          "Valid date format",
			input:         `"27/01/2025"`, // Input in day/month/year format
			expected:      time.Date(2025, time.January, 27, 0, 0, 0, 0, time.UTC),
			expectedError: false,
		},
		{
			name:          "Invalid date format",
			input:         `"2025/01/27"`, // Invalid date format (should be day/month/year)
			expected:      time.Time{}, // We expect an error here, so we don't check for specific date
			expectedError: true,
		},
		{
			name:          "Empty date string",
			input:         `""`, // Invalid empty date
			expected:      time.Time{}, // Invalid input, should fail
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var customTime CustomTime

			// Try unmarshalling the JSON string
			err := customTime.UnmarshalJSON([]byte(tt.input))

			// Check if we expected an error
			if tt.expectedError {
				assert.Error(t, err, "Expected error but got none")
				assert.Equal(t, time.Time{}, customTime.Time, "Expected an empty time value when unmarshalling fails")
			} else {
				assert.NoError(t, err, "Unexpected error while unmarshalling")
				assert.Equal(t, tt.expected, customTime.Time, "The unmarshalled time does not match the expected value")
			}
		})
	}
}

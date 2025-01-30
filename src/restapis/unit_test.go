package restapis

import (
	"time"
	"bytes"
	"encoding/json"
	"github.com/gin-gonic/gin"
	"net/http"
	"net/http/httptest"
	"testing"
)

// TestRebateProgramCreate tests the /create-rebate-program endpoint
func TestRebateProgramCreate(t *testing.T) {
	// Create a new Gin engine
	r := gin.Default()

	// Define the route
	r.POST("/create-rebate-program", createRebateProgram)

	tests := []struct {
		name           string
		body           interface{}
		expectedStatus int
		expectedBody   string
	}{
		{
			name: "Valid request",
			body: RebateProgram{
				ID:                1,
				ProgramName:       "Winter Rebate",
				RebatePercentage:  10.5,
				StartDate:         CustomTime{Time: time.Date(2025, time.January, 27, 0, 0, 0, 0, time.UTC)},
				EndDate:           CustomTime{Time: time.Date(2025, time.December, 31, 0, 0, 0, 0, time.UTC)},
				EligibilityCriteria: "Must be a registered member",
			},
			expectedStatus: 200,
			expectedBody:   `{"message":"Rebate program created successfully","rebate_program":{"id":1,"program_name":"Winter Rebate","rebate_percentage":10.5,"start_date":"2025-01-27T00:00:00Z","end_date":"2025-12-31T00:00:00Z","eligibility_criteria":"Must be a registered member"}}`,
		},
		{
			name: "Invalid date format",
			body: `{
				"id": 1,
				"program_name": "Winter Rebate",
				"rebate_percentage": 10.5,
				"start_date": "2025/01/27",  // Invalid date format (should be 27/01/2025)
				"end_date": "31/12/2025",
				"eligibility_criteria": "Must be a registered member"
			}`,
			expectedStatus: 400,
			expectedBody:   `{"error":"Invalid request body"}`,
		},
		{
			name: "Missing required field",
			body: `{
				"program_name": "Winter Rebate",
				"rebate_percentage": 10.5,
				"start_date": "27/01/2025",
				"end_date": "31/12/2025"
			}`,
			expectedStatus: 400,
			expectedBody:   `{"error":"Invalid request body"}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Prepare the request body
			var reqBody []byte
			if str, ok := tt.body.(string); ok {
				reqBody = []byte(str)
			} else {
				var err error
				reqBody, err = json.Marshal(tt.body)
				if err != nil {
					t.Fatalf("Error marshalling body: %v", err)
				}
			}

			// Create a new HTTP request
			req, _ := http.NewRequest(http.MethodPost, "/create-rebate-program", bytes.NewBuffer(reqBody))
			req.Header.Set("Content-Type", "application/json")

			// Create a new recorder to capture the response
			w := httptest.NewRecorder()

			// Call the handler
			r.ServeHTTP(w, req)

			// Check the status code
			if w.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, w.Code)
			}

			// Check the response body
			if w.Body.String() != tt.expectedBody {
				t.Errorf("Expected body %s, got %s", tt.expectedBody, w.Body.String())
			}
		})
	}
}

package jsonrpc

import (
	"testing"
)

func TestResponse_Validate(t *testing.T) {
	tests := []struct {
		name     string
		response Response
		wantErr  bool
	}{
		{
			name: "should validate successfully with correct version and result",
			response: Response{
				ID:      IDFromStr("1"),
				Version: Version2,
				Result:  "success",
			},
			wantErr: false,
		},
		{
			name: "should fail validation with incorrect version",
			response: Response{
				ID:      IDFromStr("1"),
				Version: "1.0",
				Result:  "success",
			},
			wantErr: true,
		},
		{
			name: "should fail validation with both result and error",
			response: Response{
				ID:      IDFromStr("1"),
				Version: Version2,
				Result:  "success",
				Error:   &ResponseError{Code: 1, Message: "error"},
			},
			wantErr: true,
		},
		{
			name: "should fail validation with neither result nor error",
			response: Response{
				ID:      IDFromStr("1"),
				Version: Version2,
			},
			wantErr: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := test.response.Validate()
			if (err != nil) != test.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, test.wantErr)
			}
		})
	}
}

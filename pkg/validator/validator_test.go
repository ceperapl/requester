package validator_test

import (
	"errors"
	"testing"

	"github.com/ceperapl/requester/pkg/validator"
)

var errInvalidOption = errors.New("invalid option")

// TestNew tests the New function with different options.
func TestNew(t *testing.T) {
	t.Parallel()
	// Define a test case struct
	type testCase struct {
		name    string             // The name of the test case
		opts    []validator.Option // The options to pass to the New function
		wantErr bool               // Whether an error is expected or not
	}

	// Define some test cases
	testCases := []testCase{
		{
			name:    "No options",
			opts:    nil,
			wantErr: false,
		},
		{
			name:    "With JSON names option",
			opts:    []validator.Option{validator.WithJSONNamesForStructFields()},
			wantErr: false,
		},
		{
			name:    "With invalid option",
			opts:    []validator.Option{func(v *validator.Validation) error { return errInvalidOption }},
			wantErr: true,
		},
	}

	// Iterate over the test cases
	for _, tc := range testCases {
		tc := tc
		// Run each test case as a subtest
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			// Call the New function with the test case options
			_, err := validator.New(tc.opts...)
			// Check if the error matches the expectation
			if (err != nil) != tc.wantErr {
				t.Errorf("New() error = %v, wantErr %v", err, tc.wantErr)
			}
		})
	}
}

// TestValidateStruct tests the ValidateStruct method with different structs.
func TestValidateStruct(t *testing.T) {
	t.Parallel()
	// Define a test struct with validation tags
	type testStruct struct {
		Name  string `json:"name"  validate:"required"`
		Email string `json:"email" validate:"required,email"`
		Age   int    `json:"age"   validate:"gte=18,lte=120"`
	}

	// Define a test case struct
	type testCase struct {
		name    string      // The name of the test case
		s       interface{} // The struct to validate
		wantErr bool        // Whether an error is expected or not
	}

	// Define some test cases
	testCases := []testCase{
		{
			name: "Valid struct",
			s: testStruct{
				Name:  "Alice",
				Email: "alice@example.com",
				Age:   25,
			},
			wantErr: false,
		},
		{
			name: "Invalid struct (missing name)",
			s: testStruct{
				Email: "bob@example.com",
				Age:   30,
			},
			wantErr: true,
		},
		{
			name: "Invalid struct (invalid email)",
			s: testStruct{
				Name:  "Charlie",
				Email: "charlie",
				Age:   35,
			},
			wantErr: true,
		},
		{
			name: "Invalid struct (age out of range)",
			s: testStruct{
				Name:  "Dave",
				Email: "dave@example.com",
				Age:   150,
			},
			wantErr: true,
		},
		{
			name:    "Invalid struct (nil)",
			s:       nil,
			wantErr: true,
		},
		{
			name:    "Invalid struct (not a struct)",
			s:       "not a struct",
			wantErr: true,
		},
	}

	// Create a new validation instance
	v, err := validator.New()
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	// Iterate over the test cases
	for _, tc := range testCases {
		tc := tc
		// Run each test case as a subtest
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			// Call the ValidateStruct method with the test case struct
			err := v.ValidateStruct(tc.s)
			// Check if the error matches the expectation
			if (err != nil) != tc.wantErr {
				t.Errorf("ValidateStruct() error = %v, wantErr %v", err, tc.wantErr)
			}
		})
	}
}

// TestValidateVar tests the ValidateVar method with different variables and tags.
func TestValidateVar(t *testing.T) {
	t.Parallel()
	// Define a test case struct
	type testCase struct {
		name    string      // The name of the test case
		field   interface{} // The variable to validate
		tag     string      // The validation tag to use
		wantErr bool        // Whether an error is expected or not
	}

	// Define some test cases
	testCases := []testCase{
		{
			name:    "Valid variable (string)",
			field:   "hello",
			tag:     "required,alpha",
			wantErr: false,
		},
		{
			name:    "Invalid variable (string)",
			field:   "123",
			tag:     "required,alpha",
			wantErr: true,
		},
		{
			name:    "Valid variable (int)",
			field:   42,
			tag:     "required,gt=0",
			wantErr: false,
		},
		{
			name:    "Invalid variable (int)",
			field:   -1,
			tag:     "required,gt=0",
			wantErr: true,
		},
		{
			name:    "Invalid variable (nil)",
			field:   nil,
			tag:     "required",
			wantErr: true,
		},
	}

	// Create a new validation instance
	v, err := validator.New()
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	// Iterate over the test cases
	for _, tc := range testCases {
		tc := tc
		// Run each test case as a subtest
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			// Call the ValidateVar method with the test case variable and tag
			err := v.ValidateVar(tc.field, tc.tag)
			// Check if the error matches the expectation
			if (err != nil) != tc.wantErr {
				t.Errorf("ValidateVar() error = %v, wantErr %v", err, tc.wantErr)
			}
		})
	}
}

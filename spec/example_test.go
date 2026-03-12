package spec_test

import (
	"fmt"
	"strconv"
	"testing"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/google/uuid"
	"github.com/simcap/sfuzz/assert"
	"github.com/simcap/sfuzz/spec"
)

func TestExample(t *testing.T) {
	tests := []struct {
		in     *openapi3.Schema
		out    any
		verify func(t *testing.T, s any)
	}{
		{in: new(openapi3.Schema), out: ""}, {in: nil, out: ""},
		{
			in:  &openapi3.Schema{Example: "Example"},
			out: "Example",
		},
		{
			in:     &openapi3.Schema{Format: "uuid"},
			verify: verifyUID,
		},
		{
			in:     openapi3.NewUUIDSchema(),
			verify: verifyUID,
		},
		{
			in:     openapi3.NewFloat64Schema(),
			verify: verifyNum,
		},
		{
			in:     openapi3.NewIntegerSchema(),
			verify: verifyNum,
		},
	}
	for _, tt := range tests {
		example := spec.GenerateExample(tt.in)
		if tt.verify != nil {
			tt.verify(t, example)
		} else {
			assert.Equal(t, tt.out, example)
		}
	}
}

func verifyUID(t *testing.T, s any) {
	if err := uuid.Validate(s.(string)); err != nil {
		t.Fatalf("not a valid UUID")
	}
}

func verifyNum(t *testing.T, s any) {
	_, err := strconv.ParseFloat(fmt.Sprintf("%v", s), 64)
	if err != nil {
		t.Fatalf("cannot parse as float")
	}
	_, err = strconv.ParseInt(fmt.Sprintf("%v", s), 10, 64)
	if err != nil {
		t.Fatalf("cannot parse as fintoat")
	}
}

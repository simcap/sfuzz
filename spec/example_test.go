package spec_test

import (
	"fmt"
	"strconv"
	"testing"
	"time"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/google/uuid"
	"github.com/simcap/sfuzz/assert"
	"github.com/simcap/sfuzz/spec"
)

func TestExample(t *testing.T) {
	tests := []struct {
		in     spec.Param
		out    any
		verify func(t *testing.T, s any)
	}{
		{in: spec.Param{Schema: new(openapi3.Schema)}, out: ""},
		{in: spec.Param{}, out: ""},
		{
			in:  spec.Param{Schema: &openapi3.Schema{Example: "Example"}},
			out: "Example",
		},
		{
			in:     spec.Param{Schema: &openapi3.Schema{Format: "uuid"}},
			verify: verifyUID,
		},
		{
			in:     spec.Param{Schema: openapi3.NewUUIDSchema()},
			verify: verifyUID,
		},
		{
			in:     spec.Param{Schema: openapi3.NewFloat64Schema()},
			verify: verifyNum,
		},
		{
			in:     spec.Param{Schema: openapi3.NewIntegerSchema()},
			verify: verifyNum,
		},
		{
			in: spec.Param{
				Schema: openapi3.NewAnyOfSchema(
					openapi3.NewStringSchema(),
					&openapi3.Schema{Type: &openapi3.Types{openapi3.TypeNull}},
				),
				Name: "date_from",
			},
			verify: verifyDate,
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
func verifyDate(t *testing.T, s any) {
	if _, err := time.Parse(time.RFC3339, s.(string)); err != nil {
		t.Fatalf("not a valid RFC date: %v", s)
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

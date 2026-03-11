package spec_test

import (
	"bytes"
	_ "embed"
	"testing"

	"github.com/simcap/sfuzz/assert"
	"github.com/simcap/sfuzz/spec"
)

var (
	//go:embed testdata/example.yaml
	exampleSpec []byte
	//go:embed testdata/expected.txt
	expectedExampleFuzzFile []byte
)

func TestLoadSpec(t *testing.T) {
	oapi, err := spec.NewOAPISpec(bytes.NewReader(exampleSpec))
	assert.Equal(t, err, nil)

	var b bytes.Buffer
	err = oapi.GenerateFuzzFile(&b)
	assert.Equal(t, err, nil)

	assert.EqualBytes(t, b.Bytes(), expectedExampleFuzzFile)
}

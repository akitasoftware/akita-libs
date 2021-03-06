package spec_summary

import (
	"testing"

	"github.com/stretchr/testify/assert"

	pb "github.com/akitasoftware/akita-ir/go/api_spec"
	"github.com/akitasoftware/akita-libs/test"
)

func TestSummarize(t *testing.T) {
	expected := &Summary{
		Authentications: map[string]int{
			"BASIC": 1,
		},
		HTTPMethods: map[string]int{
			"POST": 1,
		},
		Paths: map[string]int{
			"/v1/projects/{arg3}": 1,
		},
		Params: map[string]int{
			"X-My-Header": 1,
		},
		Properties: map[string]int{
			"top-level-prop":       1,
			"my-special-prop":      1,
			"other-top-level-prop": 1,
		},
		ResponseCodes: map[string]int{
			"200": 1,
		},
		DataFormats: map[string]int{
			"rfc3339": 1,
		},
		DataKinds: map[string]int{},
		DataTypes: map[string]int{
			"string": 1,
		},
	}

	m1 := test.LoadMethodFromFileOrDie("testdata/method1.pb.txt")
	assert.Equal(t, expected, Summarize(&pb.APISpec{Methods: []*pb.Method{m1}}))
}

func TestIntersect(t *testing.T) {
	m1 := test.LoadMethodFromFileOrDie("testdata/method1.pb.txt")
	m2 := test.LoadMethodFromFileOrDie("testdata/method1.pb.txt")

	setM1 := make(map[*pb.Method]struct{})
	setM1[m1] = struct{}{}

	setM2 := make(map[*pb.Method]struct{})
	setM2[m2] = struct{}{}

	setM12 := make(map[*pb.Method]struct{})
	setM12[m1] = struct{}{}
	setM12[m2] = struct{}{}

	emptyset := make(map[*pb.Method]struct{})

	assert.Equal(t, emptyset, intersect([]map[*pb.Method]struct{}{setM1, setM2}))
	assert.Equal(t, emptyset, intersect([]map[*pb.Method]struct{}{emptyset, setM2}))
	assert.Equal(t, emptyset, intersect([]map[*pb.Method]struct{}{setM1, emptyset}))
	assert.Equal(t, setM1, intersect([]map[*pb.Method]struct{}{setM1, setM12}))
	assert.Equal(t, setM1, intersect([]map[*pb.Method]struct{}{setM12, setM1}))
	assert.Equal(t, setM2, intersect([]map[*pb.Method]struct{}{setM2, setM12}))
	assert.Equal(t, setM2, intersect([]map[*pb.Method]struct{}{setM12, setM2}))
	assert.Equal(t, setM12, intersect([]map[*pb.Method]struct{}{setM12, setM12}))
}
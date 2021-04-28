package spec_util

import (
	pb "github.com/akitasoftware/akita-ir/go/api_spec"
	"github.com/pkg/errors"

	"github.com/akitasoftware/akita-libs/pbhash"
	"github.com/akitasoftware/akita-libs/visitors/go_ast"
	"github.com/akitasoftware/akita-libs/visitors/http_rest"
)

type hashOneOfVisitor struct {
	http_rest.DefaultHttpRestSpecVisitor
	err error
}

func (vis *hashOneOfVisitor) VisitData(c http_rest.HttpRestSpecVisitorContext, p *pb.Data) bool {
	oneOf := p.GetOneof()
	if oneOf == nil {
		return true
	}
	options := p.GetOneof().Options
	for k, v := range options {
		h, err := pbhash.HashProto(v)
		if err != nil {
			vis.err = err
			return false
		}
		if k != h {
			delete(options, k)
			options[h] = v
		}
	}
	return true
}

// Three maps in the IR use hashes of the values as keys (i.e. map[hash(v)] = v):
//  - Method.Args
//  - Method.Responses
//  - OneOf.Options
//
// This method traverses the spec, recomputes the hash of each value, and updates the map.
func RewriteHashKeys(spec *pb.APISpec) error {
	// Hash OneOf values in postorder, so that children are updated before computing the
	// new hash for the parent.
	v := &hashOneOfVisitor{}
	http_rest.Apply(go_ast.POSTORDER, v, spec)
	if v.err != nil {
		return errors.Wrap(v.err, "failed to compute hash of oneOf data")
	}

	// Hash Args and Responses for each method.
	for _, method := range spec.Methods {
		for _, m := range []map[string]*pb.Data{method.Args, method.Responses} {
			for k, arg := range m {
				h, err := pbhash.HashProto(arg)
				if err != nil {
					return errors.Wrap(err, "failed to compute hash of method argument")
				}
				if h != k {
					delete(m, k)
					m[h] = arg
				}
			}
		}
	}

	return nil
}

package http_rest

import (
	"reflect"
	"strconv"
	"strings"

	pb "github.com/akitasoftware/akita-ir/go/api_spec"
	"github.com/golang/glog"
	"github.com/pkg/errors"

	. "github.com/akitasoftware/akita-libs/visitors"
	"github.com/akitasoftware/akita-libs/visitors/go_ast"
)

// Represents the "name" of an argument to a method.
type ArgName interface {
	String() string

	isArgName()
}

// Returns a normalization map for the keys in the given method-argument map.
// The return value maps normalized keys to keys in the given map.
//
// In IR method objects, arguments to requests and responses are indexed by
// the arguments' hashes. However, because arguments are not normalized,
// equivalent arguments can have different hashes, so these indices are not
// useful for determining whether and how a method has changed.
func GetNormalizedArgNames(args map[string]*pb.Data, methodMeta *pb.HTTPMethodMeta) (map[ArgName]string, error) {
	// XXX Ignoring errors from the normalizer regarding non-HTTP metadata.
	normalizer := newArgNameNormalizer(methodMeta)
	Apply(normalizer, args)
	return normalizer.normalizationMap, normalizer.err
}

type argNameNormalizer struct {
	DefaultSpecVisitorImpl

	// Metadata for the method whose argument names are being normalized.
	methodMeta *pb.HTTPMethodMeta

	// The arg name being normalized.
	nonNormalizedArgName string

	// The output of the normalization.
	normalizationMap map[ArgName]string

	// Contains any arguments encountered with non-HTTP metadata.
	nonHTTPArgs []*pb.Data

	err error
}

var _ DefaultSpecVisitor = (*argNameNormalizer)(nil)

func newArgNameNormalizer(methodMeta *pb.HTTPMethodMeta) *argNameNormalizer {
	return &argNameNormalizer{
		methodMeta:       methodMeta,
		normalizationMap: make(map[ArgName]string),
	}
}

func (v *argNameNormalizer) EnterData(self interface{}, ctx SpecVisitorContext, arg *pb.Data) Cont {
	// Set our context.
	v.nonNormalizedArgName = ctx.GetPath().GetLast().OutEdge.String()

	// Make sure we have HTTP metadata for the current argument.
	argMeta := arg.GetMeta().GetHttp()
	if argMeta == nil {
		v.nonHTTPArgs = append(v.nonHTTPArgs, arg)
		return SkipChildren
	}

	return Continue
}

func (*argNameNormalizer) VisitDataChildren(self interface{}, c SpecVisitorContext, vm VisitorManager, arg *pb.Data) Cont {
	// Only visit the argument's metadata.
	return go_ast.ApplyWithContext(vm, c.EnterStruct(arg, "Meta"), arg.GetMeta())
}

func (v *argNameNormalizer) setName(name ArgName) {
	if existing, ok := v.normalizationMap[name]; ok {
		glog.Warning("Non-normalized arg names ", existing, " and ", v.nonNormalizedArgName, " have conflicting normalized name ", name, ". Keeping the former and ignoring the latter.")
		return
	}

	v.normalizationMap[name] = v.nonNormalizedArgName
}

func (v *argNameNormalizer) LeaveData(self interface{}, _ SpecVisitorContext, _ *pb.Data, cont Cont) Cont {
	v.nonNormalizedArgName = ""
	return cont
}

// == Path parameters =========================================================

type pathName struct {
	index int
}

var _ ArgName = (*pathName)(nil)

func (pathName) isArgName() {}

func (v *argNameNormalizer) EnterHTTPPath(self interface{}, _ SpecVisitorContext, path *pb.HTTPPath) Cont {
	template := v.methodMeta.GetPathTemplate()
	components := strings.Split(template, "/")
	for idx, component := range components {
		if component == "{"+path.GetKey()+"}" {
			v.setName(pathName{
				index: idx,
			})
			return SkipChildren
		}
	}

	v.err = errors.Errorf("Path parameter %q not found in %q", path.GetKey(), template)
	return SkipChildren
}

func (n pathName) String() string {
	return strconv.Itoa(n.index)
}

// == Query parameters ========================================================

type queryName struct {
	name string
}

var _ ArgName = (*queryName)(nil)

func (queryName) isArgName() {}

func (v *argNameNormalizer) EnterHTTPQuery(self interface{}, _ SpecVisitorContext, query *pb.HTTPQuery) Cont {
	v.setName(queryName{
		name: query.GetKey(),
	})
	return SkipChildren
}

func (n queryName) String() string {
	return n.name
}

// == Header parameters =======================================================

type headerName struct {
	name string
}

var _ ArgName = (*headerName)(nil)

func (headerName) isArgName() {}

func (v *argNameNormalizer) EnterHTTPHeader(self interface{}, _ SpecVisitorContext, header *pb.HTTPHeader) Cont {
	v.setName(headerName{
		name: header.GetKey(),
	})
	return SkipChildren
}

func (n headerName) String() string {
	return n.name
}

// == Cookie parameters =======================================================

type cookieName struct {
	name string
}

var _ ArgName = (*cookieName)(nil)

func (cookieName) isArgName() {}

func (v *argNameNormalizer) EnterHTTPCookie(self interface{}, _ SpecVisitorContext, cookie *pb.HTTPCookie) Cont {
	v.setName(cookieName{
		name: cookie.GetKey(),
	})
	return SkipChildren
}

func (n cookieName) String() string {
	return n.name
}

// == Body parameters =========================================================

type bodyName struct {
	// Request or response
	isResponse bool

	contentType  string
	responseCode int32
}

var _ ArgName = (*bodyName)(nil)

func (bodyName) isArgName() {}

func (v *argNameNormalizer) EnterHTTPBody(self interface{}, ctx SpecVisitorContext, body *pb.HTTPBody) Cont {
	// Assumes there is at most one per:
	// - method request x content type,
	// - method response x response code x content type.
	node, _ := ctx.GetInnermostNode(reflect.TypeOf((*pb.Data)(nil)))
	data := node.(*pb.Data)
	v.setName(bodyName{
		isResponse:   ctx.IsResponse(),
		contentType:  body.GetContentType().String(),
		responseCode: data.GetMeta().GetHttp().ResponseCode,
	})
	return SkipChildren
}

func (n bodyName) String() string {
	return "(body)"
}

// == Empty parameters ========================================================

type emptyName struct{}

var _ ArgName = (*bodyName)(nil)

func (emptyName) isArgName() {}

func (v *argNameNormalizer) EnterHTTPEmpty(self interface{}, _ SpecVisitorContext, empty *pb.HTTPEmpty) Cont {
	// Assumes there is at most one per method.
	v.setName(emptyName{})
	return SkipChildren
}

func (n emptyName) String() string {
	return ""
}

// == Auth parameters =========================================================

type authName struct{}

var _ ArgName = (*authName)(nil)

func (authName) isArgName() {}

func (v *argNameNormalizer) EnterAuth(_ SpecVisitorContext, auth *pb.HTTPAuth) Cont {
	// Assumes there is at most one per method.
	v.setName(authName{})
	return SkipChildren
}

func (n authName) String() string {
	return "Authorization"
}

// == Multipart body parameters ===============================================

type multipartName struct{}

var _ ArgName = (*multipartName)(nil)

func (multipartName) isArgName() {}

func (v *argNameNormalizer) EnterMultipart(_ SpecVisitorContext, multipart *pb.HTTPMultipart) Cont {
	// Assumes there is at most one per method.
	v.setName(multipartName{})
	return SkipChildren
}

func (n multipartName) String() string {
	return "(multipart)"
}

package http_rest

import (
	"reflect"

	"github.com/akitasoftware/akita-libs/visitors"
)

// Basically a pair of HttpRestSpecVisitorContexts. See that class for more
// information.
type SpecPairVisitorContext interface {
	visitors.PairContext

	ExtendLeftContext(leftNode interface{})
	ExtendRightContext(rightNode interface{})

	GetLeftContext() SpecVisitorContext
	GetRightContext() SpecVisitorContext
	SplitContext() (SpecVisitorContext, SpecVisitorContext)

	GetRestPaths() ([]string, []string)
	GetRestOperations() (string, string)
	setRestOperations(string, string)
	IsArg() (bool, bool)
	IsResponse() (bool, bool)
	GetValueTypes() (HttpValueType, HttpValueType)
	GetArgPaths() ([]string, []string)
	GetResponsePaths() ([]string, []string)
	GetEndpointPaths() (string, string)
	GetHosts() (string, string)

	// Returns, for each side, the innermost node being visited having the given
	// type, and the nodes' context.
	GetInnermostNode(reflect.Type) (left, right interface{}, ctxt SpecPairVisitorContext)

	appendRestPaths(string, string)
	setIsArg(bool, bool)
	setValueType(HttpValueType, HttpValueType)
	setTopLevelDataIndex(int, int)
}

type specPairVisitorContext struct {
	left  SpecVisitorContext
	right SpecVisitorContext
}

var _ SpecPairVisitorContext = (*specPairVisitorContext)(nil)

func newSpecPairVisitorContext() SpecPairVisitorContext {
	return &specPairVisitorContext{
		left:  &specVisitorContext{},
		right: &specVisitorContext{},
	}
}

func (c *specPairVisitorContext) ExtendLeftContext(leftNode interface{}) {
	extendContext(c.left, leftNode)
}

func (c *specPairVisitorContext) ExtendRightContext(rightNode interface{}) {
	extendContext(c.right, rightNode)
}

func (c *specPairVisitorContext) GetLeftContext() SpecVisitorContext {
	return c.left
}

func (c *specPairVisitorContext) GetRightContext() SpecVisitorContext {
	return c.right
}

func (c *specPairVisitorContext) SplitContext() (SpecVisitorContext, SpecVisitorContext) {
	return c.left, c.right
}

// Returns a new PairContext in which the path on one side is appended to
// indicate a traversal into the given field of the given struct.
func (c *specPairVisitorContext) EnterStruct(leftOrRight visitors.LeftOrRight, structNode interface{}, fieldName string) visitors.PairContext {
	if leftOrRight == visitors.LeftSide {
		return &specPairVisitorContext{
			left:  c.left.EnterStruct(structNode, fieldName).(*specVisitorContext),
			right: c.right,
		}
	}

	return &specPairVisitorContext{
		left:  c.left,
		right: c.right.EnterStruct(structNode, fieldName).(*specVisitorContext),
	}
}

func (c *specPairVisitorContext) EnterStructs(leftStruct interface{}, leftFieldName string, rightStruct interface{}, rightFieldName string) visitors.PairContext {
	return &specPairVisitorContext{
		left:  c.left.EnterStruct(leftStruct, leftFieldName).(*specVisitorContext),
		right: c.right.EnterStruct(rightStruct, rightFieldName).(*specVisitorContext),
	}
}

// Returns a new PairContext in which the path on one side is appended to
// indicate a traversal into the given element of the given array.
func (c *specPairVisitorContext) EnterArray(leftOrRight visitors.LeftOrRight, arrayNode interface{}, elementIndex int) visitors.PairContext {
	if leftOrRight == visitors.LeftSide {
		return &specPairVisitorContext{
			left:  c.left.EnterArray(arrayNode, elementIndex).(*specVisitorContext),
			right: c.right,
		}
	}

	return &specPairVisitorContext{
		left:  c.left,
		right: c.right.EnterArray(arrayNode, elementIndex).(*specVisitorContext),
	}
}

// Returns a new PairContext in which the paths are appended to indicate a
// traversal into the given elements of the given arrays.
func (c *specPairVisitorContext) EnterArrays(leftArray interface{}, leftIndex int, rightArray interface{}, rightIndex int) visitors.PairContext {
	return &specPairVisitorContext{
		left:  c.left.EnterArray(leftArray, leftIndex).(*specVisitorContext),
		right: c.right.EnterArray(rightArray, rightIndex).(*specVisitorContext),
	}
}

// Returns a new PairContext in which the path on one side is appended to
// indicate a traversal into the value at the given key of the given map.
func (c *specPairVisitorContext) EnterMapValue(leftOrRight visitors.LeftOrRight, mapNode, mapKey interface{}) visitors.PairContext {
	if leftOrRight == visitors.LeftSide {
		return &specPairVisitorContext{
			left:  c.left.EnterMapValue(mapNode, mapKey).(*specVisitorContext),
			right: c.right,
		}
	}
	return &specPairVisitorContext{
		left:  c.left,
		right: c.right.EnterMapValue(mapNode, mapKey).(*specVisitorContext),
	}
}

// Returns a new PairContext in which the paths are appended to indicate a
// traversal into the values at the given keys of the given maps.
func (c *specPairVisitorContext) EnterMapValues(leftMap, leftKey, rightMap, rightKey interface{}) visitors.PairContext {
	return &specPairVisitorContext{
		left:  c.left.EnterMapValue(leftMap, leftKey).(*specVisitorContext),
		right: c.right.EnterMapValue(rightMap, rightKey).(*specVisitorContext),
	}
}

func (c *specPairVisitorContext) GetPaths() (visitors.ContextPath, visitors.ContextPath) {
	return c.left.GetPath(), c.right.GetPath()
}

func (c *specPairVisitorContext) appendRestPaths(left, right string) {
	c.left.appendRestPath(left)
	c.right.appendRestPath(right)
}

func (c *specPairVisitorContext) GetRestPaths() ([]string, []string) {
	return c.left.GetRestPath(), c.right.GetRestPath()
}

func (c *specPairVisitorContext) GetRestOperations() (string, string) {
	return c.left.GetRestOperation(), c.right.GetRestOperation()
}

func (c *specPairVisitorContext) setRestOperations(left, right string) {
	c.left.setRestOperation(left)
	c.right.setRestOperation(right)
}

func (c *specPairVisitorContext) IsArg() (bool, bool) {
	return c.left.IsArg(), c.right.IsArg()
}

func (c *specPairVisitorContext) IsResponse() (bool, bool) {
	return c.left.IsResponse(), c.right.IsResponse()
}

func (c *specPairVisitorContext) GetValueTypes() (HttpValueType, HttpValueType) {
	return c.left.GetValueType(), c.right.GetValueType()
}

func (c *specPairVisitorContext) GetArgPaths() ([]string, []string) {
	return c.left.GetArgPath(), c.right.GetArgPath()
}

func (c *specPairVisitorContext) GetResponsePaths() ([]string, []string) {
	return c.left.GetResponsePath(), c.right.GetResponsePath()
}

func (c *specPairVisitorContext) GetEndpointPaths() (string, string) {
	return c.left.GetEndpointPath(), c.right.GetEndpointPath()
}

func (c *specPairVisitorContext) GetHosts() (string, string) {
	return c.left.GetHost(), c.right.GetHost()
}

func (c *specPairVisitorContext) GetInnermostNode(typ reflect.Type) (left, right interface{}, ctxt SpecPairVisitorContext) {
	left, leftCtxt := c.left.GetInnermostNode(typ)
	right, rightCtxt := c.right.GetInnermostNode(typ)
	ctxt = &specPairVisitorContext{
		left:  leftCtxt,
		right: rightCtxt,
	}
	return left, right, ctxt
}

func (c *specPairVisitorContext) setIsArg(left, right bool) {
	c.left.setIsArg(left)
	c.right.setIsArg(right)
}

func (c *specPairVisitorContext) setValueType(left, right HttpValueType) {
	c.left.setValueType(left)
	c.right.setValueType(right)
}

func (c *specPairVisitorContext) setTopLevelDataIndex(left, right int) {
	c.left.setTopLevelDataIndex(left)
	c.right.setTopLevelDataIndex(right)
}

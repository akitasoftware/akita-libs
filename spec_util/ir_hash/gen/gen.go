package main

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/format"
	"go/token"
	"os"
	"reflect"
	"sort"
	"strings"

	pb "github.com/akitasoftware/akita-ir/go/api_spec"

	"github.com/OneOfOne/xxhash"
	"github.com/golang/protobuf/proto"
	wrappers "github.com/golang/protobuf/ptypes/wrappers"
)

// Generate hashing code for Akita IR.
// Rather than (1) parsing the .proto file, or (2) parsing the Go file,
// it'll be easiest to reproduce objecthash_proto behavior if we
// (3) use reflection on the compiled object types.  The downside is
// we have to list them all here.

func main() {
	gf := &GeneratedFile{
		File: &ast.File{},
	}
	gf.SetPackageName("ir_hash")
	gf.AddImports()

	// Dependencies from "wrappers" package
	gf.AddHashFunc(reflect.TypeOf(wrappers.Int32Value{}))
	gf.AddHashFunc(reflect.TypeOf(wrappers.Int64Value{}))
	gf.AddHashFunc(reflect.TypeOf(wrappers.UInt32Value{}))
	gf.AddHashFunc(reflect.TypeOf(wrappers.UInt64Value{}))
	gf.AddHashFunc(reflect.TypeOf(wrappers.FloatValue{}))
	gf.AddHashFunc(reflect.TypeOf(wrappers.DoubleValue{}))

	// TODO: inner structs for oneOf need to be defined before their use, automate this somewhow?
	gf.AddOneOf(reflect.TypeOf(pb.Primitive_BoolValue{}))
	gf.AddOneOf(reflect.TypeOf(pb.Primitive_BytesValue{}))
	gf.AddOneOf(reflect.TypeOf(pb.Primitive_FloatValue{}))
	gf.AddOneOf(reflect.TypeOf(pb.Primitive_DoubleValue{}))
	gf.AddOneOf(reflect.TypeOf(pb.Primitive_StringValue{}))
	gf.AddOneOf(reflect.TypeOf(pb.Primitive_Int32Value{}))
	gf.AddOneOf(reflect.TypeOf(pb.Primitive_Int64Value{}))
	gf.AddOneOf(reflect.TypeOf(pb.Primitive_Uint32Value{}))
	gf.AddOneOf(reflect.TypeOf(pb.Primitive_Uint64Value{}))
	gf.AddOneOf(reflect.TypeOf(pb.FormatOption_StringFormat{}))
	gf.AddOneOf(reflect.TypeOf(pb.DataMeta_Grpc{}))
	gf.AddOneOf(reflect.TypeOf(pb.DataMeta_Http{}))
	gf.AddOneOf(reflect.TypeOf(pb.DataRef_ListRef{}))
	gf.AddOneOf(reflect.TypeOf(pb.DataRef_PrimitiveRef{}))
	gf.AddOneOf(reflect.TypeOf(pb.DataRef_StructRef{}))

	gf.AddHashFunc(reflect.TypeOf(pb.AkitaAnnotations{}))
	//gf.AddHashFunc(reflect.TypeOf(pb.APISpec{}))
	gf.AddHashFunc(reflect.TypeOf(pb.Bool{}))
	gf.AddHashFunc(reflect.TypeOf(pb.BoolType{}))
	gf.AddHashFunc(reflect.TypeOf(pb.Bytes{}))
	gf.AddHashFunc(reflect.TypeOf(pb.BytesType{}))
	//gf.AddHashFunc(reflect.TypeOf(pb.Data{}))
	//gf.AddHashFunc(reflect.TypeOf(pb.DataMeta{}))
	//gf.AddHashFunc(reflect.TypeOf(pb.DataRef{}))
	gf.AddHashFunc(reflect.TypeOf(pb.Double{}))
	gf.AddHashFunc(reflect.TypeOf(pb.DoubleType{}))
	gf.AddHashFunc(reflect.TypeOf(pb.Int32{}))
	gf.AddHashFunc(reflect.TypeOf(pb.Int32Type{}))
	gf.AddHashFunc(reflect.TypeOf(pb.Int64{}))
	gf.AddHashFunc(reflect.TypeOf(pb.Int64Type{}))
	gf.AddHashFunc(reflect.TypeOf(pb.Float{}))
	gf.AddHashFunc(reflect.TypeOf(pb.FloatType{}))
	gf.AddHashFunc(reflect.TypeOf(pb.FormatOption{}))
	gf.AddHashFunc(reflect.TypeOf(pb.Primitive{}))
	gf.AddHashFunc(reflect.TypeOf(pb.StringType{}))
	gf.AddHashFunc(reflect.TypeOf(pb.String{}))
	gf.AddHashFunc(reflect.TypeOf(pb.Uint32{}))
	gf.AddHashFunc(reflect.TypeOf(pb.Uint32Type{}))
	gf.AddHashFunc(reflect.TypeOf(pb.Uint64{}))
	gf.AddHashFunc(reflect.TypeOf(pb.Uint64Type{}))

	/*
		DataRef_ListRef
		DataRef_PrimitiveRef
		DataRef_StructRef
		DataTemplate
		DataTemplate_ListTemplate
		DataTemplate_OptionalTemplate
		DataTemplate_Ref
		DataTemplate_StructTemplate
		DataTemplate_Value
		Data_List
		Data_Oneof
		Data_Optional
		Data_Primitive
		Data_Struct
		Double
		DoubleType
		ExampleValue
		FormatOption
		FormatOption_StringFormat
		GRPCMeta
		GRPCMethodMeta
		HTTPAuth
		HTTPBody
		HTTPCookie
		HTTPEmpty
		HTTPHeader
		HTTPMeta
		HTTPMeta_Auth
		HTTPMeta_Body
		HTTPMeta_Cookie
		HTTPMeta_Empty
		HTTPMeta_Header
		HTTPMeta_Multipart
		HTTPMeta_Path
		HTTPMeta_Query
		HTTPMethodMeta
		HTTPMultipart
		HTTPPath
		HTTPQuery
		IndexedDataRef
		List
		ListRef
		ListRef_ElemRef
		ListRef_FullList
		ListRef_FullListRef
		ListTemplate
		Method
		MethodCalls
		MethodDataRef
		MethodDataRef_ArgRef
		MethodDataRef_ResponseRef
		MethodID
		MethodMeta
		MethodMeta_Grpc
		MethodMeta_Http
		MethodTemplate
		NamedDataRef
		None
		OneOf
		Optional
		OptionalTemplate
		Optional_Data
		Optional_None
		Primitive
		PrimitiveRef
		PrimitiveRef_BoolType
		PrimitiveRef_BytesType
		PrimitiveRef_DoubleType
		PrimitiveRef_FloatType
		PrimitiveRef_Int32Type
		PrimitiveRef_Int64Type
		PrimitiveRef_StringType
		PrimitiveRef_Uint32Type
		PrimitiveRef_Uint64Type
		Primitive_BoolValue
		Primitive_BytesValue
		Primitive_DoubleValue
		Primitive_FloatValue
		Primitive_Int32Value
		Primitive_Int64Value
		Primitive_StringValue
		Primitive_Uint32Value
		Primitive_Uint64Value
		Sequence
		SequenceRun
		String
		StringType
		Struct
		StructRef
		StructRef_FieldRef
		StructRef_FullStruct
		StructRef_FullStructRef
		StructTemplate
		Witness
	*/

	fset := token.NewFileSet()
	format.Node(os.Stdout, fset, gf.File)
}

type OneOfInnerStruct struct {
	Type reflect.Type
	Hash []byte
	Code func(string) ast.Stmt
}

type GeneratedFile struct {
	File *ast.File

	OneOfTypes []OneOfInnerStruct
}

func (f *GeneratedFile) SetPackageName(pn string) {
	f.File.Name = ast.NewIdent(pn)
}

func (f *GeneratedFile) AddImports() {
	f.File.Imports = []*ast.ImportSpec{
		&ast.ImportSpec{
			Name: ast.NewIdent("pb"),
			Path: &ast.BasicLit{
				Kind:  token.STRING,
				Value: "\"github.com/akitasoftware/akita-ir/go/api_spec\"",
			},
		},
		&ast.ImportSpec{
			Name: ast.NewIdent("wrappers"),
			Path: &ast.BasicLit{
				Kind:  token.STRING,
				Value: "\"github.com/golang/protobuf/ptypes/wrappers\"",
			},
		},
		&ast.ImportSpec{
			Path: &ast.BasicLit{
				Kind:  token.STRING,
				Value: "\"github.com/OneOfOne/xxhash\"",
			},
		},
	}
	f.File.Decls = append(f.File.Decls,
		&ast.GenDecl{
			Tok: token.IMPORT,
			Specs: []ast.Spec{
				f.File.Imports[0],
				f.File.Imports[1],
				f.File.Imports[2],
			},
		},
	)
}

func (f *GeneratedFile) AddFunction(decl *ast.FuncDecl) {
	f.File.Decls = append(f.File.Decls, decl)
}

// Add an inner struct type
func (f *GeneratedFile) AddOneOf(structType reflect.Type) {
	props := proto.GetProperties(structType)

	// Extract the type the inner struct points to
	nestedField := structType.Field(0)
	tag := int64(props.Prop[0].Tag)

	var innerHash ast.Stmt

	// FIXME: handle other types here?
	if nestedField.Type.Kind() == reflect.String {
		innerHash = &ast.ExprStmt{
			X: &ast.CallExpr{
				Fun: selector2("hash", "Write"),
				Args: []ast.Expr{
					&ast.CallExpr{
						Fun: ast.NewIdent("Hash_Unicode"),
						Args: []ast.Expr{
							selector2("val", nestedField.Name),
						},
					},
				},
			},
		}
	} else {
		nestedFieldType := nestedField.Type.Elem()
		innerHash = hashOtherStruct("val",
			nestedField.Name,
			nestedFieldType.Name())
	}

	// Build an if-statement from a field name within "node"
	factory := func(fieldName string) ast.Stmt {
		return &ast.IfStmt{
			Init: &ast.AssignStmt{
				Lhs: []ast.Expr{
					ast.NewIdent("val"),
					ast.NewIdent("ok"),
				},
				Tok: token.DEFINE,
				Rhs: []ast.Expr{
					&ast.TypeAssertExpr{
						X: selector2("node", fieldName),
						Type: &ast.StarExpr{
							X: selector2("pb", structType.Name()),
						},
					},
				},
			},
			Cond: ast.NewIdent("ok"),
			Body: &ast.BlockStmt{
				List: []ast.Stmt{
					hashPrecomputedInt(tag),
					innerHash,
				},
			},
		}
	}

	f.OneOfTypes = append(f.OneOfTypes,
		OneOfInnerStruct{
			// The pointer type is needed to satisfy the interface
			Type: reflect.PtrTo(structType),
			Hash: Hash_Int64(tag),
			Code: factory,
		},
	)
}

func byteArray() ast.Expr {
	return &ast.ArrayType{
		Len: nil,
		Elt: ast.NewIdent("byte"),
	}
}

func nodeArg(pack string, message string) *ast.Field {
	return &ast.Field{
		Names: []*ast.Ident{
			ast.NewIdent("node"),
		},
		Type: &ast.StarExpr{
			X: selector2(pack, message),
		},
	}
}

func selector2(first string, second string) *ast.SelectorExpr {
	return &ast.SelectorExpr{
		X:   ast.NewIdent(first),
		Sel: ast.NewIdent(second),
	}
}

func startHash() ast.Stmt {
	return &ast.AssignStmt{
		Lhs: []ast.Expr{
			ast.NewIdent("hash"),
		},
		Tok: token.DEFINE,
		Rhs: []ast.Expr{
			&ast.CallExpr{
				Fun:  selector2("xxhash", "New64"),
				Args: nil,
			},
		},
	}
}

func hashStringLiteral(hashVar string, s string) ast.Stmt {
	return &ast.ExprStmt{
		X: &ast.CallExpr{
			Fun: selector2(hashVar, "Write"),
			Args: []ast.Expr{
				&ast.CallExpr{
					Fun: &ast.ArrayType{
						Elt: ast.NewIdent("byte"),
					},
					Args: []ast.Expr{
						&ast.BasicLit{
							Kind:  token.STRING,
							Value: fmt.Sprintf("%q", s),
						},
					},
				},
			},
		},
	}
}

func hashPrecomputedInt(i int64) ast.Stmt {
	return &ast.ExprStmt{
		X: &ast.CallExpr{
			Fun: selector2("hash", "Write"),
			Args: []ast.Expr{
				&ast.IndexExpr{
					X: ast.NewIdent("intHashes"),
					Index: &ast.BasicLit{
						Kind:  token.INT,
						Value: fmt.Sprintf("%d", i),
					},
				},
			},
		},
	}
}

func hashOtherStruct(localVar string, fieldName string, fieldType string) ast.Stmt {
	return &ast.ExprStmt{
		X: &ast.CallExpr{
			Fun: selector2("hash", "Write"),
			Args: []ast.Expr{
				&ast.CallExpr{
					Fun: ast.NewIdent(fmt.Sprintf("Hash%v", fieldType)),
					Args: []ast.Expr{
						selector2(localVar, fieldName),
					},
				},
			},
		},
	}
}

func returnHash() ast.Stmt {
	return &ast.ReturnStmt{
		Results: []ast.Expr{
			&ast.CallExpr{
				Fun: selector2("hash", "Sum"),
				Args: []ast.Expr{
					ast.NewIdent("nil"),
				},
			},
		},
	}
}

func hashAfterComparison(localVar string, fieldName string,
	nullValue ast.Expr,
	tag int64,
	hashFunction string) ast.Stmt {
	return &ast.IfStmt{
		Cond: &ast.BinaryExpr{
			X:  selector2(localVar, fieldName),
			Op: token.NEQ,
			Y:  nullValue,
		},
		Body: &ast.BlockStmt{
			List: []ast.Stmt{
				hashPrecomputedInt(tag),
				&ast.ExprStmt{
					X: &ast.CallExpr{
						Fun: selector2("hash", "Write"),
						Args: []ast.Expr{
							&ast.CallExpr{
								Fun: ast.NewIdent(hashFunction),
								Args: []ast.Expr{
									selector2(localVar, fieldName),
								},
							},
						},
					},
				},
			},
		},
	}
}

func hashGolangString(localVar string, fieldName string, tag int64) ast.Stmt {
	return hashAfterComparison(localVar, fieldName,
		&ast.BasicLit{
			Kind:  token.STRING,
			Value: "\"\"",
		},
		tag, "Hash_Unicode")
}

func hashGolangBool(localVar string, fieldName string, tag int64) ast.Stmt {
	return hashAfterComparison(localVar, fieldName,
		ast.NewIdent("false"),
		tag, "Hash_Bool")
}

func hashGolangInt32(localVar string, fieldName string, tag int64) ast.Stmt {
	return hashAfterComparison(localVar, fieldName,
		&ast.BasicLit{
			Kind:  token.INT,
			Value: "0",
		},
		tag, "Hash_Int32")
}

func hashGolangInt64(localVar string, fieldName string, tag int64) ast.Stmt {
	return hashAfterComparison(localVar, fieldName,
		&ast.BasicLit{
			Kind:  token.INT,
			Value: "0",
		},
		tag, "Hash_Int64")
}

func hashGolangUint32(localVar string, fieldName string, tag int64) ast.Stmt {
	return hashAfterComparison(localVar, fieldName,
		&ast.BasicLit{
			Kind:  token.INT,
			Value: "0",
		},
		tag, "Hash_Uint32")
}

func hashGolangUint64(localVar string, fieldName string, tag int64) ast.Stmt {
	return hashAfterComparison(localVar, fieldName,
		&ast.BasicLit{
			Kind:  token.INT,
			Value: "0",
		},
		tag, "Hash_Uint64")
}

func hashGolangFloat32(localVar string, fieldName string, tag int64) ast.Stmt {
	return hashAfterComparison(localVar, fieldName,
		&ast.BasicLit{
			Kind:  token.FLOAT,
			Value: "0.0",
		},
		tag, "Hash_Float32")
}

func hashGolangFloat64(localVar string, fieldName string, tag int64) ast.Stmt {
	return hashAfterComparison(localVar, fieldName,
		&ast.BasicLit{
			Kind:  token.FLOAT,
			Value: "0.0",
		},
		tag, "Hash_Float64")
}

func type2HashFunc(t reflect.Type) string {
	switch t.Kind() {
	case reflect.String:
		return "Hash_Unicode"
	case reflect.Bool:
		return "Hash_Bool"
	case reflect.Int32:
		return "Hash_Int32"
	case reflect.Int64:
		return "Hash_Int64"
	case reflect.Uint32:
		return "Hash_Uint32"
	case reflect.Uint64:
		return "Hash_Uint64"
	case reflect.Float32:
		return "Hash_Float32"
	case reflect.Float64:
		return "Hash_Float64"
	case reflect.Ptr:
		return fmt.Sprintf("Hash%v", t.Elem().Name())
	case reflect.Slice:
		// handle [][]byte but not other [][]X
		if t.Elem().Kind() == reflect.Uint8 {
			return "Hash_Bytes"
		}
		return "UNKNOWNSLICE"
	default:
		return "UNKNOWN"
	}
}

func hashRepeatedValue(localVar string, field reflect.StructField, tag int64) ast.Stmt {
	startListHash := &ast.AssignStmt{
		Lhs: []ast.Expr{
			ast.NewIdent("listHash"),
		},
		Tok: token.DEFINE,
		Rhs: []ast.Expr{
			&ast.CallExpr{
				Fun: selector2("xxhash", "New64"),
			},
		},
	}
	finishListHash := &ast.ExprStmt{
		X: &ast.CallExpr{
			Fun: selector2("hash", "Write"),
			Args: []ast.Expr{
				&ast.CallExpr{
					Fun: selector2("listHash", "Sum"),
					Args: []ast.Expr{
						ast.NewIdent("nil"),
					},
				},
			},
		},
	}

	hashExpr := &ast.CallExpr{
		Fun: ast.NewIdent(type2HashFunc(field.Type.Elem())),
		Args: []ast.Expr{
			ast.NewIdent("v"),
		},
	}

	// Check length against zero
	return &ast.IfStmt{
		Cond: &ast.BinaryExpr{
			X: &ast.CallExpr{
				Fun: ast.NewIdent("len"),
				Args: []ast.Expr{
					selector2(localVar, field.Name),
				},
			},
			Op: token.NEQ,
			Y: &ast.BasicLit{
				Kind:  token.INT,
				Value: "0",
			},
		},
		// Loop over all the elements
		Body: &ast.BlockStmt{
			List: []ast.Stmt{
				hashPrecomputedInt(tag),
				startListHash,
				hashStringLiteral("listHash", "l"),
				&ast.RangeStmt{
					Key:   ast.NewIdent("_"),
					Value: ast.NewIdent("v"),
					Tok:   token.DEFINE,
					X:     selector2(localVar, field.Name),
					Body: &ast.BlockStmt{
						List: []ast.Stmt{
							&ast.ExprStmt{
								X: &ast.CallExpr{
									Fun: selector2("listHash", "Write"),
									Args: []ast.Expr{
										hashExpr,
									},
								},
							},
						},
					},
				},
				finishListHash,
			},
		},
	}
}

func hashMapValue(localVar string, field reflect.StructField, tag int64) ast.Stmt {
	startMapHash := &ast.AssignStmt{
		Lhs: []ast.Expr{
			ast.NewIdent("pairs"),
		},
		Tok: token.DEFINE,
		Rhs: []ast.Expr{
			&ast.CallExpr{
				Fun: ast.NewIdent("make "),
				Args: []ast.Expr{
					&ast.ArrayType{
						Elt: ast.NewIdent("KeyValuePair"),
					},
					&ast.BasicLit{
						Kind:  token.INT,
						Value: "0",
					},
					&ast.CallExpr{
						Fun: ast.NewIdent("len"),
						Args: []ast.Expr{
							selector2(localVar, field.Name),
						},
					},
				},
			},
		},
	}
	finishMapHash := &ast.ExprStmt{
		X: &ast.CallExpr{
			Fun: selector2("hash", "Write"),
			Args: []ast.Expr{
				&ast.CallExpr{
					Fun: ast.NewIdent("Hash_KeyValues"),
					Args: []ast.Expr{
						ast.NewIdent("pairs"),
					},
				},
			},
		},
	}

	keyHashExpr := &ast.CallExpr{
		Fun: ast.NewIdent(type2HashFunc(field.Type.Key())),
		Args: []ast.Expr{
			ast.NewIdent("k"),
		},
	}

	valHashExpr := &ast.CallExpr{
		Fun: ast.NewIdent(type2HashFunc(field.Type.Elem())),
		Args: []ast.Expr{
			ast.NewIdent("v"),
		},
	}

	pairConstructor := &ast.CompositeLit{
		Type: ast.NewIdent("KeyValuePair"),
		Elts: []ast.Expr{
			keyHashExpr,
			valHashExpr,
		},
	}

	// Check length against zero
	return &ast.IfStmt{
		Cond: &ast.BinaryExpr{
			X: &ast.CallExpr{
				Fun: ast.NewIdent("len"),
				Args: []ast.Expr{
					selector2(localVar, field.Name),
				},
			},
			Op: token.NEQ,
			Y: &ast.BasicLit{
				Kind:  token.INT,
				Value: "0",
			},
		},
		// Loop over all the elements and hash each
		Body: &ast.BlockStmt{
			List: []ast.Stmt{
				hashPrecomputedInt(tag),
				startMapHash,
				&ast.RangeStmt{
					Key:   ast.NewIdent("k"),
					Value: ast.NewIdent("v"),
					Tok:   token.DEFINE,
					X:     selector2(localVar, field.Name),
					Body: &ast.BlockStmt{
						List: []ast.Stmt{
							&ast.AssignStmt{
								Lhs: []ast.Expr{
									ast.NewIdent("pairs"),
								},
								Tok: token.ASSIGN,
								Rhs: []ast.Expr{
									&ast.CallExpr{
										Fun: ast.NewIdent("append"),
										Args: []ast.Expr{
											ast.NewIdent("pairs"),
											pairConstructor,
										},
									},
								},
							},
						},
					},
				},
				finishMapHash,
			},
		},
	}
}

type Field struct {
	IndexHash []byte
	Handler   ast.Stmt
}

func (f *GeneratedFile) AddHashFunc(messageType reflect.Type) {
	messageName := messageType.Name()

	props := proto.GetProperties(messageType)

	// Hash the fields of the structure in the order of
	// hash(props.Tag).  Any oneOf fields should be handled
	// inline so they get the right value.  But that means we need
	// to know what they all are!
	fieldHandlers := make([]Field, 0, messageType.NumField())
	for i := 0; i < messageType.NumField(); i++ {
		field := messageType.Field(i)
		if strings.HasPrefix(field.Name, "XXX_") {
			continue
		}
		// TODO: user ignored fields

		if field.Tag.Get("protobuf_oneof") != "" {
			// field.Type is an interface type, find all the inner struct types that
			// implement that interface and add their field handlers.
			for _, inner := range f.OneOfTypes {
				if inner.Type.Implements(field.Type) {
					fieldHandlers = append(fieldHandlers,
						Field{
							IndexHash: inner.Hash,
							Handler:   inner.Code(field.Name),
						},
					)
				}
			}

		} else {
			tag := int64(props.Prop[i].Tag)
			// ordinary member

			var s ast.Stmt
			switch field.Type.Kind() {
			case reflect.Map:
				s = hashMapValue("node", field, tag)
			case reflect.Slice:
				if field.Type.Elem().Kind() == reflect.Uint8 {
					// This is a []byte, so we only check if it's nil
					// (length-0 byte slices are not omitted.)
					s = hashAfterComparison("node", field.Name,
						ast.NewIdent("nil"),
						tag,
						"Hash_Bytes")
				} else {
					s = hashRepeatedValue("node", field, tag)
				}
			case reflect.String:
				s = hashGolangString("node", field.Name, tag)
			case reflect.Bool:
				s = hashGolangBool("node", field.Name, tag)
			case reflect.Int32:
				s = hashGolangInt32("node", field.Name, tag)
			case reflect.Int64:
				s = hashGolangInt64("node", field.Name, tag)
			case reflect.Uint32:
				s = hashGolangUint32("node", field.Name, tag)
			case reflect.Uint64:
				s = hashGolangUint64("node", field.Name, tag)
			case reflect.Float32:
				s = hashGolangFloat32("node", field.Name, tag)
			case reflect.Float64:
				s = hashGolangFloat64("node", field.Name, tag)
			case reflect.Ptr:
				// Assume everything is a struct
				s = hashAfterComparison("node", field.Name,
					ast.NewIdent("nil"),
					tag,
					fmt.Sprintf("Hash%v", field.Type.Elem().Name()))
			}
			if s != nil {
				fieldHandlers = append(fieldHandlers,
					Field{
						IndexHash: Hash_Int64(tag),
						Handler:   s,
					},
				)
			}
		}
	}

	sort.Slice(fieldHandlers, func(i, j int) bool {
		return bytes.Compare(fieldHandlers[i].IndexHash, fieldHandlers[j].IndexHash) < 0
	})

	statements := []ast.Stmt{
		startHash(),
		hashStringLiteral("hash", "d"),
	}
	for _, fh := range fieldHandlers {
		statements = append(statements, fh.Handler)
	}
	statements = append(statements, returnHash())

	body := &ast.BlockStmt{
		List: statements,
	}

	pkg := "pb"
	if strings.HasSuffix(messageType.PkgPath(), "wrappers") {
		pkg = "wrappers"
	}

	decl := &ast.FuncDecl{
		Name: ast.NewIdent(fmt.Sprintf("Hash%v", messageName)),
		Recv: nil,
		Type: &ast.FuncType{
			Params: &ast.FieldList{
				List: []*ast.Field{
					nodeArg(pkg, messageName),
				},
			},
			Results: &ast.FieldList{
				List: []*ast.Field{
					{
						Names: nil,
						Type:  byteArray(),
					},
				},
			},
		},
		Body: body,
	}
	f.AddFunction(decl)
}

// copy of the function in ir_hash, to avoid depdendency
func Hash_Int64(i int64) []byte {
	hf := xxhash.New64()
	hf.Write([]byte(`u`))
	hf.Write([]byte(fmt.Sprintf("%d", i)))
	return hf.Sum(nil)
}

package tagger

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"reflect"
	"strconv"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/pkg/errors"
)

// TagStruct add Go tags to the struct at the FilePosition, determined by the offset provided.
func TagStruct(rawPosition, tagName, prefix, suffix string) (*FilePosition, error) {
	file, err := NewFilePosition(rawPosition)
	if err != nil {
		return nil, err
	}
	return file, file.TagStruct(tagName, prefix, suffix)
}

// FilePosition specifies a filename and offset of a file.
type FilePosition struct {
	Name   string
	Offset int

	FileSet *token.FileSet
	Root    ast.Node
}

// NewFilePosition correctly instantiates a FilePosition from an offset.
// Raw position is in the form of "<go file name>:#<line number>", for example: "file.go:#123".
func NewFilePosition(rawPosition string) (*FilePosition, error) {
	parts := strings.Split(rawPosition, ":#")
	if len(parts) != 2 {
		return nil, errors.Errorf("%q: invalid file position", rawPosition)
	}
	filename := parts[0]

	for _, r := range parts[1] {
		if !unicode.IsDigit(r) {
			return nil, errors.Errorf("%q: non-numeric line number", rawPosition)
		}
	}

	offset, err := strconv.Atoi(parts[1])
	if err != nil {
		return nil, errors.Wrapf(err, "failed to parse line number %q", parts[1])
	}

	if _, err := os.Stat(filename); err != nil {
		return nil, err
	}

	filePos := FilePosition{
		Name:   filename,
		Offset: offset,
	}

	filePos.FileSet = token.NewFileSet()

	filePos.Root, err = parser.ParseFile(filePos.FileSet, filename, nil, 0)
	if err != nil {
		return nil, err
	}

	return &filePos, nil
}

// TagStruct add Go tags to the struct at the FilePosition.
func (f FilePosition) TagStruct(tagName, prefix, suffix string) error {
	node, err := getStruct(f.FileSet, f.Root, f.Offset)
	if err != nil {
		return err
	}

	for _, field := range node.Fields.List {
		tagField(field, tagName, prefix, suffix)
	}

	return nil
}

func tagField(field *ast.Field, tagName, prefix, suffix string) {
	// Don't tag line with 0 or 2 or more fields.
	if len(field.Names) < 1 || len(field.Names) > 1 {
		return
	}

	// Don't tag if the field is not exported.
	r, _ := utf8.DecodeRuneInString(field.Names[0].Name)
	if !unicode.IsUpper(r) {
		return
	}

	// Construct Tag
	tagContents := prefix + field.Names[0].String() + suffix
	fieldTag := fmt.Sprintf("`%s:\"%s\"`", tagName, tagContents)

	if field.Tag != nil {
		tag := reflect.StructTag(field.Tag.Value[1 : len(field.Tag.Value)-1])
		if _, ok := tag.Lookup(tagName); !ok {
			if strings.TrimSpace(string(tag)) == "" {
				field.Tag.Value = fieldTag
			} else {
				field.Tag.Value = field.Tag.Value[:len(field.Tag.Value)-1] + " " + fieldTag[1:]
			}
		}
	} else {
		field.Tag = &ast.BasicLit{
			Kind:     token.STRING,
			Value:    fieldTag,
			ValuePos: token.Pos(field.End()),
		}
	}
}

func getStruct(fset *token.FileSet, file ast.Node, offset int) (*ast.StructType, error) {
	var node *ast.StructType
	// Inspect the AST and print all identifiers and literals.
	ast.Inspect(file, func(n ast.Node) bool {
		x, ok := n.(*ast.StructType)
		if !ok {
			return true
		}

		structBegin := fset.Position(x.Pos()).Offset
		structEnd := fset.Position(x.End()).Offset
		if structBegin > offset || structEnd < offset {
			return true
		}

		node = x
		return false
	})

	if node == nil {
		return nil, fmt.Errorf("no struct found at offset %d", offset)
	}

	return node, nil
}

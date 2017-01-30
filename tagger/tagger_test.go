package tagger_test

import (
	"bytes"
	"go/format"
	"os"
	"path/filepath"
	"testing"

	"github.com/brycereitano/gotag/tagger"
)

var (
	testDirectoryPath = "testdata"
	testFilePath      = filepath.Join(testDirectoryPath, "foo.go")
)

func instantiateTestData(t *testing.T, body string) func() {
	cleanup := func() { os.RemoveAll("testdata") }

	err := os.Mkdir("testdata", 0700)
	if err != nil {
		t.Fatal(err)
	}
	file, err := os.Create(testFilePath)
	if err != nil {
		cleanup()
		t.Fatal(err)
	}
	_, err = file.WriteString(body)
	file.Close()

	return cleanup
}

func TestFilePosition_New(t *testing.T) {
	t.Run("ValidGoFile", func(t *testing.T) {
		cleanup := instantiateTestData(t, `package main

	type Foo struct {
		Foo string `+"`json:\"-\"`"+`
		Bar map[int]interface{}
		baz int

		Nax struct {
			Hello string
		}
	}
`)
		defer cleanup()

		position, err := tagger.NewFilePosition(testFilePath + ":#30")
		if err != nil {
			t.Fatalf("Unexpected error, %q", err)
		}

		if position.Name != testFilePath {
			t.Errorf("expected position.Name to equal %q, but got %q", testFilePath, position.Name)
		}

		if position.Offset != 30 {
			t.Errorf("expected position.Offset to equal %d, but got %d", 30, position.Offset)
		}
	})

	testCases := []struct {
		name        string
		rawPosition string
		expectedErr string
	}{
		{"NoOffset", "file.go", `"file.go": invalid file position`},
		{"InvalidLineNumber", "file.go:#a", `"file.go:#a": non-numeric line number`},
		{"MissingFile", "file.go:#30", `stat file.go: no such file or directory`},
		{"InvalidGoFile", testFilePath + ":#30", `testdata/foo.go:1:1: expected 'package', found 'EOF'`},
	}
	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			cleanup := instantiateTestData(t, "")
			defer cleanup()
			_, err := tagger.NewFilePosition(testCase.rawPosition)
			if err == nil {
				t.Error("expected to return an error, but got nil")
			}

			if testCase.expectedErr != err.Error() {
				t.Errorf("expected error to equal %q, but got %q", testCase.expectedErr, err)
			}
		})
	}
}

func TestFilePosition_TagStruct(t *testing.T) {
	testCases := []struct {
		name           string
		tag            string
		prefix         string
		suffix         string
		input          string
		expectedOutput string
	}{
		{
			name: "JSONTags",
			tag:  "json", prefix: "", suffix: "",
			input: `package main

type Foo struct {
	Foo string ` + "`json:\"-\"`" + `
	Bar map[int]interface{}
	baz int

	Nax struct {
		Hello string
	}
}`,
			expectedOutput: `package main

type Foo struct {
	Foo string              ` + "`json:\"-\"`" + `
	Bar map[int]interface{} ` + "`json:\"Bar\"`" + `
	baz int

	Nax struct {
		Hello string
	} ` + "`json:\"Nax\"`" + `
}
`,
		},
		{
			name: "XMLTags",
			tag:  "xml", prefix: "", suffix: "",
			input: `package main

type Foo struct {
	Foo string ` + "`json:\"-\"`" + `
	Bar map[int]interface{}
	baz int

	Nax struct {
		Hello string
	}
}`,
			expectedOutput: `package main

type Foo struct {
	Foo string              ` + "`json:\"-\" xml:\"Foo\"`" + `
	Bar map[int]interface{} ` + "`xml:\"Bar\"`" + `
	baz int

	Nax struct {
		Hello string
	} ` + "`xml:\"Nax\"`" + `
}
`,
		},
		{
			name: "PrefixSuffix",
			tag:  "json", prefix: "JSON", suffix: ",omitempty",
			input: `package main

type Foo struct {
	Foo string ` + "`json:\"-\"`" + `
	Bar map[int]interface{}
	baz int

	Nax struct {
		Hello string
	}
}`,
			expectedOutput: `package main

type Foo struct {
	Foo string              ` + "`json:\"-\"`" + `
	Bar map[int]interface{} ` + "`json:\"JSONBar,omitempty\"`" + `
	baz int

	Nax struct {
		Hello string
	} ` + "`json:\"JSONNax,omitempty\"`" + `
}
`,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			cleanup := instantiateTestData(t, testCase.input)
			defer cleanup()

			position, err := tagger.NewFilePosition(testFilePath + ":#30")
			if err != nil {
				t.Fatalf("Unexpected error, %q", err)
			}

			err = position.TagStruct(testCase.tag, testCase.prefix, testCase.suffix)
			if err != nil {
				t.Fatalf("Unexpected error, %q", err)
			}

			var buffer bytes.Buffer
			err = format.Node(&buffer, position.FileSet, position.Root)
			if err != nil {
				t.Fatalf("Unexpected error, %q", err)
			}

			if testCase.expectedOutput != buffer.String() {
				t.Fatalf("Expected to produce \n%q\n but got \n%q\n", testCase.expectedOutput, buffer.String())
			}
		})
	}
}

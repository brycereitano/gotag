package main

import (
	"flag"
	"fmt"
	"go/format"
	"log"
	"os"

	"github.com/brycereitano/gotag/tagger"
)

var (
	offsetFlag = flag.String("offset", "", "file and byte offset of identifier to be tagged, e.g. 'file.go:#123'. For use by editors.")
	tagFlag    = flag.String("tag", "", "the tag to add to each field of a struct")
	prefixFlag = flag.String("prefix", "", "a string to apply before the tag value")
	suffixFlag = flag.String("suffix", "", "a string to apply after the tag value")
	helpFlag   = flag.Bool("help", false, "show usage message")
)

func main() {
	flag.Parse()
	if len(flag.Args()) > 0 {
		fmt.Fprintln(os.Stderr, "surplus arguments")
		return
	}

	if *helpFlag || (*offsetFlag == "" && *tagFlag == "") {
		fmt.Fprintln(os.Stderr, Usage)
		return
	}

	file, err := ParseOffsetFlag(*offsetFlag)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return
	}

	pos, err := tagger.NewFilePosition(file.Name, file.Offset)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return
	}

	err = pos.TagStruct(*tagFlag, *prefixFlag, *suffixFlag)
	if err != nil {
		log.Fatal(err)
	}

	err = format.Node(os.Stdout, pos.FileSet, pos.Root)
	if err != nil {
		log.Fatal(err)
	}
}

const Usage = `goanno: simple golang struct annotator.
Usage:
  goanno -offset <file>:#<byte-offset>) -tag <name>
You must specify the object (defined struct) to add tags using the -offset.
Flags:
-offset    specifies the filename and byte offset of an identifier to rename.
           This form is intended for use by text editors.
-tag       what tag to add to the struct fields.
-d         display diffs instead of rewriting files
Examples:
$ gorename -offset file.go:#123 -tag json
  TODO:
`

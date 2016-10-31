# Go Tag


## Example

Input: sample.go
```golang
package main

type Foo struct {
	Foo string `json:"-"`
	Bar map[int]interface{}
	baz int

	Nax struct {
		Hello string
	}
}
```

Execution:
```
$ gotag -offset sample.go:#50 -tag json -suffix ",omitempty"
package main

type Foo struct {
    Foo string              `json:"-"`
    Bar map[int]interface{} `json:"Bar,omitempty"`
    baz int

    Nax struct {
        Hello string
    } `json:"Nax,omitempty"`
}
```

# Attribution

Based off of the `gorename` tool.

```
Copyright 2014 The Go Authors. All rights reserved.
Use of this source code is governed by a BSD-style
license that can be found in the LICENSE-golang file.

Modified source of https://github.com/golang/tools/blob/master/refactor/rename
```

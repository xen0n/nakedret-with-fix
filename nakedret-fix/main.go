package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"strconv"
	"strings"
)

type fix struct {
	lineno  uint64
	replace string
}

type fixWithFilename struct {
	filename string
	fix      fix
}

func main() {
	// stdin -> []byte
	lines, err := ioutil.ReadAll(os.Stdin)
	if err != nil {
		panic(err)
	}

	// []byte -> string -> []string
	s := string(lines)
	l := strings.Split(s, "\n")

	// []string -> []fixWithFilename
	ungroupedFixes := make([]fixWithFilename, 0, len(l))
	for _, line := range l {
		if len(line) == 0 {
			continue
		}

		elements := strings.Split(line, ":")
		if len(elements) != 3 {
			panic("malformed input")
		}

		lineno, err := strconv.ParseUint(elements[1], 10, 64)
		if err != nil {
			panic("malformed lineno")
		}

		ungroupedFixes = append(ungroupedFixes, fixWithFilename{
			filename: elements[0],
			fix: fix{
				lineno:  lineno,
				replace: elements[2],
			},
		})
	}

	// []fixWithFilename -> map[filename][]fix
	fixesByFile := make(map[string][]fix)
	for _, f := range ungroupedFixes {
		filename := f.filename
		// make use of zero value semantics
		fixesByFile[filename] = append(fixesByFile[filename], f.fix)
	}

	// fix files
	for filename, fixes := range fixesByFile {
		fmt.Println(filename, "->", len(fixes), "fixes")

		err := fixFile(filename, fixes)
		if err != nil {
			panic(err)
		}
	}
}

func fixFile(filename string, fixes []fix) error {
	// read
	var content string
	{
		f, err := os.Open(filename)
		if err != nil {
			panic(err)
		}

		contentBytes, err := ioutil.ReadAll(f)
		if err != nil {
			panic(err)
		}

		content = string(contentBytes)

		_ = f.Close()
	}

	lines := strings.Split(content, "\n")

	// apply fix
	for _, f := range fixes {
		// lineno is 1-based
		lineIdx := f.lineno - 1
		before := lines[lineIdx]
		after := strings.Replace(before, "return", f.replace, 1)
		lines[lineIdx] = after
	}

	// concatenate back into string
	contentAfter := strings.Join(lines, "\n")

	// write
	err := ioutil.WriteFile(filename, []byte(contentAfter), 0644)
	if err != nil {
		panic(err)
	}

	return nil
}

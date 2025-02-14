// Copyright 2022 Baidu Inc. All rights Reserved.
// 2022/4/4 9:20 PM, by Wang Xugang(wangxugang@baidu.com), create
//
package main

import (
	"flag"
	"fmt"
	"io/fs"
	"io/ioutil"
	"os"
	"sort"
	"strings"
)

var (
	gofiles []string
)

func main() {
	if len(os.Args) != 2 {
		fmt.Println("usage: check_import your-go-file-path")
		os.Exit(1)
	}
	fname := os.Args[1]
	errPrefix := fname + ": "
	flag.Parse()

	by, err := ioutil.ReadFile(fname)
	if len(by) == 0 {
		fmt.Printf("read file %s got empty, error=%v\n", fname, err)
		os.Exit(1)
	}

	lines := strings.Split(string(by), "\n")
	import_start := -1
	import_end := -1
	var importSystem []string
	var importGithub []string
	var importIcode []string
	for i, l := range lines {
		if import_start < 0 {
			if strings.HasPrefix(l, "import (") {
				import_start = i
			}
			continue
		}
		if strings.HasSuffix(l, ")") {
			import_end = i
			break
		}

		t := getImportType(l)
		switch t {
		case ImportTypeSystem:
			importSystem = append(importSystem, l)
		case ImportTypeGithub:
			importGithub = append(importGithub, l)
		case ImportTypeIcode:
			importIcode = append(importIcode, l)
		}
	}
	if !(import_start >= 0 && import_end >= 0 || import_start < 0 && import_end < 0) {
		fmt.Printf(errPrefix + "invalid import syntax\n")
		os.Exit(1)
	}
	// TODO: 对import "xxxx" 这种方式进行处理
	if import_start < 0 && import_end < 0 {
		fmt.Printf("%s: not import( xxx ) style, refine later\n", fname)
		return
	}

	for _, ss := range [][]string{importSystem, importGithub, importIcode} {
		sort.Slice(ss, func(i, j int) bool {
			c1 := strings.Split(ss[i], "\"")[1]
			c2 := strings.Split(ss[j], "\"")[1]
			return strings.Compare(c1, c2) < 0
		})
	}

	ls := strings.Join(importSystem, "\n")
	lg := strings.Join(importGithub, "\n")
	li := strings.Join(importIcode, "\n")
	var expectLines []string
	if len(ls) > 0 {
		expectLines = append(expectLines, ls)
	}
	if len(lg) > 0 {
		expectLines = append(expectLines, lg)
	}
	if len(li) > 0 {
		expectLines = append(expectLines, li)
	}
	expectedLine := strings.Join(expectLines, "\n\n")
	if ns := strings.Join(lines[import_start+1:import_end], "\n"); ns != expectedLine {
		pre := strings.Join(lines[:import_start+1], "\n") + "\n"
		middle := expectedLine + "\n"
		end := strings.Join(lines[import_end:], "\n")
		final := pre + middle + end
		err := ioutil.WriteFile(fname, []byte(final), 0666)
		if err != nil {
			fmt.Printf(errPrefix+"error: writefile got: %v\n", err)
			os.Exit(1)
		}
	}
}

func walk(path string, info fs.FileInfo, err error) error {
	if err != nil {
		return err
	}
	if info.Mode().IsRegular() && strings.HasSuffix(info.Name(), ".go") {
		gofiles = append(gofiles, path)
	}
	return nil
}

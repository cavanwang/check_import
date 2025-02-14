// Copyright 2022 Baidu Inc. All rights Reserved.
// 2022/4/4 9:20 PM, by Wang Xugang(wangxugang@baidu.com), create
package main

import (
	"fmt"
	"strings"
)

const (
	ImportTypeSystem  ImportType = "golang"
	ImportTypeGithub  ImportType = "github"
	ImportTypeIcode   ImportType = "gitlab"
	ImportTypeUnknown ImportType = "unknown"

	OurImportKeyword = "gitlab.bee.to"
)

type ImportType = string

func checkFileLines(lines []string) (errMsg string) {
	var lastCategory string
	slen := len(lines)
	for i := 0; i < slen; i++ {
		if lastCategory == "" {
			if strings.Index(lines[i], "import (") == 0 {
				lastCategory = "import"
			} else if strings.Index(lines[i], `import "`) == 0 { // 当行import 直接跳过不处理
				return ""
			}
			continue
		}

		endIndex, catetory := getNextCategory(lines[i:])
		if endIndex < 0 {
			return fmt.Sprintf("from line %d: invalid import order", i+1)
		}

		if catetory == "system" && lastCategory == "import" ||
			catetory == "github" && (lastCategory == "import" || lastCategory == "system") ||
			catetory == ImportTypeIcode && (lastCategory == "import" || lastCategory == "system" || lastCategory == "github") {
			lastCategory = catetory
		} else {
			return fmt.Sprintf("from line %d: not expected import type %s while last type is %s", i+1, catetory, lastCategory)
		}

		// 这里是唯一正常返回的地方
		if lines[i+endIndex+1] == ")" {
			return
		}
		// 下一次跳过单个空行
		i += endIndex + 1
	}
	if lastCategory != "" {
		return "invalid import order"
	}
	return ""
}

// 返回是system, github, icode三者中的一个，其他表示错误。endLineIndex要么是单个空行，要么是)结束行。
func getNextCategory(lines []string) (endLineIndex int, category string) {
	for i := range lines {
		l := strings.TrimSpace(lines[i])
		c := ""
		if strings.Index(l, OurImportKeyword) >= 0 {
			c = ImportTypeIcode
		} else if strings.Index(l, ".") >= 1 { // github.com gopkg.in
			c = "github"
			fmt.Printf("line '%s' is github\n", l)
		} else if l == ")" {
			c = "end"
		} else if l != "" {
			c = "system"
		}

		// 如果首行，则必须为正常导入package
		if i == 0 {
			if c == "system" || c == "github" || c == ImportTypeIcode {
				endLineIndex = i
				category = c
				continue
			}
			return -1, ""
		}

		// 后续行必须与之前的行连续为相同类型
		switch c {
		case "system":
			if category == "system" {
				endLineIndex = i
			} else {
				return -1, ""
			}
		case "github":
			if category == "github" {
				endLineIndex = i
			} else {
				return -1, ""
			}
		case ImportTypeIcode:
			if category == ImportTypeIcode {
				endLineIndex = i
			} else {
				return -1, ""
			}
		default:
			return
		}
	}

	return -1, ""
}

func getImportType(line string) string {
	if strings.Index(line, OurImportKeyword) >= 0 {
		return ImportTypeIcode
	}
	dot := strings.Index(line, ".")
	if strings.Index(line, "github.com") >= 0 || strings.Index(line, "gopkg.in") >= 0 ||
		dot > 0 && dot < strings.Index(line, "/") {
		return ImportTypeGithub
	} else if strings.TrimSpace(line) != "" {
		return ImportTypeSystem
	}
	return ImportTypeUnknown
}

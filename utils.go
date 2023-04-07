package main

import (
	"strings"
)

func pkgName(pkgPath string) string {
	index := strings.LastIndex(pkgPath, "/")
	if index == -1 {
		return pkgPath // universal package
	}
	return pkgPath[index+1:]
}

package utils

import "strings"

func SplitNameExt(fullName string) (string, string) {
	idx := strings.LastIndexByte(fullName, '.')
	if idx == -1 {
		return fullName, ""
	}
	return fullName[:idx], fullName[idx+1:]
}

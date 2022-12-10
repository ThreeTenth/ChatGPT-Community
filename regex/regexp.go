package regex

import "regexp"

// MultBlankLines 是多个空行的正则表达式
var MultBlankLines = regexp.MustCompile(`\n+`)

package extras

import "github.com/yuin/goldmark/util"

func hasSpace(line []byte) bool {
	marker := line[0]
	for i := 1; i < len(line); i++ {
		c := line[i]
		if c == marker {
			break
		}
		if util.IsSpace(c) {
			return true
		}
	}
	return false
}

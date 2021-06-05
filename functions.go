package shuttle

import (
	"strconv"
	"strings"
)

func ReadPathElement(rawPath, upstreamElement string) string {
	var nextDocument bool
	var document string
	for len(rawPath) > 0 {
		slashIndex := strings.Index(rawPath, "/")
		if slashIndex >= 0 {
			document, rawPath = rawPath[:slashIndex], rawPath[slashIndex+1:]
		} else {
			document = rawPath
			rawPath = ""
		}

		if nextDocument {
			return document
		}

		if document == upstreamElement {
			nextDocument = true
		}
	}

	return ""
}
func ReadNumericPathElement(rawPath, upstreamElement string) uint64 {
	raw := ReadPathElement(rawPath, upstreamElement)
	value, _ := strconv.ParseUint(raw, 10, 64)
	return value
}

package shared

import "strings"

const ProtocolFile = "file://"

func MaybeFile(reference string) bool {
	return strings.HasPrefix(reference, "./") || strings.HasPrefix(reference, ProtocolFile)
}

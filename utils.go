package myblockchain

import (
	"crypto/sha256"
	"fmt"

	"github.com/satori/go.uuid"
)

func ComputeHashSha256(bytes []byte) string {
	return fmt.Sprintf("%x", sha256.Sum256(bytes))
}

func UUID() string {
	out := uuid.NewV4()
	return fmt.Sprintf("%s", out)
}

func NewStringSet() StringSet {
	return StringSet{make(map[string]bool)}
}

type StringSet struct {
	set map[string]bool
}

func (set *StringSet) Add(str string) bool {
	_, found := set.set[str]
	set.set[str] = true
	return !found
}

func (set *StringSet) Keys() []string {
	var keys []string
	for k, _ := range set.set {
		keys = append(keys, k)
	}
	return keys
}

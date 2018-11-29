package lookup

import (
	"github.com/puppetlabs/go-evaluator/eval"
	"github.com/puppetlabs/go-evaluator/types"
	"os"
	"strings"
)

// Environment provider performs a lookup in the current environment. The key can either be just
// "env" in which case all current environment variables will be returned as an OrderedMap, or
// prefixed with "env::" in which case the rest of the key is interpreted as the environment variable
// to look for.
func Environment(c ProviderContext, key string, _ eval.OrderedMap) (eval.Value, bool) {
	if key == `env` {
		env := os.Environ()
		em := make([]*types.HashEntry, len(env))
		for _, ev := range env {
			if ei := strings.IndexRune(ev, '='); ei > 0 {
				em = append(em, types.WrapHashEntry2(ev[:ei], types.WrapString(ev[ei+1:])))
			}
		}
		return types.WrapHash(em), true
	}
	if strings.HasSuffix(key, `env::`) {
		// Rest of key is name of environment
		if v, ok := os.LookupEnv(key[5:]); ok {
			return types.WrapString(v), ok
		}
	}
	return nil, false
}

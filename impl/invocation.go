package impl

import (
	"github.com/puppetlabs/go-evaluator/eval"
	"github.com/puppetlabs/go-evaluator/utils"
	"github.com/puppetlabs/go-hiera/config"
	"github.com/puppetlabs/go-hiera/lookup"
	"github.com/puppetlabs/go-issues/issue"
)

const HieraCacheKey = `Hiera::Cache`
const HieraTopProviderKey = `Hiera::TopProvider`
const HieraTopProviderCacheKey = `Hiera::TopProvider::Cache`
const HieraConfigsKey = `Hiera::Config::`

type invocation struct {
	eval.Context
	nameStack []string
}

// InitContext initializes the given context with the Hiera cache. The context initialized
// with this method determines the life-cycle of that cache.
func InitContext(c eval.Context, topProvider lookup.LookupKey) {
	c.Set(HieraCacheKey, NewConcurrentMap(37))
	c.Set(HieraTopProviderKey, topProvider)
	c.Set(HieraTopProviderCacheKey, make(map[string]eval.Value, 23))
}

func NewInvocation(c eval.Context) lookup.Invocation {
	return &invocation{Context: c, nameStack: []string{}}
}

func (ic *invocation) topProvider() lookup.LookupKey {
	if v, ok := ic.Get(HieraTopProviderKey); ok {
		var tp lookup.LookupKey
		if tp, ok = v.(lookup.LookupKey); ok {
			return tp
		}
	}
	panic(eval.Error(HIERA_NOT_INITIALIZED, issue.NO_ARGS))
}

func (ic *invocation) topProviderCache() map[string]eval.Value {
	if v, ok := ic.Get(HieraTopProviderCacheKey); ok {
		var tc map[string]eval.Value
		if tc, ok = v.(map[string]eval.Value); ok {
			return tc
		}
	}
	panic(eval.Error(HIERA_NOT_INITIALIZED, issue.NO_ARGS))
}

func (ic *invocation) sharedCache() *ConcurrentMap {
	if v, ok := ic.Get(HieraCacheKey); ok {
		var sh *ConcurrentMap
		if sh, ok = v.(*ConcurrentMap); ok {
			return sh
		}
	}
	panic(eval.Error(HIERA_NOT_INITIALIZED, issue.NO_ARGS))
}

func (ic *invocation) Config(configPath string) config.ResolvedConfig {
	val, _ := ic.sharedCache().EnsureSet(HieraConfigsKey + configPath, func() (interface{}, bool) {
		return NewConfig(ic, configPath), true
	})
	return val.(config.ResolvedConfig)
}

func (ic *invocation) lookupViaCache(key lookup.Key, options eval.OrderedMap) (eval.Value, bool) {
	rootKey := key.Root()

	val, ok := ic.sharedCache().EnsureSet(rootKey, func() (interface{}, bool) {
		if v, ok := ic.topProvider()(newContext(ic, ic.topProviderCache()), rootKey, options); ok {
			return Interpolate(ic, v, true), true
		}
		return nil, false
	})
	if ok {
		return key.Dig(val.(eval.Value))
	}
	return nil, false
}

func (ic *invocation) Check(key lookup.Key, actor lookup.Producer) (eval.Value, bool) {
	if utils.ContainsString(ic.nameStack, key.String()) {
		panic(eval.Error(HIERA_ENDLESS_RECURSION, issue.H{`name_stack`: ic.nameStack}))
	}
	ic.nameStack = append(ic.nameStack, key.String())
	defer func() {
		ic.nameStack = ic.nameStack[:len(ic.nameStack)-1]
	}()
	return actor()
}

func (ic *invocation) WithDataProvider(dh lookup.DataProvider, actor lookup.Producer) (eval.Value, bool) {
	return actor()
}

func (ic *invocation) WithLocation(loc lookup.Location, actor lookup.Producer) (eval.Value, bool) {
	return actor()
}

func (ic *invocation) ReportLocationNotFound() {
}

func (ic *invocation) ReportFound(key string, value eval.Value) {
}

func (ic *invocation) ReportNotFound(key string) {
}

var notFoundSingleton = &lookup.NotFound{}

func (ic *invocation) NotFound() {
	panic(notFoundSingleton)
}

func (ic *invocation) Explain(messageProducer func() string) {
	// TODO: Add explanation support
}

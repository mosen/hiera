package impl_test

import (
	"context"
	"fmt"
	"github.com/puppetlabs/go-evaluator/eval"
	evalimpl "github.com/puppetlabs/go-evaluator/impl"
	"github.com/puppetlabs/go-evaluator/types"
	"github.com/puppetlabs/go-hiera/impl"
	"github.com/puppetlabs/go-hiera/lookup"
	"github.com/puppetlabs/go-issues/issue"

	// Ensure initialization
	_ "github.com/puppetlabs/go-evaluator/pcore"
	_ "github.com/puppetlabs/go-hiera/functions"
)

var sampleData = map[string]eval.Value{
  `first`: types.WrapString(`value of first`),
  `array`: eval.Wrap(nil, []string{`one`, `two`, `three`}),
	`hash`: eval.Wrap(nil, map[string]interface{}{`int`: 1, `string`: `one`, `array`: []string{`two`, `%{hiera('first')}`}}),
  `second`: types.WrapString(`includes '%{lookup('first')}'`),
	`ipAlias`: types.WrapString(`%{alias('array')}`),
	`ipBadAlias`: types.WrapString(`x %{alias('array')}`),
	`ipScope`: types.WrapString(`hello %{world}`),
	`ipScope2`: types.WrapString(`hello %{scope('world')}`),
	`ipLiteral`: types.WrapString(`some %{literal('literal')} text`),
	`ipBad`: types.WrapString(`hello %{bad('world')}`),
	`empty1`: types.WrapString(`start%{}end`),
	`empty2`: types.WrapString(`start%{''}end`),
	`empty3`: types.WrapString(`start%{""}end`),
	`empty4`: types.WrapString(`start%{::}end`),
	`empty5`: types.WrapString(`start%{'::'}end`),
	`empty6`: types.WrapString(`start%{"::"}end`),
}

func tp(ic lookup.ProviderContext, key string, _ eval.OrderedMap) (eval.Value, bool) {
	v, ok := sampleData[key]
	return v, ok
}

func ExampleLookup_first() {
	lookup.DoWithParent(context.Background(), tp, func(c eval.Context) {
		fmt.Println(lookup.Lookup(impl.NewInvocation(c), `first`, nil, eval.EMPTY_MAP))
	})
	// Output: value of first
}

func ExampleLookup_dottedInt() {
	lookup.DoWithParent(context.Background(), tp, func(c eval.Context) {
		fmt.Println(lookup.Lookup(impl.NewInvocation(c), `array.1`, nil, eval.EMPTY_MAP))
	})
	// Output: two
}

func ExampleLookup_dottedMix() {
	lookup.DoWithParent(context.Background(), tp, func(c eval.Context) {
		fmt.Println(lookup.Lookup(impl.NewInvocation(c), `hash.array.1`, nil, eval.EMPTY_MAP))
	})
	// Output: value of first
}

func ExampleLookup_interpolate() {
	lookup.DoWithParent(context.Background(), tp, func(c eval.Context) {
		fmt.Println(lookup.Lookup(impl.NewInvocation(c), `second`, nil, eval.EMPTY_MAP))
	})
	// Output: includes 'value of first'
}

func ExampleLookup_interpolateScope() {
	eval.Puppet.DoWithParent(context.Background(), func(c eval.Context) {
		c.DoWithScope(evalimpl.NewScope2(types.WrapStringToInterfaceMap(c, issue.H{
			`world`: `cruel world`,
		}), false), func() {
			lookup.DoWithParent(c, tp, func(c eval.Context) {
				fmt.Println(lookup.Lookup(impl.NewInvocation(c), `ipScope`, nil, eval.EMPTY_MAP))
				fmt.Println(lookup.Lookup(impl.NewInvocation(c), `ipScope2`, nil, eval.EMPTY_MAP))
			})
		})
	})
	// Output:
	// hello cruel world
	// hello cruel world
}

func ExampleLookup_interpolateEmpty() {
	lookup.DoWithParent(context.Background(), tp, func(c eval.Context) {
		fmt.Println(lookup.Lookup(impl.NewInvocation(c), `empty1`, nil, eval.EMPTY_MAP))
		fmt.Println(lookup.Lookup(impl.NewInvocation(c), `empty2`, nil, eval.EMPTY_MAP))
		fmt.Println(lookup.Lookup(impl.NewInvocation(c), `empty3`, nil, eval.EMPTY_MAP))
		fmt.Println(lookup.Lookup(impl.NewInvocation(c), `empty4`, nil, eval.EMPTY_MAP))
		fmt.Println(lookup.Lookup(impl.NewInvocation(c), `empty5`, nil, eval.EMPTY_MAP))
		fmt.Println(lookup.Lookup(impl.NewInvocation(c), `empty6`, nil, eval.EMPTY_MAP))
	})
	// Output:
	// startend
	// startend
	// startend
	// startend
	// startend
	// startend
}

func ExampleLookup_interpolateLiteral() {
	lookup.DoWithParent(context.Background(), tp, func(c eval.Context) {
		fmt.Println(lookup.Lookup(impl.NewInvocation(c), `ipLiteral`, nil, eval.EMPTY_MAP))
	})
	// Output: some literal text
}

func ExampleLookup_interpolateAlias() {
	lookup.DoWithParent(context.Background(), tp, func(c eval.Context) {
		v := lookup.Lookup(impl.NewInvocation(c), `ipAlias`, nil, eval.EMPTY_MAP)
		fmt.Printf(`%s %s`, eval.GenericValueType(v), v)
	})
	// Output: Array[Enum] ['one', 'two', 'three']
}

func ExampleLookup_interpolateBadAlias() {
	fmt.Println(lookup.TryWithParent(context.Background(), tp, func(c eval.Context) error {
		lookup.Lookup(impl.NewInvocation(c), `ipBadAlias`, nil, eval.EMPTY_MAP)
		return nil
	}))
	// Output: 'alias' interpolation is only permitted if the expression is equal to the entire string
}

func ExampleLookup_interpolateBadFunction() {
	fmt.Println(lookup.TryWithParent(context.Background(), tp, func(c eval.Context) error {
		lookup.Lookup(impl.NewInvocation(c), `ipBad`, nil, eval.EMPTY_MAP)
		return nil
	}))
	// Output: Unknown interpolation method 'bad'
}

func ExampleLookup_notFoundWithoutDefault() {
	fmt.Println(lookup.TryWithParent(context.Background(), tp, func(c eval.Context) error {
		lookup.Lookup(impl.NewInvocation(c), `nonexistent`, nil, eval.EMPTY_MAP)
		return nil
	}))
	// Output: lookup() did not find a value for the name 'nonexistent'
}

func ExampleLookup_notFoundDflt() {
	lookup.DoWithParent(context.Background(), tp, func(c eval.Context) {
		fmt.Println(lookup.Lookup(impl.NewInvocation(c), `nonexistent`, types.WrapString(`default value`), eval.EMPTY_MAP))
	})
	// Output: default value
}

func ExampleLookup_notFoundDottedIdx() {
	lookup.DoWithParent(context.Background(), tp, func(c eval.Context) {
		fmt.Println(lookup.Lookup(impl.NewInvocation(c), `array.3`, types.WrapString(`default value`), eval.EMPTY_MAP))
	})
	// Output: default value
}

func ExampleLookup_notFoundDottedMix() {
	lookup.DoWithParent(context.Background(), tp, func(c eval.Context) {
		fmt.Println(lookup.Lookup(impl.NewInvocation(c), `hash.float`, types.WrapString(`default value`), eval.EMPTY_MAP))
	})
	// Output: default value
}

func ExampleLookup_badStringDig() {
	fmt.Println(lookup.TryWithParent(context.Background(), tp, func(c eval.Context) error {
		lookup.Lookup(impl.NewInvocation(c), `hash.int.v`, nil, eval.EMPTY_MAP)
		return nil
	}))
	// Output: lookup() Got Integer when a hash-like object was expected to access value using 'v' from key 'hash.int.v'
}

func ExampleLookup_badIntDig() {
	fmt.Println(lookup.TryWithParent(context.Background(), tp, func(c eval.Context) error {
		lookup.Lookup(impl.NewInvocation(c), `hash.3`, nil, eval.EMPTY_MAP)
		return nil
	}))
	// Output: lookup() Got Hash[Enum, Data] when a hash-like object was expected to access value using '3' from key 'hash.3'
}

func ExampleLookup2_findFirst() {
	lookup.DoWithParent(context.Background(), tp, func(c eval.Context) {
		fmt.Println(lookup.Lookup2(impl.NewInvocation(c), []string{`first`, `second`}, types.DefaultAnyType(), nil, eval.EMPTY_MAP, eval.EMPTY_MAP, eval.EMPTY_MAP, nil))
	})
	// Output: value of first
}

func ExampleLookup2_findSecond() {
	lookup.DoWithParent(context.Background(), tp, func(c eval.Context) {
		fmt.Println(lookup.Lookup2(impl.NewInvocation(c), []string{`nonexisting`, `second`}, types.DefaultAnyType(), nil, eval.EMPTY_MAP, eval.EMPTY_MAP, eval.EMPTY_MAP, nil))
	})
	// Output: includes 'value of first'
}

func ExampleLookup2_notFoundWithoutDflt() {
	fmt.Println(lookup.TryWithParent(context.Background(), tp, func(c eval.Context) error {
		lookup.Lookup2(impl.NewInvocation(c), []string{`nonexisting`, `notthere`}, types.DefaultAnyType(), nil, eval.EMPTY_MAP, eval.EMPTY_MAP, eval.EMPTY_MAP, nil)
		return nil
	}))
	// Output: lookup() did not find a value for any of the names [nonexisting, notthere]
}

func ExampleLookup2_notFoundDflt() {
	lookup.DoWithParent(context.Background(), tp, func(c eval.Context) {
		fmt.Println(lookup.Lookup2(impl.NewInvocation(c), []string{`nonexisting`, `notthere`}, types.DefaultAnyType(), types.WrapString(`default value`), eval.EMPTY_MAP, eval.EMPTY_MAP, eval.EMPTY_MAP, nil))
	})
	// Output: default value
}

func ExampleContextCachedValue() {

	cachingProvider := func(ic lookup.ProviderContext, key string, options eval.OrderedMap) (eval.Value, bool) {
		if v, ok := ic.CachedValue(key); ok {
			fmt.Printf("Returning cached value for %s\n", key)
			return v, true
		}
		fmt.Printf("Creating and caching value for %s\n", key)
		v := ic.Interpolate(types.WrapString(fmt.Sprintf("generated value for %%{%s}", key)))
		ic.Cache(key, v)
		return v, true
	}

	lookup.DoWithParent(context.Background(), cachingProvider, func(c eval.Context) {
		c.DoWithScope(evalimpl.NewScope2(types.WrapStringToInterfaceMap(c, map[string]interface{}{
			`a`: `scope 'a'`,
			`b`: `scope 'b'`,
		}), false), func() {
			ic := impl.NewInvocation(c)
			fmt.Println(lookup.Lookup(ic, `a`, nil, eval.EMPTY_MAP))
			fmt.Println(lookup.Lookup(ic, `b`, nil, eval.EMPTY_MAP))
			fmt.Println(lookup.Lookup(ic, `a`, nil, eval.EMPTY_MAP))
			fmt.Println(lookup.Lookup(ic, `b`, nil, eval.EMPTY_MAP))
		})
	})
	// Output:
	// Creating and caching value for a
	// generated value for scope 'a'
	// Creating and caching value for b
	// generated value for scope 'b'
	// generated value for scope 'a'
	// generated value for scope 'b'
}

func ExampleLookup_dottedStringInt() {
	lookup.DoWithParent(context.Background(), tp, func(c eval.Context) {
		v := lookup.Lookup(impl.NewInvocation(c), `hash.array.0`, nil, nil)
		fmt.Println(v)
	})
	// Output: two
}
package safeauth

import (
	"context"
	"slices"
)

type Privilege string

type privilegeKey struct{}

/***
*
***/
func Grant(c context.Context, ps ...Privilege) context.Context {
	if c.Value(privilegeKey{}) != nil {
		panic("privilege key called multiple times")
	}
	return context.WithValue(c, privilegeKey{}, ps)
}

type checkedKey struct{}

func Check(c context.Context, ps ...Privilege) (_ context.Context, ok bool) {
	granted, ok := c.Value(privilegeKey{}).([]Privilege)
	if !ok {
		return c, false
	}
	for _, p := range ps {
		if !slices.Contains(granted, p) {
			return c, false
		}
	}
	return context.WithValue(c, checkedKey{}, struct{}{}), true
}

func Must(c context.Context) (ok bool) {
	if c.Value(checkedKey{}) == nil {
		// Potentially collect call stack information here
		return false
	}
	return true
}

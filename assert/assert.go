// Package assert contains an assertion library, providing methods to e.g.
// assert equality of expected and actual values. This makes your unit tests
// much more readable.
//
// To use it, create an Assert object and call its methods for assertion:
//
//	asrt := NewAssert(t)
//	actual := myTestCall()
//	asrt.Equal(actual, "expected", "Got unexpected value from myTestCall()")
package assert

import (
	"reflect"
	"regexp"
	"testing"
)

// Assert wraps a testing.T pointer for storing failures.
type Assert struct {
	t *testing.T
}

// NewAssert returns an Assert type that wraps t.
func NewAssert(t *testing.T) *Assert {
	return &Assert{t}
}

func isNil(i interface{}) bool {
	if i == nil {
		return true
	}
	value := reflect.ValueOf(i)
	kind := value.Kind()
	if kind >= reflect.Chan && kind <= reflect.Slice && value.IsNil() {
		return true
	}
	return false
}

// Nil asserts that actual is nil.
func (a *Assert) Nil(actual interface{}, msg string) {
	if !isNil(actual) {
		a.t.Error(msg)
	}
}

// NotNil asserts that actual is not nil.
func (a *Assert) NotNil(actual interface{}, msg string) {
	if isNil(actual) {
		a.t.Error(msg)
	}
}

// Equal asserts that actual and expected are identical. It does not assert
// deep equality.
func (a *Assert) Equal(actual interface{}, expected interface{}, msg string) {
	if actual != expected {
		a.t.Errorf("%s, actual: '%v', expected: '%v'", msg, actual, expected)
	}
}

// EqualSlice asserts that the two given slices are equal, i.e. they have the
// same size and contain the same elements, doing a shallow compare of the
// elements.
func (a *Assert) EqualSlice(actual []interface{}, expected []interface{}, msg string) {
	if actual == nil && expected != nil || actual != nil && expected == nil {
		a.t.Errorf("%s, actual: '%v', expected: '%v'", msg, actual, expected)
		return
	}

	if len(actual) != len(expected) {
		a.t.Errorf("%s, lengths differ. Actual: '%d', expected: '%d'", msg, len(actual), len(expected))
		return
	}

	for idx, elem := range actual {
		if elem != expected[idx] {
			a.t.Errorf("%s, slice content differs. Actual: '%v', expected: '%v'", msg, actual, expected)
			return
		}
	}
}

// Match asserts that target matches pattern using a regular expression.
func (a *Assert) Match(pattern string, target string, msg string) {
	match, err := regexp.MatchString(pattern, target)
	if !match || err != nil {
		a.t.Errorf("%s, '%v' does not match '%v'", msg, target, pattern)
	}
}

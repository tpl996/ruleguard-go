package target

import (
	"errors"
	"fmt"
	"sync"
	"testing"
)

func sink(args ...interface{}) {}

func ifacePtr() {
	type structType struct {
		_ *fmt.Stringer // want `\Qdon't use pointers to an interface`
	}

	type ifacePtrAlias = *fmt.Stringer // want `\Qdon't use pointers to an interface`

	{
		var x *interface{} // want `\Qdon't use pointers to an interface`
		_ = x
		_ = *x
	}

	{
		var x **interface{} // want `\Qdon't use pointers to an interface`
		_ = x
		_ = *x
	}
}

func newMutex() {
	mu2 := new(sync.Mutex) // want `\Quse zero mutex value instead, 'var mu2 sync.Mutex'`
	mu2.Lock()             // want `\Qprovide an immediate unlock; 'defer mu2.Unlock()'`
}

func lockMutex() {
	var mu sync.Mutex // OK using empty value rather than new
	mu.Lock()         // OK followed by immediate defer to unlock
	defer mu.Unlock()

}

func lockMutext() {
	var mu3 sync.RWMutex // OK using empty var rather than new
	mu3.RLock()          // want `\Qprovide an immediate unlock; 'defer mu3.RUnlock()'`

	x := 3
	if x < 4 {
		defer mu3.RUnlock()
	}
}

type t struct {
	mu  sync.Mutex
	rmu sync.RWMutex
}

func (s *t) x() {
	s.mu.Lock() // want `\Qprovide an immediate unlock; 'defer s.mu.Unlock()'`
}

func (s *t) y() {
	s.mu.Lock()
	defer s.mu.Unlock()
}

func (s *t) r() {
	s.rmu.RLock() // want `\Qprovide an immediate unlock; 'defer s.rmu.RUnlock()'`
}

func (s *t) u() {
	s.rmu.RLock()
	defer s.rmu.RUnlock()
}

func channelSize() {
	_ = make(chan int, 1)    // OK: size of 1
	_ = make(chan string, 0) // OK: explicit size of 0
	_ = make(chan float32)   // OK: unbuffered, implicit size of 0

	size := 1
	_ = make(chan int, size) // OK: can't analyze

	_ = make(chan int, 2)     // want `\Qchannels should have a size of one or be unbuffered`
	_ = make(chan []int, 128) // want `\Qchannels should have a size of one or be unbuffered`
}

func uncheckedTypeAssert() {
	var v interface{}

	_ = v.(int) // want `\Qavoid unchecked type assertions as they can panic`
	{
		x := v.(int) // want `\Qavoid unchecked type assertions as they can panic`
		_ = x
	}

	sink(v.(int))          // want `\Qavoid unchecked type assertions as they can panic`
	sink(0, v.(int))       // want `\Qavoid unchecked type assertions as they can panic`
	sink(v.(int), 0)       // want `\Qavoid unchecked type assertions as they can panic`
	sink(1, 2, v.(int), 3) // want `\Qavoid unchecked type assertions as they can panic`

	{
		type structSink struct {
			f0 interface{}
			f1 interface{}
			f2 interface{}
		}
		_ = structSink{v.(int), 0, 0}    // want `\Qavoid unchecked type assertions as they can panic`
		_ = structSink{0, v.(string), 0} // want `\Qavoid unchecked type assertions as they can panic`
		_ = structSink{0, 0, v.([]int)}  // want `\Qavoid unchecked type assertions as they can panic`

		_ = structSink{f0: v.(int)}                  // want `\Qavoid unchecked type assertions as they can panic`
		_ = structSink{f0: 0, f1: v.(int)}           // want `\Qavoid unchecked type assertions as they can panic`
		_ = structSink{f0: 0, f1: 0, f2: v.(int)}    // want `\Qavoid unchecked type assertions as they can panic`
		_ = structSink{f0: v.(string), f1: 0, f2: 0} // want `\Qavoid unchecked type assertions as they can panic`
	}

	{
		_ = []interface{}{v.(int)}       // want `\Qavoid unchecked type assertions as they can panic`
		_ = []interface{}{0, v.(int)}    // want `\Qavoid unchecked type assertions as they can panic`
		_ = []interface{}{v.(int), 0}    // want `\Qavoid unchecked type assertions as they can panic`
		_ = []interface{}{0, v.(int), 0} // want `\Qavoid unchecked type assertions as they can panic`

		_ = [...]interface{}{10: v.(int)}               // want `\Qavoid unchecked type assertions as they can panic`
		_ = [...]interface{}{10: 0, 20: v.(int)}        // want `\Qavoid unchecked type assertions as they can panic`
		_ = [...]interface{}{10: v.(int), 20: 0}        // want `\Qavoid unchecked type assertions as they can panic`
		_ = [...]interface{}{10: 0, 20: v.(int), 30: 0} // want `\Qavoid unchecked type assertions as they can panic`
	}
}

func unnecessaryElse() {
	var cond bool

	{
		var x int //want `\Qrewrite as 'x := 5; if cond { x = 10 }'`
		if cond {
			x = 10
		} else {
			x = 5
		}
		_ = x
	}
}

func avoidFailedTo(e error) {
	// OK: not a wrapping call.
	_ = fmt.Errorf("failed to succeed in life")
	_ = errors.New("failed to succeed in life")

	// Could be an error wrapping, but it's not a usual form anyway (missing ':').
	_ = fmt.Errorf("failed to succeed in life %w", e)

	_ = fmt.Errorf("failed to succeed in life: %w", e) // want `\Q"failed to" message part is redundant in error wrapping`
	_ = fmt.Errorf("failed to succeed in life: %v", e) // want `\Q"failed to" message part is redundant in error wrapping`
}

type S struct { // want `\Qdo not embed sync.Mutex`
	sync.Mutex
}
type S1 struct { // want `\Qdo not embed sync.RWMutex`
	sync.RWMutex
}

var G = make(map[string]int) // want `\Q make map without size; use 'var map[string]int'`

func unsizedMaps() {
	m := make(map[string]string) // want `\Q make map without size; use 'var map[string]string'`
	_ = make(map[string]string, len(m))

	s := make([]int, 0) // want `\Q make slice without size; use 'var []int'`
	_ = make([]int, len(s), 0)
}

// options
// Detection of a func WithXXX returning a closure is assumed to be an option to a
// constructor. This isn't perfect
type s struct{}
type OptionF func(*s)

func WithXXX(p ...int) OptionF { // want `\Q use Option interface not closures`
	return func(*s) {}
}

// Since this function signature does't start with 'With' it is assumed not to be
// an option and so is not flagged.
func NonWithXXX(p ...int) OptionF {
	return func(*s) {}
}

// this is the style way of doing things and is allowed.
type Option interface{}

func WithYYY(p ...any) Option {
	return (Option)(nil)
}

// local var
func f() int {
	var x = 0 // want `\Q var used for assignment; use 'x := 0'`

	return x
}

// not returning nil
func l() []int {
	return []int{} // want `\Q returning empty slice; use 'return nil'`
}

// emptytSlice declaration
func eSlice() []string {
	x := []string{} // want `\Q empty slice declaration; use 'var x []string'`

	return x
}

func eFunc() error {
	return nil
}

func mFunc() (int, error) {
	return 0, nil
}

func scope() {
	var err error

	if err = eFunc(); err != nil { // want `\Q reduce scope; use 'if err := eFunc()'`

	}

	if _, err = mFunc(); err != nil { // want `\Q reduce scope; use 'if _, err := mFunc()'`

	}

	if err := eFunc(); err != nil {

	}
}

func printInfo(name string, isLocal, done bool) {}

func nakedBool() {
	const (
		isLocal   = true
		done      = true
		trueConst = true
	)

	printInfo("hi", true, true)   // want `\Q do not use naked booleans in call parameters`
	printInfo("hi", true, false)  // want `\Q do not use naked booleans in call parameters`
	printInfo("hi", false, false) // want `\Q do not use naked booleans in call parameters`
	printInfo("hi", isLocal, done)
	printInfo("hi", isLocal, true) // want `\Q do not use naked booleans in call parameters`
	printInfo("hi", isLocal, trueConst)

}

type P struct{}

func varForZeroValueStruct() {
	type s struct {
		f int
	}

	var b s

	c := s{} // want `\Quse var for zero value struct; var c s`

	d := s{
		f: 20,
	}

	b = c
	c.f = d.f
	b.f++

	p := P{} // want `\Quse var for zero value struct; var p P`
	_ = p
}

func nakedBoolInClosure(t *testing.T) {
	x := 0
	t.Run("", func(t *testing.T) {
		var found bool

		if x == 0 {
			found = true
		}

		if found {
			x++
		}
	})

}

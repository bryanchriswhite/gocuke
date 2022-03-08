package gocuke

import (
	"github.com/cucumber/messages-go/v16"
	"pgregory.net/rapid"
	"reflect"
	"testing"
)

// Runner is a test runner.
type Runner struct {
	topLevelT            *testing.T
	suiteType            reflect.Type
	initScenario         func(t TestingT) Suite
	incr                 *messages.Incrementing
	paths                []string
	parallel             bool
	stepDefs             []*stepDef
	suggestions          map[string]methodSig
	supportedSpecialArgs map[reflect.Type]specialArgGetter
	beforeHooks          []*stepDef
	afterHooks           []*stepDef
	beforeStepHooks      []*stepDef
	afterStepHooks       []*stepDef
}

// NewRunner constructs a new Runner with the provided initScenario function.
// initScenario will be called for each test case returning a new suite instance
// for each test case which can be used for sharing state between steps. It
// is expected that the suite will retain a copy of the TestingT instance
// for usage in each step.
func NewRunner(t *testing.T, initScenario func(t TestingT) Suite) *Runner {
	r := &Runner{
		topLevelT:    t,
		initScenario: initScenario,
		incr:         &messages.Incrementing{},
		suggestions:  map[string]methodSig{},
		supportedSpecialArgs: map[reflect.Type]specialArgGetter{
			// TestingT
			reflect.TypeOf((*TestingT)(nil)).Elem(): func(runner *scenarioRunner) interface{} {
				return runner.t
			},
			// *rapid.T
			reflect.TypeOf(&rapid.T{}): func(runner *scenarioRunner) interface{} {
				if t, ok := runner.t.(*rapid.T); ok {
					return t
				}
				runner.t.Fatalf("expected %T, but got %T", &rapid.T{}, runner.t)
				return nil
			},
		},
	}

	return r
}

func (r *Runner) setupSuite(initScenario func(t TestingT) Suite) {
	s := initScenario(r.topLevelT)
	r.suiteType = reflect.TypeOf(s)
	r.supportedSpecialArgs[r.suiteType] = func(runner *scenarioRunner) interface{} {
		return runner.s
	}

	if before, ok := r.suiteType.MethodByName("Before"); ok {
		r.beforeHooks = append(r.beforeHooks, r.newStepDefOrHook(r.topLevelT, nil, before.Func))
	}

	if after, ok := r.suiteType.MethodByName("After"); ok {
		r.afterHooks = append(r.afterHooks, r.newStepDefOrHook(r.topLevelT, nil, after.Func))
	}

	if beforeStep, ok := r.suiteType.MethodByName("BeforeStep"); ok {
		r.beforeStepHooks = append(r.beforeStepHooks, r.newStepDefOrHook(r.topLevelT, nil, beforeStep.Func))
	}

	if afterStep, ok := r.suiteType.MethodByName("AfterStep"); ok {
		r.afterStepHooks = append(r.afterStepHooks, r.newStepDefOrHook(r.topLevelT, nil, afterStep.Func))
	}
}

// Suite is a dummy interface to mark a test suite.
type Suite interface{}

var rapidTType = reflect.TypeOf(&rapid.T{})

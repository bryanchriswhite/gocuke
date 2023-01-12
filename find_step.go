package gocuke

import (
	"testing"

	"github.com/cucumber/common/messages/go/v17"
)

func (r *Runner) findStep(t *testing.T, step *messages.PickleStep) *stepDef {
	t.Helper()

	for _, def := range r.stepDefs {
		matches := def.regex.FindSubmatch([]byte(step.Text))
		if len(matches) != 0 {
			return def
		}
	}

	sig := guessMethodSig(step)
	method, ok := r.suiteType.MethodByName(sig.name)
	if ok {
		return r.addStepDef(t, sig.regex, method.Func)
	}

	r.suggestions[sig.name] = sig

	t.Logf("can't find step definition for: %s", step.Text)
	return nil
}

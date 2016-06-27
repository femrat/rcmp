package cmp

import (
	"github.com/femrat/rcmp/report"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestDoSafetyCheck(t *testing.T) {
	var err error

	err = doSafetyCheck(nil)
	if err == nil || err.Error() != "No report is presented" {
		t.Error("need nil error")
	}

	err = doSafetyCheck([]*report.Report{
		makeReport("REP1", row("A", "11"), row("B"), row("C")),
		makeReport("REP2", row("A", "21"), row("B"), row("C")),
		makeReport("REP3", row("A", "31"), row("B")),
	})
	assert.EqualError(t, err, "The number of instance-file in `REP1' is 3, while in `REP3' is 2, not matched")

	err = doSafetyCheck([]*report.Report{
		makeReport("REP1", row("A", "11"), row("C")),
		makeReport("REP2", row("A", "21"), row("C")),
		makeReport("REP3", row("A", "31"), row("D")),
	})
	assert.EqualError(t, err, "The instance-file of line 2 in `REP1' is `C', while in `REP3' is `D'")

	err = doSafetyCheck([]*report.Report{
		makeReport("REP1", row("A", "11"), row("C"), row("C")),
		makeReport("REP2", row("A", "21"), row("C"), row("C")),
		makeReport("REP3", row("A", "31"), row("C"), row("C")),
	})
	assert.EqualError(t, err, "The instance-file C appears twice")
}

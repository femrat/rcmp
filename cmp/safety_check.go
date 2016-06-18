package cmp

import (
	"fmt"
	"github.com/femrat/rcmp/report"
)

func doSafetyCheck(rc []*report.Report) error {
	if len(rc) == 0 {
		return fmt.Errorf("No report is presented")
	}
	if len(rc) == 1 {
		return nil
	}

	for i := 1; i < len(rc); i++ {
		if len(rc[i].Rows) != len(rc[0].Rows) {
			return fmt.Errorf("The number of instance-file in `%s' is %d, while in `%s' is %d, not matched",
				rc[0].ReportFile, len(rc[0].Rows), rc[i].ReportFile, len(rc[i].Rows))
		}
	}

	exist := make(map[string]bool)

	for line, baseRow := range rc[0].Rows {
		if exist[baseRow.File] {
			return fmt.Errorf("The instance-file %s appears twice", baseRow.File)
		}
		exist[baseRow.File] = true
		for i := 1; i < len(rc); i++ {
			if baseRow.File != rc[i].Rows[line].File {
				return fmt.Errorf("The instance-file of line %d in `%s' is `%s', while in `%s' is `%s'",
					line+1, rc[0].ReportFile, baseRow.File, rc[i].ReportFile, rc[i].Rows[line].File)
			}
		}
	}

	return nil
}

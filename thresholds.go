package monitoringplugin

import (
	"fmt"
	"github.com/pkg/errors"
	"math/big"
)

// CheckThresholds contains all threshold values
type CheckThresholds struct {
	WarningMin  interface{} `json:"warningMin" xml:"warningMin"`
	WarningMax  interface{} `json:"warningMax" xml:"warningMax"`
	CriticalMin interface{} `json:"criticalMin" xml:"criticalMin"`
	CriticalMax interface{} `json:"criticalMax" xml:"criticalMax"`
}

// Validate checks if the CheckThresholds contains some invalid combination of warning and critical values
func (c *CheckThresholds) Validate() error {
	if c.WarningMin != nil && c.WarningMax != nil {
		var min, max big.Float
		_, _, err := min.Parse(fmt.Sprint(c.WarningMin), 10)
		if err != nil {
			return errors.Wrap(err, "can't parse warning min")
		}
		_, _, err = max.Parse(fmt.Sprint(c.WarningMax), 10)
		if err != nil {
			return errors.Wrap(err, "can't parse warning max")
		}

		if res := min.Cmp(&max); res == 1 {
			return errors.New("warning min and max are invalid")
		}
	}

	if c.CriticalMin != nil && c.CriticalMax != nil {
		var min, max big.Float
		_, _, err := min.Parse(fmt.Sprint(c.CriticalMin), 10)
		if err != nil {
			return errors.Wrap(err, "can't parse critical min")
		}
		_, _, err = max.Parse(fmt.Sprint(c.CriticalMax), 10)
		if err != nil {
			return errors.Wrap(err, "can't parse critical max")
		}

		if res := min.Cmp(&max); res == 1 {
			return errors.New("critical min and max are invalid")
		}
	}

	if c.CriticalMin != nil && c.WarningMin != nil {
		var wMin, cMin big.Float
		_, _, err := wMin.Parse(fmt.Sprint(c.WarningMin), 10)
		if err != nil {
			return errors.Wrap(err, "can't parse warning min")
		}
		_, _, err = cMin.Parse(fmt.Sprint(c.CriticalMin), 10)
		if err != nil {
			return errors.Wrap(err, "can't parse critical min")
		}

		if res := cMin.Cmp(&wMin); res != -1 {
			return errors.New("critical and warning min are invalid")
		}
	}

	if c.WarningMax != nil && c.CriticalMax != nil {
		var wMax, cMax big.Float
		_, _, err := wMax.Parse(fmt.Sprint(c.WarningMax), 10)
		if err != nil {
			return errors.Wrap(err, "can't parse warning min")
		}
		_, _, err = cMax.Parse(fmt.Sprint(c.CriticalMax), 10)
		if err != nil {
			return errors.Wrap(err, "can't parse critical min")
		}

		if res := cMax.Cmp(&wMax); res != 1 {
			return errors.New("critical and warning max are invalid")
		}
	}

	return nil
}

// IsEmpty checks if the thresholds are empty
func (c *CheckThresholds) IsEmpty() bool {
	return c.WarningMin == nil && c.WarningMax == nil && c.CriticalMin == nil && c.CriticalMax == nil
}

// CheckValue checks if the input is violating the thresholds
func (c *CheckThresholds) CheckValue(v interface{}) (int, error) {
	var value, wMin, wMax, cMin, cMax big.Float
	_, _, err := value.Parse(fmt.Sprint(v), 10)
	if err != nil {
		return 0, errors.Wrap(err, "value can't be parsed")
	}
	if c.CriticalMin != nil {
		_, _, err := cMin.Parse(fmt.Sprint(c.CriticalMin), 10)
		if err != nil {
			return 0, errors.Wrap(err, "critical min can't be parsed")
		}
		if cMin.Cmp(&value) != -1 {
			return CRITICAL, nil
		}
	}
	if c.CriticalMax != nil {
		_, _, err := cMax.Parse(fmt.Sprint(c.CriticalMax), 10)
		if err != nil {
			return 0, errors.Wrap(err, "critical max can't be parsed")
		}
		if cMax.Cmp(&value) != 1 {
			return CRITICAL, nil
		}
	}
	if c.WarningMin != nil {
		_, _, err := wMin.Parse(fmt.Sprint(c.WarningMin), 10)
		if err != nil {
			return 0, errors.Wrap(err, "warning min can't be parsed")
		}
		if wMin.Cmp(&value) != -1 {
			return WARNING, nil
		}
	}
	if c.WarningMax != nil {
		_, _, err := wMax.Parse(fmt.Sprint(c.WarningMax), 10)
		if err != nil {
			return 0, errors.Wrap(err, "warning max can't be parsed")
		}
		if wMax.Cmp(&value) != 1 {
			return WARNING, nil
		}
	}
	return OK, nil
}

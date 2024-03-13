package monitoringplugin

import (
	"errors"
	"fmt"
	"math/big"
	"strconv"
)

// Thresholds contains all threshold values
type Thresholds struct {
	WarningMin  interface{} `json:"warningMin" xml:"warningMin"`
	WarningMax  interface{} `json:"warningMax" xml:"warningMax"`
	CriticalMin interface{} `json:"criticalMin" xml:"criticalMin"`
	CriticalMax interface{} `json:"criticalMax" xml:"criticalMax"`
}

// NewThresholds creates a new threshold
func NewThresholds(warningMin, warningMax, criticalMin, criticalMax interface{}) Thresholds {
	return Thresholds{
		WarningMin:  warningMin,
		WarningMax:  warningMax,
		CriticalMin: criticalMin,
		CriticalMax: criticalMax,
	}
}

// Validate checks if the Thresholds contains some invalid combination of warning and critical values
func (c *Thresholds) Validate() error {
	if c.WarningMin != nil && c.WarningMax != nil {
		var min, max big.Float
		_, _, err := min.Parse(fmt.Sprint(c.WarningMin), 10)
		if err != nil {
			return fmt.Errorf("can't parse warning min: %w", err)
		}
		_, _, err = max.Parse(fmt.Sprint(c.WarningMax), 10)
		if err != nil {
			return fmt.Errorf("can't parse warning max: %w", err)
		}

		if res := min.Cmp(&max); res == 1 {
			return errors.New("warning min and max are invalid")
		}
	}

	if c.CriticalMin != nil && c.CriticalMax != nil {
		var min, max big.Float
		_, _, err := min.Parse(fmt.Sprint(c.CriticalMin), 10)
		if err != nil {
			return fmt.Errorf("can't parse critical min: %w", err)
		}
		_, _, err = max.Parse(fmt.Sprint(c.CriticalMax), 10)
		if err != nil {
			return fmt.Errorf("can't parse critical max: %w", err)
		}

		if res := min.Cmp(&max); res == 1 {
			return errors.New("critical min and max are invalid")
		}
	}

	if c.CriticalMin != nil && c.WarningMin != nil {
		var wMin, cMin big.Float
		_, _, err := wMin.Parse(fmt.Sprint(c.WarningMin), 10)
		if err != nil {
			return fmt.Errorf("can't parse warning min: %w", err)
		}
		_, _, err = cMin.Parse(fmt.Sprint(c.CriticalMin), 10)
		if err != nil {
			return fmt.Errorf("can't parse critical min: %w", err)
		}

		if res := cMin.Cmp(&wMin); res == 1 {
			return errors.New("critical and warning min are invalid")
		}
	}

	if c.WarningMax != nil && c.CriticalMax != nil {
		var wMax, cMax big.Float
		_, _, err := wMax.Parse(fmt.Sprint(c.WarningMax), 10)
		if err != nil {
			return fmt.Errorf("can't parse warning min: %w", err)
		}
		_, _, err = cMax.Parse(fmt.Sprint(c.CriticalMax), 10)
		if err != nil {
			return fmt.Errorf("can't parse critical min: %w", err)
		}

		if res := cMax.Cmp(&wMax); res == -1 {
			return errors.New("critical and warning max are invalid")
		}
	}

	return nil
}

// HasWarning checks if a warning threshold is set
func (c *Thresholds) HasWarning() bool {
	return c.WarningMax != nil || c.WarningMin != nil
}

// HasCritical checks if a critical threshold is set
func (c *Thresholds) HasCritical() bool {
	return c.CriticalMax != nil || c.CriticalMin != nil
}

// IsEmpty checks if the thresholds are empty
func (c *Thresholds) IsEmpty() bool {
	return c.WarningMin == nil && c.WarningMax == nil && c.CriticalMin == nil && c.CriticalMax == nil
}

// CheckValue checks if the input is violating the thresholds
func (c *Thresholds) CheckValue(v interface{}) (int, error) {
	var value, wMin, wMax, cMin, cMax big.Float
	_, _, err := value.Parse(fmt.Sprint(v), 10)
	if err != nil {
		return 0, fmt.Errorf("value can't be parsed: %w", err)
	}
	if c.CriticalMin != nil {
		_, _, err := cMin.Parse(fmt.Sprint(c.CriticalMin), 10)
		if err != nil {
			return 0, fmt.Errorf("critical min can't be parsed: %w", err)
		}
		if cMin.Cmp(&value) == 1 {
			return CRITICAL, nil
		}
	}
	if c.CriticalMax != nil {
		_, _, err := cMax.Parse(fmt.Sprint(c.CriticalMax), 10)
		if err != nil {
			return 0, fmt.Errorf("critical max can't be parsed: %w", err)
		}
		if cMax.Cmp(&value) == -1 {
			return CRITICAL, nil
		}
	}
	if c.WarningMin != nil {
		_, _, err := wMin.Parse(fmt.Sprint(c.WarningMin), 10)
		if err != nil {
			return 0, fmt.Errorf("warning min can't be parsed: %w", err)
		}
		if wMin.Cmp(&value) == 1 {
			return WARNING, nil
		}
	}
	if c.WarningMax != nil {
		_, _, err := wMax.Parse(fmt.Sprint(c.WarningMax), 10)
		if err != nil {
			return 0, fmt.Errorf("warning max can't be parsed: %w", err)
		}
		if wMax.Cmp(&value) == -1 {
			return WARNING, nil
		}
	}
	return OK, nil
}

func (c *Thresholds) getWarning() string {
	return getRange(c.WarningMin, c.WarningMax)
}

func (c *Thresholds) getCritical() string {
	return getRange(c.CriticalMin, c.CriticalMax)
}

func getRange(min, max interface{}) string {
	if min == nil && max == nil {
		return ""
	}

	var res string

	if min != nil {
		var minString string
		switch m := min.(type) {
		case float64:
			minString = strconv.FormatFloat(m, 'f', -1, 64)
		default:
			minString = fmt.Sprint(m)
		}
		if minString != "0" || max == nil {
			res += minString + ":"
		}
	} else {
		res += "~:"
	}

	if max != nil {
		var maxString string
		switch m := max.(type) {
		case float64:
			maxString = strconv.FormatFloat(m, 'f', -1, 64)
		default:
			maxString = fmt.Sprint(m)
		}
		res += maxString
	}

	return res
}

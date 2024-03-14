package monitoringplugin

import (
	"cmp"
	"errors"
	"fmt"
	"strings"
)

// Thresholds contains all threshold values.
type Thresholds[T cmp.Ordered] struct {
	WarningMin  T `json:"warningMin" xml:"warningMin"`
	WarningMax  T `json:"warningMax" xml:"warningMax"`
	CriticalMin T `json:"criticalMin" xml:"criticalMin"`
	CriticalMax T `json:"criticalMax" xml:"criticalMax"`

	hasWarnMin, hasWarnMax bool
	hasCritMin, hasCritMax bool
}

// NewThresholds creates a new threshold.
func NewThresholds[T cmp.Ordered](warningMin, warningMax, criticalMin,
	criticalMax T,
) Thresholds[T] {
	return Thresholds[T]{
		WarningMin:  warningMin,
		WarningMax:  warningMax,
		CriticalMin: criticalMin,
		CriticalMax: criticalMax,

		hasWarnMin: true,
		hasWarnMax: true,
		hasCritMin: true,
		hasCritMax: true,
	}
}

// UseWarning configures how to use WarningMin and WarningMax.
func (c *Thresholds[T]) UseWarning(useMin, useMax bool) *Thresholds[T] {
	c.hasWarnMin, c.hasWarnMax = useMin, useMax
	return c
}

// UseCritical configures how to use CriticalMin and CriticalMax.
func (c *Thresholds[T]) UseCritical(useMin, useMax bool) *Thresholds[T] {
	c.hasCritMin, c.hasCritMax = useMin, useMax
	return c
}

// Validate checks if the Thresholds contains some invalid combination of
// warning and critical values.
func (c *Thresholds[T]) Validate() error {
	if c.hasWarnMin && c.hasWarnMax && cmp.Compare(c.WarningMin, c.WarningMax) == 1 {
		return errors.New("warning min and max are invalid")
	}

	if c.hasCritMin && c.hasCritMax && cmp.Compare(c.CriticalMin, c.CriticalMax) == 1 {
		return errors.New("critical min and max are invalid")
	}

	if c.hasCritMin && c.hasWarnMin && cmp.Compare(c.CriticalMin, c.WarningMin) == 1 {
		return errors.New("critical and warning min are invalid")
	}

	if c.hasWarnMax && c.hasCritMax && cmp.Compare(c.CriticalMax, c.WarningMax) == -1 {
		return errors.New("critical and warning max are invalid")
	}
	return nil
}

// HasWarning checks if a warning threshold is set.
func (c *Thresholds[T]) HasWarning() bool {
	return c.hasWarnMax || c.hasWarnMin
}

// HasCritical checks if a critical threshold is set.
func (c *Thresholds[T]) HasCritical() bool {
	return c.hasCritMax || c.hasCritMin
}

// IsEmpty checks if the thresholds are empty.
func (c *Thresholds[T]) IsEmpty() bool {
	return !c.HasWarning() && !c.HasCritical()
}

// CheckValue checks if the input is violating the thresholds.
func (c *Thresholds[T]) CheckValue(value T) int {
	switch {
	case c.hasCritMin && cmp.Compare(c.CriticalMin, value) == 1:
		return CRITICAL
	case c.hasCritMax && cmp.Compare(c.CriticalMax, value) == -1:
		return CRITICAL
	case c.hasWarnMin && cmp.Compare(c.WarningMin, value) == 1:
		return WARNING
	case c.hasWarnMax && cmp.Compare(c.WarningMax, value) == -1:
		return WARNING
	}
	return OK
}

func (c *Thresholds[T]) getWarning() string {
	return getRange(c.WarningMin, c.WarningMax, c.hasWarnMin, c.hasWarnMax)
}

func (c *Thresholds[T]) getCritical() string {
	return getRange(c.CriticalMin, c.CriticalMax, c.hasCritMin, c.hasCritMax)
}

func getRange[T cmp.Ordered](min, max T, hasMin, hasMax bool) string {
	if !hasMin && !hasMax {
		return ""
	}

	var b strings.Builder
	if hasMin {
		minString := fmt.Sprint(min)
		if minString != "0" || !hasMax {
			b.WriteString(minString)
			b.WriteString(":")
		}
	} else {
		b.WriteString("~:")
	}

	if hasMax {
		b.WriteString(fmt.Sprint(max))
	}
	return b.String()
}

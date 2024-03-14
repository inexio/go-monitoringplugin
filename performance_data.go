package monitoringplugin

import (
	"bytes"
	"cmp"
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
)

func newPerformanceDataPointKey(metric, label string) performanceDataPointKey {
	return performanceDataPointKey{Metric: metric, Label: label}
}

type performanceDataPointKey struct {
	Metric string `json:"metric"`
	Label  string `json:"label,omitempty"`
}

func (self *performanceDataPointKey) String() string {
	if self.Label == "" {
		return self.Metric
	}
	return self.Metric + " and label " + self.Label
}

func newPerformanceData() performanceData {
	return performanceData{
		keys: make(map[performanceDataPointKey]int),
	}
}

// performanceData is a map where all performanceDataPoints are stored.
// It assigns labels (string) to performanceDataPoints.
type performanceData struct {
	keys   map[performanceDataPointKey]int
	points []anyDataPoint
}

type anyDataPoint interface {
	Validate() error
	HasThresholds() bool
	CheckThresholds() int
	Name() string

	key() performanceDataPointKey
	output(jsonLabel bool) []byte
}

// add adds a PerformanceDataPoint to the performanceData Map. The function
// checks if a PerformanceDataPoint is valid and if there is already another
// PerformanceDataPoint with the same metric in the performanceData map.
//
// Usage:
//
//	err := performanceData.add(NewPerformanceDataPoint("temperature", 32, "°C").SetWarn(35).SetCrit(40))
//	if err != nil {
//	  ...
//	}
func (p *performanceData) add(point anyDataPoint) error {
	if err := point.Validate(); err != nil {
		return fmt.Errorf("given performance data point is not valid: %w", err)
	}
	key := point.key()
	if _, ok := p.keys[key]; ok {
		return fmt.Errorf(
			"a performance data point with the metric '%s' does already exist", key)
	}
	p.keys[key] = len(p.points)
	p.points = append(p.points, point)
	return nil
}

func (p *performanceData) point(key performanceDataPointKey) anyDataPoint {
	if i, ok := p.keys[key]; ok {
		return p.points[i]
	}
	return nil
}

// getInfo returns all information for performance data.
func (p *performanceData) getInfo() []anyDataPoint {
	return p.points
}

// NewPerformanceDataPoint creates a new PerformanceDataPoint. Metric and value
// are mandatory but are not checked at this point, the performanceDatePoint's
// validation is checked later when it is added to the performanceData list in
// the function performanceData.add(*PerformanceDataPoint).
//
// It is possible to directly add thresholds, min and max values with the
// functions SetThresholds(Thresholds), SetMin(int) and SetMax(int).
//
// Usage:
//
//	PerformanceDataPoint := NewPerformanceDataPoint("memory_usage", 55).SetUnit("%")
func NewPerformanceDataPoint[T cmp.Ordered](metric string, value T,
) *PerformanceDataPoint[T] {
	return &PerformanceDataPoint[T]{Metric: metric, Value: value}
}

// PerformanceDataPoint contains all information of one PerformanceDataPoint.
type PerformanceDataPoint[T cmp.Ordered] struct {
	Metric     string        `json:"metric" xml:"metric"`
	Label      string        `json:"label" xml:"label"`
	Value      T             `json:"value" xml:"value"`
	Unit       string        `json:"unit" xml:"unit"`
	Thresholds Thresholds[T] `json:"thresholds" xml:"thresholds"`
	Min        T             `json:"min" xml:"min"`
	Max        T             `json:"max" xml:"max"`

	hasMin, hasMax bool
}

var (
	reInvalidMetricLabel = regexp.MustCompile("([='])")
	reInvalidUnit        = regexp.MustCompile("([0-9;'\"])")
)

// Validate validates a PerformanceDataPoint. This function is used to check if
// a PerformanceDataPoint is compatible with the documentation from
// [Monitoring Plugins Development Guidelines](https://www.monitoring-plugins.org/doc/guidelines.html)
// (valid name and unit, value is inside the range of min and max etc.)
func (p *PerformanceDataPoint[T]) Validate() error {
	if p.Metric == "" {
		return errors.New("data point metric cannot be an empty string")
	}

	if reInvalidMetricLabel.MatchString(p.Metric) {
		return errors.New("metric contains invalid character")
	}

	if reInvalidMetricLabel.MatchString(p.Label) {
		return errors.New("metric contains invalid character")
	}

	if reInvalidUnit.MatchString(p.Unit) {
		return errors.New("unit can not contain numbers, semicolon or quotes")
	}

	if p.hasMin && cmp.Compare(p.Min, p.Value) == 1 {
		return errors.New("value cannot be smaller than min")
	}

	if p.hasMax && cmp.Compare(p.Max, p.Value) == -1 {
		return errors.New("value cannot be larger than max")
	}

	if p.hasMin && p.hasMax && cmp.Compare(p.Min, p.Max) == 1 {
		return errors.New("min cannot be larger than max")
	}

	if p.HasThresholds() {
		if err := p.Thresholds.Validate(); err != nil {
			return fmt.Errorf("thresholds are invalid: %w", err)
		}
	}
	return nil
}

func (p *PerformanceDataPoint[T]) key() performanceDataPointKey {
	return newPerformanceDataPointKey(p.Metric, p.Label)
}

// SetUnit sets the unit of the performance data point
func (p *PerformanceDataPoint[T]) SetUnit(unit string) *PerformanceDataPoint[T] {
	p.Unit = unit
	return p
}

// SetMin sets minimum value.
func (p *PerformanceDataPoint[T]) SetMin(min T) *PerformanceDataPoint[T] {
	p.Min, p.hasMin = min, true
	return p
}

// SetMax sets maximum value.
func (p *PerformanceDataPoint[T]) SetMax(max T) *PerformanceDataPoint[T] {
	p.Max, p.hasMax = max, true
	return p
}

// SetLabel adds a tag to the performance data point
// If one tag is added more than once, the value before will be overwritten
func (p *PerformanceDataPoint[T]) SetLabel(label string,
) *PerformanceDataPoint[T] {
	p.Label = label
	return p
}

// SetThresholds sets the thresholds for the performance data point
func (p *PerformanceDataPoint[T]) SetThresholds(thresholds Thresholds[T],
) *PerformanceDataPoint[T] {
	p.Thresholds = thresholds
	return p
}

// This function returns the PerformanceDataPoint in the specified format that
// will be returned by the check plugin.
func (p *PerformanceDataPoint[T]) output(jsonLabel bool) []byte {
	var buffer bytes.Buffer
	if jsonLabel {
		buffer.WriteByte('\'')
		key := performanceDataPointKey{Metric: p.Metric, Label: p.Label}
		jsonKey, _ := json.Marshal(key)
		buffer.Write(jsonKey)
		buffer.WriteByte('\'')
	} else {
		buffer.WriteByte('\'')
		buffer.WriteString(p.Metric)
		if p.Label != "" {
			buffer.WriteByte('_')
			buffer.WriteString(p.Label)
		}
		buffer.WriteByte('\'')
	}
	buffer.WriteByte('=')

	buffer.WriteString(fmt.Sprint(p.Value))
	buffer.WriteString(p.Unit)

	if p.HasThresholds() || p.hasMax || p.hasMin {
		buffer.WriteByte(';')
		if p.Thresholds.HasWarning() {
			buffer.WriteString(p.Thresholds.getWarning())
		}
		buffer.WriteByte(';')
		if p.Thresholds.HasCritical() {
			buffer.WriteString(p.Thresholds.getCritical())
		}
		buffer.WriteByte(';')
		if p.hasMin {
			buffer.WriteString(fmt.Sprint(p.Min))
		}
		buffer.WriteByte(';')
		if p.hasMax {
			buffer.WriteString(fmt.Sprint(p.Max))
		}
	}
	return buffer.Bytes()
}

// HasThresholds checks if the thresholds are not empty.
func (p *PerformanceDataPoint[T]) HasThresholds() bool {
	return !p.Thresholds.IsEmpty()
}

// Name returns a human-readable name suitable for [Response.UpdateStatus].
func (p *PerformanceDataPoint[T]) Name() string {
	if p.Label == "" {
		return p.Metric
	}
	return p.Metric + " (" + p.Label + ")"
}

// CheckThresholds checks if [Value] is violating the thresholds. See
// [Thresholds.CheckValue].
func (p *PerformanceDataPoint[T]) CheckThresholds() int {
	return p.Thresholds.CheckValue(p.Value)
}

// NewThresholds is a wrapper, which creates [NewThresholds] with same type as
// [Value], [SetThresholds] it and returns pointer to [Thresholds].
func (p *PerformanceDataPoint[T]) NewThresholds(warningMin, warningMax,
	criticalMin, criticalMax T,
) *Thresholds[T] {
	th := NewThresholds(warningMin, warningMax, criticalMin, criticalMax)
	p.SetThresholds(th)
	return &p.Thresholds
}

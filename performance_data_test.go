package monitoringplugin

import (
	"fmt"
	"regexp"
	"testing"
)

func TestPerformanceDataPointCreation(t *testing.T) {
	metric := "testMetric"
	var value float64 = 10
	p := NewPerformanceDataPoint(metric, value)

	if p.Metric != metric || p.Value != value {
		t.Error("the created PerfomanceDataPoint NewPerformanceDataPoint")
	}

	unit := "%"
	p.SetUnit(unit)
	if p.Unit != unit {
		t.Error("SetUnit failed")
	}

	label := "testLabel"
	p.SetLabel(label)
	if p.Label != label {
		t.Error("SetLabel failed")
	}

	var min float64
	p.SetMin(min)
	if p.Min != min || p.Min == nil {
		t.Error("SetMin failed")
	}

	var max float64 = 100
	p.SetMax(max)
	if p.Max != max || p.Max == nil {
		t.Error("SetMax failed")
	}

	thresholds := Thresholds{
		WarningMin:  0,
		WarningMax:  10,
		CriticalMin: 0,
		CriticalMax: 20,
	}
	p.SetThresholds(thresholds)
	if p.Thresholds != thresholds {
		t.Error("SetThresholds failed")
	}
}

func TestPerformanceDataPoint_validate(t *testing.T) {
	p := NewPerformanceDataPoint("metric", 10).SetMin(0).SetMax(100)
	if err := p.Validate(); err != nil {
		t.Error("valid performance data point resulted in an error: " + err.Error())
	}

	// empty metric
	p = NewPerformanceDataPoint("", 10)
	if err := p.Validate(); err == nil {
		t.Error("invalid performance data did not return an error (case: empty metric)")
	}

	// invalid metric
	p = NewPerformanceDataPoint("metric=", 10)
	if err := p.Validate(); err == nil {
		t.Error("invalid performance data did not return an error (case: invalid metric, contains =)")
	}
	p = NewPerformanceDataPoint("'metric'", 10)

	if err := p.Validate(); err == nil {
		t.Error("invalid performance data did not return an error (case: invalid metric, contains single quotes)")
	}

	// invalid unit
	p = NewPerformanceDataPoint("metric", 10).SetUnit("unit1")
	if err := p.Validate(); err == nil {
		t.Error("invalid performance data did not return an error (case: invalid unit, contains numbers)")
	}
	p = NewPerformanceDataPoint("metric", 10).SetUnit("unit;")
	if err := p.Validate(); err == nil {
		t.Error("invalid performance data did not return an error (case: invalid unit, contains semicolon)")
	}
	p = NewPerformanceDataPoint("metric", 10).SetUnit("unit'")
	if err := p.Validate(); err == nil {
		t.Error("invalid performance data did not return an error (case: invalid unit, contains single quotes)")
	}
	p = NewPerformanceDataPoint("metric", 10).SetUnit("unit\"")
	if err := p.Validate(); err == nil {
		t.Error("invalid performance data did not return an error (case: invalid unit, contains double quotes)")
	}

	// value < min
	p = NewPerformanceDataPoint("metric", 10).SetMin(50)
	if err := p.Validate(); err == nil {
		t.Error("invalid performance data did not return an error (case: value < min)")
	}

	// value > max
	p = NewPerformanceDataPoint("metric", 10).SetMax(5)
	if err := p.Validate(); err == nil {
		t.Error("invalid performance data did not return an error (case: value < min)")
	}

	// min > max
	p = NewPerformanceDataPoint("metric", 10).SetMin(10).SetMax(5)
	if err := p.Validate(); err == nil {
		t.Error("invalid performance data did not return an error (case: max < min)")
	}
}

func TestPerformanceDataPoint_output(t *testing.T) {
	label := "metric"
	value := 10.0
	unit := "s"
	warn := 40.0
	crit := 50.0
	min := 0.0
	max := 60.0

	p := NewPerformanceDataPoint(label, value)
	regex := fmt.Sprintf("'%s'=%g", label, value)
	match, err := regexp.Match(regex, p.output(false))
	if err != nil {
		t.Error(err.Error())
	}
	if !match {
		t.Error("output string did not match regex")
	}

	p.SetUnit(unit)
	regex = fmt.Sprintf("'%s'=%g%s", label, value, unit)
	match, err = regexp.Match(regex, p.output(false))
	if err != nil {
		t.Error(err.Error())
	}
	if !match {
		t.Error("output string did not match regex")
	}

	p.SetMax(max)
	regex = fmt.Sprintf("'%s'=%g%s;;;;%g", label, value, unit, max)
	match, err = regexp.Match(regex, p.output(false))
	if err != nil {
		t.Error(err.Error())
	}
	if !match {
		t.Error("output string did not match regex")
	}

	p.SetThresholds(NewThresholds(nil, warn, nil, crit))
	regex = fmt.Sprintf("'%s'=%g%s;~:%g;~:%g;;%g", label, value, unit, warn, crit, max)
	match, err = regexp.Match(regex, p.output(false))
	if err != nil {
		t.Error(err.Error())
	}
	if !match {
		t.Error("output string did not match regex")
	}

	p.SetThresholds(NewThresholds(0, nil, -10, nil))
	regex = fmt.Sprintf("'%s'=%g%s;%d:;%d:;;%g", label, value, unit, 0, -10, max)
	match, err = regexp.Match(regex, p.output(false))
	if err != nil {
		t.Error(err.Error())
	}
	if !match {
		t.Error("output string did not match regex")
	}

	p.SetThresholds(NewThresholds(5, 10, 3, 11))
	regex = fmt.Sprintf("'%s'=%g%s;%d:%d;%d:%d;;%g", label, value, unit, 5, 10, 3, 11, max)
	match, err = regexp.Match(regex, p.output(false))
	if err != nil {
		t.Error(err.Error())
	}
	if !match {
		t.Error("output string did not match regex")
	}

	p.SetThresholds(NewThresholds(0, warn, 0, crit))
	regex = fmt.Sprintf("'%s'=%g%s;%g;%g;;%g", label, value, unit, warn, crit, max)
	match, err = regexp.Match(regex, p.output(false))
	if err != nil {
		t.Error(err.Error())
	}
	if !match {
		t.Error("output string did not match regex")
	}

	p.SetMin(min)
	regex = fmt.Sprintf("'%s'=%g%s;%g;%g;%g;%g", label, value, unit, warn, crit, min, max)
	match, err = regexp.Match(regex, p.output(false))
	if err != nil {
		t.Error(err.Error())
	}
	if !match {
		t.Error("output string did not match regex")
	}

	regex = fmt.Sprintf(`'{"metric":"%s"}'=%g%s;%g;%g;%g;%g`, label, value, unit, warn, crit, min, max)
	match, err = regexp.Match(regex, p.output(true))
	if err != nil {
		t.Error(err.Error())
	}
	if !match {
		t.Error("output string did not match regex")
	}

	tag := "tag"
	p.SetLabel(tag)
	regex = fmt.Sprintf(`'{"metric":"%s","label":"%s"}'=%g%s;%g;%g;%g;%g`, label, tag, value, unit, warn, crit, min, max)
	match, err = regexp.Match(regex, p.output(true))
	if err != nil {
		t.Error(err.Error())
	}
	if !match {
		t.Error("output string did not match regex")
	}

	regex = fmt.Sprintf(`'%s_%s'=%g%s;%g;%g;%g;%g`, label, tag, value, unit, warn, crit, min, max)
	match, err = regexp.Match(regex, p.output(false))
	if err != nil {
		t.Error(err.Error())
	}
	if !match {
		t.Error("output string did not match regex")
	}
}

func TestPerformanceData_add(t *testing.T) {
	perfData := make(performanceData)

	// valid perfdata point
	err := perfData.add(NewPerformanceDataPoint("metric", 10))
	if err != nil {
		t.Error("adding a valid performance data point resulted in an error")
		return
	}

	if _, ok := perfData[performanceDataPointKey{"metric", ""}]; !ok {
		t.Error("performance data point was not added to the map of performance data points")
	}

	err = perfData.add(NewPerformanceDataPoint("metric", 10))
	if err == nil {
		t.Error("there was no error when adding a performance data point with a metric, that already exists in performance data")
	}

	err = perfData.add(NewPerformanceDataPoint("metric", 10).SetLabel("tag1"))
	if err != nil {
		t.Error("adding a valid performance data point resulted in an error")
		return
	}

	err = perfData.add(NewPerformanceDataPoint("metric", 10).SetLabel("tag2"))
	if err != nil {
		t.Error("adding a valid performance data point resulted in an error")
		return
	}

	err = perfData.add(NewPerformanceDataPoint("metric", 10).SetLabel("tag1"))
	if err == nil {
		t.Error("there was no error when adding a performance data point with a metric and tag, that already exists in performance data")
	}
}

func TestResponse_SetPerformanceDataJsonLabel(t *testing.T) {
	perfData := make(performanceData)

	// valid perfdata point
	err := perfData.add(NewPerformanceDataPoint("metric", 10))
	if err != nil {
		t.Error("adding a valid performance data point resulted in an error")
		return
	}

	if _, ok := perfData[performanceDataPointKey{"metric", ""}]; !ok {
		t.Error("performance data point was not added to the map of performance data points")
	}

	err = perfData.add(NewPerformanceDataPoint("metric", 10))
	if err == nil {
		t.Error("there was no error when adding a performance data point with a metric, that already exists in performance data")
	}
}

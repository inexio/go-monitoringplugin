/* Copyright (c) 2020, inexio GmbH, BSD 2-Clause License */

package monitoringplugin

import (
	"fmt"
	"regexp"
	"testing"
)

func TestPerformanceDataPointCreation(t *testing.T) {
	label := "label"
	var value float64 = 10
	unit := "%"
	p := NewPerformanceDataPoint(label, value, unit)

	if p.label != label || p.value != value || p.unit != unit {
		t.Error("the created PerfomanceDataPoint NewPerformanceDataPoint")
	}

	var min float64
	p.SetMin(min)
	if p.min != min || !p.hasMin {
		t.Error("SetMin failed")
	}

	var max float64 = 100
	p.SetMax(max)
	if p.max != max || !p.hasMax {
		t.Error("SetMax failed")
	}

	var warn float64 = 70
	p.SetWarn(warn)
	if p.warn != warn || !p.hasWarn {
		t.Error("SetWarn failed")
	}

	var crit float64 = 90
	p.SetCrit(crit)
	if p.crit != crit || !p.hasCrit {
		t.Error("SetCrit failed")
	}
	return
}

func TestPerformanceDataPoint_validate(t *testing.T) {
	p := NewPerformanceDataPoint("label", 10, "").SetMin(0).SetMax(100).SetWarn(60).SetCrit(80)
	if err := p.validate(); err != nil {
		t.Error("valid performance data point resulted in an error: " + err.Error())
	}

	//empty label
	p = NewPerformanceDataPoint("", 10, "")
	if err := p.validate(); err == nil {
		t.Error("invalid performance data did not return an error (case: empty label)")
	}

	//invalid label
	p = NewPerformanceDataPoint("label=", 10, "")
	if err := p.validate(); err == nil {
		t.Error("invalid performance data did not return an error (case: invalid label, contains =)")
	}
	p = NewPerformanceDataPoint("'label'", 10, "")
	if err := p.validate(); err == nil {
		t.Error("invalid performance data did not return an error (case: invalid label, contains single quotes)")
	}

	//invalid unit
	p = NewPerformanceDataPoint("label", 10, "unit1")
	if err := p.validate(); err == nil {
		t.Error("invalid performance data did not return an error (case: invalid unit, contains numbers)")
	}
	p = NewPerformanceDataPoint("label", 10, "unit;")
	if err := p.validate(); err == nil {
		t.Error("invalid performance data did not return an error (case: invalid unit, contains semicolon)")
	}
	p = NewPerformanceDataPoint("label", 10, "unit'")
	if err := p.validate(); err == nil {
		t.Error("invalid performance data did not return an error (case: invalid unit, contains single quotes)")
	}
	p = NewPerformanceDataPoint("label", 10, "unit\"")
	if err := p.validate(); err == nil {
		t.Error("invalid performance data did not return an error (case: invalid unit, contains double quotes)")
	}

	//value < min
	p = NewPerformanceDataPoint("label", 10, "").SetMin(50)
	if err := p.validate(); err == nil {
		t.Error("invalid performance data did not return an error (case: value < min)")
	}

	//value > max
	p = NewPerformanceDataPoint("label", 10, "").SetMax(5)
	if err := p.validate(); err == nil {
		t.Error("invalid performance data did not return an error (case: value < min)")
	}

	//min > max
	p = NewPerformanceDataPoint("label", 10, "").SetMin(10).SetMax(5)
	if err := p.validate(); err == nil {
		t.Error("invalid performance data did not return an error (case: max < min)")
	}
}

func TestPerformanceDataPoint_output(t *testing.T) {
	label := "label"
	value := 10.0
	unit := "s"
	warn := 40.0
	crit := 50.0
	min := 0.0
	max := 60.0

	p := NewPerformanceDataPoint(label, value, unit)
	regex := fmt.Sprintf("'%s'=%g%s;;;;", label, value, unit)
	match, err := regexp.Match(regex, p.output(false))
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

	p.SetWarn(warn)
	regex = fmt.Sprintf("'%s'=%g%s;%g;;;%g", label, value, unit, warn, max)
	match, err = regexp.Match(regex, p.output(false))
	if err != nil {
		t.Error(err.Error())
	}
	if !match {
		t.Error("output string did not match regex")
	}

	p.SetCrit(crit)
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
	p.SetLabelTag(tag)
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
	perfData := make(PerformanceData)

	//valid perfdata point
	err := perfData.add(NewPerformanceDataPoint("label", 10, ""))
	if err != nil {
		t.Error("adding a valid performance data point resulted in an error")
		return
	}

	if _, ok := perfData["label"]; !ok {
		t.Error("performance data point was not added to the map of performance data points")
	}

	err = perfData.add(NewPerformanceDataPoint("label", 10, ""))
	if err == nil {
		t.Error("there was no error when adding a performance data point with a label, that already exists in performance data")
	}

	err = perfData.add(NewPerformanceDataPoint("label", 10, "").SetLabelTag("tag1"))
	if err != nil {
		t.Error("adding a valid performance data point resulted in an error")
		return
	}

	err = perfData.add(NewPerformanceDataPoint("label", 10, "").SetLabelTag("tag2"))
	if err != nil {
		t.Error("adding a valid performance data point resulted in an error")
		return
	}

	err = perfData.add(NewPerformanceDataPoint("label", 10, "").SetLabelTag("tag1"))
	if err == nil {
		t.Error("there was no error when adding a performance data point with a label and tag, that already exists in performance data")
	}
}

func TestResponse_SetPerformanceDataJsonLabel(t *testing.T) {
	perfData := make(PerformanceData)

	//valid perfdata point
	err := perfData.add(NewPerformanceDataPoint("label", 10, ""))
	if err != nil {
		t.Error("adding a valid performance data point resulted in an error")
		return
	}

	if _, ok := perfData["label"]; !ok {
		t.Error("performance data point was not added to the map of performance data points")
	}

	err = perfData.add(NewPerformanceDataPoint("label", 10, ""))
	if err == nil {
		t.Error("there was no error when adding a performance data point with a label, that already exists in performance data")
	}
}

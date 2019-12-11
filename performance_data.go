/* Copyright (c) 2019, inexio GmbH, BSD 2-Clause License */
package monitoringplugin

import (
	"fmt"
	"github.com/pkg/errors"
	"regexp"
)

/*
performanceData is a map where all performanceDataPoints are stored. It assigns labels (string) to performanceDataPoints.
*/
type performanceData map[string]performanceDataPoint

/*
performanceDataPoint contains all information of one performanceDataPoint.
*/
type performanceDataPoint struct {
	label string
	value float64
	unit  string
	warn  float64 //currently we do not support ranges for warning and critical thresholds, because icinga2 does not support it
	crit  float64
	min   float64
	max   float64

	hasWarn bool
	hasCrit bool
	hasMin  bool
	hasMax  bool
}

/*
Adds a performanceDataPoint to the performanceData Map.
The function checks if a performanceDataPoint is valid and if there is already another performanceDataPoint with the same label in the performanceData map.
Usage:
	err := performanceData.Add(NewPerformanceDataPoint("temperature", 32, "°C").SetWarn(35).SetCrit(40))
	if err != nil {
		...
	}
*/
func (p *performanceData) Add(point *performanceDataPoint) error {
	if err := point.validate(); err != nil {
		return errors.Wrap(err, "given performance data point is not valid")
	}
	if _, ok := (*p)[point.label]; ok {
		return errors.New("a performance data point with this label does already exist")
	}
	(*p)[point.label] = *point
	return nil
}

/*
Validates a performanceDataPoint.
This function is used to check if a performanceDataPoint is compatible with the documentation from 'http://nagios-plugins.org/doc/guidelines.html'(valid name and unit, value is inside the range of min and max etc.)
*/
func (p *performanceDataPoint) validate() error {
	if p.label == "" {
		return errors.New("data point label cannot be an empty string")
	}

	match, err := regexp.MatchString("([='])", p.label)
	if err != nil {
		return errors.Wrap(err, "error during regex match")
	}
	if match {
		return errors.New("label can not contain the equal sign or single quote (')")
	}

	match, err = regexp.MatchString("([0-9;'\"])", p.unit)
	if err != nil {
		return errors.Wrap(err, "error during regex match")
	}
	if match {
		return errors.New("unit can not contain numbers, semicolon or quotes")
	}

	if (p.hasMin && p.hasMax) && (p.min > p.max) {
		return errors.New("min cannot be larger than max")
	}
	if p.hasMin && p.value < p.min {
		return errors.New("value cannot be smaller than min")
	}
	if p.hasMax && p.value > p.max {
		return errors.New("value cannot be larger than max")
	}
	return nil
}

/*
This function creates a new performanceDataPoint. Label and value are mandatory but are not checked at this point, the performanceDatePoint's validation is checked later when it is added to the performanceData list in the function performanceData.Add(*performanceDataPoint).
It is possible to directly add warning, critical, min and max values with the functions SetWarn(int), SetCrit(int), SetMin(int) and SetMax(int).
Usage:
	performanceDataPoint := NewPerformanceDataPoint("memory_usage", 55, "%").SetWarn(80).SetCrit(90)
*/
func NewPerformanceDataPoint(label string, value float64, unit string) *performanceDataPoint {
	return &performanceDataPoint{
		label: label,
		value: value,
		unit:  unit,
	}
}

/*
This function returns the performanceDataPoint in the specified string format that will be returned by the check plugin.
*/
func (p *performanceDataPoint) outputString() string {
	var outputString string
	outputString += "'" + p.label + "'=" + fmt.Sprintf("%g", p.value) + p.unit + ";"
	if p.hasWarn {
		outputString += fmt.Sprintf("%g", p.warn)
	}
	outputString += ";"
	if p.hasCrit {
		outputString += fmt.Sprintf("%g", p.crit)
	}
	outputString += ";"
	if p.hasMin {
		outputString += fmt.Sprintf("%g", p.min)
	}
	outputString += ";"
	if p.hasMax {
		outputString += fmt.Sprintf("%g", p.max)
	}

	return outputString
}

/*
Set Min Value.
*/
func (p *performanceDataPoint) SetMin(min float64) *performanceDataPoint {
	p.hasMin = true
	p.min = min
	return p
}

/*
Set Max Value.
*/
func (p *performanceDataPoint) SetMax(max float64) *performanceDataPoint {
	p.hasMax = true
	p.max = max
	return p
}

/*
Set Warn Value.
*/
func (p *performanceDataPoint) SetWarn(warn float64) *performanceDataPoint {
	p.hasWarn = true
	p.warn = warn
	return p
}

/*
Set Crit Value.
*/
func (p *performanceDataPoint) SetCrit(crit float64) *performanceDataPoint {
	p.hasCrit = true
	p.crit = crit
	return p
}

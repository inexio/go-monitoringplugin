/* Copyright (c) 2020, inexio GmbH, BSD 2-Clause License */

package monitoringplugin

import (
	"bytes"
	"fmt"
	"github.com/pkg/errors"
	"math/big"
	"regexp"
)

type performanceDataPointKey struct {
	metric string
	label  string
}

/*
PerformanceData is a map where all performanceDataPoints are stored. It assigns labels (string) to performanceDataPoints.
*/
type PerformanceData map[performanceDataPointKey]PerformanceDataPoint

/*
PerformanceDataPoint contains all information of one PerformanceDataPoint.
*/
type PerformanceDataPoint struct {
	metric string
	value  interface{}
	unit   string
	warn   interface{} //currently we do not support ranges for warning and critical thresholds, because icinga2 does not support it
	crit   interface{}
	min    interface{}
	max    interface{}

	label string

	hasWarn bool
	hasCrit bool
	hasMin  bool
	hasMax  bool
}

/*
add adds a PerformanceDataPoint to the PerformanceData Map.
The function checks if a PerformanceDataPoint is valid and if there is already another PerformanceDataPoint with the same metric in the PerformanceData map.
Usage:
	err := PerformanceData.add(NewPerformanceDataPoint("temperature", 32, "°C").SetWarn(35).SetCrit(40))
	if err != nil {
		...
	}
*/
func (p *PerformanceData) add(point *PerformanceDataPoint) error {
	if err := point.validate(); err != nil {
		return errors.Wrap(err, "given performance data point is not valid")
	}
	key := performanceDataPointKey{point.metric, point.label}
	if _, ok := (*p)[key]; ok {
		return errors.New("a performance data point with this metric does already exist")
	}
	(*p)[key] = *point
	return nil
}

/*
Validates a PerformanceDataPoint.
This function is used to check if a PerformanceDataPoint is compatible with the documentation from 'http://nagios-plugins.org/doc/guidelines.html'(valid name and unit, value is inside the range of min and max etc.)
*/
func (p *PerformanceDataPoint) validate() error {
	if p.metric == "" {
		return errors.New("data point metric cannot be an empty string")
	}

	match, err := regexp.MatchString("([='|;])", p.metric)
	if err != nil {
		return errors.Wrap(err, "error during regex match")
	}
	if match {
		return errors.New("metric contains invalid character")
	}

	match, err = regexp.MatchString("([='|;])", p.label)
	if err != nil {
		return errors.Wrap(err, "error during regex match")
	}
	if match {
		return errors.New("metric contains invalid character")
	}

	match, err = regexp.MatchString("([0-9;'\"])", p.unit)
	if err != nil {
		return errors.Wrap(err, "error during regex match")
	}
	if match {
		return errors.New("unit can not contain numbers, semicolon or quotes")
	}

	var min, max, value big.Float
	_, _, err = value.Parse(fmt.Sprint(p.value), 10)
	if err != nil {
		return errors.Wrap(err, "can't parse value")
	}

	if p.hasMin {
		_, _, err = min.Parse(fmt.Sprint(p.min), 10)
		if err != nil {
			return errors.Wrap(err, "can't parse min")
		}
		switch min.Cmp(&value) {
		case 1:
			return errors.New("value cannot be smaller than min")
		default:
		}
	}
	if p.hasMax {
		_, _, err = max.Parse(fmt.Sprint(p.max), 10)
		if err != nil {
			return errors.Wrap(err, "can't parse max")
		}
		switch max.Cmp(&value) {
		case -1:
			return errors.New("value cannot be larger than max")
		default:
		}
	}
	if p.hasMin && p.hasMax {
		switch min.Cmp(&max) {
		case 1:
			return errors.New("min cannot be larger than max")
		default:
		}
	}
	return nil
}

/*
NewPerformanceDataPoint creates a new PerformanceDataPoint. Label and value are mandatory but are not checked at this point, the performanceDatePoint's validation is checked later when it is added to the PerformanceData list in the function PerformanceData.add(*PerformanceDataPoint).
It is possible to directly add warning, critical, min and max values with the functions SetWarn(int), SetCrit(int), SetMin(int) and SetMax(int).
Usage:
	PerformanceDataPoint := NewPerformanceDataPoint("memory_usage", 55, "%").SetWarn(80).SetCrit(90)
*/
func NewPerformanceDataPoint(label string, value interface{}, unit string) *PerformanceDataPoint {
	return &PerformanceDataPoint{
		metric: label,
		value:  value,
		unit:   unit,
		label:  "",
	}
}

/*
This function returns the PerformanceDataPoint in the specified format as a string.
*/
func (p *PerformanceDataPoint) outputString(jsonLabel bool) string {
	return string(p.output(jsonLabel))
}

/*
This function returns the PerformanceDataPoint in the specified format that will be returned by the check plugin.
*/
func (p *PerformanceDataPoint) output(jsonLabel bool) []byte {
	var buffer bytes.Buffer
	if jsonLabel {
		buffer.WriteString(`'{"metric":"`)
		buffer.WriteString(p.metric)
		buffer.WriteByte('"')
		if p.label != "" {
			buffer.WriteString(`,"metric":"`)
			buffer.WriteString(p.label)
			buffer.WriteByte('"')
		}
		buffer.WriteString("}'")
	} else {
		buffer.WriteByte('\'')
		buffer.WriteString(p.metric)
		if p.label != "" {
			buffer.WriteByte('_')
			buffer.WriteString(p.label)
		}
		buffer.WriteByte('\'')
	}
	buffer.WriteByte('=')
	buffer.WriteString(fmt.Sprint(p.value))
	buffer.WriteString(p.unit)
	buffer.WriteByte(';')
	if p.hasWarn {
		buffer.WriteString(fmt.Sprintf("%g", p.warn))
	}
	buffer.WriteByte(';')
	if p.hasCrit {
		buffer.WriteString(fmt.Sprintf("%g", p.crit))
	}
	buffer.WriteByte(';')
	if p.hasMin {
		buffer.WriteString(fmt.Sprintf("%g", p.min))
	}
	buffer.WriteByte(';')
	if p.hasMax {
		buffer.WriteString(fmt.Sprintf("%g", p.max))
	}

	return buffer.Bytes()
}

/*
SetMin sets minimum value.
*/
func (p *PerformanceDataPoint) SetMin(min float64) *PerformanceDataPoint {
	p.hasMin = true
	p.min = min
	return p
}

/*
SetMax sets maximum value.
*/
func (p *PerformanceDataPoint) SetMax(max float64) *PerformanceDataPoint {
	p.hasMax = true
	p.max = max
	return p
}

/*
SetWarn sets maximum value.
*/
func (p *PerformanceDataPoint) SetWarn(warn float64) *PerformanceDataPoint {
	p.hasWarn = true
	p.warn = warn
	return p
}

/*
SetCrit sets critical value.
*/
func (p *PerformanceDataPoint) SetCrit(crit float64) *PerformanceDataPoint {
	p.hasCrit = true
	p.crit = crit
	return p
}

/*
SetLabel adds a tag to the performance data point
If one tag is added more than once, the value before will be overwritten
*/
func (p *PerformanceDataPoint) SetLabel(label string) *PerformanceDataPoint {
	p.label = label
	return p
}

/*
PerformanceDataPointInfo has all information to one performance data point as exported variables. It is returned by
PerformanceDataPoint.GetInfo()
*/
type PerformanceDataPointInfo struct {
	Metric string `yaml:"metric" json:"metric" xml:"metric"`
	Label  string `yaml:"label" json:"label" xml:"label"`

	Value interface{} `yaml:"value" json:"value" xml:"value"`
	Unit  string      `yaml:"unit" json:"unit" xml:"unit"`
	Warn  interface{} `yaml:"warn" json:"warn" xml:"warn"`
	Crit  interface{} `yaml:"crit" json:"crit" xml:"crit"`
	Min   interface{} `yaml:"min" json:"min" xml:"min"`
	Max   interface{} `yaml:"max" json:"max" xml:"max"`
}

/*
GetInfo returns all information for a performance data point.
*/
func (p PerformanceDataPoint) GetInfo() PerformanceDataPointInfo {
	return PerformanceDataPointInfo{
		Metric: p.metric,
		Label:  p.label,
		Value:  p.value,
		Unit:   p.unit,
		Warn:   p.warn,
		Crit:   p.crit,
		Min:    p.min,
		Max:    p.max,
	}
}

/*
GetInfo returns all information for performance data.
*/
func (p PerformanceData) GetInfo() []PerformanceDataPointInfo {
	var info []PerformanceDataPointInfo
	for _, pd := range p {
		info = append(info, pd.GetInfo())
	}
	return info
}

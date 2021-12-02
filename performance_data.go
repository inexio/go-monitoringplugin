package monitoringplugin

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/pkg/errors"
	"math/big"
	"regexp"
	"strconv"
)

type performanceDataPointKey struct {
	Metric string `json:"metric"`
	Label  string `json:"label,omitempty"`
}

// performanceData is a map where all performanceDataPoints are stored.
// It assigns labels (string) to performanceDataPoints.
type performanceData map[performanceDataPointKey]PerformanceDataPoint

/*
add adds a PerformanceDataPoint to the performanceData Map.
The function checks if a PerformanceDataPoint is valid and if there is already another PerformanceDataPoint with the
same metric in the performanceData map.
Usage:
	err := performanceData.add(NewPerformanceDataPoint("temperature", 32, "°C").SetWarn(35).SetCrit(40))
	if err != nil {
		...
	}
*/
func (p *performanceData) add(point *PerformanceDataPoint) error {
	if err := point.Validate(); err != nil {
		return errors.Wrap(err, "given performance data point is not valid")
	}
	key := performanceDataPointKey{point.Metric, point.Label}
	if _, ok := (*p)[key]; ok {
		return fmt.Errorf("a performance data point with the metric '%s' does already exist", func(key performanceDataPointKey) string {
			res := key.Metric
			if key.Label != "" {
				res += " and label " + key.Label
			}
			return res
		}(key))
	}
	(*p)[key] = *point
	return nil
}

// getInfo returns all information for performance data.
func (p performanceData) getInfo() []PerformanceDataPoint {
	var info []PerformanceDataPoint
	for _, pd := range p {
		info = append(info, pd)
	}
	return info
}

// PerformanceDataPoint contains all information of one PerformanceDataPoint.
type PerformanceDataPoint struct {
	Metric     string      `json:"metric" xml:"metric"`
	Label      string      `json:"label" xml:"label"`
	Value      interface{} `json:"value" xml:"value"`
	Unit       string      `json:"unit" xml:"unit"`
	Thresholds Thresholds  `json:"thresholds" xml:"thresholds"`
	Min        interface{} `json:"min" xml:"min"`
	Max        interface{} `json:"max" xml:"max"`
}

/*
Validate validates a PerformanceDataPoint.
This function is used to check if a PerformanceDataPoint is compatible with the documentation from
'http://nagios-plugins.org/doc/guidelines.html'(valid name and unit, value is inside the range of min and max etc.)
*/
func (p *PerformanceDataPoint) Validate() error {
	if p.Metric == "" {
		return errors.New("data point metric cannot be an empty string")
	}

	match, err := regexp.MatchString("([='])", p.Metric)
	if err != nil {
		return errors.Wrap(err, "error during regex match")
	}
	if match {
		return errors.New("metric contains invalid character")
	}

	match, err = regexp.MatchString("([='])", p.Label)
	if err != nil {
		return errors.Wrap(err, "error during regex match")
	}
	if match {
		return errors.New("metric contains invalid character")
	}

	match, err = regexp.MatchString("([0-9;'\"])", p.Unit)
	if err != nil {
		return errors.Wrap(err, "error during regex match")
	}
	if match {
		return errors.New("unit can not contain numbers, semicolon or quotes")
	}

	var min, max, value big.Float
	_, _, err = value.Parse(fmt.Sprint(p.Value), 10)
	if err != nil {
		return errors.Wrap(err, "can't parse value")
	}

	if p.Min != nil {
		_, _, err = min.Parse(fmt.Sprint(p.Min), 10)
		if err != nil {
			return errors.Wrap(err, "can't parse min")
		}
		switch min.Cmp(&value) {
		case 1:
			return errors.New("value cannot be smaller than min")
		default:
		}
	}
	if p.Max != nil {
		_, _, err = max.Parse(fmt.Sprint(p.Max), 10)
		if err != nil {
			return errors.Wrap(err, "can't parse max")
		}
		switch max.Cmp(&value) {
		case -1:
			return errors.New("value cannot be larger than max")
		default:
		}
	}
	if p.Min != nil && p.Max != nil {
		switch min.Cmp(&max) {
		case 1:
			return errors.New("min cannot be larger than max")
		default:
		}
	}

	if !p.Thresholds.IsEmpty() {
		err = p.Thresholds.Validate()
		if err != nil {
			return errors.Wrap(err, "thresholds are invalid")
		}
	}

	return nil
}

/*
NewPerformanceDataPoint creates a new PerformanceDataPoint. Metric and value are mandatory but are not checked at this
point, the performanceDatePoint's validation is checked later when it is added to the performanceData list in the
function performanceData.add(*PerformanceDataPoint).
It is possible to directly add thresholds, min and max values with the functions SetThresholds(Thresholds),
SetMin(int) and SetMax(int).
Usage:
	PerformanceDataPoint := NewPerformanceDataPoint("memory_usage", 55).SetUnit("%")
*/
func NewPerformanceDataPoint(metric string, value interface{}) *PerformanceDataPoint {
	return &PerformanceDataPoint{
		Metric: metric,
		Value:  value,
	}
}

// SetUnit sets the unit of the performance data point
func (p *PerformanceDataPoint) SetUnit(unit string) *PerformanceDataPoint {
	p.Unit = unit
	return p
}

// SetMin sets minimum value.
func (p *PerformanceDataPoint) SetMin(min interface{}) *PerformanceDataPoint {
	p.Min = min
	return p
}

// SetMax sets maximum value.
func (p *PerformanceDataPoint) SetMax(max interface{}) *PerformanceDataPoint {
	p.Max = max
	return p
}

// SetLabel adds a tag to the performance data point
// If one tag is added more than once, the value before will be overwritten
func (p *PerformanceDataPoint) SetLabel(label string) *PerformanceDataPoint {
	p.Label = label
	return p
}

// SetThresholds sets the thresholds for the performance data point
func (p *PerformanceDataPoint) SetThresholds(thresholds Thresholds) *PerformanceDataPoint {
	p.Thresholds = thresholds
	return p
}

// This function returns the PerformanceDataPoint in the specified format that will be returned by the check plugin.
func (p *PerformanceDataPoint) output(jsonLabel bool) []byte {
	var buffer bytes.Buffer
	if jsonLabel {
		buffer.WriteByte('\'')
		key := performanceDataPointKey{
			Metric: p.Metric,
			Label:  p.Label,
		}
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

	switch p.Value.(type) {
	case float64:
		buffer.WriteString(strconv.FormatFloat(p.Value.(float64), 'f', -1, 64))
	default:
		buffer.WriteString(fmt.Sprint(p.Value))
	}

	buffer.WriteString(p.Unit)

	if !p.Thresholds.IsEmpty() || p.Max != nil || p.Min != nil {
		buffer.WriteByte(';')
		if p.Thresholds.HasWarning() {
			buffer.WriteString(p.Thresholds.getWarning())
		}
		buffer.WriteByte(';')
		if p.Thresholds.HasCritical() {
			buffer.WriteString(p.Thresholds.getCritical())
		}
		buffer.WriteByte(';')
		if p.Min != nil {
			switch min := p.Min.(type) {
			case float64:
				buffer.WriteString(strconv.FormatFloat(min, 'f', -1, 64))
			default:
				buffer.WriteString(fmt.Sprint(min))
			}
		}
		buffer.WriteByte(';')
		if p.Max != nil {
			switch max := p.Max.(type) {
			case float64:
				buffer.WriteString(strconv.FormatFloat(max, 'f', -1, 64))
			default:
				buffer.WriteString(fmt.Sprint(max))
			}
		}
	}

	return buffer.Bytes()
}

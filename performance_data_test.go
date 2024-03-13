package monitoringplugin

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPerformanceDataPointCreation(t *testing.T) {
	const metric = "testMetric"
	const value = float64(10)
	p := NewPerformanceDataPoint(metric, value)
	assert.Implements(t, (*anyDataPoint)(nil), p)

	require.Equal(t, metric, p.Metric)
	//nolint:testifylint // float-compare safe here
	require.Equal(t, value, p.Value)

	const unit = "%"
	p.SetUnit(unit)
	require.Equal(t, unit, p.Unit, "SetUnit failed")

	const label = "testLabel"
	p.SetLabel(label)
	require.Equal(t, label, p.Label, "SetLabel failed")

	const min = float64(0)
	p.SetMin(min)
	//nolint:testifylint // float-compare safe here
	require.Equal(t, min, p.Min, "SetMin failed")
	require.True(t, p.hasMin, "SetMin failed")

	const max = float64(100)
	p.SetMax(max)
	//nolint:testifylint // float-compare safe here
	require.Equal(t, max, p.Max, "SetMax failed")
	require.True(t, p.hasMax, "SetMax failed")

	thresholds := NewThresholds[float64](0, 10, 0, 20)
	p.SetThresholds(thresholds)
	require.Equal(t, thresholds, p.Thresholds, "SetThresholds failed")
}

func TestPerformanceDataPoint_validate(t *testing.T) {
	p := NewPerformanceDataPoint("metric", 10).SetMin(0).SetMax(100)
	require.NoError(t, p.Validate(),
		"valid performance data point resulted in an error")

	// empty metric
	p = NewPerformanceDataPoint("", 10)
	require.Error(t, p.Validate(),
		"invalid performance data did not return an error (case: empty metric)")

	// invalid metric
	p = NewPerformanceDataPoint("metric=", 10)
	require.Error(t, p.Validate(),
		"invalid performance data did not return an error (case: invalid metric, contains =)")

	p = NewPerformanceDataPoint("'metric'", 10)
	require.Error(t, p.Validate(),
		"invalid performance data did not return an error (case: invalid metric, contains single quotes)")

	// invalid unit
	p = NewPerformanceDataPoint("metric", 10).SetUnit("unit1")
	require.Error(t, p.Validate(),
		"invalid performance data did not return an error (case: invalid unit, contains numbers)")

	p = NewPerformanceDataPoint("metric", 10).SetUnit("unit;")
	require.Error(t, p.Validate(),
		"invalid performance data did not return an error (case: invalid unit, contains semicolon)")

	p = NewPerformanceDataPoint("metric", 10).SetUnit("unit'")
	require.Error(t, p.Validate(),
		"invalid performance data did not return an error (case: invalid unit, contains single quotes)")

	p = NewPerformanceDataPoint("metric", 10).SetUnit("unit\"")
	require.Error(t, p.Validate(),
		"invalid performance data did not return an error (case: invalid unit, contains double quotes)")

	// value < min
	p = NewPerformanceDataPoint("metric", 10).SetMin(50)
	require.Error(t, p.Validate(),
		"invalid performance data did not return an error (case: value < min)")

	// value > max
	p = NewPerformanceDataPoint("metric", 10).SetMax(5)
	require.Error(t, p.Validate(),
		"invalid performance data did not return an error (case: value < min)")

	// min > max
	p = NewPerformanceDataPoint("metric", 10).SetMin(10).SetMax(5)
	require.Error(t, p.Validate(),
		"invalid performance data did not return an error (case: max < min)")
}

func TestPerformanceDataPoint_output(t *testing.T) {
	const label = "metric"
	const value = float64(10.0)
	const unit = "s"
	const warn = float64(40.0)
	const crit = float64(50.0)
	const min = float64(0.0)
	const max = float64(60.0)

	p := NewPerformanceDataPoint(label, value)
	regex := fmt.Sprintf("'%s'=%g", label, value)
	require.Contains(t, string(p.output(false)), regex,
		"output string did not match regex")

	p.SetUnit(unit)
	regex = fmt.Sprintf("'%s'=%g%s", label, value, unit)
	require.Contains(t, string(p.output(false)), regex,
		"output string did not match regex")

	p.SetMax(max)
	regex = fmt.Sprintf("'%s'=%g%s;;;;%g", label, value, unit, max)
	require.Contains(t, string(p.output(false)), regex,
		"output string did not match regex")

	th := NewThresholds(0, warn, 0, crit)
	th.UseWarning(false, true).UseCritical(false, true)
	p.SetThresholds(th)
	regex = fmt.Sprintf("'%s'=%g%s;~:%g;~:%g;;%g", label, value, unit, warn, crit,
		max)
	require.Contains(t, string(p.output(false)), regex,
		"output string did not match regex")

	th = NewThresholds[float64](0, 0, -10, 0)
	th.UseWarning(true, false).UseCritical(true, false)
	p.SetThresholds(th)
	regex = fmt.Sprintf("'%s'=%g%s;%d:;%d:;;%g", label, value, unit, 0, -10, max)
	require.Contains(t, string(p.output(false)), regex,
		"output string did not match regex")

	th = NewThresholds[float64](5, 10, 3, 11)
	p.SetThresholds(th)
	regex = fmt.Sprintf("'%s'=%g%s;%d:%d;%d:%d;;%g", label, value, unit, 5, 10, 3,
		11, max)
	require.Contains(t, string(p.output(false)), regex,
		"output string did not match regex")

	th = NewThresholds(0, warn, 0, crit)
	p.SetThresholds(th)
	regex = fmt.Sprintf("'%s'=%g%s;%g;%g;;%g", label, value, unit, warn, crit,
		max)
	require.Contains(t, string(p.output(false)), regex,
		"output string did not match regex")

	p.SetMin(min)
	regex = fmt.Sprintf("'%s'=%g%s;%g;%g;%g;%g", label, value, unit, warn, crit,
		min, max)
	require.Contains(t, string(p.output(false)), regex,
		"output string did not match regex")

	regex = fmt.Sprintf(`'{"metric":"%s"}'=%g%s;%g;%g;%g;%g`, label, value, unit,
		warn, crit, min, max)
	require.Contains(t, string(p.output(true)), regex,
		"output string did not match regex")

	tag := "tag"
	p.SetLabel(tag)
	regex = fmt.Sprintf(`'{"metric":"%s","label":"%s"}'=%g%s;%g;%g;%g;%g`,
		label, tag, value, unit, warn, crit, min, max)
	require.Contains(t, string(p.output(true)), regex,
		"output string did not match regex")

	regex = fmt.Sprintf(`'%s_%s'=%g%s;%g;%g;%g;%g`, label, tag, value, unit, warn,
		crit, min, max)
	require.Contains(t, string(p.output(false)), regex,
		"output string did not match regex")
}

func TestPerformanceData_add(t *testing.T) {
	perfData := newPerformanceData()

	// valid perfdata point
	require.NoError(t, perfData.add(NewPerformanceDataPoint("metric", 10)),
		"adding a valid performance data point resulted in an error")

	key := newPerformanceDataPointKey("metric", "")
	require.Contains(t, perfData, key,
		"performance data point was not added to the map of performance data points")

	require.Error(t, perfData.add(NewPerformanceDataPoint("metric", 10)),
		"there was no error when adding a performance data point with a metric, that already exists in performance data")

	require.NoError(t,
		perfData.add(NewPerformanceDataPoint("metric", 10).SetLabel("tag1")),
		"adding a valid performance data point resulted in an error")

	require.NoError(t,
		perfData.add(NewPerformanceDataPoint("metric", 10).SetLabel("tag2")),
		"adding a valid performance data point resulted in an error")

	require.Error(t,
		perfData.add(NewPerformanceDataPoint("metric", 10).SetLabel("tag1")),
		"there was no error when adding a performance data point with a metric and tag, that already exists in performance data")
}

func TestResponse_SetPerformanceDataJsonLabel(t *testing.T) {
	perfData := newPerformanceData()

	// valid perfdata point
	require.NoError(t, perfData.add(NewPerformanceDataPoint("metric", 10)),
		"adding a valid performance data point resulted in an error")

	key := newPerformanceDataPointKey("metric", "")
	require.Contains(t, perfData, key,
		"performance data point was not added to the map of performance data points")

	require.Error(t, perfData.add(NewPerformanceDataPoint("metric", 10)),
		"there was no error when adding a performance data point with a metric, that already exists in performance data")
}

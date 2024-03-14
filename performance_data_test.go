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

	assert.False(t, p.HasThresholds())
	th := p.NewThresholds(0, 10, 0, 20)
	assert.Same(t, &p.Thresholds, th, "NewThresholds failed")
	assert.True(t, p.HasThresholds())

	p.SetThresholds(*th)
	require.Equal(t, *th, p.Thresholds, "SetThresholds failed")
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

	p.NewThresholds(0, warn, 0, crit).
		UseWarning(false, true).UseCritical(false, true)
	regex = fmt.Sprintf("'%s'=%g%s;~:%g;~:%g;;%g", label, value, unit, warn, crit,
		max)
	require.Contains(t, string(p.output(false)), regex,
		"output string did not match regex")

	p.NewThresholds(0, 0, -10, 0).
		UseWarning(true, false).UseCritical(true, false)
	regex = fmt.Sprintf("'%s'=%g%s;%d:;%d:;;%g", label, value, unit, 0, -10, max)
	require.Contains(t, string(p.output(false)), regex,
		"output string did not match regex")

	p.NewThresholds(5, 10, 3, 11)
	regex = fmt.Sprintf("'%s'=%g%s;%d:%d;%d:%d;;%g", label, value, unit, 5, 10, 3,
		11, max)
	require.Contains(t, string(p.output(false)), regex,
		"output string did not match regex")

	p.NewThresholds(0, warn, 0, crit)
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

	key := newPerformanceDataPointKey("metric", "")
	assert.Nil(t, perfData.point(key))

	// valid perfdata point
	require.NoError(t, perfData.add(NewPerformanceDataPoint("metric", 10)),
		"adding a valid performance data point resulted in an error")

	point := perfData.point(key)
	require.NotNil(t, point,
		"performance data point was not added to the map of performance data points")
	assert.Equal(t, key, point.key())

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
	point := perfData.point(key)
	require.Equal(t, key, point.key(),
		"performance data point was not added to the map of performance data points")

	require.Error(t, perfData.add(NewPerformanceDataPoint("metric", 10)),
		"there was no error when adding a performance data point with a metric, that already exists in performance data")
}

func TestPerformanceData_keepOrder(t *testing.T) {
	pointKeys := [...]performanceDataPointKey{
		{"metric", ""},
		{"metric", "tag1"},
		{"metric", "tag2"},
	}

	perfData := newPerformanceData()
	wantKeys := make([]performanceDataPointKey, 0, len(pointKeys))
	for i := range pointKeys {
		key := &pointKeys[i]
		require.NoError(t, perfData.add(
			NewPerformanceDataPoint(key.Metric, 10).SetLabel(key.Label)))
		wantKeys = append(wantKeys, newPerformanceDataPointKey(
			key.Metric, key.Label))
	}

	gotKeys := make([]performanceDataPointKey, 0, len(pointKeys))
	for _, p := range perfData.getInfo() {
		gotKeys = append(gotKeys, p.key())
	}
	assert.Equal(t, wantKeys, gotKeys, "wrong order of data points")
}

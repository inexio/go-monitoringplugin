package monitoringplugin

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestValidateThresholds(t *testing.T) {
	th1 := NewThresholds(5, 10, 3, 12)
	require.NoError(t, th1.Validate())

	th2 := NewThresholds(0, 10, 0, 12)
	require.NoError(t, th2.Validate())

	th3 := Thresholds[int]{}
	require.NoError(t, th3.Validate())

	th4 := NewThresholds(0, 3, 0, 0)
	require.NoError(t,
		th4.UseWarning(false, true).UseCritical(false, false).Validate())

	th5 := NewThresholds(2, 1, 0, 0)
	require.Error(t, th5.UseCritical(false, false).Validate())

	th6 := NewThresholds(0, 0, 2, 1)
	require.Error(t, th6.UseWarning(false, false).Validate())

	th7 := NewThresholds(1, 0, 2, 0)
	require.Error(t,
		th7.UseWarning(true, false).UseCritical(true, false).Validate())

	th8 := NewThresholds(0, 2, 0, 1)
	require.Error(t,
		th8.UseWarning(false, true).UseCritical(false, true).Validate())
}

func TestCheckThresholds(t *testing.T) {
	th1 := NewThresholds(5, 10, 3, 12)
	assert.Equal(t, OK, th1.CheckValue(6))
	assert.Equal(t, OK, th1.CheckValue(5))
	assert.Equal(t, OK, th1.CheckValue(10))
	assert.Equal(t, WARNING, th1.CheckValue(4))
	assert.Equal(t, WARNING, th1.CheckValue(11))
	assert.Equal(t, WARNING, th1.CheckValue(3))
	assert.Equal(t, WARNING, th1.CheckValue(12))
	assert.Equal(t, CRITICAL, th1.CheckValue(2))
	assert.Equal(t, CRITICAL, th1.CheckValue(13))

	th2 := NewThresholds(5, 10, 5, 12)
	assert.Equal(t, CRITICAL, th2.CheckValue(4))
}

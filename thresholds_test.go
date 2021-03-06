package monitoringplugin

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestValidateThresholds(t *testing.T) {
	th1 := Thresholds{
		WarningMin:  5,
		WarningMax:  10,
		CriticalMin: 3,
		CriticalMax: 12,
	}
	assert.NoError(t, th1.Validate())

	th2 := Thresholds{
		WarningMin:  0,
		WarningMax:  10,
		CriticalMin: 0,
		CriticalMax: 12,
	}
	assert.NoError(t, th2.Validate())

	th3 := Thresholds{}
	assert.NoError(t, th3.Validate())

	th4 := Thresholds{
		WarningMax: 3,
	}
	assert.NoError(t, th4.Validate())

	th5 := Thresholds{
		WarningMin: 2,
		WarningMax: 1,
	}
	assert.Error(t, th5.Validate())

	th6 := Thresholds{
		CriticalMin: 2,
		CriticalMax: 1,
	}
	assert.Error(t, th6.Validate())

	th7 := Thresholds{
		WarningMin:  1,
		CriticalMin: 2,
	}
	assert.Error(t, th7.Validate())

	th8 := Thresholds{
		WarningMax:  2,
		CriticalMax: 1,
	}
	assert.Error(t, th8.Validate())
}

func TestCheckThresholds(t *testing.T) {
	th1 := Thresholds{
		WarningMin:  5,
		WarningMax:  10,
		CriticalMin: 3,
		CriticalMax: 12,
	}

	res, err := th1.CheckValue(6)
	assert.NoError(t, err)
	assert.Equal(t, OK, res)

	res, err = th1.CheckValue(5)
	assert.NoError(t, err)
	assert.Equal(t, OK, res)

	res, err = th1.CheckValue(10)
	assert.NoError(t, err)
	assert.Equal(t, OK, res)

	res, err = th1.CheckValue(4)
	assert.NoError(t, err)
	assert.Equal(t, WARNING, res)

	res, err = th1.CheckValue(11)
	assert.NoError(t, err)
	assert.Equal(t, WARNING, res)

	res, err = th1.CheckValue(3)
	assert.NoError(t, err)
	assert.Equal(t, WARNING, res)

	res, err = th1.CheckValue(12)
	assert.NoError(t, err)
	assert.Equal(t, WARNING, res)

	res, err = th1.CheckValue(2)
	assert.NoError(t, err)
	assert.Equal(t, CRITICAL, res)

	res, err = th1.CheckValue(13)
	assert.NoError(t, err)
	assert.Equal(t, CRITICAL, res)

	th2 := Thresholds{
		WarningMin:  5,
		WarningMax:  10,
		CriticalMin: 5,
		CriticalMax: 12,
	}

	res, err = th2.CheckValue(4)
	assert.NoError(t, err)
	assert.Equal(t, CRITICAL, res)
}

package monitoringplugin

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestValidateThresholds(t *testing.T) {
	th1 := CheckThresholds{
		WarningMin:  5,
		WarningMax:  10,
		CriticalMin: 3,
		CriticalMax: 12,
	}
	assert.NoError(t, th1.Validate())

	th2 := CheckThresholds{}
	assert.NoError(t, th2.Validate())

	th3 := CheckThresholds{
		WarningMax: 3,
	}
	assert.NoError(t, th3.Validate())

	th4 := CheckThresholds{
		WarningMin: 2,
		WarningMax: 1,
	}
	assert.Error(t, th4.Validate())

	th5 := CheckThresholds{
		CriticalMin: 2,
		CriticalMax: 1,
	}
	assert.Error(t, th5.Validate())

	th6 := CheckThresholds{
		WarningMin:  1,
		CriticalMin: 2,
	}
	assert.Error(t, th6.Validate())

	th7 := CheckThresholds{
		WarningMax:  2,
		CriticalMax: 1,
	}
	assert.Error(t, th7.Validate())
}

func TestCheckThresholds(t *testing.T) {
	th1 := CheckThresholds{
		WarningMin:  5,
		WarningMax:  10,
		CriticalMin: 3,
		CriticalMax: 12,
	}

	res, err := th1.CheckValue(6)
	assert.NoError(t, err)
	assert.Equal(t, OK, res)

	res, err = th1.CheckValue(4)
	assert.NoError(t, err)
	assert.Equal(t, WARNING, res)

	res, err = th1.CheckValue(11)
	assert.NoError(t, err)
	assert.Equal(t, WARNING, res)

	res, err = th1.CheckValue(2)
	assert.NoError(t, err)
	assert.Equal(t, CRITICAL, res)

	res, err = th1.CheckValue(13)
	assert.NoError(t, err)
	assert.Equal(t, CRITICAL, res)
}

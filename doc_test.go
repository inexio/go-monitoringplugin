package monitoringplugin_test

import (
	"fmt"

	"github.com/inexio/go-monitoringplugin/v2"
)

func Example_basicUsage() {
	// Creating response with a default ok message, that will be displayed when
	// the checks exits with status ok.
	response := monitoringplugin.NewResponse("everything checked!")

	// Set output delimiter (default is \n)
	// response.SetOutputDelimiter(" / ")

	// Updating check plugin status and adding message to the output (status only
	// changes if the new status is worse than the current one).

	// check status stays ok
	response.UpdateStatus(monitoringplugin.OK, "something is ok!")
	// check status updates to critical
	response.UpdateStatus(monitoringplugin.CRITICAL,
		"something else is critical!")
	// check status stays critical, but message will be added to the output
	response.UpdateStatus(monitoringplugin.WARNING, "something else is warning!")

	// adding performance data
	p1 := monitoringplugin.NewPerformanceDataPoint("response_time", 10).
		SetUnit("s").SetMin(0)
	p1.NewThresholds(0, 10, 0, 20)
	if err := response.AddPerformanceDataPoint(p1); err != nil {
		// error handling
	}

	p2 := monitoringplugin.NewPerformanceDataPoint("memory_usage", 50.6).
		SetUnit("%").SetMin(0).SetMax(100)
	p2.NewThresholds(0, 80, 0, 90)
	if err := response.AddPerformanceDataPoint(p2); err != nil {
		// error handling
	}

	fmt.Println(response.GetInfo().RawOutput)
	// Output:
	// CRITICAL: something else is critical!
	// something else is warning!
	// something is ok! | 'response_time'=10s;10;20;0; 'memory_usage'=50.6%;80;90;0;100
}

# go-monitoringplugin

[![Go Report Card](https://goreportcard.com/badge/github.com/inexio/go-monitoringplugin)](https://goreportcard.com/report/github.com/inexio/go-monitoringplugin)
[![GitHub license](https://img.shields.io/badge/license-BSD-blue.svg)](https://github.com/inexio/go-monitoringplugin/blob/master/LICENSE)
[![GoDoc doc](https://img.shields.io/badge/godoc-reference-blue)](https://godoc.org/github.com/inexio/go-monitoringplugin)

## Description

Golang package for writing monitoring check plugins for
[nagios](https://www.nagios.org/), [icinga2](https://icinga.com/),
[zabbix](https://www.zabbix.com/), [checkmk](https://checkmk.com/), etc.

The package complies with the [Monitoring Plugins Development Guidelines](https://www.monitoring-plugins.org/doc/guidelines.html).

## Example / Usage

``` go
package main

import "github.com/inexio/go-monitoringplugin/v2"

func main() {
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

	err = response.AddPerformanceDataPoint(
		monitoringplugin.NewPerformanceDataPoint("memory_usage", 50.6).
			SetUnit("%").SetMin(0).SetMax(100).
			SetThresholds(monitoringplugin.NewThresholds(0, 80.0, 0, 90.0)))
	if err != nil {
		// error handling
	}

	response.OutputAndExit()
	/* exits program with exit code 2 and outputs:
	CRITICAL: something is ok!
	something else is critical!
	something else is warning! | 'response_time'=10s;10;20;0; 'memory_usage'=50%;80;90;0;100
	*/
}
```

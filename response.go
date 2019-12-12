/* Copyright (c) 2019, inexio GmbH, BSD 2-Clause License */

//Package monitoringplugin provides for writing monitoring check plugins for nagios, icinga2, zabbix, etc
package monitoringplugin

import (
	"fmt"
	"os"
	"strings"
)

const (
	OK       = 0
	WARNING  = 1
	CRITICAL = 2
	UNKNOWN  = 3
)

/*
Response is the main type that is responsible for the check plugin response. It stores the current status code, output messages, performance data and the output message delimiter.
*/
type response struct {
	statusCode       int
	defaultOkMessage string
	outputMessages   []string
	performanceData  performanceData
	outputDelimiter  string
}

/*
NewResponse(string) creates a new response and sets the default OK message to the given string. The default OK message will be displayed together with the other output messages, but only if the status is still OK when the check exits.
*/
func NewResponse(defaultOkMessage string) *response {
	response := &response{
		statusCode:       OK,
		defaultOkMessage: defaultOkMessage,
		outputDelimiter:  "\n",
	}
	response.performanceData = make(performanceData)
	return response
}

/*
AddPerformanceDataPoints(*performanceDataPoint) adds a performanceDataPoint to the performanceData map, using performanceData.Add(*performanceDataPoint).
Usage:
	err := response.AddPerformanceDataPoint(NewPerformanceDataPoint("temperature", 32, "°C").SetWarn(35).SetCrit(40))
	if err != nil {
		...
	}
*/
func (r *response) AddPerformanceDataPoint(point *performanceDataPoint) error {
	return r.performanceData.Add(point)
}

/*
UpdateStatus(int, string) updates the exit status of the response and adds a statusMessage to the outputMessages that will be displayed when the check exits.
See updateStatusCode(int) for a detailed description of the algorithm that is used to update the status code.
*/
func (r *response) UpdateStatus(statusCode int, statusMessage string) {
	r.updateStatusCode(statusCode)
	if statusMessage != "" {
		r.outputMessages = append(r.outputMessages, statusMessage)
	}
}

/*
Returns the current status code.
*/
func (r *response) GetStatusCode() int {
	return r.statusCode
}

/*
This function updates the statusCode of the response. The status code is mapped to a state like this:
0 = OK
1 = WARNING
2 = CRITICAL
3 = UNKNOWN
Everything else is also mapped to UNKNOWN.

UpdateStatus uses the following algorithm to update the exit status:
CRITICAL > UNKNOWN > WARNING > OK
Everything "left" from the current status code is seen as worse than the current one. If the function wants to set a status code, it will only update it if the new status code is "left" of the current one.
Example:
	//current status code = 1
	response.updateStatusCode(0) //nothing changes
	response.updateStatusCode(2) //status code changes to CRITICAL (=2)

	//now current status code = 2
	response.updateStatusCode(3) //nothing changes, because CRITICAL is worse than UNKNOWN

*/
func (r *response) updateStatusCode(statusCode int) {
	if r.statusCode == CRITICAL { //critical is the worst status code; if its critical, do not change anything
		return
	}
	if statusCode == CRITICAL {
		r.statusCode = statusCode
		return
	}
	if statusCode < OK || statusCode > UNKNOWN {
		statusCode = UNKNOWN
	}
	if statusCode > r.statusCode {
		r.statusCode = statusCode
	}
}

/*
This function is used to set the delimiter that is used to separate the outputMessages that will be displayed when the check plugin exits. The default value is a linebreak (\n)
It can be set to any string.
Example:
	response.SetOutputDelimiter(" / ")
	//this results in the output having the following format:
	//OK: defaultOkMessage / outputMessage1 / outputMessage2 / outputMessage3 | performanceData
*/
func (r *response) SetOutputDelimiter(delimiter string) {
	r.outputDelimiter = delimiter
}

/*
Sets the outputDelimiter to "\n". (See Response.SetOutputDelimiter(string))
*/
func (r *response) OutputDelimiterMultiline() {
	r.SetOutputDelimiter("\n")
}

/*
This function is used to map the status code to a string.
*/
func statusCode2Text(statusCode int) string {
	switch {
	case statusCode == 0:
		return "OK"
	case statusCode == 1:
		return "WARNING"
	case statusCode == 2:
		return "CRITICAL"
	default:
		return "UNKNOWN"
	}
}

/*
This function returns the output string that will be returned by the check plugin.
*/
func (r *response) outputString() string {
	outputString := statusCode2Text(r.statusCode) + ": "
	if r.statusCode == OK {
		outputString += r.defaultOkMessage
		if len(r.outputMessages) > 0 {
			outputString += r.outputDelimiter
		}
	}
	outputString += strings.Join(r.outputMessages, r.outputDelimiter)
	return outputString
}

/*
This function generates the output string and prints it to stdout. After that the check plugin exits with the current exit code.
Example:
	response := NewResponse("everything checked!")
	defer response.OutputAndExit()

	//check plugin logic...
*/
func (r *response) OutputAndExit() {
	fmt.Print(r.outputString())

	firstPoint := true
	for _, perfDataPoint := range r.performanceData {
		if firstPoint {
			fmt.Print(" | ")
			firstPoint = false
		} else {
			fmt.Print(" ")
		}
		fmt.Print(perfDataPoint.outputString())
	}
	fmt.Println()

	os.Exit(r.statusCode)
}

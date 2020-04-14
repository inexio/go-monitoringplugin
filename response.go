/* Copyright (c) 2019, inexio GmbH, BSD 2-Clause License */

//Package monitoringplugin provides for writing monitoring check plugins for nagios, icinga2, zabbix, etc
package monitoringplugin

import (
	"fmt"
	"os"
	"strings"
)

const (
	// OK check plugin status = OK
	OK = 0
	// WARNING check plugin status = WARNING
	WARNING = 1
	// CRITICAL check plugin status = CRITICAL
	CRITICAL = 2
	// UNKNOWN check plugin status = UNKNOWN
	UNKNOWN = 3
)

/*
Response is the main type that is responsible for the check plugin Response. It stores the current status code, output messages, performance data and the output message delimiter.
*/
type Response struct {
	statusCode               int
	defaultOkMessage         string
	outputMessages           []string
	performanceData          PerformanceData
	outputDelimiter          string
	performanceDataJSONLabel bool
}

/*
NewResponse creates a new Response and sets the default OK message to the given string. The default OK message will be displayed together with the other output messages, but only if the status is still OK when the check exits.
*/
func NewResponse(defaultOkMessage string) *Response {
	response := &Response{
		statusCode:       OK,
		defaultOkMessage: defaultOkMessage,
		outputDelimiter:  "\n",
	}
	response.performanceData = make(PerformanceData)
	return response
}

/*
AddPerformanceDataPoint adds a PerformanceDataPoint to the PerformanceData map, using PerformanceData.add(*PerformanceDataPoint).
Usage:
	err := Response.AddPerformanceDataPoint(NewPerformanceDataPoint("temperature", 32, "°C").SetWarn(35).SetCrit(40))
	if err != nil {
		...
	}
*/
func (r *Response) AddPerformanceDataPoint(point *PerformanceDataPoint) error {
	return r.performanceData.add(point)
}

/*
UpdateStatus updates the exit status of the Response and adds a statusMessage to the outputMessages that will be displayed when the check exits.
See updateStatusCode(int) for a detailed description of the algorithm that is used to update the status code.
*/
func (r *Response) UpdateStatus(statusCode int, statusMessage string) {
	r.updateStatusCode(statusCode)
	if statusMessage != "" {
		r.outputMessages = append(r.outputMessages, statusMessage)
	}
}

/*
GetStatusCode returns the current status code.
*/
func (r *Response) GetStatusCode() int {
	return r.statusCode
}

/*
SetPerformanceDataJSONLabel updates the JSON label.
*/
func (r *Response) SetPerformanceDataJSONLabel(jsonLabel bool) {
	r.performanceDataJSONLabel = jsonLabel
}

/*
This function updates the statusCode of the Response. The status code is mapped to a state like this:
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
	Response.updateStatusCode(0) //nothing changes
	Response.updateStatusCode(2) //status code changes to CRITICAL (=2)

	//now current status code = 2
	Response.updateStatusCode(3) //nothing changes, because CRITICAL is worse than UNKNOWN

*/
func (r *Response) updateStatusCode(statusCode int) {
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
SetOutputDelimiter is used to set the delimiter that is used to separate the outputMessages that will be displayed when the check plugin exits. The default value is a linebreak (\n)
It can be set to any string.
Example:
	Response.SetOutputDelimiter(" / ")
	//this results in the output having the following format:
	//OK: defaultOkMessage / outputMessage1 / outputMessage2 / outputMessage3 | PerformanceData
*/
func (r *Response) SetOutputDelimiter(delimiter string) {
	r.outputDelimiter = delimiter
}

/*
OutputDelimiterMultiline sets the outputDelimiter to "\n". (See Response.SetOutputDelimiter(string))
*/
func (r *Response) OutputDelimiterMultiline() {
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
func (r *Response) outputString() string {
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
OutputAndExit generates the output string and prints it to stdout. After that the check plugin exits with the current exit code.
Example:
	Response := NewResponse("everything checked!")
	defer Response.OutputAndExit()

	//check plugin logic...
*/
func (r *Response) OutputAndExit() {
	fmt.Print(r.outputString())

	firstPoint := true
	for _, perfDataPoint := range r.performanceData {
		if firstPoint {
			fmt.Print(" | ")
			firstPoint = false
		} else {
			fmt.Print(" ")
		}
		fmt.Print(perfDataPoint.outputString(r.performanceDataJSONLabel))
	}
	fmt.Println()

	os.Exit(r.statusCode)
}

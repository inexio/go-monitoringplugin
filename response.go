/* Copyright (c) 2019, inexio GmbH, BSD 2-Clause License */

//Package monitoringplugin provides for writing monitoring check plugins for nagios, icinga2, zabbix, etc
package monitoringplugin

import (
	"bytes"
	"fmt"
	"os"
	"sort"
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
	statusCode                 int
	defaultOkMessage           string
	outputMessages             []OutputMessage
	performanceData            PerformanceData
	outputDelimiter            string
	performanceDataJSONLabel   bool
	printPerformacneData       bool
	sortOutputMessagesByStatus bool
}

/*
OutputMessage represents a message of the response.
*/
type OutputMessage struct {
	Status  int
	Message string
}

/*
NewResponse creates a new Response and sets the default OK message to the given string. The default OK message will be displayed together with the other output messages, but only if the status is still OK when the check exits.
*/
func NewResponse(defaultOkMessage string) *Response {
	response := &Response{
		statusCode:           OK,
		defaultOkMessage:     defaultOkMessage,
		outputDelimiter:      "\n",
		printPerformacneData: true,
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
		r.outputMessages = append(r.outputMessages, OutputMessage{statusCode, statusMessage})
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
UpdateStatusIf calls UpdateStatus(statusCode, statusMessage) if the given condition is true.
*/
func (r *Response) UpdateStatusIf(condition bool, statusCode int, statusMessage string) bool {
	if condition {
		r.UpdateStatus(statusCode, statusMessage)
	}
	return condition
}

/*
UpdateStatusIfNot calls UpdateStatus(statusCode, statusMessage) if the given condition is false.
*/
func (r *Response) UpdateStatusIfNot(condition bool, statusCode int, statusMessage string) bool {
	if !condition {
		r.UpdateStatus(statusCode, statusMessage)
	}
	return !condition
}

/*
UpdateStatusOnError calls UpdateStatus(statusCode, statusMessage) if the given error is not nil.
*/
func (r *Response) UpdateStatusOnError(err error, statusCode int, statusMessage string, includeErrorMessage bool) bool {
	x := err != nil
	if x {
		msg := statusMessage
		if includeErrorMessage {
			if msg != "" {
				msg = fmt.Sprintf("%s (error: %s)", msg, err)
			} else {
				msg = err.Error()
			}
		}
		r.UpdateStatus(statusCode, msg)
	}
	return x
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
PrintPerformanceData activates or deactivates printing performance data
*/
func (r *Response) PrintPerformanceData(b bool) {
	r.printPerformacneData = b
}

/*
SortOutputMessagesByStatus sorts the output messages according to their status.
*/
func (r *Response) SortOutputMessagesByStatus(b bool) {
	r.sortOutputMessagesByStatus = b
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
String2StatusCode returns the status code for a string.
OK -> 1, WARNING -> 2, CRITICAL -> 3, UNKNOWN and everything else -> 4 (case insensitive)
*/
func String2StatusCode(s string) int {
	switch {
	case strings.EqualFold("OK", s):
		return 0
	case strings.EqualFold("WARNING", s):
		return 1
	case strings.EqualFold("CRITICAL", s):
		return 2
	default:
		return 3
	}
}

/*
This function returns the output that will be returned by the check plugin as a string.
*/
func (r *Response) outputString() string {
	return string(r.output())
}

/*
This function returns the output that will be returned by the check plugin.
*/
func (r *Response) output() []byte {
	var buffer bytes.Buffer
	buffer.WriteString(statusCode2Text(r.statusCode))
	buffer.WriteString(": ")
	if r.statusCode == OK {
		buffer.WriteString(r.defaultOkMessage)
		if len(r.outputMessages) > 0 {
			buffer.WriteString(r.outputDelimiter)
		}
	}
	var messages []OutputMessage
	if r.sortOutputMessagesByStatus {
		messages = r.getOutputMessagesSortedByStatus()
	} else {
		messages = r.outputMessages
	}
	for c, x := range messages {
		if c != 0 {
			buffer.WriteString(r.outputDelimiter)
		}
		buffer.WriteString(x.Message)
	}

	if r.printPerformacneData {
		firstPoint := true
		for _, perfDataPoint := range r.performanceData {
			if firstPoint {
				buffer.WriteString(" | ")
				firstPoint = false
			} else {
				buffer.WriteByte(' ')
			}
			buffer.Write(perfDataPoint.output(r.performanceDataJSONLabel))
		}
	}
	return buffer.Bytes()
}

func (r *Response) getOutputMessagesSortedByStatus() []OutputMessage {
	var messages []OutputMessage
	// generate copy of messages
	messages = append(messages, r.outputMessages...)
	sort.Slice(messages, func(i, j int) bool {
		if messages[i].Status == CRITICAL {
			return true
		}
		return messages[i].Status > messages[j].Status
	})
	return messages
}

/*
OutputAndExit generates the output string and prints it to stdout. After that the check plugin exits with the current exit code.
Example:
	Response := NewResponse("everything checked!")
	defer Response.OutputAndExit()

	//check plugin logic...
*/
func (r *Response) OutputAndExit() {
	fmt.Printf("%s\n", r.output())
	os.Exit(r.statusCode)
}

/*
ResponseInfo has all available information for a response.
*/
type ResponseInfo struct {
	StatusCode      int
	PerformanceData []PerformanceDataPointInfo
	RawOutput       string
	Messages        []OutputMessage
}

/*
GetInfo returns all information for a response.
*/
func (r *Response) GetInfo() ResponseInfo {
	return ResponseInfo{RawOutput: string(r.output()), StatusCode: r.statusCode, PerformanceData: r.performanceData.GetInfo(), Messages: r.getOutputMessagesSortedByStatus()}
}

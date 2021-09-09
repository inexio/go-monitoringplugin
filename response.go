//Package monitoringplugin provides types for writing monitoring check plugins for nagios, icinga2, zabbix, etc
package monitoringplugin

import (
	"bytes"
	"fmt"
	"github.com/pkg/errors"
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

const (
	// InvalidCharacterRemove removes invalid character in output message.
	InvalidCharacterRemove = iota + 1
	// InvalidCharacterReplace replaces invalid character in output message with another character.
	// Only valid if replace character is set
	InvalidCharacterReplace
	// InvalidCharacterRemoveMessage removes the message with the invalid character.
	// StatusCode of the message will still be set.
	InvalidCharacterRemoveMessage
	// InvalidCharacterReplaceWithError replaces the whole message with an error message if an invalid character is found.
	InvalidCharacterReplaceWithError
	// InvalidCharacterReplaceWithErrorAndSetUNKNOWN replaces the whole message with an error message if an invalid character is found.
	// Also sets the status code to UNKNOWN.
	InvalidCharacterReplaceWithErrorAndSetUNKNOWN
)

// OutputMessage represents a message of the response. It contains a message and a status code.
type OutputMessage struct {
	Status  int    `yaml:"status" json:"status" xml:"status"`
	Message string `yaml:"message" json:"message" xml:"message"`
}

// Response is the main type that is responsible for the check plugin Response.
// It stores the current status code, output messages, performance data and the output message delimiter.
type Response struct {
	statusCode                  int
	defaultOkMessage            string
	outputMessages              []OutputMessage
	performanceData             performanceData
	outputDelimiter             string
	performanceDataJSONLabel    bool
	printPerformanceData        bool
	sortOutputMessagesByStatus  bool
	invalidCharacterBehaviour   int
	invalidCharacterReplaceChar string
}

/*
NewResponse creates a new Response and sets the default OK message to the given string.
The default OK message will be displayed together with the other output messages, but only
if the status is still OK when the check exits.
*/
func NewResponse(defaultOkMessage string) *Response {
	response := &Response{
		statusCode:                 OK,
		defaultOkMessage:           defaultOkMessage,
		outputDelimiter:            "\n",
		printPerformanceData:       true,
		sortOutputMessagesByStatus: true,
		invalidCharacterBehaviour:  InvalidCharacterRemove,
	}
	response.performanceData = make(performanceData)
	return response
}

/*
AddPerformanceDataPoint adds a PerformanceDataPoint to the performanceData map,
using performanceData.add(*PerformanceDataPoint).
Usage:
	err := Response.AddPerformanceDataPoint(NewPerformanceDataPoint("temperature", 32, "°C").SetWarn(35).SetCrit(40))
	if err != nil {
		...
	}
*/
func (r *Response) AddPerformanceDataPoint(point *PerformanceDataPoint) error {
	err := r.performanceData.add(point)
	if err != nil {
		return errors.Wrap(err, "failed to add performance data point")
	}

	if !point.Thresholds.IsEmpty() {
		name := point.Metric
		if point.Label != "" {
			name += " (" + point.Label + ")"
		}
		err = r.CheckThresholds(point.Thresholds, point.Value, name)
		if err != nil {
			return errors.Wrap(err, "failed to check thresholds")
		}
	}

	return nil
}

/*
UpdateStatus updates the exit status of the Response and adds a statusMessage to the outputMessages that
will be displayed when the check exits.
See updateStatusCode(int) for a detailed description of the algorithm that is used to update the status code.
*/
func (r *Response) UpdateStatus(statusCode int, statusMessage string) {
	r.updateStatusCode(statusCode)
	if statusMessage != "" {
		r.outputMessages = append(r.outputMessages, OutputMessage{statusCode, statusMessage})
	}
}

// GetStatusCode returns the current status code.
func (r *Response) GetStatusCode() int {
	return r.statusCode
}

// SetPerformanceDataJSONLabel updates the JSON metric.
func (r *Response) SetPerformanceDataJSONLabel(jsonLabel bool) {
	r.performanceDataJSONLabel = jsonLabel
}

// SetInvalidCharacterBehavior sets the desired behavior if an invalid character is found in a message.
// Default is InvalidCharacterRemove.
func (r *Response) SetInvalidCharacterBehavior(behavior int, replaceCharacter string) error {
	switch behavior {
	case InvalidCharacterRemove, InvalidCharacterRemoveMessage, InvalidCharacterReplaceWithError, InvalidCharacterReplaceWithErrorAndSetUNKNOWN:
		r.invalidCharacterBehaviour = behavior
	case InvalidCharacterReplace:
		r.invalidCharacterBehaviour = behavior
		if replaceCharacter == "" {
			return errors.New("empty replace character set")
		}
		r.invalidCharacterReplaceChar = replaceCharacter
	default:
		return errors.New("unknown behavior")
	}
	return nil
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
Everything "left" from the current status code is seen as worse than the current one.
If the function wants to set a status code, it will only update it if the new status code is "left" of the current one.
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

// UpdateStatusIf calls UpdateStatus(statusCode, statusMessage) if the given condition is true.
func (r *Response) UpdateStatusIf(condition bool, statusCode int, statusMessage string) bool {
	if condition {
		r.UpdateStatus(statusCode, statusMessage)
	}
	return condition
}

// UpdateStatusIfNot calls UpdateStatus(statusCode, statusMessage) if the given condition is false.
func (r *Response) UpdateStatusIfNot(condition bool, statusCode int, statusMessage string) bool {
	if !condition {
		r.UpdateStatus(statusCode, statusMessage)
	}
	return !condition
}

// UpdateStatusOnError calls UpdateStatus(statusCode, statusMessage) if the given error is not nil.
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
SetOutputDelimiter is used to set the delimiter that is used to separate the outputMessages that will be displayed when
the check plugin exits. The default value is a linebreak (\n)
It can be set to any string.
Example:
	Response.SetOutputDelimiter(" / ")
	//this results in the output having the following format:
	//OK: defaultOkMessage / outputMessage1 / outputMessage2 / outputMessage3 | performanceData
*/
func (r *Response) SetOutputDelimiter(delimiter string) {
	r.outputDelimiter = delimiter
}

// OutputDelimiterMultiline sets the outputDelimiter to "\n". (See Response.SetOutputDelimiter(string))
func (r *Response) OutputDelimiterMultiline() {
	r.SetOutputDelimiter("\n")
}

// PrintPerformanceData activates or deactivates printing performance data
func (r *Response) PrintPerformanceData(b bool) {
	r.printPerformanceData = b
}

// SortOutputMessagesByStatus sorts the output messages according to their status.
func (r *Response) SortOutputMessagesByStatus(b bool) {
	r.sortOutputMessagesByStatus = b
}

// This function returns the output that will be returned by the check plugin as a string.
func (r *Response) outputString() string {
	return string(r.output())
}

// This function returns the output that will be returned by the check plugin.
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

	for c, x := range r.outputMessages {
		if c != 0 {
			buffer.WriteString(r.outputDelimiter)
		}
		buffer.WriteString(x.Message)
	}

	if r.printPerformanceData {
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

func (r *Response) validate() {
	if strings.Contains(r.defaultOkMessage, "|") {
		switch r.invalidCharacterBehaviour {
		case InvalidCharacterReplace:
			r.defaultOkMessage = strings.ReplaceAll(r.defaultOkMessage, "|", r.invalidCharacterReplaceChar)
		case InvalidCharacterRemoveMessage:
			r.defaultOkMessage = ""
		case InvalidCharacterReplaceWithError:
			r.defaultOkMessage = "default output message contains invalid character"
		case InvalidCharacterReplaceWithErrorAndSetUNKNOWN:
			r.statusCode = UNKNOWN
			r.outputMessages = []OutputMessage{{
				Status:  UNKNOWN,
				Message: "default output message contains invalid character",
			}}
			r.outputMessages = nil
			return
		default: // InvalidCharacterRemove
			r.defaultOkMessage = strings.ReplaceAll(r.defaultOkMessage, "|", "")
		}
	}
	r.validateMessages()
	if r.sortOutputMessagesByStatus {
		r.sortMessagesByStatus()
	}
}

func (r *Response) validateMessages() {
	var messages []OutputMessage
out:
	for _, message := range r.outputMessages {
		if !strings.Contains(message.Message, "|") {
			messages = append(messages, message)
		} else {
			switch r.invalidCharacterBehaviour {
			case InvalidCharacterReplace:
				newMessage := strings.ReplaceAll(message.Message, "|", r.invalidCharacterReplaceChar)
				if newMessage != "" {
					messages = append(messages, OutputMessage{
						Status:  message.Status,
						Message: newMessage,
					})
				}
			case InvalidCharacterRemoveMessage:
				// done
			case InvalidCharacterReplaceWithErrorAndSetUNKNOWN:
				r.statusCode = UNKNOWN
				message.Status = UNKNOWN
				fallthrough
			case InvalidCharacterReplaceWithError:
				messages = []OutputMessage{{
					Status:  message.Status,
					Message: "output message contains invalid character",
				}}
				break out
			default: // InvalidCharacterRemove
				newMessage := strings.ReplaceAll(message.Message, "|", "")
				if newMessage != "" {
					messages = append(messages, OutputMessage{
						Status:  message.Status,
						Message: newMessage,
					})
				}
			}
		}
	}
	r.outputMessages = messages
}

func (r *Response) sortMessagesByStatus() {
	sort.Slice(r.outputMessages, func(i, j int) bool {
		if r.outputMessages[i].Status == CRITICAL {
			return true
		}
		return r.outputMessages[i].Status > r.outputMessages[j].Status
	})
}

/*
OutputAndExit generates the output string and prints it to stdout.
After that the check plugin exits with the current exit code.
Example:
	Response := NewResponse("everything checked!")
	defer Response.OutputAndExit()

	//check plugin logic...
*/
func (r *Response) OutputAndExit() {
	r.validate()
	fmt.Println(r.outputString())
	os.Exit(r.statusCode)
}

// ResponseInfo has all available information for a response. It also contains the RawOutput.
type ResponseInfo struct {
	StatusCode      int                    `yaml:"status_code" json:"status_code" xml:"status_code"`
	PerformanceData []PerformanceDataPoint `yaml:"performance_data" json:"performance_data" xml:"performance_data"`
	RawOutput       string                 `yaml:"raw_output" json:"raw_output" xml:"raw_output"`
	Messages        []OutputMessage        `yaml:"messages" json:"messages" xml:"messages"`
}

// GetInfo returns all information for a response.
func (r *Response) GetInfo() ResponseInfo {
	r.validate()
	return ResponseInfo{
		RawOutput:       r.outputString(),
		StatusCode:      r.statusCode,
		PerformanceData: r.performanceData.getInfo(),
		Messages:        r.outputMessages,
	}
}

// CheckThresholds checks if the value exceeds the given thresholds and updates the response
func (r *Response) CheckThresholds(thresholds Thresholds, value interface{}, name string) error {
	res, err := thresholds.CheckValue(value)
	if err != nil {
		return errors.Wrap(err, "failed to check value against threshold")
	}
	if res != OK {
		r.UpdateStatus(res, name+" is outside of threshold")
	}
	return nil
}

/*
String2StatusCode returns the status code for a string.
OK -> 1, WARNING -> 2, CRITICAL -> 3, UNKNOWN and everything else -> 4 (case insensitive)
*/
func String2StatusCode(s string) int {
	switch {
	case strings.EqualFold("OK", s):
		return OK
	case strings.EqualFold("WARNING", s):
		return WARNING
	case strings.EqualFold("CRITICAL", s):
		return CRITICAL
	default:
		return UNKNOWN
	}
}

// This function is used to map the status code to a string.
func statusCode2Text(statusCode int) string {
	switch {
	case statusCode == OK:
		return "OK"
	case statusCode == WARNING:
		return "WARNING"
	case statusCode == CRITICAL:
		return "CRITICAL"
	default:
		return "UNKNOWN"
	}
}

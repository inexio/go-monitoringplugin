package monitoringplugin

import (
	"bytes"
	"github.com/stretchr/testify/assert"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"testing"
)

func TestOKResponse(t *testing.T) {
	defaultMessage := "OKTest"
	if os.Getenv("EXECUTE_PLUGIN") == "1" {
		r := NewResponse(defaultMessage)
		r.OutputAndExit()
	}
	cmd := exec.Command(os.Args[0], "-test.run=TestOKResponse")
	cmd.Env = append(os.Environ(), "EXECUTE_PLUGIN=1")
	var outputB bytes.Buffer
	cmd.Stdout = &outputB
	err := cmd.Run()

	if err != nil {
		if exitError, ok := err.(*exec.ExitError); ok {
			t.Error("OkResponse is expected to return exit status 0, but exited with exit code " + strconv.Itoa(exitError.ExitCode()))
		} else {
			t.Error("cmd.Run() Command resulted in an error that can not be converted to exec.ExitEror! error: " + err.Error())
		}
		return
	}

	output := outputB.String()
	match, err := regexp.MatchString("^OK: "+defaultMessage+"\n$", output)
	if err != nil {
		t.Error(err.Error())
	}
	if !match {
		t.Error("ok result output message did not match to the expected regex")
	}

	return
}

func TestWARNINGResponse(t *testing.T) {
	failureResponse(t, 1)
	return
}

func TestCRITICALResponse(t *testing.T) {
	failureResponse(t, 2)
	return
}

func TestUNKNOWNResponse(t *testing.T) {
	failureResponse(t, 3)
	return
}

func TestStatusHierarchy(t *testing.T) {
	r := NewResponse("")
	if r.statusCode != OK {
		t.Error("status code is supposed to be OK when a new Response is created")
	}

	r.UpdateStatus(WARNING, "")
	if r.statusCode != WARNING {
		t.Error("status code did not update from OK to WARNING after UpdateStatus(WARNING) is called!")
	}

	r.UpdateStatus(OK, "")
	if r.statusCode != WARNING {
		t.Error("status code did change from WARNING to " + strconv.Itoa(r.statusCode) + " after UpdateStatus(OK) was called! The function should not affect the status code, because WARNING is worse than OK")
	}

	r.UpdateStatus(CRITICAL, "")
	if r.statusCode != CRITICAL {
		t.Error("status code did not update from WARNING to CRITICAL after UpdateStatus(WARNING) is called!")
	}

	r.UpdateStatus(OK, "")
	if r.statusCode != CRITICAL {
		t.Error("status code did change from CRITICAL to " + strconv.Itoa(r.statusCode) + " after UpdateStatus(OK) was called! The function should not affect the status code, because CRITICAL is worse than OK")
	}

	r.UpdateStatus(WARNING, "")
	if r.statusCode != CRITICAL {
		t.Error("status code did change from CRITICAL to " + strconv.Itoa(r.statusCode) + " after UpdateStatus(WARNING) was called! The function should not affect the status code, because CRITICAL is worse than WARNING")
	}

	r.UpdateStatus(UNKNOWN, "")
	if r.statusCode != CRITICAL {
		t.Error("status code did change from CRITICAL to " + strconv.Itoa(r.statusCode) + " after UpdateStatus(UNKNOWN) was called! The function should not affect the status code, because CRITICAL is worse than UNKNOWN")
	}

	r = NewResponse("")
	r.UpdateStatus(UNKNOWN, "")
	if r.statusCode != UNKNOWN {
		t.Error("status code did not update from OK to UNKNOWN after UpdateStatus(UNKNOWN) is called!")
	}

	r.UpdateStatus(WARNING, "")
	if r.statusCode != UNKNOWN {
		t.Error("status code did change from UNKNOWN to " + strconv.Itoa(r.statusCode) + " after UpdateStatus(WARNING) was called! The function should not affect the status code, because UNKNOWN is worse than WARNING")
	}

	r.UpdateStatus(CRITICAL, "")
	if r.statusCode != CRITICAL {
		t.Error("status code is did not change from UNKNOWN to CRITICAL after UpdateStatus(CRITICAL) was called! The function should affect the status code, because CRITICAL is worse than UNKNOWN")
	}
}

func TestOutputMessages(t *testing.T) {
	defaultMessage := "default"
	if os.Getenv("EXECUTE_PLUGIN") == "1" {
		r := NewResponse(defaultMessage)
		r.UpdateStatus(0, "message1")
		r.UpdateStatus(0, "message2")
		r.UpdateStatus(0, "message3")
		r.UpdateStatus(0, "message4")
		r.OutputAndExit()
		return
	}
	if os.Getenv("EXECUTE_PLUGIN") == "2" {
		r := NewResponse(defaultMessage)
		r.UpdateStatus(1, "message1")
		r.UpdateStatus(0, "message2")
		r.UpdateStatus(0, "message3")
		r.UpdateStatus(0, "message4")
		r.SetOutputDelimiter(" / ")
		r.OutputAndExit()
		return
	}
	cmd := exec.Command(os.Args[0], "-test.run=TestOutputMessages")
	cmd.Env = append(os.Environ(), "EXECUTE_PLUGIN=1")
	var outputB bytes.Buffer
	cmd.Stdout = &outputB
	err := cmd.Run()

	if err != nil {
		t.Error("an error occurred during cmd.Run(), but the Response was expected to exit with exit code 0")
		return
	}

	output := outputB.String()

	match, err := regexp.MatchString("^OK: "+defaultMessage+"\nmessage1\nmessage2\nmessage3\nmessage4\n$", output)
	if err != nil {
		t.Error(err.Error())
	}
	if !match {
		t.Error("output did not match to the expected regex")
	}

	cmd = exec.Command(os.Args[0], "-test.run=TestOutputMessages")
	cmd.Env = append(os.Environ(), "EXECUTE_PLUGIN=2")
	var outputB2 bytes.Buffer
	cmd.Stdout = &outputB2
	err = cmd.Run()

	if err != nil {
		if exitError, ok := err.(*exec.ExitError); ok {
			if exitError.ExitCode() != 1 {
				t.Error("the command is expected to return exit status 1, but exited with exit code " + strconv.Itoa(exitError.ExitCode()))
			}
		} else {
			t.Errorf("cmd.Run() Command resulted in an error that can not be converted to exec.ExitEror! error: " + err.Error())
		}
	} else {
		t.Error("the command exited with exitcode 0 but is expected to exit with exitcode 1")
	}

	output = outputB2.String()
	match, err = regexp.MatchString("^WARNING: message1 / message2 / message3 / message4\n$", output)
	if err != nil {
		t.Error(err.Error())
	}

	if !match {
		t.Error("output did not match to the expected regex")
	}
}

func TestResponse_UpdateStatusIf(t *testing.T) {
	r := NewResponse("")
	r.UpdateStatusIf(false, 1, "")
	assert.True(t, r.statusCode == 0)
	r.UpdateStatusIf(true, 1, "")
	assert.True(t, r.statusCode == 1)
}

func TestResponse_UpdateStatusIfNot(t *testing.T) {
	r := NewResponse("")
	r.UpdateStatusIfNot(true, 1, "")
	assert.True(t, r.statusCode == 0)
	r.UpdateStatusIfNot(false, 1, "")
	assert.True(t, r.statusCode == 1)
}

func TestString2StatusCode(t *testing.T) {
	assert.True(t, String2StatusCode("ok") == 0)
	assert.True(t, String2StatusCode("OK") == 0)
	assert.True(t, String2StatusCode("Ok") == 0)
	assert.True(t, String2StatusCode("warning") == 1)
	assert.True(t, String2StatusCode("WARNING") == 1)
	assert.True(t, String2StatusCode("Warning") == 1)
	assert.True(t, String2StatusCode("critical") == 2)
	assert.True(t, String2StatusCode("CRITICAL") == 2)
	assert.True(t, String2StatusCode("Critical") == 2)
	assert.True(t, String2StatusCode("unknown") == 3)
	assert.True(t, String2StatusCode("UNKNOWN") == 3)
	assert.True(t, String2StatusCode("Unknown") == 3)
}

func TestOutputPerformanceData(t *testing.T) {
	p1 := NewPerformanceDataPoint("label1", 10).
		SetUnit("%").
		SetMin(0).
		SetMax(100).
		SetThresholds(
			NewThresholds(0, 80, 0, 90))
	p2 := NewPerformanceDataPoint("label2", 20).
		SetUnit("%").
		SetMin(0).
		SetMax(100).
		SetThresholds(
			NewThresholds(0, 80, 0, 90))
	p3 := NewPerformanceDataPoint("label3", 30).
		SetUnit("%").
		SetMin(0).
		SetMax(100).
		SetThresholds(
			NewThresholds(0, 80, 0, 90))

	defaultMessage := "OKTest"
	if os.Getenv("EXECUTE_PLUGIN") == "1" {
		r := NewResponse(defaultMessage)
		err := r.AddPerformanceDataPoint(p1)
		if err != nil {
			r.UpdateStatus(3, "error during add performance data point")
		}
		err = r.AddPerformanceDataPoint(p2)
		if err != nil {
			r.UpdateStatus(3, "error during add performance data point")
		}
		err = r.AddPerformanceDataPoint(p3)
		if err != nil {
			r.UpdateStatus(3, "error during add performance data point")
		}
		r.OutputAndExit()
	}
	cmd := exec.Command(os.Args[0], "-test.run=TestOutputPerformanceData")
	cmd.Env = append(os.Environ(), "EXECUTE_PLUGIN=1")
	var outputB bytes.Buffer
	cmd.Stdout = &outputB
	err := cmd.Run()
	if err != nil {
		t.Error("cmd.Run() returned an exitcode != 0, but exit code 0 was expected")
	}

	output := outputB.String()
	if !strings.HasPrefix(output, "OK: "+defaultMessage+" | ") {
		t.Error("output did not match the expected regex")
	}
}

func TestOutputPerformanceDataThresholdsExceeded(t *testing.T) {
	p1 := NewPerformanceDataPoint("label1", 10).
		SetUnit("%").
		SetMin(0).
		SetMax(100).
		SetThresholds(
			NewThresholds(0, 80, 0, 90))
	p2 := NewPerformanceDataPoint("label2", 20).
		SetUnit("%").
		SetMin(0).
		SetMax(100).
		SetThresholds(
			NewThresholds(0, 80, 0, 90))
	p3 := NewPerformanceDataPoint("label3", 85).
		SetUnit("%").
		SetMin(0).
		SetMax(100).
		SetThresholds(
			NewThresholds(0, 80, 0, 90))

	defaultMessage := "OKTest"
	if os.Getenv("EXECUTE_PLUGIN") == "1" {
		r := NewResponse(defaultMessage)
		err := r.AddPerformanceDataPoint(p1)
		if err != nil {
			r.UpdateStatus(3, "error during add performance data point")
		}
		err = r.AddPerformanceDataPoint(p2)
		if err != nil {
			r.UpdateStatus(3, "error during add performance data point")
		}
		err = r.AddPerformanceDataPoint(p3)
		if err != nil {
			r.UpdateStatus(3, "error during add performance data point")
		}
		r.OutputAndExit()
	}
	cmd := exec.Command(os.Args[0], "-test.run=TestOutputPerformanceDataThresholdsExceeded")
	cmd.Env = append(os.Environ(), "EXECUTE_PLUGIN=1")
	var outputB bytes.Buffer
	cmd.Stdout = &outputB
	err := cmd.Run()
	if err == nil {
		t.Error("cmd.Run() returned an exitcode = 0, but exit code 1 was expected")
	} else if err.Error() != "exit status 1" {
		t.Error("cmd.Run() returned an exitcode != 1, but exit code 1 was expected")
	}

	output := outputB.String()
	if !strings.HasPrefix(output, "WARNING: label3 is outside of WARNING threshold | ") {
		t.Error("output did not match the expected regex")
	}
}

func failureResponse(t *testing.T, exitCode int) {
	var status string
	switch exitCode {
	case 0:
		t.Error("exitcode in failureResponse function cannot be 0, because it is not meant to be used for a successful cmd")
		return
	case 1:
		status = "WARNING"
	case 2:
		status = "CRITICAL"
	default:
		status = "UNKNOWN"
	}

	message := status + "Test"
	if os.Getenv("EXECUTE_PLUGIN") == "1" {
		r := NewResponse("")
		r.UpdateStatus(exitCode, message)
		r.OutputAndExit()
	}
	cmd := exec.Command(os.Args[0], "-test.run=Test"+status+"Response")
	cmd.Env = append(os.Environ(), "EXECUTE_PLUGIN=1")

	var outputB bytes.Buffer
	cmd.Stdout = &outputB
	err := cmd.Run()

	if err != nil {
		if exitError, ok := err.(*exec.ExitError); ok {
			if exitError.ExitCode() != exitCode {
				t.Error(status + " Response is expected to return exit status " + strconv.Itoa(exitCode) + ", but exited with exit code " + strconv.Itoa(exitError.ExitCode()))
			}
		} else {
			t.Errorf("cmd.Run() Command resulted in an error that can not be converted to exec.ExitEror! error: " + err.Error())
		}
	} else {
		t.Error("the command exited with exitcode 0 but is expected to exit with exitcode " + strconv.Itoa(exitCode))
	}

	output := outputB.String()
	match, err := regexp.MatchString("^"+status+": "+message+"\n$", output)
	if err != nil {
		t.Error(err.Error())
	}
	if !match {
		t.Error(status + " result output message did not match to the expected regex")
	}
	return
}

func TestResponse_SortOutputMessagesByStatus(t *testing.T) {
	r := NewResponse("defaultMessage")
	r.UpdateStatus(OK, "message1")
	r.UpdateStatus(WARNING, "message2")
	r.UpdateStatus(UNKNOWN, "message3")
	r.UpdateStatus(CRITICAL, "message4")
	r.UpdateStatus(WARNING, "message5")
	r.UpdateStatus(CRITICAL, "message6")
	r.UpdateStatus(UNKNOWN, "message7")
	r.UpdateStatus(OK, "message8")
	r.validate()
	for x, message := range r.outputMessages {
		for _, m := range r.outputMessages[x:] {
			assert.True(t, message.Status >= m.Status || message.Status == CRITICAL, "sorting did not work")
		}
	}
}

func TestResponse_InvalidCharacter(t *testing.T) {
	r := NewResponse("checked")
	r.UpdateStatus(WARNING, "test|")
	r.validate()
	res := r.GetInfo()
	assert.True(t, res.RawOutput == "WARNING: test")
}

func TestResponse_InvalidCharacterReplace(t *testing.T) {
	r := NewResponse("checked")
	r.UpdateStatus(OK, "test|2")
	err := r.SetInvalidCharacterBehavior(InvalidCharacterReplace, "-")
	assert.NoError(t, err)
	r.validate()
	res := r.GetInfo()
	assert.True(t, res.RawOutput == "OK: checked\ntest-2")
}

func TestResponse_InvalidCharacterReplaceError(t *testing.T) {
	r := NewResponse("checked")
	r.UpdateStatus(OK, "test|")
	err := r.SetInvalidCharacterBehavior(InvalidCharacterReplace, "")
	assert.Error(t, err)
}

func TestResponse_InvalidCharacterRemoveMessage(t *testing.T) {
	r := NewResponse("checked")
	r.UpdateStatus(OK, "test|")
	err := r.SetInvalidCharacterBehavior(InvalidCharacterRemoveMessage, "")
	assert.NoError(t, err)
	r.validate()
	res := r.GetInfo()
	assert.True(t, res.RawOutput == "OK: checked")
}

func TestResponse_InvalidCharacterReplaceWithError(t *testing.T) {
	r := NewResponse("checked")
	r.UpdateStatus(WARNING, "test|")
	err := r.SetInvalidCharacterBehavior(InvalidCharacterReplaceWithError, "")
	assert.NoError(t, err)
	r.validate()
	res := r.GetInfo()
	assert.True(t, res.RawOutput == "WARNING: output message contains invalid character")
}

func TestResponse_InvalidCharacterReplaceWithErrorMultipleMessages(t *testing.T) {
	r := NewResponse("checked")
	r.UpdateStatus(WARNING, "test|")
	r.UpdateStatus(WARNING, "test|2")
	err := r.SetInvalidCharacterBehavior(InvalidCharacterReplaceWithError, "")
	assert.NoError(t, err)
	r.validate()
	res := r.GetInfo()
	assert.True(t, res.RawOutput == "WARNING: output message contains invalid character")
}

func TestResponse_InvalidCharacterReplaceWithErrorAndSetUnknown(t *testing.T) {
	r := NewResponse("checked")
	r.UpdateStatus(WARNING, "test|")
	err := r.SetInvalidCharacterBehavior(InvalidCharacterReplaceWithErrorAndSetUNKNOWN, "")
	assert.NoError(t, err)
	r.validate()
	res := r.GetInfo()
	assert.True(t, res.RawOutput == "UNKNOWN: output message contains invalid character")
}

func TestResponse_InvalidCharacterReplaceWithErrorAndSetUnknownMultipleMessages(t *testing.T) {
	r := NewResponse("checked")
	r.UpdateStatus(WARNING, "test|")
	r.UpdateStatus(WARNING, "test|2")
	err := r.SetInvalidCharacterBehavior(InvalidCharacterReplaceWithErrorAndSetUNKNOWN, "")
	assert.NoError(t, err)
	r.validate()
	res := r.GetInfo()
	assert.True(t, res.RawOutput == "UNKNOWN: output message contains invalid character")
}

func TestResponse_InvalidCharacterDefaultMessage(t *testing.T) {
	r := NewResponse("test|")
	r.validate()
	res := r.GetInfo()
	assert.True(t, res.RawOutput == "OK: test")
}

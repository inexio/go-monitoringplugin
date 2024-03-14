package monitoringplugin

import (
	"bytes"
	"os"
	"os/exec"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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
	if err := cmd.Run(); err != nil {
		var exitError *exec.ExitError
		require.ErrorAs(t, err, &exitError,
			"cmd.Run() Command resulted in an error that can not be converted to exec.ExitEror! error: %s",
			err)
		require.Equal(t, 0, exitError.ExitCode(),
			"OkResponse is expected to return exit status 0, but exited with exit code %v",
			exitError.ExitCode())
	}

	output := outputB.String()
	require.Regexp(t, "^OK: "+defaultMessage+"\n$", output,
		"ok result output message did not match to the expected regex")
}

func TestWARNINGResponse(t *testing.T) {
	failureResponse(t, 1)
}

func TestCRITICALResponse(t *testing.T) {
	failureResponse(t, 2)
}

func TestUNKNOWNResponse(t *testing.T) {
	failureResponse(t, 3)
}

func TestStatusHierarchy(t *testing.T) {
	r := NewResponse("")
	require.Equal(t, OK, r.statusCode,
		"status code is supposed to be OK when a new Response is created")

	r.UpdateStatus(WARNING, "")
	require.Equal(t, WARNING, r.statusCode,
		"status code did not update from OK to WARNING after UpdateStatus(WARNING) is called!")

	r.UpdateStatus(OK, "")
	require.Equal(t, WARNING, r.statusCode,
		"status code did change from WARNING to %v after UpdateStatus(OK) was called! The function should not affect the status code, because WARNING is worse than OK",
		r.statusCode)

	r.UpdateStatus(CRITICAL, "")
	require.Equal(t, CRITICAL, r.statusCode,
		"status code did not update from WARNING to CRITICAL after UpdateStatus(WARNING) is called!")

	r.UpdateStatus(OK, "")
	require.Equal(t, CRITICAL, r.statusCode,
		"status code did change from CRITICAL to %v after UpdateStatus(OK) was called! The function should not affect the status code, because CRITICAL is worse than OK",
		r.statusCode)

	r.UpdateStatus(WARNING, "")
	require.Equal(t, CRITICAL, r.statusCode,
		"status code did change from CRITICAL to %v after UpdateStatus(WARNING) was called! The function should not affect the status code, because CRITICAL is worse than WARNING",
		r.statusCode)

	r.UpdateStatus(UNKNOWN, "")
	require.Equal(t, CRITICAL, r.statusCode,
		"status code did change from CRITICAL to %v after UpdateStatus(UNKNOWN) was called! The function should not affect the status code, because CRITICAL is worse than UNKNOWN",
		r.statusCode)

	r = NewResponse("")
	r.UpdateStatus(UNKNOWN, "")
	require.Equal(t, UNKNOWN, r.statusCode,
		"status code did not update from OK to UNKNOWN after UpdateStatus(UNKNOWN) is called!")

	r.UpdateStatus(WARNING, "")
	require.Equal(t, UNKNOWN, r.statusCode,
		"status code did change from UNKNOWN to %v after UpdateStatus(WARNING) was called! The function should not affect the status code, because UNKNOWN is worse than WARNING",
		r.statusCode)

	r.UpdateStatus(CRITICAL, "")
	require.Equal(t, CRITICAL, r.statusCode,
		"status code is did not change from UNKNOWN to CRITICAL after UpdateStatus(CRITICAL) was called! The function should affect the status code, because CRITICAL is worse than UNKNOWN")
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
	}
	if os.Getenv("EXECUTE_PLUGIN") == "2" {
		r := NewResponse(defaultMessage)
		r.UpdateStatus(1, "message1")
		r.UpdateStatus(0, "message2")
		r.UpdateStatus(0, "message3")
		r.UpdateStatus(0, "message4")
		r.SetOutputDelimiter(" / ")
		r.OutputAndExit()
	}
	cmd := exec.Command(os.Args[0], "-test.run=TestOutputMessages")
	cmd.Env = append(os.Environ(), "EXECUTE_PLUGIN=1")
	var outputB bytes.Buffer
	cmd.Stdout = &outputB
	require.NoError(t, cmd.Run(),
		"an error occurred during cmd.Run(), but the Response was expected to exit with exit code 0")

	output := outputB.String()
	require.Regexp(t,
		"^OK: "+defaultMessage+"\nmessage1\nmessage2\nmessage3\nmessage4\n$",
		output, "output did not match to the expected regex")

	cmd = exec.Command(os.Args[0], "-test.run=TestOutputMessages")
	cmd.Env = append(os.Environ(), "EXECUTE_PLUGIN=2")
	var outputB2 bytes.Buffer
	cmd.Stdout = &outputB2

	err := cmd.Run()
	require.Error(t, err,
		"the command exited with exitcode 0 but is expected to exit with exitcode 1")
	var exitError *exec.ExitError
	require.ErrorAs(t, err, &exitError,
		"cmd.Run() Command resulted in an error that can not be converted to exec.ExitEror! error: %v",
		err.Error())
	require.Equal(t, 1, exitError.ExitCode(),
		"the command is expected to return exit status 1, but exited with exit code %v",
		exitError.ExitCode())

	output = outputB2.String()
	require.Regexp(t, "^WARNING: message1 / message2 / message3 / message4\n$",
		output, "output did not match to the expected regex")
}

func TestResponse_UpdateStatusIf(t *testing.T) {
	r := NewResponse("")
	r.UpdateStatusIf(false, 1, "")
	assert.Equal(t, 0, r.statusCode)
	r.UpdateStatusIf(true, 1, "")
	assert.Equal(t, 1, r.statusCode)
}

func TestResponse_UpdateStatusIfNot(t *testing.T) {
	r := NewResponse("")
	r.UpdateStatusIfNot(true, 1, "")
	assert.Equal(t, 0, r.statusCode)
	r.UpdateStatusIfNot(false, 1, "")
	assert.Equal(t, 1, r.statusCode)
}

func TestString2StatusCode(t *testing.T) {
	assert.Equal(t, 0, String2StatusCode("ok"))
	assert.Equal(t, 0, String2StatusCode("OK"))
	assert.Equal(t, 0, String2StatusCode("Ok"))
	assert.Equal(t, 1, String2StatusCode("warning"))
	assert.Equal(t, 1, String2StatusCode("WARNING"))
	assert.Equal(t, 1, String2StatusCode("Warning"))
	assert.Equal(t, 2, String2StatusCode("critical"))
	assert.Equal(t, 2, String2StatusCode("CRITICAL"))
	assert.Equal(t, 2, String2StatusCode("Critical"))
	assert.Equal(t, 3, String2StatusCode("unknown"))
	assert.Equal(t, 3, String2StatusCode("UNKNOWN"))
	assert.Equal(t, 3, String2StatusCode("Unknown"))
}

func TestOutputPerformanceData(t *testing.T) {
	p1 := NewPerformanceDataPoint("label1", 10).
		SetUnit("%").
		SetMin(0).
		SetMax(100)
	p1.NewThresholds(0, 80, 0, 90)

	p2 := NewPerformanceDataPoint("label2", 20).
		SetUnit("%").
		SetMin(0).
		SetMax(100)
	p2.NewThresholds(0, 80, 0, 90)

	p3 := NewPerformanceDataPoint("label3", 30).
		SetUnit("%").
		SetMin(0).
		SetMax(100)
	p3.NewThresholds(0, 80, 0, 90)

	defaultMessage := "OKTest"
	if os.Getenv("EXECUTE_PLUGIN") == "1" {
		r := NewResponse(defaultMessage)
		if err := r.AddPerformanceDataPoint(p1); err != nil {
			r.UpdateStatus(UNKNOWN, "error during add performance data point")
		}
		if err := r.AddPerformanceDataPoint(p2); err != nil {
			r.UpdateStatus(UNKNOWN, "error during add performance data point")
		}
		if err := r.AddPerformanceDataPoint(p3); err != nil {
			r.UpdateStatus(UNKNOWN, "error during add performance data point")
		}
		r.OutputAndExit()
	}

	cmd := exec.Command(os.Args[0], "-test.run=TestOutputPerformanceData")
	cmd.Env = append(os.Environ(), "EXECUTE_PLUGIN=1")
	var outputB bytes.Buffer
	cmd.Stdout = &outputB
	require.NoError(t, cmd.Run(),
		"cmd.Run() returned an exitcode != 0, but exit code 0 was expected")

	output := outputB.String()
	require.True(t, strings.HasPrefix(output, "OK: "+defaultMessage+" | "),
		"output did not match the expected regex")
}

func TestOutputPerformanceDataThresholdsExceeded(t *testing.T) {
	p1 := NewPerformanceDataPoint("label1", 10).
		SetUnit("%").
		SetMin(0).
		SetMax(100)
	p1.NewThresholds(0, 80, 0, 90)

	p2 := NewPerformanceDataPoint("label2", 20).
		SetUnit("%").
		SetMin(0).
		SetMax(100)
	p2.NewThresholds(0, 80, 0, 90)

	p3 := NewPerformanceDataPoint("label3", 85).
		SetUnit("%").
		SetMin(0).
		SetMax(100)
	p3.NewThresholds(0, 80, 0, 90)

	defaultMessage := "OKTest"
	if os.Getenv("EXECUTE_PLUGIN") == "1" {
		r := NewResponse(defaultMessage)
		if err := r.AddPerformanceDataPoint(p1); err != nil {
			r.UpdateStatus(UNKNOWN, "error during add performance data point")
		}
		if err := r.AddPerformanceDataPoint(p2); err != nil {
			r.UpdateStatus(UNKNOWN, "error during add performance data point")
		}
		if err := r.AddPerformanceDataPoint(p3); err != nil {
			r.UpdateStatus(UNKNOWN, "error during add performance data point")
		}
		r.OutputAndExit()
	}

	cmd := exec.Command(os.Args[0], "-test.run=TestOutputPerformanceDataThresholdsExceeded")
	cmd.Env = append(os.Environ(), "EXECUTE_PLUGIN=1")
	var outputB bytes.Buffer
	cmd.Stdout = &outputB
	err := cmd.Run()
	require.Error(t, err,
		"cmd.Run() returned an exitcode = 0, but exit code 1 was expected")
	var exitError *exec.ExitError
	require.ErrorAs(t, err, &exitError)
	require.Equal(t, 1, exitError.ExitCode(),
		"cmd.Run() returned an exitcode != 1, but exit code 1 was expected")

	output := outputB.String()
	require.True(t,
		strings.HasPrefix(output,
			"WARNING: label3 is outside of WARNING threshold | "),
		"output did not match the expected regex")
}

func failureResponse(t *testing.T, exitCode int) {
	require.NotEqual(t, 0,
		"exitcode in failureResponse function cannot be 0, because it is not meant to be used for a successful cmd")

	var status string
	switch exitCode {
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
	require.Error(t, err,
		"the command exited with exitcode 0 but is expected to exit with exitcode %v",
		exitCode)

	var exitError *exec.ExitError
	require.ErrorAs(t, err, &exitError,
		"cmd.Run() Command resulted in an error that can not be converted to exec.ExitEror! error: %s",
		err.Error())
	require.Equal(t, exitCode, exitError.ExitCode(),
		"%v Response is expected to return exit status %v, but exited with exit code %v",
		status, exitCode, exitError.ExitCode())

	output := outputB.String()
	require.Regexp(t, "^"+status+": "+message+"\n$", output,
		"%s result output message did not match to the expected regex", status)
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
			assert.True(t, message.Status >= m.Status || message.Status == CRITICAL,
				"sorting did not work")
		}
	}
}

func TestResponse_InvalidCharacter(t *testing.T) {
	r := NewResponse("checked")
	r.UpdateStatus(WARNING, "test|")
	r.validate()
	assert.Equal(t, "WARNING: test", r.GetInfo().RawOutput)
}

func TestResponse_InvalidCharacterReplace(t *testing.T) {
	r := NewResponse("checked")
	r.UpdateStatus(OK, "test|2")
	require.NoError(t, r.SetInvalidCharacterBehavior(
		InvalidCharacterReplace, "-"))
	r.validate()
	assert.Equal(t, "OK: checked\ntest-2", r.GetInfo().RawOutput)
}

func TestResponse_InvalidCharacterReplaceError(t *testing.T) {
	r := NewResponse("checked")
	r.UpdateStatus(OK, "test|")
	assert.Error(t, r.SetInvalidCharacterBehavior(InvalidCharacterReplace, ""))
}

func TestResponse_InvalidCharacterRemoveMessage(t *testing.T) {
	r := NewResponse("checked")
	r.UpdateStatus(OK, "test|")
	require.NoError(t, r.SetInvalidCharacterBehavior(
		InvalidCharacterRemoveMessage, ""))
	r.validate()
	assert.Equal(t, "OK: checked", r.GetInfo().RawOutput)
}

func TestResponse_InvalidCharacterReplaceWithError(t *testing.T) {
	r := NewResponse("checked")
	r.UpdateStatus(WARNING, "test|")
	require.NoError(t, r.SetInvalidCharacterBehavior(
		InvalidCharacterReplaceWithError, ""))
	r.validate()
	assert.Equal(t, "WARNING: output message contains invalid character",
		r.GetInfo().RawOutput)
}

func TestResponse_InvalidCharacterReplaceWithErrorMultipleMessages(
	t *testing.T,
) {
	r := NewResponse("checked")
	r.UpdateStatus(WARNING, "test|")
	r.UpdateStatus(WARNING, "test|2")
	require.NoError(t, r.SetInvalidCharacterBehavior(
		InvalidCharacterReplaceWithError, ""))
	r.validate()
	assert.Equal(t, "WARNING: output message contains invalid character",
		r.GetInfo().RawOutput)
}

func TestResponse_InvalidCharacterReplaceWithErrorAndSetUnknown(t *testing.T) {
	r := NewResponse("checked")
	r.UpdateStatus(WARNING, "test|")
	require.NoError(t, r.SetInvalidCharacterBehavior(
		InvalidCharacterReplaceWithErrorAndSetUNKNOWN, ""))
	r.validate()
	assert.Equal(t, "UNKNOWN: output message contains invalid character",
		r.GetInfo().RawOutput)
}

func TestResponse_InvalidCharacterReplaceWithErrorAndSetUnknownMultipleMessages(
	t *testing.T,
) {
	r := NewResponse("checked")
	r.UpdateStatus(WARNING, "test|")
	r.UpdateStatus(WARNING, "test|2")
	require.NoError(t, r.SetInvalidCharacterBehavior(
		InvalidCharacterReplaceWithErrorAndSetUNKNOWN, ""))
	r.validate()
	assert.Equal(t, "UNKNOWN: output message contains invalid character",
		r.GetInfo().RawOutput)
}

func TestResponse_InvalidCharacterDefaultMessage(t *testing.T) {
	r := NewResponse("test|")
	r.validate()
	assert.Equal(t, "OK: test", r.GetInfo().RawOutput)
}

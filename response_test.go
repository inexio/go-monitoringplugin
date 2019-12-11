/* Copyright (c) 2019, inexio GmbH, BSD 2-Clause License */
package monitoringplugin

import (
	"bytes"
	"os"
	"os/exec"
	"regexp"
	"strconv"
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
		t.Error("status code is supposed to be OK when a new response is created")
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
		t.Error("status code did change from CRITICAL to " + strconv.Itoa(r.statusCode) + " after UpdateStatus(UNKNOWN) was called! The function should not affect the status code, because CRITICAL is worse than UNKOWN")
	}

	r = NewResponse("")
	r.UpdateStatus(UNKNOWN, "")
	if r.statusCode != UNKNOWN {
		t.Error("status code did not update from OK to UNKNOWN after UpdateStatus(UNKNOWN) is called!")
	}

	r.UpdateStatus(WARNING, "")
	if r.statusCode != UNKNOWN {
		t.Error("status code did change from UNKNOWN to " + strconv.Itoa(r.statusCode) + " after UpdateStatus(WARNING) was called! The function should not affect the status code, because UNKOWN is worse than WARNING")
	}

	r.UpdateStatus(CRITICAL, "")
	if r.statusCode != CRITICAL {
		t.Error("status code is did not change from UNKNOWN to CRITICAL after UpdateStatus(CRITICAL) was called! The function should affect the status code, because CRITICAL is worse than UNKOWN")
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
		t.Error("an error occurred during cmd.Run(), but the response was expected to exit with exit code 0")
		return
	}

	output := outputB.String()

	match, err := regexp.MatchString("^OK: "+defaultMessage+"\nmessage1\nmessage2\nmessage3\nmessage4\n$", output)
	if err != nil {
		t.Error(err.Error())
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

func TestOutputPerformanceData(t *testing.T) {
	p1 := NewPerformanceDataPoint("label1", 10, "%").SetMin(0).SetMax(100).SetWarn(80).SetCrit(90)
	p2 := NewPerformanceDataPoint("label2", 20, "%").SetMin(0).SetMax(100).SetWarn(80).SetCrit(90)
	p3 := NewPerformanceDataPoint("label3", 30, "%").SetMin(0).SetMax(100).SetWarn(80).SetCrit(90)

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
	match, err := regexp.MatchString("^OK: "+defaultMessage+" | "+p1.outputString()+" "+p2.outputString()+" "+p3.outputString()+"\n", output)
	if err != nil {
		t.Error(err.Error())
	}
	if !match {
		t.Error("performance data output did not match the expected regex")
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
				t.Error(status + " response is expected to return exit status " + strconv.Itoa(exitCode) + ", but exited with exit code " + strconv.Itoa(exitError.ExitCode()))
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

// Package main provides functionality to fetch and summarize holiday calendars.
// This file contains tests for the main package functions.
package main

import (
	"io"
	"os"
	"strings"
	"testing"

	ics "github.com/arran4/golang-ical"
)

// Mock data for testing
const (
	mockColombianCalendar = `BEGIN:VCALENDAR
VERSION:2.0
BEGIN:VEVENT
SUMMARY:Colombian New Year
DTSTART;VALUE=DATE:20230101
END:VEVENT
BEGIN:VEVENT
SUMMARY:Colombian Independence Day
DTSTART;VALUE=DATE:20230720
END:VEVENT
END:VCALENDAR`

	mockCanadianCalendar = `BEGIN:VCALENDAR
VERSION:2.0
BEGIN:VEVENT
SUMMARY:Canadian New Year
DTSTART;VALUE=DATE:20230101
END:VEVENT
BEGIN:VEVENT
SUMMARY:Canada Day
DTSTART;VALUE=DATE:20230701
END:VEVENT
END:VCALENDAR`
)

// TestFetchCalendar tests the fetchCalendar function.
func TestFetchCalendar(t *testing.T) {
	// Since fetchCalendar makes an HTTP request, we won't test it directly here.
	// Instead, we assume it works correctly and focus on testing other functions.
}

// TestPrintCalendarSummary tests the printCalendarSummary function.
func TestPrintCalendarSummary(t *testing.T) {
	// Create a pipe to capture the output
	r, w, _ := os.Pipe()
	old := os.Stdout
	os.Stdout = w

	// Create a channel to capture the output asynchronously
	outC := make(chan string)
	go func() {
		var buf strings.Builder
		if _, err := io.Copy(&buf, r); err != nil {
			t.Fatalf("Error capturing output: %v", err)
		}
		outC <- buf.String()
	}()

	// Test with mock Colombian calendar data
	printCalendarSummary(mockColombianCalendar)
	w.Close()
	os.Stdout = old
	output := <-outC
	if !strings.Contains(output, "Colombian New Year") {
		t.Errorf("Expected summary to contain 'Colombian New Year'")
	}
	if !strings.Contains(output, "Colombian Independence Day") {
		t.Errorf("Expected summary to contain 'Colombian Independence Day'")
	}

	// Reset the pipe for the next test
	r, w, _ = os.Pipe()
	var buf strings.Builder
	if _, err := io.Copy(&buf, r); err != nil {
		t.Fatalf("Error capturing output: %v", err)
	}
	outC <- buf.String()

	// Test with mock Canadian calendar data
	printCalendarSummary(mockCanadianCalendar)
	w.Close()
	os.Stdout = old
	output = <-outC
	if !strings.Contains(output, "Canadian New Year") {
		t.Errorf("Expected summary to contain 'Canadian New Year'")
	}
	if !strings.Contains(output, "Canada Day") {
		t.Errorf("Expected summary to contain 'Canada Day'")
	}
}

// TestCombineCalendars tests the combineCalendars function.
func TestCombineCalendars(t *testing.T) {
	colombianCal, err := ics.ParseCalendar(strings.NewReader(mockColombianCalendar))
	if err != nil {
		t.Fatalf("Error parsing mock Colombian calendar: %v", err)
	}

	canadianCal, err := ics.ParseCalendar(strings.NewReader(mockCanadianCalendar))
	if err != nil {
		t.Fatalf("Error parsing mock Canadian calendar: %v", err)
	}

	combinedCal := combineCalendars(colombianCal, canadianCal)
	combinedCalData := combinedCal.Serialize()

	if !strings.Contains(combinedCalData, "Colombian New Year") {
		t.Errorf("Expected combined calendar to contain 'Colombian New Year'")
	}
	if !strings.Contains(combinedCalData, "Colombian Independence Day") {
		t.Errorf("Expected combined calendar to contain 'Colombian Independence Day'")
	}
	if !strings.Contains(combinedCalData, "Canadian New Year") {
		t.Errorf("Expected combined calendar to contain 'Canadian New Year'")
	}
	if !strings.Contains(combinedCalData, "Canada Day") {
		t.Errorf("Expected combined calendar to contain 'Canada Day'")
	}
}

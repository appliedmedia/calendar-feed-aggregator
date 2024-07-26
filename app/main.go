// main.go
// Package holidays provides functionality to fetch and summarize holiday calendars.
package main

import (
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"sort"
	"strings"
	"sync"

	"github.com/appliedmedia/calendar-feed-aggregator/fetcher"

	ics "github.com/arran4/golang-ical"
	"github.com/gin-gonic/gin"
	"gopkg.in/yaml.v2"
)

const (
	// ColombianHolidaysURL is the URL to fetch Colombian holidays in iCalendar format.
	ColombianHolidaysURL = "https://www.officeholidays.com/ics/ics_country.php?tbl_country=Colombia"
	// CanadianHolidaysURL is the URL to fetch Canadian holidays in iCalendar format.
	CanadianHolidaysURL = "https://www.officeholidays.com/ics/ics_country.php?tbl_country=Canada"
)

// fetchCalendar fetches the calendar data from the given URL and returns it as a string.
//
// Parameters:
// - url: The URL from which to fetch the calendar data.
//
// Returns:
// - A string containing the calendar data.
// - An error if there was an issue fetching or reading the data.
func fetchCalendar(url string) (string, error) {
	resp, err := http.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	return string(body), nil
}

// printCalendarSummary parses the calendar data and prints a summary of the events.
//
// Parameters:
// - calendarData: A string containing the calendar data to be parsed and summarized.
func printCalendarSummary(calendarData string) {
	cal, err := ics.ParseCalendar(strings.NewReader(calendarData))
	if err != nil {
		fmt.Println("Error parsing calendar:", err)
		return
	}

	totalLines := len(strings.Split(calendarData, "\n"))
	events := cal.Events()
	totalEvents := len(events)
	fmt.Printf("Total number of lines: %d\n", totalLines)
	fmt.Printf("Total number of events: %d\n", totalEvents)

	if totalEvents == 0 {
		return
	}

	printEventSummary(events, 0, "First")

	if totalEvents > 2 {
		middleIndex := totalEvents / 2
		printEventSummary(events, middleIndex, "Middle")
	}

	if totalEvents >= 2 {
		printEventSummary(events, totalEvents-1, "Last")
	}
}

// printEventSummary prints the summary of a specific event given its index and position.
//
// Parameters:
// - events: A slice of pointers to VEvent objects.
// - index: The index of the event to be summarized.
// - position: A string indicating the position of the event (e.g., "First", "Middle", "Last").
func printEventSummary(events []*ics.VEvent, index int, position string) {
	event := events[index]
	if summary := event.GetProperty(ics.ComponentPropertySummary); summary != nil {
		fmt.Printf("%s Event (Entry #%d): SUMMARY: %s\n", position, index+1, summary.Value)
	}
}

// combineCalendars combines two iCalendar objects into one, sorting the events chronologically.
//
// Parameters:
// - cal1: The first iCalendar object.
// - cal2: The second iCalendar object.
//
// Returns:
// - A new iCalendar object containing all events from both input calendars, sorted chronologically.
func combineCalendars(cal1, cal2 *ics.Calendar) *ics.Calendar {
	combinedCal := ics.NewCalendar()

	events := append(cal1.Events(), cal2.Events()...)

	sort.Slice(events, func(i, j int) bool {
		startTimeI := events[i].GetProperty(ics.ComponentPropertyDtStart).Value
		startTimeJ := events[j].GetProperty(ics.ComponentPropertyDtStart).Value
		return startTimeI < startTimeJ
	})

	for _, event := range events {
		combinedCal.AddVEvent(event)
	}

	return combinedCal
}

// main is the entry point of the program.
func mainVersion1() {
	fmt.Println("Calendar Feed Aggregator")

	colombianFeed, err := fetchCalendar(ColombianHolidaysURL)
	if err != nil {
		fmt.Println("Error fetching Colombian holidays:", err)
		return
	}

	canadianFeed, err := fetchCalendar(CanadianHolidaysURL)
	if err != nil {
		fmt.Println("Error fetching Canadian holidays:", err)
		return
	}

	fmt.Println("Colombian Holidays Feed Summary:")
	printCalendarSummary(colombianFeed)

	fmt.Println("Canadian Holidays Feed Summary:")
	printCalendarSummary(canadianFeed)

	colombianCal, err := ics.ParseCalendar(strings.NewReader(colombianFeed))
	if err != nil {
		fmt.Println("Error parsing Colombian calendar:", err)
		return
	}

	canadianCal, err := ics.ParseCalendar(strings.NewReader(canadianFeed))
	if err != nil {
		fmt.Println("Error parsing Canadian calendar:", err)
		return
	}

	combinedCal := combineCalendars(colombianCal, canadianCal)
	combinedCalData := combinedCal.Serialize()

	fmt.Println("Combined Holidays Feed Summary:")
	printCalendarSummary(combinedCalData)
}

type Config struct {
	ICS struct {
		Header string `yaml:"header"`
		Footer string `yaml:"footer"`
	} `yaml:"ics"`
}

var config Config

func init() {
	data, err := ioutil.ReadFile("conf.yaml")
	if err != nil {
		log.Fatalf("Error reading config file: %v", err)
	}
	if err := yaml.Unmarshal(data, &config); err != nil {
		log.Fatalf("Error parsing config file: %v", err)
	}
}

func writeICSHeader(c *gin.Context) {
	c.Writer.Write([]byte(config.ICS.Header))
}

func writeICSFooter(c *gin.Context) {
	c.Writer.Write([]byte(config.ICS.Footer))
}

// aggregateICS handles the aggregation of ICS files and streams the combined events.
func aggregateICS(c *gin.Context) {
	icsURLs := []string{ColombianHolidaysURL, CanadianHolidaysURL}
	eventChan := make(chan string)
	var wg sync.WaitGroup

	// Fetch calendars concurrently
	for _, url := range icsURLs {
		wg.Add(1)
		go func(url string) {
			defer wg.Done()
			fetcher.FetchICS(url, eventChan)
		}(url)
	}

	// Close the channel once all goroutines are done
	go func() {
		wg.Wait()
		close(eventChan)
	}()

	// Write the ICS header
	writeICSHeader(c)

	// Stream events to the client
	c.Stream(func(w io.Writer) bool {
		if event, ok := <-eventChan; ok {
			c.Writer.Write([]byte(event))
			return true
		}
		return false
	})

	// Write the ICS footer
	writeICSFooter(c)
}

func main() {
	r := gin.Default()
	r.GET("/aggregate_ics", aggregateICS)
	r.Run(":8080")
}

// End, main.go

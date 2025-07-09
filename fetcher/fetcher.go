// fetcher.go

package fetcher

import (
	"bufio"
	"io"
	"net/http"
)

// FetchICS handles the reading of ICS files and streams the individual events back.
// TODO: If there's a VTIMEZONE in a file, may need to add TZID to the VEVENTS, such as:
// BEGIN:VTIMEZONE
// TZID:America/New_York
// ...
// END:VTIMEZONE
// BEGIN:VEVENT
// DTSTART;TZID=America/New_York:20231010T090000
// DTEND;TZID=America/New_York:20231010T100000
// SUMMARY:Event in New York
// ...
// END:VEVENT
func FetchICS(url string, eventChan chan<- string) {
	resp, err := http.Get(url)
	if err != nil {
		eventChan <- "Error fetching URL: " + url
		return
	}
	defer resp.Body.Close()

	reader := bufio.NewReader(resp.Body)
	var event string
	inEvent := false

	for {
		line, err := reader.ReadString('\n')
		if err != nil && err != io.EOF {
			eventChan <- "Error reading response body: " + url
			return
		}
		if err == io.EOF {
			break
		}

		// Check for the beginning of an event
		if line == "BEGIN:VEVENT\r\n" || line == "BEGIN:VEVENT\n" {
			inEvent = true
			event = line
			continue
		}

		// Accumulate event data if within an event
		if inEvent {
			event += line
			if line == "END:VEVENT\r\n" || line == "END:VEVENT\n" {
				eventChan <- event
				inEvent = false
				event = ""
			}
		}
	}

	// Send any remaining event data
	if event != "" {
		eventChan <- event
	}
}

// End, fetcher.go

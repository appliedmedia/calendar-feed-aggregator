// fetcher.go

package fetcher

import (
	"io"
	"sync"

	"github.com/gin-gonic/gin"
)

const (
	// ColombianHolidaysURL is the URL to fetch Colombian holidays in iCalendar format.
	ColombianHolidaysURL = "https://www.officeholidays.com/ics/ics_country.php?tbl_country=Colombia"
	// CanadianHolidaysURL is the URL to fetch Canadian holidays in iCalendar format.
	CanadianHolidaysURL = "https://www.officeholidays.com/ics/ics_country.php?tbl_country=Canada"
)

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
			fetchICS(url, eventChan)
		}(url)
	}

	// Close the channel once all goroutines are done
	go func() {
		wg.Wait()
		close(eventChan)
	}()

	// Stream events to the client
	c.Stream(func(w io.Writer) bool {
		if event, ok := <-eventChan; ok {
			c.Writer.Write([]byte(event))
			return true
		}
		return false
	})
}

func main() {
	r := gin.Default()
	r.GET("/aggregate_ics", aggregateICS)
	r.Run(":8080")
}

// End, fetcher.go

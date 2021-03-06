// Package zoom provides a way to fetch the next Zoom meeting in your Google calendar.
package zoom

import (
	"bytes"
	"fmt"
	"net/url"
	"regexp"
	"strconv"
	"time"

	humanize "github.com/dustin/go-humanize"
	"github.com/pkg/errors"
	calendar "google.golang.org/api/calendar/v3"
)

const googleCalendarDateTimeFormat = time.RFC3339

var zoomURLRegexp = regexp.MustCompile(`https://.*?\.zoom\.us/(?:j/(\d+)|my/(\S+))`)

// NextEvent returns the next calendar event in your primary calendar.
// It will list at most 10 events, and select the first one with a Zoom URL if one exists.
func NextEvent(service *calendar.Service) (*calendar.Event, error) {
	t := time.Now().Format(time.RFC3339)

	events, err := service.Events.
		List("primary").
		ShowDeleted(false).
		SingleEvents(true).
		TimeMin(t).
		MaxResults(10).
		OrderBy("startTime").
		Do()
	if err != nil {
		return nil, errors.WithStack(err)
	}

	if len(events.Items) == 0 {
		return nil, nil
	}

	for _, event := range events.Items {
		if _, ok := MeetingURLFromEvent(event); ok {
			return event, nil
		}
	}

	// We couldn't find an event with a Zoom URL, so just return the first event.
	return events.Items[0], nil
}

// MeetingURLFromEvent returns a URL if the event is a Zoom meeting.
func MeetingURLFromEvent(event *calendar.Event) (*url.URL, bool) {
	matches := zoomURLRegexp.FindAllStringSubmatch(event.Location+" "+event.Description, -1)
	if len(matches) == 0 || len(matches[0]) == 0 {
		return nil, false
	}

	// By default, match the whole URL.
	stringURL := matches[0][0]

	// If we have a meeting ID in the URL, then use zoommtg:// instead of the HTTPS URL.
	if len(matches[0]) >= 2 {
		if _, err := strconv.Atoi(matches[0][1]); err == nil {
			stringURL = "zoommtg://zoom.us/join?confno=" + matches[0][1]
		}
	}

	parsedURL, err := url.Parse(stringURL)
	if err != nil {
		return nil, false
	}
	return parsedURL, true
}

// IsMeetingSoon returns true if the meeting is less than 5 minutes from now.
func IsMeetingSoon(event *calendar.Event) bool {
	startTime, err := MeetingStartTime(event)
	if err != nil {
		return false
	}
	minutesUntilStart := time.Until(startTime).Minutes()
	return -5 < minutesUntilStart && minutesUntilStart < 5
}

// HumanizedStartTime converts the event's start time to a human-friendly statement.
func HumanizedStartTime(event *calendar.Event) string {
	startTime, err := MeetingStartTime(event)
	if err != nil {
		return err.Error()
	}
	return humanize.Time(startTime)
}

// MeetingStartTime returns the calendar event's start time.
func MeetingStartTime(event *calendar.Event) (time.Time, error) {
	if event == nil || event.Start == nil || event.Start.DateTime == "" {
		return time.Time{}, errors.New("event does not have a start datetime")
	}
	return time.Parse(googleCalendarDateTimeFormat, event.Start.DateTime)
}

// MeetingSummary generates a one-line summary of the meeting as a string.
func MeetingSummary(event *calendar.Event) string {
	if event == nil {
		return ""
	}

	var output bytes.Buffer

	if event.Summary != "" {
		fmt.Fprintf(&output, "Your next meeting is %q", event.Summary)
	} else {
		fmt.Fprint(&output, "You have a meeting coming up")
	}

	if event.Organizer != nil && event.Organizer.DisplayName != "" {
		fmt.Fprintf(&output, ", organized by %s.", event.Organizer.DisplayName)
	} else if event.Creator != nil && event.Creator.DisplayName != "" {
		fmt.Fprintf(&output, ", created by %s.", event.Creator.DisplayName)
	} else {
		fmt.Fprintf(&output, ".")
	}

	return output.String()
}

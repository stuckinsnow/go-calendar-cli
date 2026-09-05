package demo

import "time"

// Calendar colours reused by the fixtures below.
const (
	colorWork     = "#7D56F4"
	colorFocus    = "#43BF6D"
	colorPersonal = "#EE6FF8"
	colorFamily   = "#F5A623"
	colorTravel   = "#00B8D4"
	colorAlert    = "#FF5F87"
)

// weeklyTemplate describes an event that repeats on given weekdays.
type weeklyTemplate struct {
	Title    string
	Location string
	Calendar string
	Color    string
	Hour     int
	Minute   int
	Minutes  int
	Weekdays []time.Weekday
}

func (t weeklyTemplate) at() time.Duration {
	return time.Duration(t.Hour)*time.Hour + time.Duration(t.Minute)*time.Minute
}

func (t weeklyTemplate) occursOn(d time.Weekday) bool {
	for _, w := range t.Weekdays {
		if w == d {
			return true
		}
	}
	return false
}

var weekdays = []time.Weekday{time.Monday, time.Tuesday, time.Wednesday, time.Thursday, time.Friday}

var weeklyTemplates = []weeklyTemplate{
	{Title: "Standup", Location: "Google Meet", Calendar: "Work", Color: colorWork,
		Hour: 9, Minute: 30, Minutes: 15, Weekdays: weekdays},
	{Title: "Deep work block", Calendar: "Work", Color: colorFocus,
		Hour: 10, Minutes: 120, Weekdays: []time.Weekday{time.Monday, time.Wednesday, time.Friday}},
	{Title: "Sprint planning", Location: "Google Meet", Calendar: "Work", Color: colorTravel,
		Hour: 11, Minutes: 45, Weekdays: []time.Weekday{time.Monday}},
	{Title: "1:1 with Dana", Location: "Room 4B", Calendar: "Work", Color: colorAlert,
		Hour: 14, Minutes: 30, Weekdays: []time.Weekday{time.Tuesday}},
	{Title: "Architecture review", Location: "Zoom", Calendar: "Work", Color: colorWork,
		Hour: 15, Minute: 30, Minutes: 60, Weekdays: []time.Weekday{time.Thursday}},
	{Title: "Climbing", Location: "The Depot", Calendar: "Personal", Color: colorPersonal,
		Hour: 18, Minute: 30, Minutes: 90, Weekdays: []time.Weekday{time.Tuesday, time.Thursday}},
	{Title: "Farmers market", Calendar: "Personal", Color: colorFamily,
		Hour: 10, Minute: 30, Minutes: 60, Weekdays: []time.Weekday{time.Saturday}},
}

// oneOff describes a single event positioned relative to today.
type oneOff struct {
	Title       string
	Location    string
	Calendar    string
	Color       string
	Description string
	DayOffset   int
	Hour        int
	Minute      int
	Minutes     int
	AllDay      bool
}

var oneOffs = []oneOff{
	{Title: "Coffee with Sam", Location: "Neighbourhood Roast", Calendar: "Personal", Color: colorPersonal,
		DayOffset: 0, Hour: 9, Minutes: 45, Description: "Catch-up before the week starts."},
	{Title: "Grocery run", Calendar: "Personal", Color: colorFocus, DayOffset: 0, Hour: 17, Minute: 30, Minutes: 45},
	{Title: "Quarterly review", Calendar: "Work", Color: colorAlert, DayOffset: -3,
		Hour: 13, Minutes: 90, Description: "Slides due the evening before."},
	{Title: "Book club", Location: "The Gatehouse", Calendar: "Personal", Color: colorFocus,
		DayOffset: -6, Hour: 19, Minutes: 120},
	{Title: "Dentist", Location: "Ash Lane Clinic", Calendar: "Personal", Color: colorPersonal,
		DayOffset: 2, Hour: 8, Minute: 15, Minutes: 45},
	{Title: "Release v2.4", Calendar: "Work", Color: colorAlert, DayOffset: 4,
		Hour: 16, Minutes: 60, Description: "Freeze at noon, deploy window 16:00–17:00."},
	{Title: "Team offsite", Location: "Bristol", Calendar: "Work", Color: colorWork,
		DayOffset: 9, AllDay: true, Description: "Two nights; travel booked separately."},
	{Title: "Anna's birthday", Calendar: "Family", Color: colorFamily, DayOffset: 13, AllDay: true},
	{Title: "Flight to Lisbon", Location: "LHR Terminal 5", Calendar: "Travel", Color: colorTravel,
		DayOffset: 21, Hour: 6, Minutes: 180, Description: "Check-in opens 24h before departure."},
	{Title: "Conference talk", Location: "Lisbon Congress Centre", Calendar: "Work", Color: colorWork,
		DayOffset: 23, Hour: 14, Minutes: 45, Description: "30 minutes plus questions."},
}

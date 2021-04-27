package main

import "time"

// DaySchedule contains the workday information (start/end of whole day and breaks) as timestamp object
type DaySchedule struct {
	WorkStart time.Time
	WorkEnd   time.Time

	Break1Start time.Time
	Break1End   time.Time

	Break2Start time.Time
	Break2End   time.Time

	Break3Start time.Time
	Break3End   time.Time
}

// GetTotWorkTime returns the total amount (included breaks) of hours worked that day
func (ds *DaySchedule) GetTotWorkTime() time.Duration {
	return ds.WorkEnd.Sub(ds.WorkStart) - ds.GetTotBreakTime()
}

// GetTotBreakTime computes the total amount of time spent on break that day
func (ds *DaySchedule) GetTotBreakTime() time.Duration {

	break1Time := ds.Break1End.Sub(ds.Break1Start)
	break2Time := ds.Break2End.Sub(ds.Break2Start)
	break3Time := ds.Break3End.Sub(ds.Break3Start)

	return break1Time + break2Time + break3Time
}

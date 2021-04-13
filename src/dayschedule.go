package main

import "time"

type DayScheduleStr struct {
	WorkStart string
	WorkEnd   string

	Break1Start string
	Break1End   string

	Break2Start string
	Break2End   string

	Break3Start string
	Break3End   string
}

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

func (ds *DaySchedule) IsRegularBreak() bool {
	return ds.GetTotBreakTime().Minutes() == 60
}

func (ds *DaySchedule) GetTotWorkTime() time.Duration {
	return ds.WorkEnd.Sub(ds.WorkStart) - ds.GetTotBreakTime()
}

func (ds *DaySchedule) GetTotBreakTime() time.Duration {

	break1Time := ds.Break1End.Sub(ds.Break1Start)
	break2Time := ds.Break2End.Sub(ds.Break2Start)

	return break1Time + break2Time
}

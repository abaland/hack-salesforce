package main

import "time"

type DayScheduleStr struct {
	WorkStart string
	WorkEnd   string

	Break1Start string
	Break1End   string

	Break2Start string
	Break2End   string
}

type DaySchedule struct {
	WorkStart time.Time
	WorkEnd   time.Time

	Break1Start time.Time
	Break1End   time.Time

	Break2Start time.Time
	Break2End   time.Time
}

func (ds *DaySchedule) IsRegularBreak() bool {

	break1Time := ds.Break1End.Sub(ds.Break1Start).Minutes()
	break2Time := ds.Break2End.Sub(ds.Break2Start).Minutes()
	return break1Time+break2Time == 60
}

func (ds *DaySchedule) GetTotWorkTime() time.Duration {

	workTime := ds.WorkEnd.Sub(ds.WorkStart)
	break1Time := ds.Break1End.Sub(ds.Break1Start)
	break2Time := ds.Break2End.Sub(ds.Break2Start)
	workTime = workTime - break1Time - break2Time

	return workTime
}

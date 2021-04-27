package main

import (
	"reflect"
	"testing"
	"time"
)

// DayScheduleStr contains the workday information (start/end of whole day and breaks) as string
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

// ParseDailySchedule parses daily modal info and return DaySchedule instance containing parsed information
func (dss DayScheduleStr) ParseDailySchedule() DaySchedule {

	workStart, _ := time.Parse(SalesforceTimeFormat, dss.WorkStart)
	workEnd, _ := time.Parse(SalesforceTimeFormat, dss.WorkEnd)

	break1Start, _ := time.Parse(SalesforceTimeFormat, dss.Break1Start)
	break1End, _ := time.Parse(SalesforceTimeFormat, dss.Break1End)

	break2Start, _ := time.Parse(SalesforceTimeFormat, dss.Break2Start)
	break2End, _ := time.Parse(SalesforceTimeFormat, dss.Break2End)

	break3Start, _ := time.Parse(SalesforceTimeFormat, dss.Break3Start)
	break3End, _ := time.Parse(SalesforceTimeFormat, dss.Break3End)

	workSchedule := DaySchedule{
		WorkStart:   workStart,
		WorkEnd:     workEnd,
		Break1Start: break1Start,
		Break1End:   break1End,
		Break2Start: break2Start,
		Break2End:   break2End,
		Break3Start: break3Start,
		Break3End:   break3End,
	}
	return workSchedule
}

func Test_GetTotBreakTime(t *testing.T) {

	tests := []struct {
		name string
		args DayScheduleStr
		want time.Duration
	}{
		{
			name: "RegularLunchBreak",
			args: DayScheduleStr{
				WorkStart:   "09:00",
				WorkEnd:     "18:00",
				Break1Start: "13:00",
				Break1End:   "14:00",
				Break2Start: "",
				Break2End:   "",
				Break3Start: "",
				Break3End:   "",
			},
			want: 60 * time.Minute,
		},
		{
			name: "LongLunchBreak",
			args: DayScheduleStr{
				WorkStart:   "09:00",
				WorkEnd:     "18:00",
				Break1Start: "12:45",
				Break1End:   "14:15",
				Break2Start: "",
				Break2End:   "",
				Break3Start: "",
				Break3End:   "",
			},
			want: 90 * time.Minute,
		},
		{
			name: "DoubleLunchBreak",
			args: DayScheduleStr{
				WorkStart:   "09:00",
				WorkEnd:     "18:00",
				Break1Start: "08:00",
				Break1End:   "09:00",
				Break2Start: "13:00",
				Break2End:   "14:00",
				Break3Start: "",
				Break3End:   "",
			},
			want: 120 * time.Minute,
		},
		{
			name: "TripleLunchBreak",
			args: DayScheduleStr{
				WorkStart:   "09:00",
				WorkEnd:     "18:00",
				Break1Start: "08:00",
				Break1End:   "09:00",
				Break2Start: "13:00",
				Break2End:   "14:00",
				Break3Start: "16:00",
				Break3End:   "17:00",
			},
			want: 180 * time.Minute,
		},
	}
	for _, tt := range tests {
		ds := tt.args.ParseDailySchedule()
		totBreakTime := ds.GetTotBreakTime()
		if !reflect.DeepEqual(totBreakTime, tt.want) {
			t.Errorf("%q. ParseJson() = %v, want %v", tt.name, ds, tt.want)
		}
	}
}

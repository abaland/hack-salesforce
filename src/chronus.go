package main

import (
	"github.com/sclevine/agouti"
	"strconv"
	"time"
)

const (
	ChronusSleepTime      = 2 * time.Second
	ChronusShortSleepTime = 1 * time.Millisecond

	ChronusLoginUrl = `https://chronus-ext.tis.co.jp/Lysithea/Logon`

	// Html Attribute Name In Login Menu
	ChronusUserNameSelector    = `document.FORM_COMMON.PersonCode.value`
	ChronusPasswordSelector    = `document.FORM_COMMON.Password.value`
	ChronusLoginSubmitSelector = `a`

	ChronusCalendarFrameName    = "MENU"
	ChronusDayScheduleFrameName = "OPERATION"
	ChronusClickableDays        = `td.calCellNotRegistration a.calLinkWeekDay`
	ChronusMaxDays              = 31 // After removing weekend, 22 is max?

	ChronusWorkStartTimeSelector = `input[type="text"][name="StartTime"]`
	ChronusWorkEndTimeSelector   = `input[type="text"][name="EndTime"]`
	ChronusTimeFormat            = "1504" // min:sec
)

type chronus struct {
	Account account
	Page    *agouti.Page
}

func (d *Driver) NewChronus(username, password string) (*chronus, error) {
	page, err := d.NewPage()
	if err != nil {
		return nil, err
	}
	if err := page.Navigate(ChronusLoginUrl); err != nil {
		return nil, err
	}
	return &chronus{
		Account: account{
			UserName: username,
			Password: password,
		},
		Page: page,
	}, nil
}

func (ch *chronus) Login() error {
	// ID, Passの要素を取得し、値を設定
	noScriptArgs := map[string]interface{}{}
	_ = ch.Page.RunScript(ChronusUserNameSelector+"= \""+ch.Account.UserName+"\"", noScriptArgs, nil)
	_ = ch.Page.RunScript(ChronusPasswordSelector+"= \""+ch.Account.Password+"\"", noScriptArgs, nil)
	// formをサブミット
	if err := ch.Page.Find(ChronusLoginSubmitSelector).Click(); err != nil {
		return err
	}

	time.Sleep(ChronusSleepTime)
	return nil

}

func (ds *DaySchedule) ToChronus() DayScheduleStr {
	return DayScheduleStr{
		ds.WorkStart.Format(ChronusTimeFormat),
		ds.WorkEnd.Format(ChronusTimeFormat),
		ds.Break1Start.Format(ChronusTimeFormat),
		ds.Break1End.Format(ChronusTimeFormat),
		ds.Break2Start.Format(ChronusTimeFormat),
		ds.Break2End.Format(ChronusTimeFormat),
	}
}

func (ch *chronus) RegisterWork(workMonth []workday) error {

	_ = ch.Page.ConfirmPopup()
	time.Sleep(ChronusShortSleepTime)

	// First switch frame to focus on the calendar one (without that, we cannot select the items inside)
	calendarFrame := ch.Page.FindByName(ChronusCalendarFrameName)
	dayFrame := ch.Page.FindByName(ChronusDayScheduleFrameName)
	_ = calendarFrame.SwitchToFrame()
	editableDays := ch.Page.All(ChronusClickableDays)

	for i := 0; i < ChronusMaxDays; i++ {
		dayAsText, _ := editableDays.At(i).Text()
		dayAsInt, _ := strconv.Atoi(dayAsText)

		for _, workDay := range workMonth {
			if workDay.DayIdx == dayAsInt {

				chronusSchedule := workDay.WorkSchedule.ToChronus()

				print(workDay.Day)
				_ = editableDays.At(i).Click()

				_ = ch.Page.SwitchToRootFrame()
				_ = dayFrame.SwitchToFrame()

				_ = ch.Page.Find(ChronusWorkStartTimeSelector).Fill(chronusSchedule.WorkStart)
				_ = ch.Page.Find(ChronusWorkEndTimeSelector).Fill(chronusSchedule.WorkEnd)

				_ = ch.Page.SwitchToRootFrame()
				_ = calendarFrame.SwitchToFrame()

			}
		}
	}

	return nil
}

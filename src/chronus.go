package main

import (
	"fmt"
	"github.com/sclevine/agouti"
	"math"
	"strconv"
	"strings"
	"time"
)

const (
	ChronusLoginUrl = `https://chronus-ext.tis.co.jp/Lysithea/Logon`

	// ChronusUserNameScript defines script to run to update chronus username field
	ChronusUserNameScript = `document.FORM_COMMON.PersonCode.value="%s"`
	// ChronusPasswordScript defines script to run to update chronus password field
	ChronusPasswordScript = `document.FORM_COMMON.Password.value="%s"`
	// ChronusLoginSubmitQuery defines selector to get login button in login page
	ChronusLoginSubmitQuery = `a`

	ChronusCalendarFrameName    = "MENU"
	ChronusDayScheduleFrameName = "OPERATION"
	ChronusClickableDays        = `td.calCellNotRegistration a.calLinkWeekDay`
	ChronusMaxDays              = 31 // After removing weekend, 22 is max?

	ChronusWorkStartTimeSelector = `input[type="text"][name="StartTime"]`
	ChronusWorkEndTimeSelector   = `input[type="text"][name="EndTime"]`

	// All 3 breaks have the same selector!
	ChronusWorkBreakStartSelector = `td input.InputTxtR[name="PrivateStart"]`
	ChronusWorkBreakEndSelector   = `td input.InputTxtR[name="PrivateEnd"]`

	ChronusCommentSelector = `input[type="text"][name="Comment"]`

	ChronusShukouCode            = "00003L3:他社出向業務"
	ChronusProjectSelectSelector = `select[name="CostNoItem"]`
	ChronusProjectHourSelector   = `input[type="text"][name="CostQuantity"]`

	ChronusScanStartSelector   = `input[type="text"][name="StartTimeStamp"]`
	ChronusWorkTypeSelector    = `select[name="AllowanceItem"]`
	ChronusWorkTypeCompanyName = `出社`
	ChronusWorkTypeRemoteName  = `フルテレワーク`

	ChronusRegisterScript          = `top.dosubmitRegister()`
	ChronusCalendarRefreshSelector = `img[src="../gif/saihyoji.gif"]`

	// ChronusRegisteredTopBarScript using the .Text() property did not get the text correctly, so run JS to do it
	ChronusRegisteredTopBarScript  = "return document.querySelector(`td[align=\"CENTER\"]`).innerText"
	ChronusRegisteredErrorSelector = `font[color="red"]`

	ChronusTimeFormat = "1504" // min:sec
)

type chronus struct {
	Page *agouti.Page
}

// NewChronus creates a new page from the browser driver and opens the Chronus webpage, returning it into a
// chronus instance
func (d *Driver) NewChronus() (*chronus, error) {
	page, err := d.NewPage()
	if err != nil {
		return nil, err
	}
	if err := page.Navigate(ChronusLoginUrl); err != nil {
		return nil, err
	}
	return &chronus{
		Page: page,
	}, nil
}

// Login uses credentials received to log into chronus
func (ch *chronus) Login(credentials Credentials) error {
	// ID, Passの要素を取得し、値を設定
	_ = ch.Page.RunScript(fmt.Sprintf(ChronusUserNameScript, credentials.User), nil, nil)
	_ = ch.Page.RunScript(fmt.Sprintf(ChronusPasswordScript, credentials.Password), nil, nil)
	// formをサブミット
	if err := ch.Page.Find(ChronusLoginSubmitQuery).Click(); err != nil {
		return err
	}

	return nil

}

// fmtDuration formats time.Duration object into string formatted for Chronus input (7h30->0730)
func fmtDuration(d time.Duration) string {
	d = d.Round(time.Minute)
	h := d / time.Hour
	d -= h * time.Hour
	m := d / time.Minute
	return fmt.Sprintf("%02d%02d", h, m)
}

// GetChronusBreaks adjusts break times from a DaySchedule instead to make them fit into Chronus' 中断 slots.
// Chronus contains a non-editable non-visible lunch break from 12:00 to 13:00, so one of the input breaks need to move
// to fill that slot.
// In order, the steps taken are:
// 1) remove empty breaks,
// 2) get break closest to lunch time and move its started at 12:00
// 3) leave other breaks as is
func (ds *DaySchedule) GetChronusBreaks() ([][]time.Time, error) {

	var err error

	lunchTime, _ := time.Parse(ChronusTimeFormat, "1230")
	chronusBreaks := [][]time.Time{
		{ds.Break1Start, ds.Break1End},
		{ds.Break2Start, ds.Break2End},
		{ds.Break3Start, ds.Break3End},
	}

	// Remove all empty breaks
	firstEmptyIdx := len(chronusBreaks)
	for i := 0; i < len(chronusBreaks); i++ {
		if chronusBreaks[i][0].Hour() == 0 {
			firstEmptyIdx = i
			break
		}
	}
	chronusBreaks = chronusBreaks[:firstEmptyIdx]

	// Move break closest to lunch and set it to lunch time
	lunchClosestIdx := 0
	lunchDistance := 24 * 60.
	for i := 0; i < len(chronusBreaks); i++ {
		lunchTimeDistance1 := math.Abs(lunchTime.Sub(chronusBreaks[i][0]).Minutes())
		lunchTimeDistance2 := math.Abs(lunchTime.Sub(chronusBreaks[i][1]).Minutes())
		if math.Min(lunchTimeDistance1, lunchTimeDistance2) < lunchDistance {
			lunchClosestIdx = i
			lunchDistance = math.Min(lunchTimeDistance1, lunchTimeDistance2)
		}
	}

	closestBreakDuration := chronusBreaks[lunchClosestIdx][1].Sub(chronusBreaks[lunchClosestIdx][0])
	if closestBreakDuration.Minutes() == 60 {
		chronusBreaks = append(chronusBreaks[:lunchClosestIdx], chronusBreaks[lunchClosestIdx+1:]...)
	} else if closestBreakDuration.Minutes() < 60 {
		err = fmt.Errorf("lunch break less than 60min. Giving up")
	} else {
		chronusBreaks[lunchClosestIdx][0], _ = time.Parse(ChronusTimeFormat, "1200")
		chronusBreaks[lunchClosestIdx][1] = chronusBreaks[lunchClosestIdx][0].Add(closestBreakDuration)
	}

	return chronusBreaks, err
}

// RegisterWorkOneDay fills in information for a specific day in Chronus and submits the form.
func (ch *chronus) RegisterWorkOneDay(workDay workday) error {

	ws := workDay.WorkSchedule

	// Fill-in top of the page
	err := ch.Page.Find(ChronusWorkStartTimeSelector).Fill(ws.WorkStart.Format(ChronusTimeFormat))
	if err != nil {
		fmt.Println(err.Error())
	}
	err = ch.Page.Find(ChronusWorkEndTimeSelector).Fill(ws.WorkEnd.Format(ChronusTimeFormat))
	if err != nil {
		fmt.Println(err.Error())
	}

	// Fill-in リモート・出社
	scanStart, err := ch.Page.Find(ChronusScanStartSelector).Attribute("value")
	if scanStart == "" {
		err = ch.Page.Find(ChronusWorkTypeSelector).Select(ChronusWorkTypeRemoteName)
		if err != nil {
			fmt.Println(err.Error())
		}
	} else {
		err = ch.Page.Find(ChronusWorkTypeSelector).Select(ChronusWorkTypeCompanyName)
		if err != nil {
			fmt.Println(err.Error())
		}
	}

	// Fill in 中断 section
	totBreakTime := ws.GetTotBreakTime()
	if totBreakTime.Minutes() != 60 {

		breakStartInputs := ch.Page.All(ChronusWorkBreakStartSelector)
		breakEndInputs := ch.Page.All(ChronusWorkBreakEndSelector)
		breaksInfo, err := ws.GetChronusBreaks()
		for breakIdx, breakInfo := range breaksInfo {

			breakStartTimeStr := breakInfo[0].Format(ChronusTimeFormat)
			breakEndTimeStr := breakInfo[1].Format(ChronusTimeFormat)

			err = breakStartInputs.At(breakIdx).Fill(breakStartTimeStr)
			if err != nil {
				fmt.Println(err.Error())
			}
			err = breakEndInputs.At(breakIdx).Fill(breakEndTimeStr)
			if err != nil {
				fmt.Println(err.Error())
			}
		}
	}

	// Fill in bottom section
	err = ch.Page.All(ChronusProjectSelectSelector).At(0).Select(ChronusShukouCode)
	if err != nil {
		fmt.Println(err.Error())
	}
	err = ch.Page.All(ChronusProjectHourSelector).At(0).Fill(fmtDuration(ws.GetTotWorkTime()))
	if err != nil {
		fmt.Println(err.Error())
	}

	// Fill in 備考 section
	err = ch.Page.Find(ChronusCommentSelector).Fill(workDay.WorkComment)
	if err != nil {
		fmt.Println(err.Error())
	}

	// Submit
	err = ch.Page.RunScript(ChronusRegisterScript, nil, nil)

	return nil
}

// isChronusLoginFinished checks browser page to see if the page following log-in is done loading
func isChronusLoginFinished(Page *agouti.Page) bool {
	_ = Page.ConfirmPopup()
	count, _ := Page.FindByName(ChronusCalendarFrameName).Count()
	return count > 0
}

// isDayRegistered checks browser page to see if the chronus day we just registered is done being registered.
// An error to register is also considered as "done". The goal here is to make sure that whatever we input is done
// processing
func isDayRegistered(Page *agouti.Page) bool {
	var topBarText string
	_ = Page.RunScript(ChronusRegisteredTopBarScript, nil, &topBarText)
	hasSuccess := strings.Contains(topBarText, "△")

	hasErrorCount, _ := Page.FindByID(ChronusRegisteredErrorSelector).Count()
	return hasErrorCount > 0 || hasSuccess
}

// ReclickOnDay switches back to the calendar frame and redoes a click on the day selected.
// This function was added due to inconsistent behavior between runs where the day page did not show up after clicking
// for unknown reason
func (ch *chronus) ReclickOnDay(calendarFrame *agouti.Selection, dayFrame *agouti.Selection, editableDay *agouti.Selection) error {

	time.Sleep(100 * time.Millisecond)

	// Switches back to calendar frame
	err := ch.Page.SwitchToRootFrame()
	if err != nil {
		fmt.Println(err.Error())
	}
	err = calendarFrame.SwitchToFrame()
	if err != nil {
		fmt.Println(err.Error())
	}

	// Reclick on day
	err = editableDay.Click()
	if err != nil {
		fmt.Println(err.Error())
	}

	// Switches back to day frame
	err = ch.Page.SwitchToRootFrame()
	if err != nil {
		fmt.Println(err.Error())
	}
	err = dayFrame.SwitchToFrame()
	if err != nil {
		fmt.Println(err.Error())
	}

	return nil
}

// RegisterWork registers unregistered days in the chronus calendar using the input workMonth
func (ch *chronus) RegisterWork(workMonth []workday) error {

	_ = sleepUntil(isChronusLoginFinished, ch.Page, SalesforceMaxSleepTime)
	time.Sleep(100 * time.Millisecond)

	// First switch frame to focus on the calendar one (without that, we cannot select the items inside)
	calendarFrame := ch.Page.FindByName(ChronusCalendarFrameName)
	dayFrame := ch.Page.FindByName(ChronusDayScheduleFrameName)
	err := calendarFrame.SwitchToFrame()
	if err != nil {
		fmt.Println(err.Error())
	}
	editableDays := ch.Page.All(ChronusClickableDays)

	// We fill chronus backwards due to At() refreshing everytime the selection, thus skipping every other item
	for i := ChronusMaxDays; i >= 0; i-- {
		editableDay := editableDays.At(i)
		dayAsText, err := editableDay.Text()
		if err != nil {
			// Error here are due to out-of-range days (just skip those)
			continue
		}
		dayAsInt, err := strconv.Atoi(dayAsText)

		for _, workDay := range workMonth {
			if workDay.DayIdx == dayAsInt && workDay.WorkSchedule.WorkEnd.Hour() > 0 {

				// Clicks on day in calendar frame, then move to the day page
				fmt.Println(workDay.Day)
				err = editableDay.Click()
				if err != nil {
					fmt.Println(err.Error())
				}

				err = ch.Page.SwitchToRootFrame()
				if err != nil {
					fmt.Println(err.Error())
				}
				err = dayFrame.SwitchToFrame()
				if err != nil {
					fmt.Println(err.Error())
				}

				// Handles potential timing error
				count, _ := ch.Page.Find(ChronusWorkStartTimeSelector).Count()
				if count == 0 {
					fmt.Println("Reclicking!")
					err := ch.ReclickOnDay(calendarFrame, dayFrame, editableDay)
					if err != nil {
						return err
					}
				}

				// If workschedule has this day as non-empty, fill it
				if workDay.WorkSchedule.WorkStart.Hour() != 0 {
					err = ch.RegisterWorkOneDay(workDay)
					if err != nil {
						fmt.Println(err.Error())
					}
				}

				_ = sleepUntil(isDayRegistered, ch.Page, SalesforceMaxSleepTime)

				// Moves back to calendar frame
				err = ch.Page.SwitchToRootFrame()
				if err != nil {
					fmt.Println(err.Error())
				}
				err = calendarFrame.SwitchToFrame()
				if err != nil {
					fmt.Println(err.Error())
				}

				// Refresh calendar frame
				err = ch.Page.Find(ChronusCalendarRefreshSelector).Click()
				if err != nil {
					fmt.Println(err.Error())
				}

				break
			}
		}

	}

	return nil
}

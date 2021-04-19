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

func (ch *chronus) Login(credentials Credentials) error {
	// ID, Passの要素を取得し、値を設定
	noScriptArgs := map[string]interface{}{}
	_ = ch.Page.RunScript(ChronusUserNameSelector+"= \""+credentials.User+"\"", noScriptArgs, nil)
	_ = ch.Page.RunScript(ChronusPasswordSelector+"= \""+credentials.Password+"\"", noScriptArgs, nil)
	// formをサブミット
	if err := ch.Page.Find(ChronusLoginSubmitSelector).Click(); err != nil {
		return err
	}

	return nil

}

func isChronusLoginFinished(Page *agouti.Page) bool {
	_ = Page.ConfirmPopup()
	count, _ := Page.FindByName(ChronusCalendarFrameName).Count()
	return count > 0
}

func (ds *DaySchedule) ToChronus() DayScheduleStr {
	return DayScheduleStr{
		ds.WorkStart.Format(ChronusTimeFormat),
		ds.WorkEnd.Format(ChronusTimeFormat),
		ds.Break1Start.Format(ChronusTimeFormat),
		ds.Break1End.Format(ChronusTimeFormat),
		ds.Break2Start.Format(ChronusTimeFormat),
		ds.Break2End.Format(ChronusTimeFormat),
		ds.Break3Start.Format(ChronusTimeFormat),
		ds.Break3End.Format(ChronusTimeFormat),
	}
}

func fmtDuration(d time.Duration) string {
	d = d.Round(time.Minute)
	h := d / time.Hour
	d -= h * time.Hour
	m := d / time.Minute
	return fmt.Sprintf("%02d%02d", h, m)
}

func (ds *DaySchedule) GetChronusBreaks() ([][]time.Time, error) {

	var err error

	lunchTime, _ := time.Parse("1504", "1230")
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
		chronusBreaks[lunchClosestIdx][0], _ = time.Parse("1504", "1200")
		chronusBreaks[lunchClosestIdx][1] = chronusBreaks[lunchClosestIdx][0].Add(closestBreakDuration)
	}

	return chronusBreaks, err
}

func (ch *chronus) RegisterWorkOneDay(workDay workday) error {

	ws := workDay.WorkSchedule
	chronusSchedule := ws.ToChronus()

	// Fill-in top of the page
	err := ch.Page.Find(ChronusWorkStartTimeSelector).Fill(chronusSchedule.WorkStart)
	if err != nil {
		fmt.Println(err.Error())
	}
	err = ch.Page.Find(ChronusWorkEndTimeSelector).Fill(chronusSchedule.WorkEnd)
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

			breakStartTimeStr := breakInfo[0].Format("1504")
			breakEndTimeStr := breakInfo[1].Format("1504")

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

func isDayRegistered(Page *agouti.Page) bool {
	var topBarText string
	_ = Page.RunScript(ChronusRegisteredTopBarScript, nil, &topBarText)
	hasSuccess := strings.Contains(topBarText, "△")

	hasErrorCount, _ := Page.FindByID(ChronusRegisteredErrorSelector).Count()
	return hasErrorCount > 0 || hasSuccess
}

func (ch *chronus) RegisterWork(workMonth []workday) error {

	_ = sleepUntil(isChronusLoginFinished, ch.Page, SalesforceMaxSleepTime)
	time.Sleep(100 * time.Millisecond)
	fmt.Println("Login finished")

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
		dayAsText, err := editableDays.At(i).Text()
		if err != nil {
			// Error here are due to out-of-range days (just skip those)
			continue
		}
		dayAsInt, err := strconv.Atoi(dayAsText)

		for _, workDay := range workMonth {
			if workDay.DayIdx == dayAsInt && workDay.WorkSchedule.WorkEnd.Hour() > 0 {

				fmt.Println(workDay.Day)
				err = editableDays.At(i).Click()
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

				if workDay.WorkSchedule.WorkStart.Hour() != 0 {
					err = ch.RegisterWorkOneDay(workDay)
					if err != nil {
						fmt.Println(err.Error())
					}
				}

				_ = sleepUntil(isDayRegistered, ch.Page, SalesforceMaxSleepTime)

				err = ch.Page.SwitchToRootFrame()
				if err != nil {
					fmt.Println(err.Error())
				}
				err = calendarFrame.SwitchToFrame()
				if err != nil {
					fmt.Println(err.Error())
				}

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

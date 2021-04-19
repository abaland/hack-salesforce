package main

import (
	"fmt"
	"github.com/sclevine/agouti"
	"time"
)

const (
	SalesforceLoginUrl = `https://login.salesforce.com/`

	// SalesforceMaxSleepTime defines maximum amount of time we should wait for an event
	//to complete before raising an error
	SalesforceMaxSleepTime = 10 * time.Second

	// SalesforceShortSleepTime defines amount of time we wait for UI updates that do not require server-side processing
	SalesforceShortSleepTime = 100 * time.Nanosecond

	// SalesforceUserNameId defines ID of html tag containing the username in login page
	SalesforceUserNameId = "username"
	// SalesforcePasswordId defines ID of html tag containing the password in login page
	SalesforcePasswordId = "password"
	// SalesforceLoginId defines ID of html tag for Login page form submit button
	SalesforceLoginId = "Login"

	// WorkTabId defines ID of html <li> tag containing the 勤務表 link
	WorkTabId = "01r7F0000017C6B_Tab"

	// tableDataMaskId defines ID of div that appears in front of calendar to block input when changing months
	tableDataMaskId = "shim"

	//MonthListId defines ID of <select> tag that allows user to change input month.
	MonthListId = "yearMonthList"
	// MonthListOptionsQuery defines selectorQuery used to get the list of all available months to chose from
	MonthListOptionsQuery = "#yearMonthList option"

	// workStatusQueryPattern defines pattern for selectorQuery used to get specific day status
	// (day off, half day, ...). It assumes it has an input parameter %s, used to differentiate the days
	workStatusQueryPattern = "tr#dateRow%s td.vstatus"
	// DayOffValue corresponds to the text written in the td.vstatus tag in case the user took a day off.
	DayOffValue = "年次有給休暇"

	// workModalButtonQueryPattern defines pattern for selectorQuery used to open the daily modal info (daily 出社 cell)
	workModalButtonQueryPattern = "tr#dateRow%s td.vst"
	// workStartTimeScript defines script to run to get work start time once the daily modal info has been opened.
	workStartTimeScript = "return document.getElementById('startTime').value"
	// workEndTimeScript defines script to run to get work end time once the daily modal info has been opened.
	workEndTimeScript = "return document.getElementById('endTime').value"
	// breakStartTimeScriptPattern defines pattern for script to run to get n-th break start time once the daily modal info has been opened.
	breakStartTimeScriptPattern = "return document.getElementById('startRest%d').value"
	// breakEndTimeScriptPattern defines pattern for script to run to get n-th break start time once the daily modal info has been opened.
	breakEndTimeScriptPattern = "return document.getElementById('endRest%d').value"

	// commentQueryPattern defines query to run to get the daily 備考 value
	commentQueryPattern = "tr#dateRow%s td.vnote"

	// cancelButtonId and closeButtonId define the ID of the buttons that need to be pressed to close the daily modal.
	//  The former is there for the current month, the latter one for past month.
	cancelButtonId = "dlgInpTimeCancel" // This month
	closeButtonId  = "dlgInpTimeClose"  // Last month

	// MonthListTextFormat specifies format of the year-month to select in the calendar page.
	MonthListTextFormat = "2006年01月"
	// SalesforceDayFormat specifies format used to select day-specific items in chronus
	SalesforceDayFormat = "2006-01-02"
	// SalesforceTimeFormat specifies format of the timestamp (HH:MM) to parse in the calendar page
	SalesforceTimeFormat = "15:04" // min:sec
)

// workday holds various information about a salesforce parsed workday.
type workday struct {
	DayIdx int    // day index (1->31)
	Day    string // string representation of date corresponding to that day

	DayOff bool // whether the day was a 年休 or not

	WorkComment  string      // 備考 value for that day
	WorkSchedule DaySchedule // schedule containing work start/end and break starts/ends for that day
}

// salesforce contains basic information about the chrome salesforce instance
type salesforce struct {
	Page *agouti.Page
}

// NewSalesForce creates a new page from the browser driver and opens the Salesforce webpage, returning it into a
// salesforce instance
func (d *Driver) NewSalesForce() (*salesforce, error) {
	page, err := d.NewPage()
	if err != nil {
		return nil, err
	}
	if err := page.Navigate(SalesforceLoginUrl); err != nil {
		return nil, err
	}
	return &salesforce{
		Page: page,
	}, nil
}

// Login uses credentials received to log into salesforce
func (sf *salesforce) Login(credentials Credentials) error {
	// ID, Passの要素を取得し、値を設定
	_ = sf.Page.FindByID(SalesforceUserNameId).Fill(credentials.User)
	_ = sf.Page.FindByID(SalesforcePasswordId).Fill(credentials.Password)

	// formをサブミット
	if err := sf.Page.FindByID(SalesforceLoginId).Submit(); err != nil {
		return err
	}

	return nil

}

// ParseDailySchedule parses daily modal info and return DaySchedule instance containing parsed information
func (sf *salesforce) ParseDailySchedule() DaySchedule {

	var startText, endText string
	_ = sf.Page.RunScript(workStartTimeScript, nil, &startText)
	_ = sf.Page.RunScript(workEndTimeScript, nil, &endText)

	var break1StartText, break1EndText string
	_ = sf.Page.RunScript(fmt.Sprintf(breakStartTimeScriptPattern, 1), nil, &break1StartText)
	_ = sf.Page.RunScript(fmt.Sprintf(breakEndTimeScriptPattern, 1), nil, &break1EndText)

	var break2StartText, break2EndText string
	_ = sf.Page.RunScript(fmt.Sprintf(breakStartTimeScriptPattern, 2), nil, &break2StartText)
	_ = sf.Page.RunScript(fmt.Sprintf(breakEndTimeScriptPattern, 2), nil, &break2EndText)

	var break3StartText, break3EndText string
	_ = sf.Page.RunScript(fmt.Sprintf(breakStartTimeScriptPattern, 3), nil, &break3StartText)
	_ = sf.Page.RunScript(fmt.Sprintf(breakEndTimeScriptPattern, 3), nil, &break3EndText)

	workStart, _ := time.Parse(SalesforceTimeFormat, startText)
	workEnd, _ := time.Parse(SalesforceTimeFormat, endText)

	break1Start, _ := time.Parse(SalesforceTimeFormat, break1StartText)
	break1End, _ := time.Parse(SalesforceTimeFormat, break1EndText)

	break2Start, _ := time.Parse(SalesforceTimeFormat, break2StartText)
	break2End, _ := time.Parse(SalesforceTimeFormat, break2EndText)

	break3Start, _ := time.Parse(SalesforceTimeFormat, break3StartText)
	break3End, _ := time.Parse(SalesforceTimeFormat, break3EndText)

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

// isSalesforceLoginFinished checks browser page to see if the page following log-in is done being displayed
func isSalesforceLoginFinished(Page *agouti.Page) bool {
	count, _ := Page.FindByID(WorkTabId).Count()
	return count > 0
}

// isScheduleLoaded checks browser page to see if the list of available months for the calendar page is fully displayed
func isScheduleLoaded(Page *agouti.Page) bool {
	count, _ := Page.All(MonthListOptionsQuery).Count()
	return count > 2
}

// isMonthLoaded checks browser page to see if the calendar is done displaying information about the selected month
func isMonthLoaded(Page *agouti.Page) bool {
	displayValue, _ := Page.FindByID(tableDataMaskId).CSS("display")
	return displayValue == "none"
}

// ParseWork reads parses the work month from the salesforce calendar page and returns a list of workday instances
// Requirements: log-in done.
// Detailed actions: opens the 勤務表 tab, select the correct month in the calendar, and parses all days in the calendar
func (sf *salesforce) ParseWork() ([]workday, error) {

	// month to process
	processDate := time.Now()

	var workMonth []workday

	_ = sleepUntil(isSalesforceLoginFinished, sf.Page, SalesforceMaxSleepTime)

	// 勤務表タブをクリック
	err := sf.Page.FindByID(WorkTabId).Click()
	if err != nil {
		return workMonth, err
	}

	_ = sleepUntil(isScheduleLoaded, sf.Page, SalesforceMaxSleepTime)

	// 月を選択
	err = sf.Page.FindByID(MonthListId).Select(processDate.Format(MonthListTextFormat))
	if err != nil {
		return workMonth, err
	}

	// ちょっと待つ
	_ = sleepUntil(isMonthLoaded, sf.Page, SalesforceMaxSleepTime)

	// 勤務表の行情報イテレーション
	startDate := time.Date(processDate.Year(), processDate.Month(), 1, 0, 0, 0, 0, time.UTC)
	for d := startDate; d.Month() == processDate.Month(); d = d.AddDate(0, 0, 1) {

		day := d.Format(SalesforceDayFormat)
		workStatusSelector := fmt.Sprintf(workStatusQueryPattern, day)
		workStatusText, _ := sf.Page.Find(workStatusSelector).Attribute("title")
		dayOffBool := workStatusText == DayOffValue

		commentID := fmt.Sprintf(commentQueryPattern, day)
		workComment, _ := sf.Page.Find(commentID).Attribute("title")

		workdayDetails := workday{
			DayIdx:      d.Day(),
			Day:         day,
			DayOff:      dayOffBool,
			WorkComment: workComment,
		}
		// In case the day is not a 年休, opens daily modal and parses its information
		if !dayOffBool {
			workModalButtonSelect := fmt.Sprintf(workModalButtonQueryPattern, day)
			_ = sf.Page.Find(workModalButtonSelect).Click()
			time.Sleep(SalesforceShortSleepTime)

			// The following values are in input.value, not in an attribute, so we need JS to get them
			workdayDetails.WorkSchedule = sf.ParseDailySchedule()

			// Button changes if past month data is displayed => try both cases
			_ = sf.Page.FindByID(cancelButtonId).Click()
			_ = sf.Page.FindByID(closeButtonId).Click()
			time.Sleep(SalesforceShortSleepTime)
		}

		workMonth = append(workMonth, workdayDetails)
	}

	_ = sf.Page.CloseWindow()
	return workMonth, nil
}

package main

import (
	"fmt"
	"github.com/sclevine/agouti"
	"time"
)

const (
	SalesforceSleepTime      = 3 * time.Second
	SalesforceShortSleepTime = 500 * time.Nanosecond

	SalesforceLoginUrl = `https://login.salesforce.com/`

	// Login menu info
	SalesforceUserNameID = "username"
	SalesforcePasswordID = "password"
	SalesforceLoginID    = "Login"

	// Main Menu Info
	WorkTabID = "01r7F0000017C6B_Tab"

	// Schedule Menu info
	MonthListID               = "yearMonthList"
	workStatusSelectorPattern = "tr#dateRow%s td.vstatus"
	DayOffValue               = "年次有給休暇"

	workModalButtonSelectPattern = "tr#dateRow%s td.vst"
	workStartTimeScript          = "return document.getElementById('startTime').value"
	workEndTimeScript            = "return document.getElementById('endTime').value"

	break1StartTimeScript = "return document.getElementById('startRest1').value"
	break1EndTimeScript   = "return document.getElementById('endRest1').value"
	break2StartTimeScript = "return document.getElementById('startRest2').value"
	break2EndTimeScript   = "return document.getElementById('endRest2').value"
	break3StartTimeScript = "return document.getElementById('startRest3').value"
	break3EndTimeScript   = "return document.getElementById('endRest3').value"

	commentIDPattern = "tr#dateRow%s td.vnote"

	cancelButtonID = "dlgInpTimeCancel" // This month
	closeButtonID  = "dlgInpTimeClose"  // Last month

	MonthListTextFormat = "2006年01月"

	SalesforceTimeFormat = "15:04" // min:sec
)

type account struct {
	UserName string
	Password string
}

type workday struct {
	DayIdx int
	Day    string

	DayOff       bool
	RegularBreak bool

	WorkComment  string
	WorkSchedule DaySchedule
}

type salesforce struct {
	Account account
	Page    *agouti.Page
}

func (d *Driver) NewSalesForce(username, password string) (*salesforce, error) {
	page, err := d.NewPage()
	if err != nil {
		return nil, err
	}
	if err := page.Navigate(SalesforceLoginUrl); err != nil {
		return nil, err
	}
	return &salesforce{
		Account: account{
			UserName: username,
			Password: password,
		},
		Page: page,
	}, nil
}

func (sf *salesforce) Login() error {
	// ID, Passの要素を取得し、値を設定
	_ = sf.Page.FindByID(SalesforceUserNameID).Fill(sf.Account.UserName)
	_ = sf.Page.FindByID(SalesforcePasswordID).Fill(sf.Account.Password)

	// formをサブミット
	if err := sf.Page.FindByID(SalesforceLoginID).Submit(); err != nil {
		return err
	}

	time.Sleep(SalesforceSleepTime)
	return nil

}

func (ds *DaySchedule) FromSalesforce(dss DayScheduleStr) {

	workStart, _ := time.Parse(SalesforceTimeFormat, dss.WorkStart)
	workEnd, _ := time.Parse(SalesforceTimeFormat, dss.WorkEnd)
	ds.WorkStart = workStart
	ds.WorkEnd = workEnd

	break1Start, _ := time.Parse(SalesforceTimeFormat, dss.Break1Start)
	break1End, _ := time.Parse(SalesforceTimeFormat, dss.Break1End)
	ds.Break1Start = break1Start
	ds.Break1End = break1End

	break2Start, _ := time.Parse(SalesforceTimeFormat, dss.Break2Start)
	break2End, _ := time.Parse(SalesforceTimeFormat, dss.Break2End)
	ds.Break2Start = break2Start
	ds.Break2End = break2End

	break3Start, _ := time.Parse(SalesforceTimeFormat, dss.Break3Start)
	break3End, _ := time.Parse(SalesforceTimeFormat, dss.Break3End)
	ds.Break3Start = break3Start
	ds.Break3End = break3End

	return
}

func (sf *salesforce) ParseWork() ([]workday, error) {

	// month to process
	processDate := time.Now()

	var workMonth []workday

	// 勤務表タブをクリック
	if err := sf.Page.FindByID(WorkTabID).Click(); err != nil {
		return workMonth, err
	}

	// ちょっと待つ
	time.Sleep(SalesforceSleepTime)

	// 月を選択
	err := sf.Page.FindByID(MonthListID).Select(processDate.Format(MonthListTextFormat))
	if err != nil {
		return workMonth, err
	}

	// ちょっと待つ
	time.Sleep(SalesforceSleepTime)

	// 勤務表の行情報イテレーション
	startDate := time.Date(processDate.Year(), processDate.Month(), 1, 0, 0, 0, 0, time.UTC)
	for d := startDate; d.Month() == processDate.Month(); d = d.AddDate(0, 0, 1) {

		// Open day modal
		day := d.Format("2006-01-02")

		workStatusSelector := fmt.Sprintf(workStatusSelectorPattern, day)
		workStatusText, _ := sf.Page.Find(workStatusSelector).Attribute("title")
		dayOffBool := workStatusText == DayOffValue

		commentID := fmt.Sprintf(commentIDPattern, day)
		workComment, _ := sf.Page.Find(commentID).Attribute("title")

		workdayDetails := workday{
			DayIdx:      d.Day(),
			Day:         day,
			DayOff:      dayOffBool,
			WorkComment: workComment,
		}
		if !dayOffBool {
			workModalButtonSelect := fmt.Sprintf(workModalButtonSelectPattern, day)
			_ = sf.Page.Find(workModalButtonSelect).Click()
			time.Sleep(SalesforceShortSleepTime)

			// The following values are in input.value, not in an attribute, so we need JS to get them
			var startText, endText string
			_ = sf.Page.RunScript(workStartTimeScript, nil, &startText)
			_ = sf.Page.RunScript(workEndTimeScript, nil, &endText)

			var break1StartText, break1EndText string
			_ = sf.Page.RunScript(break1StartTimeScript, nil, &break1StartText)
			_ = sf.Page.RunScript(break1EndTimeScript, nil, &break1EndText)

			var break2StartText, break2EndText string
			_ = sf.Page.RunScript(break2StartTimeScript, nil, &break2StartText)
			_ = sf.Page.RunScript(break2EndTimeScript, nil, &break2EndText)

			var break3StartText, break3EndText string
			_ = sf.Page.RunScript(break3StartTimeScript, nil, &break3StartText)
			_ = sf.Page.RunScript(break3EndTimeScript, nil, &break3EndText)

			workSchedule := DaySchedule{}
			workSchedule.FromSalesforce(DayScheduleStr{
				WorkStart:   startText,
				WorkEnd:     endText,
				Break1Start: break1StartText,
				Break1End:   break1EndText,
				Break2Start: break2StartText,
				Break2End:   break2EndText,
				Break3Start: break3StartText,
				Break3End:   break3EndText,
			})
			workdayDetails.WorkSchedule = workSchedule
			workdayDetails.RegularBreak = workSchedule.IsRegularBreak()

			// Button changes if past month data is displayed => try both cases
			_ = sf.Page.FindByID(cancelButtonID).Click()
			_ = sf.Page.FindByID(closeButtonID).Click()
			time.Sleep(SalesforceShortSleepTime)
		}

		workMonth = append(workMonth, workdayDetails)
	}

	_ = sf.Page.CloseWindow()
	return workMonth, nil
}

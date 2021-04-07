package main

import (
	"github.com/sclevine/agouti"
	"time"
)

const (
	SleepTime = 3 * time.Second

	SalesforceLoginUrl = `https://login.salesforce.com/`

	// Html Attribute Name In Login Menu
	UserNameID = "username"
	PasswordID = "password"
	LoginID    = "Login"

	// Html Attribute Name In Main Menu
	WorkTabID   = "01r7F0000017C6B_Tab"
	MonthListID = "yearMonthList"
	WorkRowId   = "dateRow"
	DayOffValue = "年次有給休暇"

	MonthListTextFormat = "2006年01月"
	InputTimeFormat     = "15:04"
)

type account struct {
	UserName string
	Password string
}

type workday struct {
	Day       string
	DayOff    bool
	StartTime string
	EndTime   string
}

type salesforce struct {
	Account   account
	Page      *agouti.Page
	WorkMonth []workday
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
	_ = sf.Page.FindByID(UserNameID).Fill(sf.Account.UserName)
	_ = sf.Page.FindByID(PasswordID).Fill(sf.Account.Password)

	// formをサブミット
	if err := sf.Page.FindByID(LoginID).Submit(); err != nil {
		return err
	}

	time.Sleep(SleepTime)
	return nil

}

func (sf *salesforce) RegisterWork() error {

	today := time.Now()

	// 勤務表タブをクリック
	if err := sf.Page.FindByID(WorkTabID).Click(); err != nil {
		return err
	}

	// ちょっと待つ
	time.Sleep(SleepTime)

	// 月を選択
	err := sf.Page.FindByID(MonthListID).Select(today.Format(MonthListTextFormat))
	if err != nil {
		return err
	}

	// ちょっと待つ
	time.Sleep(SleepTime)

	// 勤務表の行情報イテレーション
	startDate := time.Date(today.Year(), today.Month()-1, 1, 0, 0, 0, 0, time.UTC)
	for d := startDate; d.Month() == startDate.Month(); d = d.AddDate(0, 0, 1) {

		day := d.Format("2006-01-02")
		workStatusSelector := "tr #" + WorkRowId + day + " td.vstatus"
		workStartSelector := "tr #" + WorkRowId + day + " td.vst"
		workEndSelector := "tr #" + WorkRowId + day + " td.vet"

		workStatusText, _ := sf.Page.Find(workStatusSelector).Attribute("title")
		dayOffBool := workStatusText == DayOffValue
		startText, _ := sf.Page.Find(workStartSelector).Text()
		endText, _ := sf.Page.Find(workEndSelector).Text()

		workdayDetails := workday{
			Day:       day,
			DayOff:    dayOffBool,
			StartTime: startText,
			EndTime:   endText,
		}

		sf.WorkMonth = append(sf.WorkMonth, workdayDetails)
	}

	return nil
}

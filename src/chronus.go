package main

import (
	"github.com/sclevine/agouti"
	"time"
)

const (
	ChronusSleepTime = 3 * time.Second

	ChronusLoginUrl = `https://chronus-ext.tis.co.jp/Lysithea/Logon`

	// Html Attribute Name In Login Menu
	ChronusUserNameSelector    = `document.FORM_COMMON.PersonCode.value`
	ChronusPasswordSelector    = `document.FORM_COMMON.Password.value`
	ChronusLoginSubmitSelector = `a`
)

type chronus struct {
	Account   account
	Page      *agouti.Page
	WorkMonth []workday
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

func (ch *chronus) RegisterWork() error {

	//today := time.Now()
	//
	//// 勤務表タブをクリック
	//if err := sf.Page.FindByID(WorkTabID).Click(); err != nil {
	//	return err
	//}
	//
	//// ちょっと待つ
	time.Sleep(60 * time.Second)
	//
	//// 月を選択
	//err := sf.Page.FindByID(MonthListID).Select(today.Format(MonthListTextFormat))
	//if err != nil {
	//	return err
	//}
	//
	//// ちょっと待つ
	//time.Sleep(ChronusSleepTime)
	//
	//// 勤務表の行情報イテレーション
	//startDate := time.Date(today.Year(), today.Month()-1, 1, 0, 0, 0, 0, time.UTC)
	//for d := startDate; d.Month() == startDate.Month(); d = d.AddDate(0, 0, 1) {
	//
	//	day := d.Format("2006-01-02")
	//	workStatusSelector := "tr #" + WorkRowId + day + " td.vstatus"
	//	workStartSelector := "tr #" + WorkRowId + day + " td.vst"
	//	workEndSelector := "tr #" + WorkRowId + day + " td.vet"
	//
	//	workStatusText, _ := sf.Page.Find(workStatusSelector).Attribute("title")
	//	dayOffBool := workStatusText == DayOffValue
	//	startText, _ := sf.Page.Find(workStartSelector).Text()
	//	endText, _ := sf.Page.Find(workEndSelector).Text()
	//
	//	workdayDetails := workday{
	//		Day:       day,
	//		DayOff:    dayOffBool,
	//		StartTime: startText,
	//		EndTime:   endText,
	//	}
	//
	//	sf.WorkMonth = append(sf.WorkMonth, workdayDetails)
	//}

	return nil
}

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
	time.Sleep(60 * time.Second)

	return nil
}

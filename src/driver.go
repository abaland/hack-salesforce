package main

import (
	"github.com/sclevine/agouti"
	"time"
)

type Driver struct {
	*agouti.WebDriver
}

func NewChromeDriver(c agouti.Option) *Driver {
	return &Driver{agouti.ChromeDriver(c)}
}

type fn func(Page *agouti.Page) bool

func sleepUntil(condition fn, Page *agouti.Page, maxSleep time.Duration) bool {
	tStart := time.Now().Nanosecond()
	for int64(time.Now().Nanosecond()-tStart) < maxSleep.Nanoseconds() {
		if condition(Page) {
			return true
		}
		time.Sleep(100 * time.Millisecond)
	}
	return false
}

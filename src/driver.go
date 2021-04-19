package main

import (
	"github.com/sclevine/agouti"
	"time"
)

type Driver struct {
	*agouti.WebDriver
}

// NewChromeDriver returns a Driver instance to control a Chrome instance
func NewChromeDriver(c agouti.Option) *Driver {
	return &Driver{agouti.ChromeDriver(c)}
}

// boolPageCond defines function that make an assertion regarding the current state of a chrome page
type boolPageCond func(Page *agouti.Page) bool

// sleepUntil makes the program pause until a web page reaches a certain state.
// To avoid infinite wait, a maxSleep parameter needs to be provided.
func sleepUntil(condition boolPageCond, Page *agouti.Page, maxSleep time.Duration) bool {
	tStart := time.Now().Nanosecond()
	for int64(time.Now().Nanosecond()-tStart) < maxSleep.Nanoseconds() {
		if condition(Page) {
			return true
		}
		time.Sleep(100 * time.Millisecond)
	}
	return false
}

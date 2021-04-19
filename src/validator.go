package main

import (
	"errors"
)

// validate checks that Credentials information fit expectations.
// it does not check whether credentials are correct.
func (c *Credentials) validate() error {
	if c.User == "" {
		return errors.New("empty UserName from ini file")
	}
	if c.Password == "" {
		return errors.New("empty Password from ini file")
	}
	return nil
}

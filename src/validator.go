package main

import (
	"errors"
)

func (c *Config) validate() error {
	if c.User == "" {
		return errors.New("empty UserName from ini file.")
	}
	if c.Password == "" {
		return errors.New("empty Password from ini file.")
	}
	return nil
}

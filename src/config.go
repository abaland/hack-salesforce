package main

import "gopkg.in/go-ini/ini.v1"

type Config struct {
	User     string
	Password string
}

const (
	salesforceSection     = "salesforce"
	salesforceUserKey     = "user"
	salesforcePasswordKey = "password"
	chronusSection        = "chronus"
	chronusUserKey        = "user"
	chronusPasswordKey    = "password"
)

func ParseCredentials(c *ini.File, section string, userKey string, passwordKey string) *Config {
	return &Config{
		User:     c.Section(section).Key(userKey).String(),
		Password: c.Section(section).Key(passwordKey).String(),
	}
}

func NewConfig(configPath string) (*Config, *Config, error) {
	c, err := ini.Load(configPath)
	if err != nil {
		return nil, nil, err
	}
	salesforceCredentials := ParseCredentials(c, salesforceSection, salesforceUserKey, salesforcePasswordKey)
	chronusCredentials := ParseCredentials(c, chronusSection, chronusUserKey, chronusPasswordKey)
	return salesforceCredentials, chronusCredentials, nil
}

package main

import "gopkg.in/go-ini/ini.v1"

// Credentials holds username and password information for a specific website
type Credentials struct {
	User     string
	Password string
}

const (
	// Information about how to parse config file for salesforce credentials
	salesforceSection     = "salesforce"
	salesforceUserKey     = "user"
	salesforcePasswordKey = "password"

	// Information about how to parse config file for chronus credentials
	chronusSection     = "chronus"
	chronusUserKey     = "user"
	chronusPasswordKey = "password"
)

// ParseCredentials extracts username and password information from a given config object
func ParseCredentials(c *ini.File, section string, userKey string, passwordKey string) *Credentials {
	return &Credentials{
		User:     c.Section(section).Key(userKey).String(),
		Password: c.Section(section).Key(passwordKey).String(),
	}
}

// NewConfig reads .ini config files and generates salesforce and chronus credentials objects
func NewConfig(configPath string) (*Credentials, *Credentials, error) {
	c, err := ini.Load(configPath)
	if err != nil {
		return nil, nil, err
	}
	salesforceCredentials := ParseCredentials(c, salesforceSection, salesforceUserKey, salesforcePasswordKey)
	chronusCredentials := ParseCredentials(c, chronusSection, chronusUserKey, chronusPasswordKey)
	return salesforceCredentials, chronusCredentials, nil
}

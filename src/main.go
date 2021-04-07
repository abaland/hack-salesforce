package main

import (
	"encoding/json"
	"flag"
	"github.com/sclevine/agouti"
	"io/ioutil"
)

var (
	configPath string
)

func main() {
	flag.StringVar(&configPath, "config_path", "", "ini config path")
	flag.Parse()

	logger := NewLogger()
	logger.Info("start.")

	// Setting
	salesforceConfig, chronusConfig, err := NewConfig(configPath)
	if err != nil {
		logger.Errorf("NewConfig Error:%v", err)
		return
	}
	if err = salesforceConfig.validate(); err != nil {
		logger.Errorf("config validate Error:%v", err)
		return
	}
	if err = chronusConfig.validate(); err != nil {
		logger.Errorf("config validate Error:%v", err)
		return
	}

	// Driver Start
	driver := NewChromeDriver(agouti.Desired(agouti.Capabilities{}))
	if err := driver.Start(); err != nil {
		logger.Errorf("Failed to start:%v", err)
		return
	}
	defer driver.Stop()
	sf, err := driver.NewSalesForce(salesforceConfig.User, salesforceConfig.Password)
	if err != nil {
		logger.Errorf("Failed to create instance:%v", err)
		return
	}
	err = sf.Login()
	if err != nil {
		logger.Errorf("Failed to login:%v", err)
		return
	}

	// Setting Daily Works
	if err := sf.ParseWork(); err != nil {
		logger.Errorf("Failed to RegisterWork:%v", err)
		return
	}
	logger.Infof("finish to parse work.")

	// Write file to json TODO Add Chronus registration
	file, _ := json.MarshalIndent(sf.WorkMonth, "", " ")
	_ = ioutil.WriteFile("test.json", file, 0644)

	ch, err := driver.NewChronus(chronusConfig.User, chronusConfig.Password)
	if err != nil {
		logger.Errorf("Failed to create chronus instance:%v", err)
		return
	}
	err = ch.Login()
	if err != nil {
		logger.Errorf("Failed to login:%v", err)
		return
	}

	//
	if err := ch.RegisterWork(); err != nil {
		logger.Errorf("Failed to RegisterWork:%v", err)
		return
	}
	logger.Infof("finish to register work.")

	logger.Info("finish.")
}

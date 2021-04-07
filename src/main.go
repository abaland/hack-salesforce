package main

import (
	"encoding/json"
	"github.com/sclevine/agouti"
	"io/ioutil"

	"flag"
)

type Project struct {
	Name     string `json:"name"`
	Duration string `json:"duration"`
}

var (
	configPath string
	jsonFile   string
)

func main() {
	flag.StringVar(&configPath, "config_path", "", "ini config path")
	flag.StringVar(&jsonFile, "jsonfile", "", "json file for dailywork")
	flag.Parse()

	logger := NewLogger()
	logger.Info("start.")

	// Setting
	config, err := NewConfig(configPath)
	if err != nil {
		logger.Errorf("NewConfig Error:%v", err)
		return
	}
	if err = config.validate(); err != nil {
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
	sf, err := driver.NewSalesForce(config.User, config.Password)
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
	if err := sf.RegisterWork(); err != nil {
		logger.Errorf("Failed to RegisterWork:%v", err)
		return
	}
	logger.Infof("finish to parse work.")

	// Write file to json TODO Add Chronus registration
	file, _ := json.MarshalIndent(sf.WorkMonth, "", " ")
	_ = ioutil.WriteFile("test.json", file, 0644)

	logger.Info("finish.")
}

package settings

/*	License: GPLv3
	Authors:
		Mirko Brombin <send@mirko.pm>
		Vanilla OS Contributors <https://github.com/vanilla-os/>
	Copyright: 2023
	Description: VSO is a utility which allows you to perform maintenance
	tasks on your Vanilla OS installation.
*/

import (
	"fmt"

	"github.com/spf13/viper"
)

type Config struct {
	Updates UpdatesConfig `json:"updates"`
}

type UpdatesConfig struct {
	Schedule string `json:"schedule"`
	Smart    bool   `json:"smart"`
}

var Cnf *Config

func init() {
	viper.AddConfigPath("/etc/vso/")
	viper.AddConfigPath("config/")
	viper.SetConfigName("config")
	viper.SetConfigType("json")
	err := viper.ReadInConfig()

	if err != nil {
		panic("Config error!")
	}

	err = viper.Unmarshal(&Cnf)

	if err != nil {
		panic("Config error!\n" + err.Error())
	}
}

func GetConfig() *Config {
	return Cnf
}

func GetConfigAsKV() map[string]string {
	list := make(map[string]string)
	list["updates.schedule"] = Cnf.Updates.Schedule
	list["updates.smart"] = fmt.Sprintf("%t", Cnf.Updates.Smart)
	return list
}

func GetConfigValue(key string) interface{} {
	return viper.Get(key)
}

func SetConfigValue(key string, value interface{}) {
	if key == "updates.schedule" {
		if value != "daily" && value != "weekly" && value != "monthly" {
			fmt.Println("Invalid value for updates.schedule!")
			return
		}
	} else if key == "updates.smart" {
		if value != true && value != false && value != "true" && value != "false" {
			fmt.Println("Invalid value for updates.smart!")
			return
		}
	}

	if value == "true" {
		value = true
	} else if value == "false" {
		value = false
	}

	viper.Set(key, value)
}

func SaveConfig() error {
	return viper.WriteConfig()
}

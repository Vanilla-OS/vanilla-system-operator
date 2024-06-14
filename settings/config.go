package settings

/*	License: GPLv3
	Authors:
		Mirko Brombin <send@mirko.pm>
		Vanilla OS Contributors <https://github.com/vanilla-os/>
	Copyright: 2024
	Description: VSO is a utility which allows you to perform maintenance
	tasks on your Vanilla OS installation.
*/

import (
	"fmt"
	"os"
	"slices"

	"github.com/spf13/viper"
)

type Config struct {
	Updates UpdatesConfig `json:"updates"`
}

type UpdatesConfig struct {
	Schedule string `json:"schedule"`
	Smart    bool   `json:"smart"`
}

var Cnf Config

const (
	ScheduleNever   = "never"
	ScheduleDaily   = "daily"
	ScheduleWeekly  = "weekly"
	ScheduleMonthly = "monthly"
)

var (
	validUpdatesScheduleKeys = []string{ScheduleNever, ScheduleDaily, ScheduleWeekly, ScheduleMonthly}
	validUpdatesSmartKeys    = []string{"true", "false"}
)

func init() {
	viper.AddConfigPath("/usr/share/vso/")
	viper.AddConfigPath("/etc/vso/")
	viper.AddConfigPath("config/")
	viper.SetConfigName("config")
	viper.SetConfigType("json")
	err := viper.ReadInConfig()
	if err != nil {
		panic("Config error!")
	}

	if os.Getenv("VSO_DEBUG") != "" {
		fmt.Println("Config loaded from " + viper.ConfigFileUsed())
	}

	err = viper.Unmarshal(&Cnf)
	if err != nil {
		panic("Config error!\n" + err.Error())
	}
}

func GetConfig() Config {
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

func SetConfigValue(key string, value string) error {
	switch key {
	case "updates.schedule":
		if !slices.Contains(validUpdatesScheduleKeys, value) {
			return fmt.Errorf("invalid value for updates.schedule")
		}
	case "updates.smart":
		if !slices.Contains(validUpdatesSmartKeys, value) {
			return fmt.Errorf("invalid value for updates.smart")
		}
	}

	var anyValue any
	switch value {
	case "true":
		anyValue = true
	case "false":
		anyValue = false
	default:
		anyValue = value
	}

	viper.Set(key, anyValue)

	err := SaveConfig()
	if err != nil {
		return err
	}

	return nil
}

func SaveConfig() error {
	return viper.WriteConfig()
}

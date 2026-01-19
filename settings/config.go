package settings

/*	License: GPLv3
	Authors:
		Mirko Brombin <brombin94@gmail.com>
		Pietro di Caprio <pietro@fabricators.ltd>
		Vanilla OS Contributors <https://github.com/vanilla-os/>
	Copyright: 2024
	Description: VSO is a utility which allows you to perform maintenance
	tasks on your Vanilla OS installation.
*/

import (
	"encoding/json"
	"fmt"
	"os"
	"slices"

	"github.com/vanilla-os/sdk/pkg/v1/conf"
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

func LoadConfig() error {
	var err error
	Cnf = Config{
		Updates: UpdatesConfig{
			Schedule: ScheduleWeekly,
			Smart:    true,
		},
	}

	loadedCnf, err := conf.NewBuilder[Config]("vso").
		WithType("json").
		Build()

	if err != nil {
		return nil
	}

	Cnf = *loadedCnf
	return nil
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
	switch key {
	case "updates.schedule":
		return Cnf.Updates.Schedule
	case "updates.smart":
		return Cnf.Updates.Smart
	}
	return nil
}

func SetConfigValue(key string, value string) error {
	switch key {
	case "updates.schedule":
		if !slices.Contains(validUpdatesScheduleKeys, value) {
			return fmt.Errorf("invalid value for updates.schedule")
		}
		Cnf.Updates.Schedule = value
	case "updates.smart":
		if !slices.Contains(validUpdatesSmartKeys, value) {
			return fmt.Errorf("invalid value for updates.smart")
		}
		Cnf.Updates.Smart = (value == "true")
	default:
		return fmt.Errorf("unknown key: %s", key)
	}

	return SaveConfig()
}

func SaveConfig() error {
	path := "/etc/vso/config.json"

	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()

	encoder := json.NewEncoder(f)
	encoder.SetIndent("", "  ")
	return encoder.Encode(Cnf)
}

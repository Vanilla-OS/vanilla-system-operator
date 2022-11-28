package settings

/*	License: GPLv3
	Authors:
		Mirko Brombin <send@mirko.pm>
		Pietro di Caprio <pietro@fabricators.ltd>
	Copyright: 2022
	Description: VSO is an utility which allows you to perform maintenance
	tasks on your Vanilla OS installation.
*/

import (
	"github.com/spf13/viper"
)

type Config struct {
	Updates UpdatesConfig `json:"updates"`
}

type UpdatesConfig struct {
	Schedule string `json:"schedule"`
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

func GetConfigKeys() []string {
	return viper.AllKeys()
}

func GetConfigValue(key string) interface{} {
	return viper.Get(key)
}

func SetConfigValue(key string, value interface{}) {
	viper.Set(key, value)
}

func SaveConfig() error {
	return viper.WriteConfig()
}

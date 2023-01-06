package cmd

/*	License: GPLv3
	Authors:
		Mirko Brombin <send@mirko.pm>
		Pietro di Caprio <pietro@fabricators.ltd>
	Copyright: 2023
	Description: VSO is a utility which allows you to perform maintenance
	tasks on your Vanilla OS installation.
*/

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/vanilla-os/vso/core"
	"github.com/vanilla-os/vso/settings"
)

func configUsage(*cobra.Command) error {
	fmt.Print(`Description: 
	Configure VSO

Usage:
  	vso config [options] [command]

Options:
	--help/-h		show this message
	--assume-yes/-y		assume yes to all questions

Commands:
	show			show current configuration
	get <key>		get a configuration value
	set <key> <value>	set a configuration value

Examples:
	vso config get updates::schedule
	vso config set updates::schedule weekly
`)
	return nil
}

func NewConfigCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "config",
		Short: "Configure VSO",
		RunE:  config,
	}
	cmd.SetUsageFunc(configUsage)
	return cmd
}

func config(cmd *cobra.Command, args []string) error {
	if len(args) == 0 {
		fmt.Println("No command specified.")
		return configUsage(cmd)
	}

	switch args[0] {
	case "show":
		for _, key := range settings.GetConfigKeys() {
			fmt.Printf("%s: %v\n", key, settings.GetConfigValue(key))
		}
	case "get":
		if len(args) < 2 {
			fmt.Println("No key specified.")
			return configUsage(cmd)
		}
		fmt.Println(settings.GetConfigValue(args[1]))
	case "set":
		if !core.RootCheck(true) {
			return nil
		}
		if len(args) < 3 {
			fmt.Println("No key or value specified.")
			return configUsage(cmd)
		}
		settings.SetConfigValue(args[1], args[2])
		settings.SaveConfig()
	default:
		fmt.Println("Invalid command.")
		return configUsage(cmd)
	}

	return nil
}

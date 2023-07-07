package cmd

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

	"github.com/spf13/cobra"
	"github.com/vanilla-os/orchid/cmdr"
	"github.com/vanilla-os/vso/settings"
)

func NewConfigCommand() *cmdr.Command {
	// Root command
	cmd := cmdr.NewCommand(
		"config",
		vso.Trans("config.description"),
		vso.Trans("config.description"),
		nil,
	)

	// Show subcommand
	showCmd := cmdr.NewCommand(
		"show",
		vso.Trans("config.show.description"),
		vso.Trans("config.show.description"),
		showConfig,
	)

	// Get subcommand
	getCmd := cmdr.NewCommand(
		"get",
		vso.Trans("config.get.description"),
		vso.Trans("config.get.description"),
		getConfig,
	)

	getCmd.WithStringFlag(
		cmdr.NewStringFlag(
			"key",
			"k",
			vso.Trans("config.get.options.key"),
			"",
		),
	)

	// Set subcommand
	setCmd := cmdr.NewCommand(
		"set",
		vso.Trans("config.set.description"),
		vso.Trans("config.set.description"),
		setConfig,
	)

	setCmd.WithStringFlag(
		cmdr.NewStringFlag(
			"key",
			"k",
			vso.Trans("config.set.options.key"),
			"",
		),
	)
	setCmd.WithStringFlag(
		cmdr.NewStringFlag(
			"value",
			"v",
			vso.Trans("config.set.options.value"),
			"",
		),
	)

	// Add subcommands to root command
	cmd.AddCommand(showCmd)
	cmd.AddCommand(getCmd)
	cmd.AddCommand(setCmd)

	return cmd
}

func showConfig(cmd *cobra.Command, args []string) error {
	cnf := settings.GetConfigAsKV()
	for key, value := range cnf {
		fmt.Println(key, ":", value)
	}

	return nil
}

func getConfig(cmd *cobra.Command, args []string) error {
	key, _ := cmd.Flags().GetString("key")
	if key == "" {
		cmdr.Error.Println(vso.Trans("config.get.error.noKey"))
		return nil
	}

	value := settings.GetConfigValue(key)
	if value == nil {
		cmdr.Error.Println(vso.Trans("config.get.error.noValue"))
		return nil
	}

	fmt.Println(value)
	return nil
}

func setConfig(cmd *cobra.Command, args []string) error {
	key, _ := cmd.Flags().GetString("key")
	if key == "" {
		cmdr.Error.Println(vso.Trans("config.set.error.noKey"))
		return nil
	}

	value, _ := cmd.Flags().GetString("value")
	if value == "" {
		cmdr.Error.Println(vso.Trans("config.set.error.noValue"))
		return nil
	}

	settings.SetConfigValue(key, value)
	cmdr.Info.Println(vso.Trans("config.set.success"), key, value)
	return nil
}

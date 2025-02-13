package main

/*	License: GPLv3
	Authors:
		Mirko Brombin <send@mirko.pm>
		Pietro di Caprio <pietro@fabricators.ltd>
	Copyright: 2024
	Description: VSO is a utility which allows you to perform maintenance
	tasks on your Vanilla OS installation.
*/

import (
	"embed"
	"os"

	"github.com/vanilla-os/apx/v2/core"
	"github.com/vanilla-os/apx/v2/settings"
	"github.com/vanilla-os/orchid/cmdr"
	"github.com/vanilla-os/vanilla-system-operator/cmd"
)

var Version = "development"

//go:embed locales/*.yml
var fs embed.FS
var vso *cmdr.App

func main() {
	cnf := settings.NewApxConfig(
		"/usr/share/vso/apx",
		"/usr/share/apx/distrobox/distrobox",
		"btrfs",
	)
	core.NewApx(cnf)

	vso = cmd.New(Version, fs)

	// root command
	root := cmd.NewRootCommand(Version)
	vso.CreateRootCommand(root)
	vso.CreateRootCommand(root, vso.Trans("vso.msg.help"), vso.Trans("vso.msg.version"))

	msgs := cmdr.UsageStrings{
		Usage:                vso.Trans("vso.msg.usage"),
		Aliases:              vso.Trans("vso.msg.aliases"),
		Examples:             vso.Trans("vso.msg.examples"),
		AvailableCommands:    vso.Trans("vso.msg.availableCommands"),
		AdditionalCommands:   vso.Trans("vso.msg.additionalCommands"),
		Flags:                vso.Trans("vso.msg.flags"),
		GlobalFlags:          vso.Trans("vso.msg.globalFlags"),
		AdditionalHelpTopics: vso.Trans("vso.msg.additionalHelpTopics"),
		MoreInfo:             vso.Trans("vso.msg.moreInfo"),
	}
	vso.SetUsageStrings(msgs)

	// commands
	tasks := cmd.NewTasksCommand()
	root.AddCommand(tasks)

	upgrade := cmd.NewUpgradeCommand()
	root.AddCommand(upgrade)

	config := cmd.NewConfigCommand()
	root.AddCommand(config)

	// pico
	pico := cmd.NewPicoCommand()
	root.AddCommand(pico...)

	// waydroid
	// disabled until properly implemented again
	// way := cmd.NewWayCommand()
	// root.AddCommand(way...)

	// run the app
	err := vso.Run()
	if err != nil {
		cmdr.Error.Println(err)
		os.Exit(1)
	}
}

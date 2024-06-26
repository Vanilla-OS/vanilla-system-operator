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

	"github.com/vanilla-os/apx/v2/core"
	"github.com/vanilla-os/apx/v2/settings"
	"github.com/vanilla-os/orchid/cmdr"
	"github.com/vanilla-os/vanilla-system-operator/cmd"
)

var Version = "2.0.1"

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
	way := cmd.NewWayCommand()
	root.AddCommand(way...)

	// run the app
	err := vso.Run()
	if err != nil {
		cmdr.Error.Println(err)
	}
}

package main

/*	License: GPLv3
	Authors:
		Mirko Brombin <send@mirko.pm>
		Pietro di Caprio <pietro@fabricators.ltd>
	Copyright: 2023
	Description: VSO is a utility which allows you to perform maintenance
	tasks on your Vanilla OS installation.
*/

import (
	"embed"

	"github.com/vanilla-os/orchid/cmdr"
	"github.com/vanilla-os/vso/cmd"
)

var (
	Version = "2.0.0-alpha.1"
)

//go:embed locales/*.yml
var fs embed.FS
var vso *cmdr.App

func main() {
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

	// run the app
	err := vso.Run()
	if err != nil {
		cmdr.Error.Println(err)
	}
}

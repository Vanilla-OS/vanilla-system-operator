package cmd

/*	License: GPLv3
	Authors:
		Mirko Brombin <send@mirko.pm>
		Pietro di Caprio <pietro@fabricators.ltd>
	Copyright: 2024
	Description: The Vanilla System Operator is a package manager,
	a system updater and a task automator.
*/

import (
	"embed"

	"github.com/vanilla-os/orchid/cmdr"
)

var vso *cmdr.App

func New(version string, fs embed.FS) *cmdr.App {
	vso = cmdr.NewApp("vso", version, fs)
	return vso
}

func NewRootCommand(version string) *cmdr.Command {
	root := cmdr.NewCommand(
		"vso",
		vso.Trans("vso.description"),
		vso.Trans("vso.description"),
		nil,
	)
	root.Version = version

	return root
}

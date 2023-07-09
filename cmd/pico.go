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
	"github.com/vanilla-os/vso/core"
)

func NewPicoCommand() []*cmdr.Command {
	handleFunc := func() func(cmd *cobra.Command, args []string) error {
		return func(cmd *cobra.Command, args []string) error {
			if !core.PicoExists() {
				cmdr.Error.Println("The Pico subsystem is not initialized. Please run `vso init-pico` to initialize it.")
				return nil
			}

			return fmt.Errorf("not implemented")
		}
	}

	installCmd := cmdr.NewCommand(
		"install",
		vso.Trans("install.description"),
		vso.Trans("install.description"),
		handleFunc(),
	)
	removeCmd := cmdr.NewCommand(
		"remove",
		vso.Trans("remove.description"),
		vso.Trans("remove.description"),
		handleFunc(),
	)
	updateCmd := cmdr.NewCommand(
		"update",
		vso.Trans("update.description"),
		vso.Trans("update.description"),
		handleFunc(),
	)
	upgradeCmd := cmdr.NewCommand(
		"upgrade",
		vso.Trans("upgrade.description"),
		vso.Trans("upgrade.description"),
		handleFunc(),
	)
	searchCmd := cmdr.NewCommand(
		"search",
		vso.Trans("search.description"),
		vso.Trans("search.description"),
		handleFunc(),
	)
	shellCmd := cmdr.NewCommand(
		"shell",
		vso.Trans("shell.description"),
		vso.Trans("shell.description"),
		handleFunc(),
	)
	initCmd := cmdr.NewCommand(
		"init",
		vso.Trans("init.description"),
		vso.Trans("init.description"),
		handleFunc(),
	)

	return []*cmdr.Command{
		installCmd,
		removeCmd,
		updateCmd,
		upgradeCmd,
		searchCmd,
		shellCmd,
		initCmd,
	}
}

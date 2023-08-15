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
				cmdr.Error.Println("The Pico subsystem is not initialized. Please run `vso pico-init` to initialize it.")
				return nil
			}

			return runPicoCmd(cmd.Name(), args)
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
	runCmd := cmdr.NewCommand(
		"run",
		vso.Trans("run.description"),
		vso.Trans("run.description"),
		handleFunc(),
	)

	initCmd := cmdr.NewCommand(
		"pico-init",
		vso.Trans("init.description"),
		vso.Trans("init.description"),
		picoInit,
	)
	initCmd.WithBoolFlag(
		cmdr.NewBoolFlag(
			"force",
			"f",
			vso.Trans("init.options.force"),
			false,
		),
	)

	return []*cmdr.Command{
		installCmd,
		removeCmd,
		updateCmd,
		upgradeCmd,
		searchCmd,
		shellCmd,
		runCmd,
		initCmd,
	}
}

func picoInit(cmd *cobra.Command, args []string) error {
	force, _ := cmd.Flags().GetBool("force")

	if core.PicoExists() && !force {
		cmdr.Error.Println("The Pico subsystem is already initialized. Use the --force flag to force the initialization.")
		return nil
	}

	err := core.PicoInit()
	if err != nil {
		return err
	}

	cmdr.Success.Println("The Pico subsystem has been initialized successfully.")
	return nil
}

func runPicoCmd(command string, args []string) error {
	pico, err := core.GetPico()
	if err != nil {
		return err
	}

	if command == "shell" {
		return pico.Enter()
	} else if command == "run" {
		_, err := pico.Exec(false, args...)
		return err
	}

	pkgManager, err := pico.Stack.GetPkgManager()
	if err != nil {
		return err
	}

	var realCommand string
	switch command {
	case "autoremove":
		realCommand = pkgManager.CmdAutoRemove
	case "clean":
		realCommand = pkgManager.CmdClean
	case "install":
		realCommand = pkgManager.CmdInstall
	case "list":
		realCommand = pkgManager.CmdList
	case "purge":
		realCommand = pkgManager.CmdPurge
	case "remove":
		realCommand = pkgManager.CmdRemove
	case "search":
		realCommand = pkgManager.CmdSearch
	case "show":
		realCommand = pkgManager.CmdShow
	case "update":
		realCommand = pkgManager.CmdUpdate
	case "upgrade":
		realCommand = pkgManager.CmdUpgrade
	default:
		return fmt.Errorf("unknown command: %s", command)
	}

	finalArgs := pkgManager.GenCmd(realCommand, args...)

	_, err = pico.Exec(false, finalArgs...)
	if err != nil {
		return err
	}

	return nil
}

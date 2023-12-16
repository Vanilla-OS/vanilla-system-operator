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
	"errors"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/vanilla-os/orchid/cmdr"
	"github.com/vanilla-os/vso/core"
)

func NewPicoCommand() []*cmdr.Command {
	handleFunc := func() func(cmd *cobra.Command, args []string) error {
		return func(cmd *cobra.Command, args []string) error {
			if !core.PicoExists() {
				cmdr.Error.Println(vso.Trans("pico.error.notInitialized"))
				return nil
			}

			return runPicoCmd(cmd.Name(), cmd, args)
		}
	}

	installCmd := cmdr.NewCommand(
		"install",
		vso.Trans("pico.install.description"),
		vso.Trans("pico.install.description"),
		handleFunc(),
	)
	removeCmd := cmdr.NewCommand(
		"remove",
		vso.Trans("pico.remove.description"),
		vso.Trans("pico.remove.description"),
		handleFunc(),
	)
	sideloadCmd := cmdr.NewCommand(
		"sideload",
		vso.Trans("pico.sideload.description"),
		vso.Trans("pico.sideload.description"),
		handleFunc(),
	)
	updateCmd := cmdr.NewCommand(
		"update",
		vso.Trans("pico.update.description"),
		vso.Trans("pico.update.description"),
		handleFunc(),
	)
	upgradeCmd := cmdr.NewCommand(
		"upgrade",
		vso.Trans("pico.upgrade.description"),
		vso.Trans("pico.upgrade.description"),
		handleFunc(),
	)
	searchCmd := cmdr.NewCommand(
		"search",
		vso.Trans("pico.search.description"),
		vso.Trans("pico.search.description"),
		handleFunc(),
	)
	shellCmd := cmdr.NewCommand(
		"shell",
		vso.Trans("pico.shell.description"),
		vso.Trans("pico.shell.description"),
		handleFunc(),
	)
	runCmd := cmdr.NewCommand(
		"run",
		vso.Trans("pico.run.description"),
		vso.Trans("pico.run.description"),
		handleFunc(),
	)
	runCmd.Flags().SetInterspersed(false)

	exportCmd := cmdr.NewCommand(
		"export",
		vso.Trans("pico.export.description"),
		vso.Trans("pico.export.description"),
		picoExport,
	)
	exportCmd.WithStringFlag(
		cmdr.NewStringFlag(
			"app",
			"a",
			vso.Trans("pico.export.options.app.description"),
			"",
		),
	)
	exportCmd.WithStringFlag(
		cmdr.NewStringFlag(
			"bin",
			"b",
			vso.Trans("pico.export.options.bin.description"),
			"",
		),
	)
	unexportCmd := cmdr.NewCommand(
		"unexport",
		vso.Trans("pico.unexport.description"),
		vso.Trans("pico.unexport.description"),
		picoUnexport,
	)
	unexportCmd.WithStringFlag(
		cmdr.NewStringFlag(
			"app",
			"a",
			vso.Trans("pico.unexport.options.app.description"),
			"",
		),
	)
	unexportCmd.WithStringFlag(
		cmdr.NewStringFlag(
			"bin",
			"b",
			vso.Trans("pico.unexport.options.bin.description"),
			"",
		),
	)

	initCmd := cmdr.NewCommand(
		"pico-init",
		vso.Trans("pico.init.description"),
		vso.Trans("pico.init.description"),
		picoInit,
	)
	initCmd.WithBoolFlag(
		cmdr.NewBoolFlag(
			"force",
			"f",
			vso.Trans("pico.init.options.force.description"),
			false,
		),
	)

	return []*cmdr.Command{
		installCmd,
		removeCmd,
		sideloadCmd,
		updateCmd,
		upgradeCmd,
		searchCmd,
		shellCmd,
		runCmd,
		exportCmd,
		unexportCmd,
		initCmd,
	}
}

func picoInit(cmd *cobra.Command, args []string) error {
	force, _ := cmd.Flags().GetBool("force")

	if core.PicoExists() {
		if !force {
			cmdr.Error.Println(vso.Trans("pico.error.alreadyInitialized"))
			return nil
		}

		err := core.PicoDelete()
		if err != nil {
			return err
		}
	}

	err := core.PicoInit()
	if err != nil {
		return err
	}

	cmdr.Success.Println(vso.Trans("pico.info.initialized"))
	return nil
}

func picoExport(cmd *cobra.Command, args []string) error {
	app, _ := cmd.Flags().GetString("app")
	bin, _ := cmd.Flags().GetString("bin")

	if app == "" && bin == "" {
		cmdr.Error.Println(vso.Trans("pico.error.noAppNameOrBin"))
		return nil
	}

	err := core.PicoExport(app, bin)
	if err != nil {
		var errMsg string
		if app != "" {
			errMsg = vso.Trans("pico.error.exportingApp", err.Error())
		} else {
			errMsg = vso.Trans("pico.error.exportingBin", err.Error())
		}
		return errors.New(errMsg)
	}

	if app != "" {
		cmdr.Success.Printf(vso.Trans("pico.info.exported"), app)
	} else {
		cmdr.Success.Printf(vso.Trans("pico.info.exported"), bin)
	}
	return nil
}

func picoUnexport(cmd *cobra.Command, args []string) error {
	app, _ := cmd.Flags().GetString("app")
	bin, _ := cmd.Flags().GetString("bin")

	if app == "" && bin == "" {
		cmdr.Error.Println(vso.Trans("pico.error.noAppNameOrBin"))
		return nil
	}

	err := core.PicoUnexport(app, bin)
	if err != nil {
		var errMsg string
		if app != "" {
			errMsg = vso.Trans("pico.error.unexportingApp", err.Error())
		} else {
			errMsg = vso.Trans("pico.error.unexportingBin", err.Error())
		}
		return errors.New(errMsg)
	}

	if app != "" {
		cmdr.Success.Printf(vso.Trans("pico.info.unexported"), app)
	} else {
		cmdr.Success.Printf(vso.Trans("pico.info.unexported"), bin)
	}
	return nil
}

func runPicoCmd(command string, cmd *cobra.Command, args []string) error {
	pico, err := core.GetPico()
	if err != nil {
		return err
	}

	switch command {
	case "shell":
		return pico.Enter()
	case "run":
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
	case "sideload":
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
		return fmt.Errorf(vso.Trans("pico.error.unknownCommand"), command)
	}

	if command == "remove" {
		exportedN, err := pico.UnexportDesktopEntries(args...)
		if err == nil {
			cmdr.Info.Printfln(vso.Trans("pico.info.unexportedApps"), exportedN)
		}
	}

	if command == "sideload" {
		// we need to handle possible broken/missing dependencies while
		// sideloading, so we use the --fix-broken flag, assuming VSO is
		// being used in Vanilla OS (which is the only supported use case)
		args = append([]string{"--fix-broken"}, args...)
	}
	finalArgs := pkgManager.GenCmd(realCommand, args...)
	_, err = pico.Exec(false, finalArgs...)
	if err != nil {
		return err
	}

	if command == "install" || command == "sideload" {
		exportedN, err := pico.ExportDesktopEntries(args...)
		if err == nil {
			cmdr.Info.Printfln(vso.Trans("pico.info.exportedApps"), exportedN)
		}
	}

	return nil
}

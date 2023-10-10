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

func NewWayCommand() []*cmdr.Command {
	// Root command
	cmd := cmdr.NewCommand(
		"android",
		vso.Trans("waydroid.description"),
		vso.Trans("waydroid.description"),
		nil,
	)

	// Subcommands
	installCmd := cmdr.NewCommand(
		"install",
		vso.Trans("waydroid.export.description"),
		vso.Trans("waydroid.export.description"),
		wayInstall,
	)

	searchCmd := cmdr.NewCommand(
		"search",
		vso.Trans("waydroid.search.description"),
		vso.Trans("waydroid.search.descrioption"),
		waySearch,
	)

	initCmd := cmdr.NewCommand(
		"init",
		vso.Trans("waydroid.init.description"),
		vso.Trans("waydroid.init.description"),
		wayInit,
	)
	initCmd.WithBoolFlag(
		cmdr.NewBoolFlag(
			"force",
			"f",
			vso.Trans("waydroid.init.options.force.description"),
			false,
		),
	)

	launchCmd := cmdr.NewCommand(
		"launch",
		vso.Trans("waydroid.launch.description"),
		vso.Trans("waydroid.launch.description"),
		wayLaunch,
	)

	launcherCmd := cmdr.NewCommand(
		"launcher",
		vso.Trans("waydroid.launcher.description"),
		vso.Trans("waydroid.launcher.description"),
		wayLauncher,
	)

	// Add subcommands to root
	cmd.AddCommand(installCmd)
	cmd.AddCommand(searchCmd)
	cmd.AddCommand(initCmd)
	cmd.AddCommand(launchCmd)
	cmd.AddCommand(launcherCmd)

	return []*cmdr.Command{cmd}
}

func wayInit(cmd *cobra.Command, args []string) error {
	force, _ := cmd.Flags().GetBool("force")

	if core.WayExists() && !force {
		cmdr.Error.Println(vso.Trans("waydroid.error.alreadyInitialized"))
		return nil
	}

	err := core.WayInit()
	if err != nil {
		return err
	}

	cmdr.Success.Println(vso.Trans("waydroid.info.initialized"))
	return nil
}

func wayInstall(cmd *cobra.Command, args []string) error {
	way, err := core.GetWay()
	if err != nil {
		return err
	}

	if len(args) != 1 {
		return fmt.Errorf("can only install one apk per run") // TODO: improve this error message
	}

	finalArgs := []string{"ewaydroid", "app", "install", args[0]}
	fmt.Print(finalArgs)
	_, err = way.Exec(false, finalArgs...)
	return err
}

func waySearch(cmd *cobra.Command, args []string) error {
	if len(args) > 1 {
		return fmt.Errorf("too many arguments")
	} else if len(args) < 1 {
		return fmt.Errorf("not enough arguments")
	}
	err := core.SearchPackage(args[0])
	return err
}

func wayLaunch(cmd *cobra.Command, args []string) error {
	way, err := core.GetWay()
	if err != nil {
		return err
	}

	finalArgs := []string{"ewaydroid", "app", "launch"}
	_, err = way.Exec(false, finalArgs...)
	return err
}

func wayLauncher(cmd *cobra.Command, args []string) error {
	way, err := core.GetWay()
	if err != nil {
		return err
	}

	finalArgs := []string{"ewaydroid", "show-full-ui"}
	_, err = way.Exec(false, finalArgs...)
	return err
}

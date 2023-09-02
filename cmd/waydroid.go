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

	finalArgs := []string{"ewaydroid", "app", "install"}
	_, err = way.Exec(false, finalArgs...)
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

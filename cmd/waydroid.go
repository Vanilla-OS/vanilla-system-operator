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
	"strings"
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
	installCmd.WithBoolFlag(
		cmdr.NewBoolFlag(
			"local",
			"l",
			vso.Trans("waydroid.export.options.local.description"), false,
		),
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

	removeCmd := cmdr.NewCommand(
		"remove",
		vso.Trans("waydroid.remove.description"),
		vso.Trans("waydroid.remove.description"),
		wayRemove,
	)

	searchCmd := cmdr.NewCommand(
		"search",
		vso.Trans("waydroid.search.description"),
		vso.Trans("waydroid.search.description"),
		waySearch,
	)

	syncCmd := cmdr.NewCommand(
		"sync",
		vso.Trans("waydroid.sync.description"),
		vso.Trans("waydroid.sync.description"),
		waySync,
	)

	// Add subcommands to root
	cmd.AddCommand(installCmd)
	cmd.AddCommand(initCmd)
	cmd.AddCommand(launchCmd)
	cmd.AddCommand(launcherCmd)
	cmd.AddCommand(removeCmd)
	cmd.AddCommand(searchCmd)
	cmd.AddCommand(syncCmd)

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
	if len(args) == 0 {
		return fmt.Errorf("no arguments provided")
	}
	localFlag, _ := cmd.Flags().GetBool("local")
	way, err := core.GetWay()
	if err != nil {
		return err
	}
	var apk string
	if !localFlag {
		apk, err = core.FetchPackage(strings.Join(args, " ")) // Can only install one thing at once, so might as well merge everything as one term
		if err != nil {
			return err
		}
	} else {
		apk = args[0]
	}

	finalArgs := []string{"ewaydroid", "app", "install", apk}
	//fmt.Println(finalArgs)
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

func wayRemove(cmd *cobra.Command, args []string) error {
	way, err := core.GetWay()
	if err != nil {
		return err
	}
	search := strings.Join(args, " ")
	packages, err := core.GetWayPackages(way)
	var rem []string
	for _, pkg := range packages {
		if strings.Contains(pkg[0], search) || strings.Contains(pkg[1], search) {
			fmt.Printf("Removing package %s (%s)\n", pkg[0], pkg[1])
			rem = pkg
		}
	}
	finalArgs := []string{"ewaydroid", "app", "remove", rem[1]}
	_, err = way.Exec(false, finalArgs...)
	return err
}

func waySearch(cmd *cobra.Command, args []string) error {
	err := core.SearchPackage(strings.Join(args, " ")) // Can only search for one thing at once, so might as well merge everything as one term
	return err
}

func waySync(cmd *cobra.Command, args []string) error {
	err := core.GetRepos()
	if err != nil {
		return err
	}
	err = core.SyncIndex(true)
	return err
}

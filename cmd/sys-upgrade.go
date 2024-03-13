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
	"os"

	"github.com/spf13/cobra"
	"github.com/vanilla-os/orchid/cmdr"
	"github.com/vanilla-os/vso/core"
)

func NewUpgradeCommand() *cmdr.Command {
	// Root command
	cmd := cmdr.NewCommand(
		"sys-upgrade",
		vso.Trans("sysUpgrade.description"),
		vso.Trans("sysUpgrade.description"),
		nil,
	)

	// Subcommands
	check := cmdr.NewCommand(
		"check",
		vso.Trans("sysUpgrade.check.description"),
		vso.Trans("sysUpgrade.check.description"),
		checkUpgrade,
	)

	check.WithBoolFlag(
		cmdr.NewBoolFlag(
			"as-exit-code",
			"x",
			vso.Trans("sysUpgrade.check.asExitCode"),
			false,
		),
	)
	check.WithBoolFlag(
		cmdr.NewBoolFlag(
			"json",
			"j",
			vso.Trans("sysUpgrade.check.json.description"),
			false,
		),
	)

	upgrade := cmdr.NewCommand(
		"upgrade",
		vso.Trans("sysUpgrade.sysUpgrade.description"),
		vso.Trans("sysUpgrade.sysUpgrade.description"),
		upgrade,
	)

	upgrade.WithBoolFlag(
		cmdr.NewBoolFlag(
			"now",
			"n",
			vso.Trans("sysUpgrade.sysUpgrade.now"),
			false,
		),
	)

	// Add subcommands to root
	cmd.AddCommand(check)
	cmd.AddCommand(upgrade)

	return cmd
}

func checkUpgrade(cmd *cobra.Command, args []string) error {
	asExitCode, _ := cmd.Flags().GetBool("as-exit-code")
	json, _ := cmd.Flags().GetBool("json")

	if asExitCode && json {
		cmdr.Error.Println(vso.Trans("sysUpgrade.check.error.asExitCodeAndJson"))
		return nil
	}

	if !asExitCode && !json {
		cmdr.Info.Println(vso.Trans("sysUpgrade.check.info.checking"))
	}

	if asExitCode {
		status, err := core.HasUpdates()
		if err != nil {
			cmdr.Error.Println(vso.Trans("sysUpgrade.check.error.asExitCodeAndJson"))
			return err
		}
		if status {
			os.Exit(1)
		}
		os.Exit(0)
	}

	var err error
	if !json {
		err = core.RunUpgradeCheck()
		if err != nil {
			cmdr.Error.Println(err)
			return nil
		}
	} else {
		_, err = core.RunUpgradeCheckJSON()
	}
	if err != nil {
		cmdr.Error.Println(err)
		return nil
	}

	return nil
}

func upgrade(cmd *cobra.Command, args []string) error {
	now, _ := cmd.Flags().GetBool("now")

	if now {
		err := core.TryUpdate(true)
		if err != nil {
			cmdr.Error.Println(vso.Trans("sysUpgrade.sysUpgrade.error.updating"))
			return nil
		}

		return nil
	}

	if core.NeedUpdate() {
		err := core.TryUpdate(false)
		if err != nil {
			cmdr.Error.Println(vso.Trans("sysUpgrade.sysUpgrade.error.updating"))
			return nil
		}

		return nil
	}

	cmdr.Info.Println(vso.Trans("sysUpgrade.sysUpgrade.info.noUpdates"))
	return nil
}

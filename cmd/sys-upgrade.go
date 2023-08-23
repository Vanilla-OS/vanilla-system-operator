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
	"os"
	"strings"

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
			vso.Trans("sysUpgrade.check.json"),
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

	status, updates, err := core.HasUpdates()

	if asExitCode {
		if status {
			os.Exit(1)
		}
		os.Exit(0)
	}

	if !json {
		if err != nil {
			cmdr.Error.Println(err)
			return nil
		}

		if !status {
			cmdr.Info.Println(vso.Trans("sysUpgrade.check.info.noUpdates"))
			return nil
		}

		fmt.Println("--------------------------------------------")
		fmt.Println(vso.Trans("sysUpgrade.check.info.updatesAvailable"))
		fmt.Println(strings.Join(updates, "\n"))
		fmt.Println("--------------------------------------------")

		return nil
	} else {
		if err != nil {
			cmdr.Error.Println(err)
			return nil
		}

		if !status {
			fmt.Println(`{"status": "ok", "updates": []}`)
			return nil
		}

		fmt.Println(`{"status": "ok", "updates": [`)
		for i, update := range updates {
			if i == len(updates)-1 {
				fmt.Printf(`	"%s"`, update)
			} else {
				fmt.Printf(`	"%s",`, update)
			}
		}
		fmt.Println("\n]}")
		return nil
	}
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

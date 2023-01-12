package cmd

/*	License: GPLv3
	Authors:
		Mirko Brombin <send@mirko.pm>
		Pietro di Caprio <pietro@fabricators.ltd>
	Copyright: 2022
	Description: VSO is a utility which allows you to perform maintenance
	tasks on your Vanilla OS installation.
*/

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/vanilla-os/vso/core"
)

func checkUpdateUsage(*cobra.Command) error {
	fmt.Print(`Description: 
	Check for system updates

Usage:
  	vso update-check [options]

Options:
	--help/-h		show this message
	--as-exit-code		return 0 if no updates are available, 1 otherwise

Examples:
	vso update-check
`)
	return nil
}

func NewCheckUpdateCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update-check",
		Short: "Check for system updates",
		RunE:  checkUpdate,
	}
	cmd.SetUsageFunc(checkUpdateUsage)
	cmd.Flags().BoolP("as-exit-code", "", false, "return 0 if no updates are available, 1 otherwise")
	return cmd
}

func checkUpdate(cmd *cobra.Command, args []string) error {
	if !core.RootCheck(true) {
		return nil
	}

	if !cmd.Flag("as-exit-code").Changed {
		fmt.Println("Checking for updates...")
	}
	status, updates, err := core.HasUpdates()

	if cmd.Flag("as-exit-code").Changed {
		if status {
			os.Exit(1)
		} else {
			os.Exit(0)
		}
	}

	if !status {
		fmt.Println("Your system is up-to-date.")
		return nil
	}

	if err != nil {
		fmt.Println("An error occurred while checking for updates.")
		return err
	}

	fmt.Println("--------------------------------------------")
	fmt.Println("The following packages have pending updates:")
	fmt.Println(strings.Join(updates, "\n"))
	fmt.Println("--------------------------------------------")

	return nil
}

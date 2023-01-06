package cmd

/*	License: GPLv3
	Authors:
		Mirko Brombin <send@mirko.pm>
		Pietro di Caprio <pietro@fabricators.ltd>
	Copyright: 2023
	Description: VSO is a utility which allows you to perform maintenance
	tasks on your Vanilla OS installation.
*/

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/vanilla-os/vso/core"
)

func triggerUpdateUsage(*cobra.Command) error {
	fmt.Print(`Description: 
	Trigger a system update

Usage:
  	vso trigger-update [flags]

Flags:
	--help/-h		show this message
	--now			trigger a system update immediately

Examples:
	vso trigger-update --now
`)
	return nil
}

func NewTriggerUpdateCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "trigger-update",
		Short: "Trigger a system update",
		RunE:  triggerUpdate,
	}
	cmd.SetUsageFunc(triggerUpdateUsage)
	cmd.Flags().BoolP("now", "n", false, "trigger an update now")
	return cmd
}

func triggerUpdate(cmd *cobra.Command, args []string) error {
	if !core.RootCheck(true) {
		return nil
	}

	msg_fail_update := `An error occurred while trying to update the system.
	Please try again later and report the issue if it persists.
	
	ERR_CODE: %s`

	if cmd.Flags().Changed("now") {
		fmt.Println("Triggering update now...")
		err := core.TryUpdate()
		if err != nil {
			fmt.Println(msg_fail_update, "CMD_TR_UP_NOW")
			return err
		}
		return nil
	}

	if core.NeedUpdate() {
		err := core.TryUpdate()
		if err != nil {
			fmt.Println(msg_fail_update, "CMD_TR_UP_NEED")
			return err
		}
		return nil
	}

	fmt.Println("No update needed.")
	return nil
}

package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/vanilla-os/vso/core"
)

func rotateTasksUsage(*cobra.Command) error {
	fmt.Print(`Description: 
	  Rotate tasks

Usage:
	vso rotate-tasks [options]

Options:
	--help/-h		show this message

Examples:
	vso rotate-tasks
`)

	return nil
}

func NewRotateTasksCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "rotate-tasks",
		Short: "Rotate tasks",
		RunE:  rotateTasks,
	}
	cmd.SetUsageFunc(rotateTasksUsage)
	cmd.Flags().StringP("private-event", "e", "", "private event to rotate tasks for (boot, shutdown, login, logout)")

	return cmd
}

func rotateTasks(cmd *cobra.Command, args []string) error {
	event, err := cmd.Flags().GetString("private-event")
	if err != nil {
		event = "no-system-event"
	}

	return core.Rotate(event)
}

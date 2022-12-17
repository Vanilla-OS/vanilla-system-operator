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

	"github.com/spf13/cobra"
	"github.com/vanilla-os/vso/core"
)

func deleteTaskUsage(*cobra.Command) error {
	fmt.Print(`Description: 
  Delete a task

Usage:
	vso delete-task [task] [options]

Options:
	--help/-h		show this message

Examples:
	vso delete-task my-task
	vso delete-task "my task"
`)

	return nil
}

func NewDeleteTaskCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete-task",
		Short: "Delete a task",
		RunE:  deleteTask,
	}
	cmd.SetUsageFunc(deleteTaskUsage)

	return cmd
}

func deleteTask(cmd *cobra.Command, args []string) error {
	if len(args) != 1 {
		return fmt.Errorf("invalid number of arguments")
	}

	err := core.DeleteTaskByUnitName(args[0])
	if err != nil {
		fmt.Println("Task", args[0], "does not exist")
		return err
	}

	fmt.Println("Task", args[0], "deleted")

	return nil
}

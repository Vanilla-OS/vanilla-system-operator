package cmd

/*	License: GPLv3
	Authors:
		Mirko Brombin <send@mirko.pm>
		Pietro di Caprio <pietro@fabricators.ltd>
	Copyright: 2022
	Description: VSO is an utility which allows you to perform maintenance
	tasks on your Vanilla OS installation.
*/

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/vanilla-os/vso/core"
)

func listTasksUsage(*cobra.Command) error {
	fmt.Print(`Description: 
  List all tasks

Usage:
	vso list-tasks [options] [arguments]

Options:
	--help/-h		show this message

Examples:
	vso list-tasks
`)

	return nil
}

func NewListTasksCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list-tasks",
		Short: "List all tasks",
		RunE:  listTasks,
	}
	cmd.SetUsageFunc(listTasksUsage)

	return cmd
}

func listTasks(cmd *cobra.Command, args []string) error {
	tasks, err := core.ListDetailed()
	if err != nil {
		return err
	}

	fmt.Println("Found", len(tasks), "tasks:")
	fmt.Println("--------------------")

	for _, task := range tasks {
		fmt.Println("- " + task.Name)
		fmt.Println("  Description: " + task.Description)
		fmt.Println("  LastExecution: " + task.LastExecution.String())
		fmt.Println("--------------------")
	}

	return nil
}

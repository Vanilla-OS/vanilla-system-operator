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

func listTasksUsage(*cobra.Command) error {
	fmt.Print(`Description: 
  List all tasks

Usage:
	vso list-tasks [options]

Options:
	--help/-h		show this message
	--json/-j		output in JSON format

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
	cmd.Flags().BoolP("json", "j", false, "output in JSON format")
	cmd.SetUsageFunc(listTasksUsage)

	return cmd
}

func listTasks(cmd *cobra.Command, args []string) error {
	json, err := cmd.Flags().GetBool("json")
	if err != nil {
		return err
	}

	if json {
		return listTasksJson()
	}

	tasks, err := core.ListTasksDetailed()
	if err != nil {
		return err
	}

	fmt.Println("Found", len(tasks), "tasks:")
	fmt.Println("--------------------")

	for _, task := range tasks {
		relations := task.Relations()
		dependencies := task.Dependencies()

		fmt.Println("- " + task.Name)
		fmt.Println("  Description: " + task.Description)
		fmt.Println("  LastExecution: " + task.LastExecution.String())
		if len(relations) > 0 {
			fmt.Println("  Relations:")
			for _, relation := range relations {
				fmt.Println("    - " + relation.Name)
			}
		}
		if len(dependencies) > 0 {
			fmt.Println("  Dependencies:")
			for _, dependency := range dependencies {
				fmt.Println("    - " + dependency.Name)
			}
		}
		fmt.Println("--------------------")
	}

	return nil
}

func listTasksJson() error {
	json, err := core.ListTasksJson()
	if err != nil {
		return err
	}

	fmt.Println(string(json))

	return nil
}

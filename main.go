package main

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
	"github.com/vanilla-os/vso/cmd"
)

var (
	Version = "0.1"
)

func help(cmd *cobra.Command, args []string) {
	fmt.Println(`Usage:
	vso [options] [command] [arguments]

Options:
	-h, --help            	Show this help message and exit

Commands:
	config              	Configure VSO
	create-task             Create a new task
	delete-task             Delete a task
	developer-program   	Join the developers program
	help                	Show this help message and exit
	list-tasks          	List all tasks
	rotate-tasks		Rotate tasks
	trigger-update	  	Trigger a system update
	version             	Show version and exit`)
}

func newVsoCommand() *cobra.Command {
	return &cobra.Command{
		Use:     "vso",
		Short:   "VSO is an utility which allows you to perform maintenance tasks on your Vanilla OS installation.",
		Version: Version,
	}
}

func main() {
	rootCmd := newVsoCommand()

	rootCmd.AddCommand(cmd.NewCreateTaskCommand())
	rootCmd.AddCommand(cmd.NewDeleteTaskCommand())
	rootCmd.AddCommand(cmd.NewConfigCommand())
	rootCmd.AddCommand(cmd.NewDevProgramCommand())
	rootCmd.AddCommand(cmd.NewListTasksCommand())
	rootCmd.AddCommand(cmd.NewRotateTasksCommand())
	rootCmd.AddCommand(cmd.NewTriggerUpdateCommand())
	rootCmd.SetHelpFunc(help)
	rootCmd.Execute()
}

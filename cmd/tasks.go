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
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/vanilla-os/orchid/cmdr"
	"github.com/vanilla-os/vso/core"
)

func NewTasksCommand() *cmdr.Command {
	// Root command
	cmd := cmdr.NewCommand(
		"tasks",
		vso.Trans("tasks.description"),
		vso.Trans("tasks.description"),
		nil,
	)

	// List subcommand
	listCmd := cmdr.NewCommand(
		"list",
		vso.Trans("tasks.list.description"),
		vso.Trans("tasks.list.description"),
		listTasks,
	)

	listCmd.WithBoolFlag(
		cmdr.NewBoolFlag(
			"json",
			"j",
			vso.Trans("tasks.list.options.json.description"),
			false,
		),
	)

	// New subcommand
	newCmd := cmdr.NewCommand(
		"new",
		vso.Trans("tasks.new.description"),
		vso.Trans("tasks.new.description"),
		newTask,
	)
	newCmd.WithBoolFlag(
		cmdr.NewBoolFlag(
			"assume-yes",
			"y",
			vso.Trans("tasks.new.options.assumeYes.description"),
			false,
		),
	)
	newCmd.WithBoolFlag(
		cmdr.NewBoolFlag(
			"need-confirm",
			"",
			vso.Trans("tasks.new.options.needConfirm.description"),
			false,
		),
	)
	newCmd.WithStringFlag(
		cmdr.NewStringFlag(
			"name",
			"",
			vso.Trans("tasks.new.options.name.description"),
			"",
		),
	)
	newCmd.WithStringFlag(
		cmdr.NewStringFlag(
			"description",
			"",
			vso.Trans("tasks.new.options.description.description"),
			"",
		),
	)
	newCmd.WithStringFlag(
		cmdr.NewStringFlag(
			"command",
			"",
			vso.Trans("tasks.new.options.command.description"),
			"",
		),
	)
	newCmd.WithStringFlag(
		cmdr.NewStringFlag(
			"every",
			"",
			vso.Trans("tasks.new.options.every.description"),
			"",
		),
	)
	newCmd.WithStringFlag(
		cmdr.NewStringFlag(
			"at",
			"",
			vso.Trans("tasks.new.options.at.description"),
			"",
		),
	)
	newCmd.WithBoolFlag(
		cmdr.NewBoolFlag(
			"on-network",
			"",
			vso.Trans("tasks.new.options.onNetwork.description"),
			false,
		),
	)
	newCmd.WithBoolFlag(
		cmdr.NewBoolFlag(
			"on-disconnect",
			"",
			vso.Trans("tasks.new.options.onDisconnect.description"),
			false,
		),
	)
	newCmd.WithBoolFlag(
		cmdr.NewBoolFlag(
			"on-battery",
			"",
			vso.Trans("tasks.new.options.onBattery.description"),
			false,
		),
	)
	newCmd.WithBoolFlag(
		cmdr.NewBoolFlag(
			"on-low-battery",
			"",
			vso.Trans("tasks.new.options.onLowBattery.description"),
			false,
		),
	)
	newCmd.WithBoolFlag(
		cmdr.NewBoolFlag(
			"on-charge",
			"",
			vso.Trans("tasks.new.options.onCharge.description"),
			false,
		),
	)
	newCmd.WithBoolFlag(
		cmdr.NewBoolFlag(
			"on-full-battery",
			"",
			vso.Trans("tasks.new.options.onFullBattery.description"),
			false,
		),
	)
	newCmd.WithStringFlag(
		cmdr.NewStringFlag(
			"on-condition-command",
			"",
			vso.Trans("tasks.new.options.onConditionCommand.description"),
			"",
		),
	)
	newCmd.WithStringFlag(
		cmdr.NewStringFlag(
			"on-process",
			"",
			vso.Trans("tasks.new.options.onProcess.description"),
			"",
		),
	)
	newCmd.WithStringFlag(
		cmdr.NewStringFlag(
			"on-internet-usage",
			"",
			vso.Trans("tasks.new.options.onInternetUsage.description"),
			"",
		),
	)
	newCmd.WithBoolFlag(
		cmdr.NewBoolFlag(
			"on-high-internet-usage",
			"",
			vso.Trans("tasks.new.options.onHighInternetUsage.description"),
			false,
		),
	)
	newCmd.WithStringFlag(
		cmdr.NewStringFlag(
			"on-memory-usage",
			"",
			vso.Trans("tasks.new.options.onMemoryUsage.description"),
			"",
		),
	)
	newCmd.WithBoolFlag(
		cmdr.NewBoolFlag(
			"on-high-memory-usage",
			"",
			vso.Trans("tasks.new.options.onHighMemoryUsage.description"),
			false,
		),
	)
	newCmd.WithStringFlag(
		cmdr.NewStringFlag(
			"on-cpu-usage",
			"",
			vso.Trans("tasks.new.options.onCpuUsage.description"),
			"",
		),
	)
	newCmd.WithBoolFlag(
		cmdr.NewBoolFlag(
			"on-high-cpu-usage",
			"",
			vso.Trans("tasks.new.options.onHighCpuUsage.description"),
			false,
		),
	)
	newCmd.WithStringFlag(
		cmdr.NewStringFlag(
			"on-cpu-temp",
			"",
			vso.Trans("tasks.new.options.onCpuTemp.description"),
			"",
		),
	)
	newCmd.WithStringFlag(
		cmdr.NewStringFlag(
			"on-device-connect",
			"",
			vso.Trans("tasks.new.options.onDeviceConnect.description"),
			"",
		),
	)
	newCmd.WithStringFlag(
		cmdr.NewStringFlag(
			"on-device-disconnect",
			"",
			vso.Trans("tasks.new.options.onDeviceDisconnect.description"),
			"",
		),
	)

	// Rm subcommand
	rmTaskCmd := cmdr.NewCommand(
		"rm",
		vso.Trans("tasks.rm.description"),
		vso.Trans("tasks.rm.description"),
		removeTask,
	)

	rmTaskCmd.WithStringFlag(
		cmdr.NewStringFlag(
			"name",
			"n",
			vso.Trans("tasks.rm.options.name.description"),
			"",
		),
	)
	rmTaskCmd.WithBoolFlag(
		cmdr.NewBoolFlag(
			"force",
			"f",
			vso.Trans("tasks.rm.options.force.description"),
			false,
		),
	)

	// Rotate subcommand
	rotateCmd := cmdr.NewCommand(
		"rotate",
		vso.Trans("tasks.rotate.description"),
		vso.Trans("tasks.rotate.description"),
		rotateTasks,
	)

	rotateCmd.WithStringFlag(
		cmdr.NewStringFlag(
			"private-event",
			"e",
			vso.Trans("tasks.rotate.options.privateEvent.description"),
			"",
		),
	)

	// Add subcommands to stacks
	cmd.AddCommand(listCmd)
	cmd.AddCommand(newCmd)
	cmd.AddCommand(rmTaskCmd)
	cmd.AddCommand(rotateCmd)

	return cmd
}

func listTasks(cmd *cobra.Command, args []string) error {
	jsonFlag, _ := cmd.Flags().GetBool("json")

	if !jsonFlag {
		tasks, err := core.ListTasksDetailed()
		if err != nil {
			return err
		}

		tasksCount := len(tasks)
		if tasksCount == 0 {
			cmdr.Info.Println(vso.Trans("tasks.list.noTasks"))
			return nil
		}

		cmdr.Info.Printfln(vso.Trans("tasks.list.info.foundTasks"), tasksCount)

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
	} else {
		json, err := core.ListTasksJson()
		if err != nil {
			return err
		}

		fmt.Println(string(json))
	}

	return nil
}

func newTask(cmd *cobra.Command, args []string) error {
	assumeYes, _ := cmd.Flags().GetBool("assume-yes")
	name, _ := cmd.Flags().GetString("name")
	description, _ := cmd.Flags().GetString("description")
	nconfirm, _ := cmd.Flags().GetBool("need-confirm")
	command, _ := cmd.Flags().GetString("command")
	every, _ := cmd.Flags().GetString("every")
	at, _ := cmd.Flags().GetString("at")
	onNetwork, _ := cmd.Flags().GetBool("on-network")
	onDisconnect, _ := cmd.Flags().GetBool("on-disconnect")
	onBattery, _ := cmd.Flags().GetBool("on-battery")
	onLowBattery, _ := cmd.Flags().GetBool("on-low-battery")
	onCharge, _ := cmd.Flags().GetBool("on-charge")
	onFullBattery, _ := cmd.Flags().GetBool("on-full-battery")
	onConditionCommand, _ := cmd.Flags().GetString("on-condition-command")
	onProcess, _ := cmd.Flags().GetString("on-process")
	onInternetUsage, _ := cmd.Flags().GetInt("on-internet-usage")
	onHighInternetUsage, _ := cmd.Flags().GetBool("on-high-internet-usage")
	onMemoryUsage, _ := cmd.Flags().GetInt("on-memory-usage")
	onHighMemoryUsage, _ := cmd.Flags().GetBool("on-high-memory-usage")
	onCPUUsage, _ := cmd.Flags().GetInt("on-cpu-usage")
	onHighCPUUsage, _ := cmd.Flags().GetBool("on-high-cpu-usage")
	onCPUTemp, _ := cmd.Flags().GetInt("on-cpu-temp")
	onDeviceConnect, _ := cmd.Flags().GetString("on-device-connect")
	onDeviceDisconnect, _ := cmd.Flags().GetString("on-device-disconnect")

	if name == "" {
		if !assumeYes {
			cmdr.Info.Println(vso.Trans("tasks.stacks.new.info.askName"))
			reader := bufio.NewReader(os.Stdin)
			name, _ = reader.ReadString('\n')
			if name == "" {
				cmdr.Error.Println(vso.Trans("tasks.stacks.new.error.emptyName"))
				return nil
			}
		} else {
			cmdr.Error.Println(vso.Trans("tasks.stacks.new.error.noName"))
			return nil
		}
	}

	if description == "" {
		if !assumeYes {
			cmdr.Info.Println(vso.Trans("tasks.stacks.new.info.askDescription"))
			reader := bufio.NewReader(os.Stdin)
			description, _ = reader.ReadString('\n')
			if description == "" {
				cmdr.Error.Println(vso.Trans("tasks.stacks.new.error.emptyDescription"))
				return nil
			}
		} else {
			cmdr.Error.Println(vso.Trans("tasks.stacks.new.error.noDescription"))
			return nil
		}
	}

	if command == "" {
		if !assumeYes {
			cmdr.Info.Println(vso.Trans("tasks.stacks.new.info.askCommand"))
			reader := bufio.NewReader(os.Stdin)
			command, _ = reader.ReadString('\n')
			if command == "" {
				cmdr.Error.Println(vso.Trans("tasks.stacks.new.error.emptyCommand"))
				return nil
			}
		} else {
			cmdr.Error.Println(vso.Trans("tasks.stacks.new.error.noCommand"))
			return nil
		}
	}

	task := core.Task{
		Name:                 name,
		Description:          description,
		NeedConfirm:          nconfirm,
		Command:              command,
		Every:                every,
		At:                   at,
		OnNetwork:            onNetwork,
		OnDisconnect:         onDisconnect,
		OnBattery:            onBattery,
		OnLowBattery:         onLowBattery,
		OnCharge:             onCharge,
		OnFullBattery:        onFullBattery,
		OnConditionCommand:   onConditionCommand,
		OnProcess:            onProcess,
		OnInternetUsage:      onInternetUsage,
		OnHighInternetUsage:  onHighInternetUsage,
		OnMemoryUsage:        onMemoryUsage,
		OnHighMemoryUsage:    onHighMemoryUsage,
		OnCPUUsage:           onCPUUsage,
		OnHighCPUUsage:       onHighCPUUsage,
		OnCPUTemp:            onCPUTemp,
		OnDeviceConnected:    onDeviceConnect,
		OnDeviceDisconnected: onDeviceDisconnect,
	}

	err := task.Save()
	if err != nil {
		return err
	}

	fmt.Printf(vso.Trans("tasks.stacks.new.info.taskCreated"), name)
	return nil
}

func removeTask(cmd *cobra.Command, args []string) error {
	taskName, _ := cmd.Flags().GetString("name")
	if taskName == "" {
		cmdr.Error.Println(vso.Trans("tasks.rm.error.noName"))
		return nil
	}

	task, err := core.LoadTaskByUnitName(taskName)
	if err != nil {
		cmdr.Error.Println(vso.Trans("tasks.rm.error.notFound"))
		return nil
	}

	force, _ := cmd.Flags().GetBool("force")
	if !force {
		cmdr.Info.Printf(vso.Trans("tasks.rm.info.askConfirmation"), taskName)
		var confirmation string
		fmt.Scanln(&confirmation)
		if strings.ToLower(confirmation) != "y" {
			cmdr.Info.Println(vso.Trans("vso.info.aborting"))
			return nil
		}
	}

	err = task.Delete()
	if err != nil {
		cmdr.Error.Println(vso.Trans("tasks.rm.error.cannotDelete"))
		return err
	}

	cmdr.Info.Println(vso.Trans("tasks.rm.info.taskDeleted"))

	return nil
}

func rotateTasks(cmd *cobra.Command, args []string) error {
	event, err := cmd.Flags().GetString("private-event")
	if err != nil {
		event = "no-system-event"
	}

	return core.RotateTasks(event)
}

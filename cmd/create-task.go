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

func createTaskUsage(*cobra.Command) error {
	fmt.Print(`Description: 
	  Create a new task

Usage:
  	vso create-task [flags] [arguments]

Flags:
	--help/-h		show this message

Arguments:
	--name/-n		name of the task
	--description/-d	description of the task
	--need-confirm		ask for confirmation before executing the task
	--command/-c		command to execute
	--every/-e		execute the task every X (minutes, hours, days)
	--at/-a			execute the task at a specific time (hh:mm)
	--on-boot		execute the task on boot
	--on-network		execute the task on network connection
	--on-disconnect	execute the task on network disconnection
	--on-battery		execute the task on battery
	--on-low-battery	execute the task on low battery (20%)
	--on-charge		execute the task on charge
	--on-full-battery	execute the task on full battery
	--on-condition-command	execute the task on condition command
	--on-process		execute the task when a process came up
	--on-internet-usage	execute the task when internet usage is higher than Xkb/s
	--on-high-internet-usage	execute the task when internet usage is higher than 500kb/s
	--on-memory-usage	execute the task when memory usage is higher than X%
	--on-high-memory-usage	execute the task when memory usage is higher than 50%
	--on-cpu-usage		execute the task when cpu usage is higher than X%
	--on-high-cpu-usage	execute the task when cpu usage is higher than 50%
	--on-cpu-temp		execute the task when cpu temperature is higher than 60°C	

Examples:
	vso create-task -n "Battery fully charged" -d "notify at full charge" -c "notify-send 'Battery fully charged'" --on-full-battery
	vso create-task -n "Launch-terminal" -d "Launch terminal at Settings launch" -c "kgx" --on-process gnome-control-center
`)
	return nil
}

func NewCreateTaskCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create-task",
		Short: "Create a new task",
		RunE:  createTask,
	}
	cmd.SetUsageFunc(createTaskUsage)
	cmd.Flags().StringP("name", "n", "", "name of the task")
	cmd.Flags().StringP("description", "d", "", "description of the task")
	cmd.Flags().Bool("need-confirm", false, "ask for confirmation before executing the task")
	cmd.Flags().StringP("command", "c", "", "command to execute")
	//cmd.Flags().String("after-task", "", "execute the task after another task")
	//cmd.Flags().String("after-task-success", "", "execute the task after another task if it was successful")
	//cmd.Flags().String("after-task-failure", "", "execute the task after another task if it failed")
	cmd.Flags().StringP("every", "e", "", "execute the task every X (minutes, hours, days)")
	cmd.Flags().StringP("at", "a", "", "execute the task at a specific time (hh:mm)")
	cmd.Flags().Bool("on-boot", false, "execute the task on boot")
	cmd.Flags().Bool("on-network", false, "execute the task on network connection")
	cmd.Flags().Bool("on-disconnect", false, "execute the task on network disconnection")
	cmd.Flags().Bool("on-battery", false, "execute the task on battery")
	cmd.Flags().Bool("on-low-battery", false, "execute the task on low battery (20%)")
	cmd.Flags().Bool("on-charge", false, "execute the task on charge")
	cmd.Flags().Bool("on-full-battery", false, "execute the task on full battery")
	cmd.Flags().String("on-condition-command", "", "execute the task on condition command")
	cmd.Flags().String("on-process", "", "execute the task when a process came up")
	cmd.Flags().Int("on-internet-usage", 0, "execute the task when internet usage is higher than Xkb/s")
	cmd.Flags().Bool("on-high-internet-usage", false, "execute the task when internet usage is higher than 500kb/s")
	cmd.Flags().Int("on-memory-usage", 0, "execute the task when memory usage is higher than X%")
	cmd.Flags().Bool("on-high-memory-usage", false, "execute the task when memory usage is higher than 50%")
	cmd.Flags().Int("on-cpu-usage", 0, "execute the task when cpu usage is higher than X%")
	cmd.Flags().Bool("on-high-cpu-usage", false, "execute the task when cpu usage is higher than 50%")
	cmd.Flags().Int("on-cpu-temp", 0, "execute the task when cpu temperature is higher than 60°C")

	return cmd
}

func createTask(cmd *cobra.Command, args []string) error {
	if core.RootCheck(false) {
		return fmt.Errorf("do not run this command as root")
	}

	err := core.TasksInit()
	if err != nil {
		return err
	}

	// checking for mandatory arguments
	name, _ := cmd.Flags().GetString("name")
	description, _ := cmd.Flags().GetString("description")
	nconfirm, _ := cmd.Flags().GetBool("need-confirm")
	command, _ := cmd.Flags().GetString("command")
	if name == "" || description == "" || command == "" {
		return fmt.Errorf("missing mandatory arguments")
	}

	// checking for optional arguments
	//afterTask, _ := cmd.Flags().GetString("after-task")
	//afterTaskSuccess, _ := cmd.Flags().GetString("after-task-success")
	//afterTaskFailure, _ := cmd.Flags().GetString("after-task-failure")
	every, _ := cmd.Flags().GetString("every")
	at, _ := cmd.Flags().GetString("at")
	onBoot, _ := cmd.Flags().GetBool("on-boot")
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

	// creating the task
	task := core.Task{
		Name:        name,
		Description: description,
		NeedConfirm: nconfirm,
		Command:     command,

		AfterTask:        "",
		AfterTaskSuccess: "",
		AfterTaskFailure: "",
		Every:            every,
		At:               at,

		OnBoot:              onBoot,
		OnNetwork:           onNetwork,
		OnDisconnect:        onDisconnect,
		OnBattery:           onBattery,
		OnLowBattery:        onLowBattery,
		OnCharge:            onCharge,
		OnFullBattery:       onFullBattery,
		OnConditionCommand:  onConditionCommand,
		OnProcess:           onProcess,
		OnInternetUsage:     onInternetUsage,
		OnHighInternetUsage: onHighInternetUsage,
		OnMemoryUsage:       onMemoryUsage,
		OnHighMemoryUsage:   onHighMemoryUsage,
		OnCPUUsage:          onCPUUsage,
		OnHighCPUUsage:      onHighCPUUsage,
		OnCPUTemp:           onCPUTemp,
	}

	// saving the task
	err = task.Save()
	if err != nil {
		return err
	}

	fmt.Printf("Task %s created\n", name)
	return nil
}

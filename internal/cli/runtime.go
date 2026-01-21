package cli

/*	License: GPLv3
	Authors:
		Mirko Brombin <brombin94@gmail.com>
		Pietro di Caprio <pietro@fabricators.ltd>
		Vanilla OS Contributors <https://github.com/vanilla-os/>
	Copyright: 2024
	Description: VSO is a utility which allows you to perform maintenance
	tasks on your Vanilla OS installation.
*/

import (
	"fmt"
	"strings"

	apxCore "github.com/vanilla-os/apx/v2/core"
	"github.com/vanilla-os/vanilla-system-operator/core"
	"github.com/vanilla-os/vanilla-system-operator/settings"
)

// Helpers

func ensurePicoInitialized() (*apxCore.SubSystem, error) {
	pico, err := core.GetPico()
	if err != nil {
		return nil, err
	}
	if !core.PicoExists() {
		VSO.Log.Error(VSO.LC.Get("vso.cmd.pico.error.notInitialized"))
		return nil, nil
	}
	return pico, nil
}

func getPicoAndPkgManager() (*apxCore.SubSystem, *apxCore.PkgManager, error) {
	pico, err := ensurePicoInitialized()
	if err != nil || pico == nil {
		return nil, nil, err
	}

	pkgManager, err := pico.Stack.GetPkgManager()
	if err != nil {
		return nil, nil, err
	}
	return pico, pkgManager, nil
}

// Tasks

func (c *TasksListCmd) Run() error {
	if !c.Json {
		tasks, err := core.ListTasksDetailed()
		if err != nil {
			return err
		}

		tasksCount := len(tasks)
		if tasksCount == 0 {
			VSO.Log.Info(VSO.LC.Get("vso.cmd.tasks.list.noTasks"))
			return nil
		}

		VSO.Log.Info(fmt.Sprintf(VSO.LC.Get("vso.cmd.tasks.list.info.foundTasks"), tasksCount))

		headers := []string{
			VSO.LC.Get("vso.cmd.tasks.labels.name"),
			VSO.LC.Get("vso.cmd.tasks.labels.description"),
			VSO.LC.Get("vso.cmd.tasks.labels.relations"),
			VSO.LC.Get("vso.cmd.tasks.labels.dependencies"),
			VSO.LC.Get("vso.cmd.tasks.labels.lastExecution"),
		}
		var data [][]string

		for _, task := range tasks {
			data = append(data, []string{
				task.Name,
				task.Description,
				fmt.Sprintf("%d", len(task.Relations())),
				fmt.Sprintf("%d", len(task.Dependencies())),
				task.LastExecution.Format("2006-01-02 15:04:05"),
			})
		}
		VSO.CLI.Table(headers, data)
	} else {
		json, err := core.ListTasksJson()
		if err != nil {
			return err
		}
		fmt.Println(string(json))
	}
	return nil
}

func (c *TasksNewCmd) Run() error {
	name := c.Name
	description := c.Description
	command := c.Command

	if name == "" {
		if !c.AssumeYes {
			res, err := VSO.CLI.PromptText(VSO.LC.Get("vso.cmd.tasks.stacks.new.info.askName"), "")
			if err != nil || res == "" {
				VSO.Log.Error(VSO.LC.Get("vso.cmd.tasks.stacks.new.error.emptyName"))
				return nil
			}
			name = res
		} else {
			VSO.Log.Error(VSO.LC.Get("vso.cmd.tasks.stacks.new.error.noName"))
			return nil
		}
	}

	if description == "" {
		if !c.AssumeYes {
			res, err := VSO.CLI.PromptText(VSO.LC.Get("vso.cmd.tasks.stacks.new.info.askDescription"), "")
			if err != nil || res == "" {
				VSO.Log.Error(VSO.LC.Get("vso.cmd.tasks.stacks.new.error.emptyDescription"))
				return nil
			}
			description = res
		} else {
			VSO.Log.Error(VSO.LC.Get("vso.cmd.tasks.stacks.new.error.noDescription"))
			return nil
		}
	}

	if command == "" {
		if !c.AssumeYes {
			res, err := VSO.CLI.PromptText(VSO.LC.Get("vso.cmd.tasks.stacks.new.info.askCommand"), "")
			if err != nil || res == "" {
				VSO.Log.Error(VSO.LC.Get("vso.cmd.tasks.stacks.new.error.emptyCommand"))
				return nil
			}
			command = res
		} else {
			VSO.Log.Error(VSO.LC.Get("vso.cmd.tasks.stacks.new.error.noCommand"))
			return nil
		}
	}

	task := core.Task{
		Name:                 name,
		Description:          description,
		NeedConfirm:          c.NeedConfirm,
		Command:              command,
		Every:                c.Every,
		At:                   c.At,
		OnNetwork:            c.OnNetwork,
		OnDisconnect:         c.OnDisconnect,
		OnBattery:            c.OnBattery,
		OnLowBattery:         c.OnLowBattery,
		OnCharge:             c.OnCharge,
		OnFullBattery:        c.OnFullBattery,
		OnConditionCommand:   c.OnConditionCommand,
		OnProcess:            c.OnProcess,
		OnInternetUsage:      c.OnInternetUsage,
		OnHighInternetUsage:  c.OnHighInternetUsage,
		OnMemoryUsage:        c.OnMemoryUsage,
		OnHighMemoryUsage:    c.OnHighMemoryUsage,
		OnCPUUsage:           c.OnCpuUsage,
		OnHighCPUUsage:       c.OnHighCpuUsage,
		OnCPUTemp:            c.OnCpuTemp,
		OnDeviceConnected:    c.OnDeviceConnect,
		OnDeviceDisconnected: c.OnDeviceDisconnect,
	}

	err := task.Save()
	if err != nil {
		return err
	}

	VSO.Log.Info(fmt.Sprintf(VSO.LC.Get("vso.cmd.tasks.stacks.new.info.taskCreated"), name))
	return nil
}

func (c *TasksRmCmd) Run() error {
	taskName := c.Name
	if taskName == "" {
		VSO.Log.Error(VSO.LC.Get("vso.cmd.tasks.rm.error.noName"))
		return nil
	}

	task, err := core.LoadTaskByUnitName(taskName)
	if err != nil {
		VSO.Log.Error(VSO.LC.Get("vso.cmd.tasks.rm.error.notFound"))
		return nil
	}

	if !c.Force {
		res, err := VSO.CLI.ConfirmAction(
			fmt.Sprintf(VSO.LC.Get("vso.cmd.tasks.rm.info.askConfirmation"), taskName),
			"y", "N",
			false,
		)
		if err != nil || !res {
			VSO.Log.Info(VSO.LC.Get("vso.info.aborting"))
			return nil
		}
	}

	err = task.Delete()
	if err != nil {
		VSO.Log.Error(VSO.LC.Get("vso.cmd.tasks.rm.error.cannotDelete"))
		return err
	}

	VSO.Log.Info(VSO.LC.Get("vso.cmd.tasks.rm.info.taskDeleted"))
	return nil
}

func (c *TasksRotateCmd) Run() error {
	event := c.PrivateEvent
	if event == "" {
		event = "no-system-event"
	}
	return core.RotateTasks(event, c.Silent)
}

// Upgrade

func (c *UpgradeCheckCmd) Run() error {
	if c.Json {
		_, err := core.RunUpgradeCheckJSON()
		if err != nil {
			return err
		}
		return nil
	}

	VSO.Log.Info(VSO.LC.Get("vso.cmd.sysUpgrade.check.info.checking"))
	err := core.RunUpgradeCheck()
	if err != nil {
		VSO.Log.Error(err.Error())
		return nil
	}
	return nil
}

func (c *UpgradeCmd) Run() error {
	if c.Now {
		err := core.TryUpdate(true)
		if err != nil {
			VSO.Log.Error(VSO.LC.Get("vso.cmd.sysUpgrade.error.updating"))
			return nil
		}
		return nil
	}

	if core.NeedUpdate() {
		err := core.TryUpdate(false)
		if err != nil {
			VSO.Log.Error(VSO.LC.Get("vso.cmd.sysUpgrade.error.updating"))
			return nil
		}
		return nil
	} else {
		status, err := core.HasUpdates()
		if err != nil {
			VSO.Log.Error(fmt.Sprintf(VSO.LC.Get("vso.cmd.sysUpgrade.error.onHasUpdate"), err))
			return nil
		}
		if status {
			schedule := settings.GetConfigValue("updates.schedule")
			if schedule == settings.ScheduleNever {
				VSO.Log.Info(VSO.LC.Get("vso.cmd.sysUpgrade.info.willNeverUpdate"))
			} else {
				scheduleStr, ok := schedule.(string)
				if !ok {
					scheduleStr = "unknown"
				}
				scheduleTrans := VSO.LC.Get(fmt.Sprintf("vso.cmd.sysUpgrade.schedule.%s", scheduleStr))
				VSO.Log.Info(fmt.Sprintf(VSO.LC.Get("vso.cmd.sysUpgrade.info.willUpdateLater"), scheduleTrans))
			}
			return nil
		}
	}

	VSO.Log.Info(VSO.LC.Get("vso.cmd.sysUpgrade.info.noUpdates"))
	return nil
}

// Config

func (c *ConfigGetCmd) Run() error {
	if c.Key == "" {
		VSO.Log.Error(VSO.LC.Get("vso.cmd.config.get.error.noKey"))
		return nil
	}
	key := c.Key
	val := settings.GetConfigValue(key)
	if val == nil {
		VSO.Log.Error(fmt.Sprintf(VSO.LC.Get("vso.cmd.config.get.error.noValue"), key))
		return nil
	}
	fmt.Println(val)
	return nil
}

func (c *ConfigSetCmd) Run() error {
	if c.Key == "" {
		VSO.Log.Error(VSO.LC.Get("vso.cmd.config.set.error.noKey"))
		return nil
	}
	if c.Value == "" {
		VSO.Log.Error(VSO.LC.Get("vso.cmd.config.set.error.noValue"))
		return nil
	}
	key := c.Key
	value := c.Value

	err := settings.SetConfigValue(key, value)
	if err != nil {
		VSO.Log.Error(fmt.Sprintf(VSO.LC.Get("vso.cmd.config.set.error.failed"), err.Error()))
		return nil
	}
	VSO.Log.Info(VSO.LC.Get("vso.cmd.config.set.success"))
	return nil
}

func (c *ConfigShowCmd) Run() error {
	kv := settings.GetConfigAsKV()
	for k, v := range kv {
		fmt.Printf("%s: %s\n", k, v)
	}
	return nil
}

// Pico

func (c *PicoInitCmd) Run() error {
	if !core.PicoExists() {
		VSO.Log.Info(VSO.LC.Get("vso.cmd.pico.info.initializing"))
		err := core.PicoInit()
		if err != nil {
			return err
		}
	} else if c.Force {
		VSO.Log.Info(VSO.LC.Get("vso.cmd.pico.info.deleting"))
		err := core.PicoDelete()
		if err != nil {
			return err
		}
		VSO.Log.Info(VSO.LC.Get("vso.cmd.pico.info.initializing"))
		err = core.PicoInit()
		if err != nil {
			return err
		}
	} else {
		VSO.Log.Error(VSO.LC.Get("vso.cmd.pico.error.alreadyInitialized"))
		return nil
	}

	VSO.Log.Info(VSO.LC.Get("vso.cmd.pico.info.initialized"))
	return nil
}

func (c *PicoInstallCmd) Run() error {
	if len(c.Args) == 0 {
		return fmt.Errorf("package argument required")
	}

	pico, pkgManager, err := getPicoAndPkgManager()
	if err != nil {
		return err
	}
	if pico == nil {
		return nil
	}

	exportedN, err := pico.ExportDesktopEntries(c.Args...)
	if err == nil && exportedN > 0 {
		VSO.Log.Info(fmt.Sprintf(VSO.LC.Get("vso.cmd.pico.info.exportedApps"), exportedN))
	}

	finalArgs := pkgManager.GenCmd(pkgManager.CmdInstall, c.Args...)
	_, err = pico.Exec(false, false, finalArgs...)
	return err
}

func (c *PicoRemoveCmd) Run() error {
	pico, pkgManager, err := getPicoAndPkgManager()
	if err != nil {
		return err
	}
	if pico == nil {
		return nil
	}

	exportedN, err := pico.UnexportDesktopEntries(c.Args...)
	if err == nil && exportedN > 0 {
		VSO.Log.Info(fmt.Sprintf(VSO.LC.Get("vso.cmd.pico.info.unexportedApps"), exportedN))
	}

	finalArgs := pkgManager.GenCmd(pkgManager.CmdRemove, c.Args...)
	_, err = pico.Exec(false, false, finalArgs...)
	return err
}

func (c *PicoRunCmd) Run() error {
	pico, err := ensurePicoInitialized()
	if err != nil {
		return err
	}
	if pico == nil {
		return nil
	}

	_, err = pico.Exec(false, false, c.Args...)
	return err
}

func (c *PicoExportCmd) Run() error {
	if c.App == "" && c.Bin == "" {
		VSO.Log.Error(VSO.LC.Get("vso.cmd.pico.error.noAppNameOrBin"))
		return nil
	}

	pico, err := ensurePicoInitialized()
	if err != nil {
		return err
	}
	if pico == nil {
		return nil
	}

	err = core.PicoExport(c.App, c.Bin)
	if err != nil {
		var errMsg string
		if c.App != "" {
			errMsg = fmt.Sprintf(VSO.LC.Get("vso.cmd.pico.error.exportingApp"), err.Error())
		} else {
			errMsg = fmt.Sprintf(VSO.LC.Get("vso.cmd.pico.error.exportingBin"), err.Error())
		}
		return fmt.Errorf(errMsg)
	}

	if c.App != "" {
		VSO.Log.Info(fmt.Sprintf(VSO.LC.Get("vso.cmd.pico.info.exported"), c.App))
	} else {
		VSO.Log.Info(fmt.Sprintf(VSO.LC.Get("vso.cmd.pico.info.exported"), c.Bin))
	}
	return nil
}

func (c *PicoUnexportCmd) Run() error {
	if c.App == "" && c.Bin == "" {
		VSO.Log.Error(VSO.LC.Get("vso.cmd.pico.error.noAppNameOrBin"))
		return nil
	}

	pico, err := ensurePicoInitialized()
	if err != nil {
		return err
	}
	if pico == nil {
		return nil
	}

	err = core.PicoUnexport(c.App, c.Bin)
	if err != nil {
		var errMsg string
		if c.App != "" {
			errMsg = fmt.Sprintf(VSO.LC.Get("vso.cmd.pico.error.unexportingApp"), err.Error())
		} else {
			errMsg = fmt.Sprintf(VSO.LC.Get("vso.cmd.pico.error.unexportingBin"), err.Error())
		}
		return fmt.Errorf(errMsg)
	}

	if c.App != "" {
		VSO.Log.Info(fmt.Sprintf(VSO.LC.Get("vso.cmd.pico.info.unexported"), c.App))
	} else {
		VSO.Log.Info(fmt.Sprintf(VSO.LC.Get("vso.cmd.pico.info.unexported"), c.Bin))
	}
	return nil
}

func (c *PicoShellCmd) Run() error {
	pico, err := ensurePicoInitialized()
	if err != nil {
		return err
	}
	if pico == nil {
		return nil
	}

	return pico.Enter()
}

func (c *PicoSideloadCmd) Run() error {
	pico, pkgManager, err := getPicoAndPkgManager()
	if err != nil {
		return err
	}
	if pico == nil {
		return nil
	}

	updateArgs := pkgManager.GenCmd(pkgManager.CmdUpdate)
	_, err = pico.Exec(false, false, updateArgs...)
	if err != nil {
		return err
	}

	args := append([]string{"--fix-broken"}, c.Args...)
	finalArgs := pkgManager.GenCmd(pkgManager.CmdInstall, args...)
	_, err = pico.Exec(false, false, finalArgs...)
	if err != nil {
		return err
	}

	packageNames := []string{}
	for _, v := range c.Args { // Iterate original args (files)
		out, _ := pico.Exec(true, false, "dpkg-deb", "--info", v)
		for _, w := range strings.Split(out, "\n") {
			wordsInLine := strings.Split(w, " ")
			if len(wordsInLine) > 1 && wordsInLine[1] == "Package:" {
				packageName := strings.TrimSpace(wordsInLine[2])
				packageNames = append(packageNames, packageName)
			}
		}
	}

	exportedN, err := pico.ExportDesktopEntries(packageNames...)
	if err == nil && exportedN > 0 {
		VSO.Log.Info(fmt.Sprintf(VSO.LC.Get("vso.cmd.pico.info.exportedApps"), exportedN))
	}
	return nil
}

func (c *PicoUpdateCmd) Run() error {
	pico, pkgManager, err := getPicoAndPkgManager()
	if err != nil {
		return err
	}
	if pico == nil {
		return nil
	}

	finalArgs := pkgManager.GenCmd(pkgManager.CmdUpdate)
	_, err = pico.Exec(false, false, finalArgs...)
	return err
}

func (c *PicoUpgradeCmd) Run() error {
	pico, pkgManager, err := getPicoAndPkgManager()
	if err != nil {
		return err
	}
	if pico == nil {
		return nil
	}

	finalArgs := pkgManager.GenCmd(pkgManager.CmdUpgrade)
	_, err = pico.Exec(false, false, finalArgs...)
	return err
}

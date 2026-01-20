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
	"github.com/vanilla-os/sdk/pkg/v1/app"
	"github.com/vanilla-os/sdk/pkg/v1/cli"
)

var VSO *app.App

type RootCmd struct {
	cli.Base
	Version string

	Tasks   TasksCmd   `cmd:"tasks" help:"pr:vso.cmd.tasks"`
	Upgrade UpgradeCmd `cmd:"upgrade" help:"pr:vso.cmd.sysUpgrade"`
	Config  ConfigCmd  `cmd:"config" help:"pr:vso.cmd.config"`
	Pico    PicoCmd    `cmd:"pico" help:"pr:vso.cmd.pico"`
}

// Tasks

type TasksCmd struct {
	cli.Base
	List   TasksListCmd   `cmd:"list" help:"pr:vso.cmd.tasks.list"`
	New    TasksNewCmd    `cmd:"new" help:"pr:vso.cmd.tasks.new"`
	Rm     TasksRmCmd     `cmd:"rm" help:"pr:vso.cmd.tasks.rm"`
	Rotate TasksRotateCmd `cmd:"rotate" help:"pr:vso.cmd.tasks.rotate"`
}

type TasksListCmd struct {
	cli.Base
	Json bool `flag:"short:j, long:json, name:pr:vso.cmd.tasks.list.options.json"`
}

type TasksNewCmd struct {
	cli.Base
	AssumeYes           bool   `flag:"short:y, long:assume-yes, name:pr:vso.cmd.tasks.new.options.assumeYes"`
	NeedConfirm         bool   `flag:"long:need-confirm, name:pr:vso.cmd.tasks.new.options.needConfirm"`
	Name                string `flag:"long:name, name:pr:vso.cmd.tasks.new.options.name"`
	Description         string `flag:"long:description, name:pr:vso.cmd.tasks.new.options.description"`
	Command             string `flag:"long:command, name:pr:vso.cmd.tasks.new.options.command"`
	Every               string `flag:"long:every, name:pr:vso.cmd.tasks.new.options.every"`
	At                  string `flag:"long:at, name:pr:vso.cmd.tasks.new.options.at"`
	OnNetwork           bool   `flag:"long:on-network, name:pr:vso.cmd.tasks.new.options.onNetwork"`
	OnDisconnect        bool   `flag:"long:on-disconnect, name:pr:vso.cmd.tasks.new.options.onDisconnect"`
	OnBattery           bool   `flag:"long:on-battery, name:pr:vso.cmd.tasks.new.options.onBattery"`
	OnLowBattery        bool   `flag:"long:on-low-battery, name:pr:vso.cmd.tasks.new.options.onLowBattery"`
	OnCharge            bool   `flag:"long:on-charge, name:pr:vso.cmd.tasks.new.options.onCharge"`
	OnFullBattery       bool   `flag:"long:on-full-battery, name:pr:vso.cmd.tasks.new.options.onFullBattery"`
	OnConditionCommand  string `flag:"long:on-condition-command, name:pr:vso.cmd.tasks.new.options.onConditionCommand"`
	OnProcess           string `flag:"long:on-process, name:pr:vso.cmd.tasks.new.options.onProcess"`
	OnInternetUsage     int    `flag:"long:on-internet-usage, name:pr:vso.cmd.tasks.new.options.onInternetUsage"`
	OnHighInternetUsage bool   `flag:"long:on-high-internet-usage, name:pr:vso.cmd.tasks.new.options.onHighInternetUsage"`
	OnMemoryUsage       int    `flag:"long:on-memory-usage, name:pr:vso.cmd.tasks.new.options.onMemoryUsage"`
	OnHighMemoryUsage   bool   `flag:"long:on-high-memory-usage, name:pr:vso.cmd.tasks.new.options.onHighMemoryUsage"`
	OnCpuUsage          int    `flag:"long:on-cpu-usage, name:pr:vso.cmd.tasks.new.options.onCpuUsage"`
	OnHighCpuUsage      bool   `flag:"long:on-high-cpu-usage, name:pr:vso.cmd.tasks.new.options.onHighCpuUsage"`
	OnCpuTemp           int    `flag:"long:on-cpu-temp, name:pr:vso.cmd.tasks.new.options.onCpuTemp"`
	OnDeviceConnect     string `flag:"long:on-device-connect, name:pr:vso.cmd.tasks.new.options.onDeviceConnect"`
	OnDeviceDisconnect  string `flag:"long:on-device-disconnect, name:pr:vso.cmd.tasks.new.options.onDeviceDisconnect"`
}

type TasksRmCmd struct {
	cli.Base
	Name  string `flag:"short:n, long:name, name:pr:vso.cmd.tasks.rm.options.name"`
	Force bool   `flag:"short:f, long:force, name:pr:vso.cmd.tasks.rm.options.force"`
}

type TasksRotateCmd struct {
	cli.Base
	PrivateEvent string `flag:"short:e, long:private-event, name:pr:vso.cmd.tasks.rotate.options.privateEvent"`
	Silent       bool   `flag:"short:s, long:silent, name:pr:vso.cmd.tasks.rotate.options.silent"`
}

// Upgrade

type UpgradeCmd struct {
	cli.Base
	Check UpgradeCheckCmd `cmd:"check" help:"pr:vso.cmd.sysUpgrade.check"`
	Now   bool            `flag:"long:now, name:pr:vso.cmd.sysUpgrade.now"`
}

type UpgradeCheckCmd struct {
	cli.Base
	Json bool `flag:"long:json, name:pr:vso.cmd.sysUpgrade.check.json"`
}

// Config

type ConfigCmd struct {
	cli.Base
	Get  ConfigGetCmd  `cmd:"get" help:"pr:vso.cmd.config.get"`
	Set  ConfigSetCmd  `cmd:"set" help:"pr:vso.cmd.config.set"`
	Show ConfigShowCmd `cmd:"show" help:"pr:vso.cmd.config.show"`
}

type ConfigGetCmd struct {
	cli.Base
	Key string `arg:"" name:"key" help:"pr:vso.arg.key"`
}

type ConfigSetCmd struct {
	cli.Base
	Key   string `arg:"" name:"key" help:"pr:vso.arg.key"`
	Value string `arg:"" name:"value" help:"pr:vso.arg.value"`
}

type ConfigShowCmd struct {
	cli.Base
}

// Pico

type PicoCmd struct {
	cli.Base
	Init     PicoInitCmd     `cmd:"init" help:"pr:vso.cmd.pico.init"`
	Install  PicoInstallCmd  `cmd:"install" help:"pr:vso.cmd.pico.install"`
	Remove   PicoRemoveCmd   `cmd:"remove" help:"pr:vso.cmd.pico.remove"`
	Run      PicoRunCmd      `cmd:"run" help:"pr:vso.cmd.pico.run"`
	Export   PicoExportCmd   `cmd:"export" help:"pr:vso.cmd.pico.export"`
	Unexport PicoUnexportCmd `cmd:"unexport" help:"pr:vso.cmd.pico.unexport"`
	Shell    PicoShellCmd    `cmd:"shell" help:"pr:vso.cmd.pico.shell"`
	Sideload PicoSideloadCmd `cmd:"sideload" help:"pr:vso.cmd.pico.sideload"`
	Update   PicoUpdateCmd   `cmd:"update" help:"pr:vso.cmd.pico.update"`
	Upgrade  PicoUpgradeCmd  `cmd:"upgrade" help:"pr:vso.cmd.pico.upgrade"`
}

type PicoInitCmd struct {
	cli.Base
	Force bool `flag:"short:f, long:force, name:pr:vso.cmd.pico.init.options.force"`
}

type PicoInstallCmd struct {
	cli.Base
	Args []string `arg:"" optional:"" name:"package" help:"pr:vso.arg.package"`
}

type PicoRemoveCmd struct {
	cli.Base
	Args []string `arg:"" optional:"" name:"package" help:"pr:vso.arg.package"`
}

type PicoRunCmd struct {
	cli.Base
	NoReset bool     `flag:"long:no-reset, name:pr:vso.cmd.pico.run.options.noReset"`
	Args    []string `arg:"" optional:"" name:"command" help:"pr:vso.arg.command"`
}

type PicoExportCmd struct {
	cli.Base
	App       string `flag:"long:app, name:pr:vso.cmd.pico.export.options.app"`
	Bin       string `flag:"long:bin, name:pr:vso.cmd.pico.export.options.bin"`
	BinOutput string `flag:"long:bin-output, name:pr:vso.cmd.pico.export.options.binOutput"`
}

type PicoUnexportCmd struct {
	cli.Base
	App       string `flag:"long:app, name:pr:vso.cmd.pico.unexport.options.app"`
	Bin       string `flag:"long:bin, name:pr:vso.cmd.pico.unexport.options.bin"`
	BinOutput string `flag:"long:bin-output, name:pr:vso.cmd.pico.unexport.options.binOutput"`
}

type PicoShellCmd struct {
	cli.Base
}

type PicoSideloadCmd struct {
	cli.Base
	Args []string `arg:"" optional:"" name:"package" help:"pr:vso.arg.package"`
}

type PicoUpdateCmd struct {
	cli.Base
}

type PicoUpgradeCmd struct {
	cli.Base
}

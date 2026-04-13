package core

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
	"os/exec"

	"github.com/vanilla-os/apx/v3/core"
)

func GetNative() (*core.SubSystem, error) {
	subsystem, err := core.LoadSubSystem("vso-native", false)
	if err != nil {
		return nil, err
	}

	return subsystem, nil
}

func NativeExists() bool {
	_, err := GetNative()
	return err == nil
}

func NativeInit() error {
	stack, _ := core.LoadStack("vso-native")
	subsystem, err := core.NewSubSystem(
		"vso-native",
		stack,
		"",
		true,
		true,
		false,
		false,
		true,
		"",
		"--pre-init-hooks",
		"/run/host/usr/share/vso/scripts/pre-init-hook",
		"--init-hooks",
		"/run/host/usr/share/vso/scripts/init-hook",
	)

	if err != nil {
		return err
	}

	err = subsystem.Create()
	if err != nil {
		return err
	}

	terminal_setup_cmd := exec.Command("setup-vso-terminal", "--non-interactive")
	err = terminal_setup_cmd.Run()
	return err
}

func NativeDelete() error {
	subsystem, err := GetNative()
	if err != nil {
		return err
	}

	err = subsystem.Remove()
	if err != nil {
		return err
	}

	return nil
}

func NativeExport(app string, binary string) error {
	subsystem, err := GetNative()
	if err != nil {
		return err
	}

	if app != "" {
		err = subsystem.ExportDesktopEntry(app)
	} else {
		err = subsystem.ExportBin(binary, "")
	}

	return err
}

func NativeUnexport(app string, binary string) error {
	subsystem, err := GetNative()
	if err != nil {
		return err
	}

	if app != "" {
		err = subsystem.UnexportDesktopEntry(app)
	} else {
		err = subsystem.UnexportBin(binary, "")
	}

	return err
}

func NativeUpgrade() error {
	subsystem, err := GetNative()
	if err != nil {
		return err
	}

	pkgManager, err := subsystem.Stack.GetPkgManager()
	if err != nil {
		return err
	}

	finalArgs := pkgManager.GenCmd(pkgManager.CmdUpdate, []string{}...)
	_, err = subsystem.Exec(false, false, finalArgs...)
	if err != nil {
		return err
	}

	finalArgs = pkgManager.GenCmd(pkgManager.CmdUpgrade, []string{}...)
	_, err = subsystem.Exec(false, false, finalArgs...)
	if err != nil {
		return err
	}

	return nil
}

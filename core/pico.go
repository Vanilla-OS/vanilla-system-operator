package core

import (
	"github.com/vanilla-os/apx/core"
)

func GetPico() (*core.SubSystem, error) {
	subsystem, err := core.LoadSubSystem("vso-pico")
	if err != nil {
		return nil, err
	}

	return subsystem, nil
}

func PicoExists() bool {
	_, err := GetPico()
	return err == nil
}

func PicoInit() error {
	var stack *core.Stack
	if !core.StackExists("vso-pico") {

		if !core.PkgManagerExists("vso-apt") {
			core.NewPkgManager(
				"vso-apt",
				true,
				"autoremove",
				"clean",
				"install",
				"list",
				"purge",
				"remove",
				"search",
				"show",
				"update",
				"upgrade",
				true,
			)
		}

		stack = core.NewStack(
			"vso-pico",
			"registry.vanillaos.org/vanillaos/pico:main",
			[]string{"systemd"},
			"vso-apt",
			true,
		)
	} else {
		stack, _ = core.LoadStack("vso-pico")
	}

	subsystem, err := core.NewSubSystem(
		"vso-pico",
		stack,
		true,
		true,
	)

	if err != nil {
		return err
	}

	subsystem.Create()
	return nil
}

func PicoDelete() error {
	subsystem, err := GetPico()
	if err != nil {
		return err
	}

	err = subsystem.Remove()
	if err != nil {
		return err
	}

	return nil
}

package core

import (
	"github.com/vanilla-os/apx/core"
)

func GetWay() (*core.SubSystem, error) {
	subsystem, err := core.LoadSubSystem("vso-waydroid", true)
	if err != nil {
		return nil, err
	}

	return subsystem, nil
}

func WayExists() bool {
	_, err := GetWay()
	return err == nil
}

func WayInit() error {
	stack, _ := core.LoadStack("vso-waydroid")
	subsystem, err := core.NewSubSystem(
		"vso-waydroid",
		stack,
		true,
		true,
		true,
		true,
	)

	if err != nil {
		return err
	}

	err = subsystem.Create()
	if err != nil {
		return err
	}

	_, err = subsystem.Exec(false, "ewaydroid", "--version")
	return err
}

func WayDelete() error {
	subsystem, err := GetWay()
	if err != nil {
		return err
	}

	err = subsystem.Remove()
	if err != nil {
		return err
	}

	return nil
}

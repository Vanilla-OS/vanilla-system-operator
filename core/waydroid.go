package core

import (
	"github.com/vanilla-os/apx/core"
	"strings"
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

func GetWayPackages(subsystem *core.SubSystem) ([][]string, error) {
	out, err := subsystem.Exec(true, "ewaydroid", "app", "list")
	if err != nil {
		return nil, err
	}
	out = strings.ReplaceAll(out, "Checking if Waydroid was initialized...\n", "")
	lines := strings.Split(out, "\n")
	var pairs = [][]string{{}}
	i := 0
	for _, line := range lines {
		if strings.Contains(line, "android.intent") {
			i = i + 1
			pairs = append(pairs, []string{})
			continue
		}
		if !strings.Contains(line, "categories") {
			processedLine := strings.Split(line, ":")
			if len(processedLine) > 1 {
				pairs[i] = append(pairs[i], strings.Trim(processedLine[1], " "))
			}

		}
	}

	return pairs[:len(pairs)-1], nil
}

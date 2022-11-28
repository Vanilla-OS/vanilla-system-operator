package core

import (
	"fmt"
	"os"
	"os/exec"
	"time"

	"github.com/vanilla-os/vso/settings"
)

var (
	checkLogPath = "/var/log/vso-check.log"
)

// NeedUpdate checks if the system needs to be updated according to the latest
// update log compared to the VSO configuation
func NeedUpdate() bool {
	res := false
	currentTime := time.Now()
	schedule := settings.GetConfigValue("updates.schedule")
	latestCheck := getLatestCheck()
	if latestCheck == nil {
		return true
	}

	switch schedule {
	case "daily":
		if currentTime.Sub(*latestCheck).Hours() >= 24 {
			res = true
		}
	case "weekly":
		if currentTime.Sub(*latestCheck).Hours() >= 168 {
			res = true
		}
	case "monthly":
		if currentTime.Sub(*latestCheck).Hours() >= 720 {
			res = true
		}
	}

	return res
}

// getLatestCheck returns the latest check time from the log file, it also
// write the current time to the log file if it doesn't exist
func getLatestCheck() *time.Time {
	var latestCheck time.Time
	if _, err := os.Stat(checkLogPath); os.IsNotExist(err) {
		latestCheck = time.Now()
		writeLatestCheck(latestCheck)
	} else {
		file, err := os.Open(checkLogPath)
		if err != nil {
			return nil
		}
		defer file.Close()

		_, err = fmt.Fscan(file, &latestCheck)
		if err != nil {
			return nil
		}
	}

	return &latestCheck
}

// writeLatestCheck writes the latest check time to the log file
func writeLatestCheck(t time.Time) error {
	file, err := os.OpenFile(checkLogPath, os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = fmt.Fprint(file, t)
	if err != nil {
		return err
	}

	return nil
}

// TryUpdate tries to update the system via ABRoot
func TryUpdate() error {
	writeLatestCheck(time.Now())

	cmd := exec.Command("abroot", "exec", "--assume-yes", "apt", "update", "&&", "apt", "upgrade", "-y", "&&", "apt", "autoremove", "-y")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err := cmd.Run()
	if err != nil {
		return err
	}

	SendNotification("Update", "System updated successfully, restart to apply changes.")
	return nil
}

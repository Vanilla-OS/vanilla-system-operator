package core

import (
	"fmt"
	"os"
	"os/exec"
	"time"

	"github.com/vanilla-os/vso/settings"
)

var (
	checkLogPath   = "/var/log/vso-check.log"
	abrootLockPath = "/tmp/abroot-transactions.lock"
)

// AreABRootTransactionsLocked checks if there are any abroot transactions
// currently running
func AreABRootTransactionsLocked() bool {
	_, err := os.Stat(abrootLockPath)

	return err == nil
}

// NeedUpdate checks if the system needs to be updated according to the latest
// update log compared to the VSO configuation
func NeedUpdate() bool {
	if AreABRootTransactionsLocked() {
		fmt.Println("ABRoot transactions are currently locked. Skipping update check.")
		return false
	}

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

		content, err := os.ReadFile(checkLogPath)
		if err != nil {
			return nil
		}

		latestCheck, err = time.Parse("2006-01-02 15:04:05.999999999 -0700 MST m=+0.000000000", string(content))
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
	if !SmartUpdate() {
		fmt.Println("Smart update detected device is being used, skipping update.")
		return nil
	}

	if os.Getenv("VSO_DEBUG_SMARTUPDATE") != "" {
		fmt.Println("Smart update is in debug mode, skipping update.")
		return nil
	}

	writeLatestCheck(time.Now())

	file, err := os.Create("/tmp/" + time.Now().Format("20060102150405") + "-script")
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = file.WriteString("#!/bin/bash\napt update && apt upgrade -y && apt autoremove -y")
	if err != nil {
		return err
	}

	cmd := exec.Command("abroot", "exec", "--assume-yes", "sh", file.Name())
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err = cmd.Run()
	if err != nil {
		return err
	}

	SendNotification("Update", "System updated successfully, restart to apply changes.")
	return nil
}

// SmartUpdate checks if the device is currently being used, then returns true
// if the device is not being used
func SmartUpdate() bool {
	if settings.GetConfigValue("updates.smart") == false {
		fmt.Println("Smart update is disabled, skipping check.")
		return true
	}

	fmt.Println("Starting smart update check...")

	commonChecks := GetCommonChecks()

	// battery check
	if commonChecks.IsLaptop && commonChecks.LowBattery {
		fmt.Printf("Low battery detected, skipping update.")
		return false
	}

	// internet usage check
	if commonChecks.HighInternetUsage {
		if commonChecks.InternetUsage == 0 {
			fmt.Println("Something goes wrong during internet usage check, skipping update.")
			return false
		}
		fmt.Printf("Connection being used (%d), skipping update.", commonChecks.InternetUsage)
		return false
	}

	// ram check (false if exceeds 50%)
	if commonChecks.HighMemoryUsage {
		if commonChecks.MemoryUsage == 0 {
			fmt.Println("Something goes wrong during memory usage check, skipping update.")
			return false
		}
		fmt.Printf("Memory usage detected (%d), skipping update.", commonChecks.MemoryUsage)
		return false
	}

	// cpu check (false if exceeds 50%)
	if commonChecks.HighCPUUsage {
		if commonChecks.CPUUsage == 0 {
			fmt.Println("Something goes wrong during CPU usage check, skipping update.")
			return false
		}
		fmt.Printf("CPU usage detected (%d), skipping update.", commonChecks.CPUUsage)
		return false
	}

	return true
}

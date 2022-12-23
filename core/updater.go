package core

import (
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
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
	if err != nil {
		return false
	}

	return true
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
	fmt.Println("Starting smart update check...")

	if settings.GetConfigValue("updates.smart") == "false" {
		return true
	}

	commonChecks := GetCommonChecks()

	// battery check
	if commonChecks.LowBattery {
		fmt.Println("Low battery detected, skipping update.")
		return false
	}

	// internet usage check (false if exceeds 500kb/s)
	cmd := exec.Command("ifstat", "1", "1")
	out, err := cmd.Output()
	if err != nil {
		fmt.Println("Error while checking internet usage, skipping update. (1)")
		return false
	}

	out = out[2:]

	re := regexp.MustCompile(`\d+\.\d+`)
	numbers := re.FindAllString(string(out), -1)

	for _, number := range numbers {
		if number > "500.0" {
			fmt.Println("Internet usage detected, skipping update.")
			return false
		}
	}

	// ram check (false if exceeds 50%)
	cmd = exec.Command("cat", "/proc/meminfo")
	out, err = cmd.Output()
	if err != nil {
		fmt.Println("Error while checking ram usage, skipping update. (1)")
		return false
	}

	var memAvailable, memTotal int
	for _, line := range strings.Split(string(out), "\n") {
		if strings.Contains(line, "MemAvailable") {
			re := regexp.MustCompile(`\d+`)
			memAvailable, err = strconv.Atoi(re.FindString(line))
			if err != nil {
				fmt.Println("Error while checking ram usage, skipping update. (2)")
				return false
			}
		} else if strings.Contains(line, "MemTotal") {
			re := regexp.MustCompile(`\d+`)
			memTotal, err = strconv.Atoi(re.FindString(line))
			if err != nil {
				fmt.Println("Error while checking ram usage, skipping update. (3)")
				return false
			}
		}
	}

	if memAvailable < memTotal/2 {
		fmt.Println("Ram usage detected, skipping update.")
		return false
	}

	// cpu check (false if exceeds 50%)
	cmd = exec.Command("top", "-bn1")
	out, err = cmd.Output()
	if err != nil {
		fmt.Println("Error while checking cpu usage, skipping update. (1)")
		return false
	}

	var cpuUsage float64
	for _, line := range strings.Split(string(out), "\n") {
		if strings.Contains(line, "Cpu(s)") {
			re := regexp.MustCompile(`\d+\.\d+`)
			cpuUsage, err = strconv.ParseFloat(re.FindString(line), 64)
			if err != nil {
				fmt.Println("Error while checking cpu usage, skipping update. (2)")
				return false
			}
		}
	}

	if cpuUsage > 50.0 {
		fmt.Println("Cpu usage detected, skipping update.")
		return false
	}

	fmt.Println("Smart update check passed.")
	return true
}

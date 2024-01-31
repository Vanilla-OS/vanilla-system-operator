package core

/*	License: GPLv3
	Authors:
		Mirko Brombin <send@mirko.pm>
		Pietro di Caprio <pietro@fabricators.ltd>
	Copyright: 2024
	Description: The Vanilla System Operator is a package manager,
	a system updater and a task automator.
*/

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/vanilla-os/vso/settings"
)

var (
	checkLogPath       = "/var/log/vso-check.log"
	firstSetupDonePath = "/etc/vanilla-first-setup-done"
	abrootLockPaths    = []string{
		"/tmp/ABSystem.Upgrade.lock",
		"/tmp/ABSystem.Upgrade.user.lock",
	}
)

// AreABRootTransactionsLocked checks if there are any abroot transactions
// currently running
func AreABRootTransactionsLocked() bool {
	for _, abrootLockPath := range abrootLockPaths {
		_, err := os.Stat(abrootLockPath)
		if err == nil {
			return true
		}
	}
	return false
}

// NeedUpdate checks if the system needs to be updated according to the latest
// update log compared to the VSO configuation
func NeedUpdate() bool {
	if AreABRootTransactionsLocked() {
		fmt.Println("ABRoot transactions are currently locked. Skipping update check.")
		return false
	}

	if !okToUpdate() {
		fmt.Println("Deferring update until first-setup is complete.")
		return false
	}

	res := false
	currentTime := time.Now()
	schedule := settings.GetConfigValue("updates.schedule")
	latestCheck := getLatestCheck()
	if latestCheck == nil {
		fmt.Println("No previous update check found. Triggering update.")
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

	status, err := HasUpdates()
	if err != nil {
		return false
	}
	if status {
		res = true
	}

	return res
}

func runABRootCheck(asJson, showStdout bool) (bool, map[string]any, error) {
	update_cmd := exec.Command("abroot", "upgrade", "--check-only")

	var out []byte
	var err error
	if asJson {
		update_cmd.Env = append(update_cmd.Environ(), "ABROOT_JSON_OUTPUT=1")
	}

	if showStdout {
		update_cmd.Stdout = os.Stdout
		update_cmd.Stderr = os.Stderr
		err = update_cmd.Run()
	} else {
		out, err = update_cmd.Output()
	}

	if err != nil {
		if exitError, ok := err.(*exec.ExitError); ok {
			if exitError.ExitCode() == 1 {
				return false, map[string]any{}, nil
			}
			return false, map[string]any{}, err
		}
	}

	if !asJson || showStdout {
		return true, map[string]any{}, nil
	}

	capturedOutput := map[string]any{}
	err = json.Unmarshal(out, &capturedOutput)
	if err != nil {
		return false, map[string]any{}, err
	}

	return capturedOutput["hasUpdate"].(bool), capturedOutput, nil
}

// HasUpdates checks if the system has updates available
func HasUpdates() (bool, error) {
	hasUpdate, _, err := runABRootCheck(false, false)
	if err != nil {
		return false, err
	}

	return hasUpdate, nil
}

// RunUpgradeCheck asks ABRoot to check for updates and passes its output to stdout
func RunUpgradeCheck() error {
	_, _, err := runABRootCheck(false, true)
	return err
}

// RunUpgradeCheckJSON asks ABRoot to check for updates and return a JSON-formatted result
func RunUpgradeCheckJSON() (bool, error) {
	hasUpdate, _, err := runABRootCheck(true, true)
	return hasUpdate, err
}

func okToUpdate() bool {
	// check vso logs for previous run
	if _, err := os.Stat(checkLogPath); os.IsNotExist(err) {
		// vso has never run, check if first-setup has been done
		if _, err := os.Stat(firstSetupDonePath); os.IsNotExist(err) {
			// first setup has not completed. don't try to update
			return false
		}

	}
	return true
}

// getLatestCheck returns the latest check time from the log file, it also
// write the current time to the log file if it doesn't exist
func getLatestCheck() *time.Time {
	var latestCheck time.Time

	if _, err := os.Stat(checkLogPath); os.IsNotExist(err) {
		latestCheck = time.Now()
		writeLatestCheck(latestCheck)
		return &latestCheck
	}

	file, err := os.Open(checkLogPath)
	if err != nil {
		fmt.Println("Failed to open update log file:", err)
		return nil
	}
	defer file.Close()

	content, err := os.ReadFile(checkLogPath)
	if err != nil {
		fmt.Println("Failed to read update log file:", err)
		return nil
	}

	timeString := string(content)
	timeArr := strings.Split(timeString, " ")
	latestCheck, err = time.Parse("2006-01-02 15:04:05", timeArr[0]+" "+timeArr[1])
	if err != nil {
		fmt.Println("Failed to parse update log file:", err)
		return nil
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
func TryUpdate(force bool) error {
	if !okToUpdate() {
		fmt.Println("Deferring update until first-setup is done.")
		return nil

	}
	if !force && !SmartUpdate() {
		fmt.Println("Smart update detected device is being used, skipping update.")
		return nil
	}

	if os.Getenv("VSO_DEBUG_SMARTUPDATE") != "" {
		fmt.Println("Smart update is in debug mode, skipping update.")
		return nil
	}

	writeLatestCheck(time.Now())

	cmd := exec.Command("abroot", "upgrade")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err := cmd.Run()
	if err != nil {
		return fmt.Errorf("failed to update system: %s", err)
	}

	// FIXME: the pico upgrade could fail if APT is locked by another process
	// or it simply fails, but since we are also upgrading the system, we
	// cannot fail the whole operation, so just ignore the error for now
	PicoUpgrade()

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

	// ram check
	if commonChecks.HighMemoryUsage {
		if commonChecks.MemoryUsage == 0 {
			fmt.Println("Something goes wrong during memory usage check, skipping update.")
			return false
		}
		fmt.Printf("Memory usage detected (%d), skipping update.", commonChecks.MemoryUsage)
		return false
	}

	// cpu check
	if commonChecks.HighCPUUsage {
		if commonChecks.CPUUsage == 0 {
			fmt.Println("Something goes wrong during CPU usage check, skipping update.")
			return false
		}
		fmt.Printf("CPU usage detected (%d), skipping update.", commonChecks.CPUUsage)
		return false
	}

	// cpu temp check
	if commonChecks.CPUTemp >= 70 {
		fmt.Printf("High CPU temperature detected (%d), skipping update.", commonChecks.CPUTemp)
		return false
	}

	return true
}

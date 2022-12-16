package core

import (
	"os/exec"
	"strconv"
	"strings"
	"time"
)

type CommonChecks struct {
	Network     bool
	Battery     bool
	LowBattery  bool
	FullBattery bool
}

// GetCommonChecks checks network and battery
func GetCommonChecks() *CommonChecks {
	cChecks := CommonChecks{}

	// Network
	cmd := exec.Command("sh", "-c", "ping -q -w 1 -c 1 `ip r | grep default | cut -d ' ' -f 3` > /dev/null && exit 0 || exit 1")
	err := cmd.Run()
	if err == nil {
		cChecks.Network = true
	}

	// Battery
	cmd = exec.Command("acpi")
	output, err := cmd.CombinedOutput()
	if err == nil {
		if strings.Contains(string(output), "Discharging") {
			cChecks.Battery = true
		}
		if strings.Contains(string(output), "Charging") || strings.Contains(string(output), "Not charging") {
			cChecks.Battery = false
		}
		percentage := ""
		for _, char := range string(output) {
			if char >= '0' && char <= '9' {
				percentage += string(char)
			}
			if char == '%' {
				break
			}
		}
		if percentage != "" {
			perc, err := strconv.Atoi(percentage)
			if err == nil {
				if perc <= 30 {
					cChecks.LowBattery = true
				}
				if perc == 100 {
					cChecks.FullBattery = true
				}
			}
		}
	}

	// fmt.Printf(
	// 	"Checks:\tNetwork: %t\n\tBattery: %t\n\tLow battery: %t\n\tFull battery: %t\n",
	// 	cChecks.Network, cChecks.Battery, cChecks.LowBattery, cChecks.FullBattery,
	// )

	return &cChecks
}

// ItsBeen checks if it's been a certain amount of time since the last time
func ItsBeen(last time.Time, duration string) bool {
	d, err := time.ParseDuration(duration)
	if err != nil {
		return false
	} else if time.Since(last) >= d {
		return true
	}
	return false
}

// ItsTime checks if it's a certain time of day
func ItsTime(targetTime string) bool {
	t, err := time.Parse("15:04", targetTime)
	if err != nil {
		return false
	}

	now := time.Now()

	if now.Hour() != t.Hour() {
		return false
	}

	if now.Minute() == t.Minute() {
		return true
	}

	return false

}

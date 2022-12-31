package core

import (
	"fmt"
	"os"
	"os/exec"
	"path"
	"strconv"
	"strings"
	"time"
)

// powerPath is the path of batteries in Linux
// after 3.20
const powerPath = "/sys/class/power_supply/"

type CommonChecks struct {
	Network     bool
	Battery     bool
	LowBattery  bool
	FullBattery bool
}

// IsLaptop checks if the system is a laptop by looking for the chassis type
func IsLaptop() bool {
	file, err := os.Open("/sys/devices/virtual/dmi/id/chassis_type")
	if err != nil {
		return false
	}
	defer file.Close()

	var chassisType int
	_, err = fmt.Fscanf(file, "%d", &chassisType)
	if err != nil {
		return false
	}

	if chassisType == 8 || chassisType == 9 || chassisType == 10 || chassisType == 14 {
		return true
	}

	return false
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
	cmd = exec.Command("ls", powerPath)
	output, err := cmd.CombinedOutput()
	if err == nil {
		devices := strings.Split(string(output), "\n")
		for _, dev := range devices {
			if strings.Contains(dev, "BAT") {
				devPath := path.Join(powerPath, dev)
				statusPath := path.Join(devPath, "status")
				capacityPath := path.Join(devPath, "capacity")

				statusCmd := exec.Command("cat", statusPath)
				statusOutput, err := statusCmd.CombinedOutput()
				if err == nil {
					if strings.Contains(string(statusOutput), "Discharging") {
						cChecks.Battery = true
					}
				}
				capacityCmd := exec.Command("cat", capacityPath)
				capacityOutput, err := capacityCmd.CombinedOutput()
				if err == nil {
					percent, err := strconv.Atoi(strings.TrimSpace(string(capacityOutput)))
					if err == nil {
						if percent <= 30 {
							cChecks.LowBattery = true
						}
						if percent == 100 {
							cChecks.FullBattery = true
						}
					}
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

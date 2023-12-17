package core

/*	License: GPLv3
	Authors:
		Mirko Brombin <send@mirko.pm>
		Vanilla OS Contributors <https://github.com/vanilla-os/>
	Copyright: 2023
	Description: VSO is a utility which allows you to perform maintenance
	tasks on your Vanilla OS installation.
*/

import (
	"fmt"
	"os"
	"os/exec"
	"path"
	"regexp"
	"strconv"
	"strings"
	"time"
)

type CommonChecks struct {
	Network           bool
	Battery           bool
	LowBattery        bool
	FullBattery       bool
	IsLaptop          bool
	HighInternetUsage bool
	InternetUsage     int
	HighMemoryUsage   bool
	MemoryUsage       int
	HighCPUUsage      bool
	CPUUsage          int
	CPUTemp           int
}

// GetCommonChecks checks network and battery
func GetCommonChecks() *CommonChecks {
	cChecks := CommonChecks{}

	// Network
	cChecks.Network = IsNetworkUp()

	// Battery
	cChecks.Battery, cChecks.LowBattery, cChecks.FullBattery = GetBatteryStats()

	// IsLaptop
	cChecks.IsLaptop = IsLaptop()

	// Internet usage
	cChecks.HighInternetUsage, cChecks.InternetUsage = IsInternetUnderHighUsage()

	// Memory usage
	cChecks.HighMemoryUsage, cChecks.MemoryUsage = IsMemoryUnderHighUsage()

	// CPU usage
	cChecks.HighCPUUsage, cChecks.CPUUsage = IsCPUUnderHighUsage()

	// CPU temp
	cChecks.CPUTemp = GetCPUTemp()

	if os.Getenv("VSO_VERBOSE") != "" {
		fmt.Printf("Network: %t\n", cChecks.Network)
		fmt.Printf("Battery: %t\n", cChecks.Battery)
		fmt.Printf("Low battery: %t\n", cChecks.LowBattery)
		fmt.Printf("Full battery: %t\n", cChecks.FullBattery)
		fmt.Printf("Is laptop: %t\n", cChecks.IsLaptop)
		fmt.Printf("High internet usage: %t\n", cChecks.HighInternetUsage)
		fmt.Printf("Internet usage: %d\n", cChecks.InternetUsage)
		fmt.Printf("High memory usage: %t\n", cChecks.HighMemoryUsage)
		fmt.Printf("Memory usage: %d\n", cChecks.MemoryUsage)
		fmt.Printf("High CPU usage: %t\n", cChecks.HighCPUUsage)
		fmt.Printf("CPU usage: %d\n", cChecks.CPUUsage)
		fmt.Printf("CPU temp: %d\n", cChecks.CPUTemp)
	}

	return &cChecks
}

// IsNetworkUp checks if the network is up
func IsNetworkUp() bool {
	cmd := exec.Command("sh", "-c", "ping -q -w 1 -c 1 `ip r | grep default | cut -d ' ' -f 3` > /dev/null && exit 0 || exit 1")
	err := cmd.Run()
	return err == nil
}

// GetBatteryStats gets the battery stats
func GetBatteryStats() (bool, bool, bool) {
	battery, lowBattery, fullBattery := false, false, false

	files, err := os.ReadDir("/sys/class/power_supply/")
	if err == nil {
		for _, file := range files {
			if strings.Contains(file.Name(), "BAT") {
				status, err := os.ReadFile(path.Join("/sys/class/power_supply/", file.Name(), "status"))
				if err == nil {
					if strings.Contains(string(status), "Discharging") {
						battery = true
					}
				}

				capacity, err := os.ReadFile(path.Join("/sys/class/power_supply/", file.Name(), "capacity"))
				if err == nil {
					percent, err := strconv.Atoi(strings.TrimSpace(string(capacity)))
					if err == nil {
						if percent <= 30 {
							lowBattery = true
						}
						if percent == 100 {
							fullBattery = true
						}
					}
				}
			}
		}
	}

	return battery, lowBattery, fullBattery
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

// IsInternetUnderHighUsage checks if the internet is being used (false if exceeds 500kb/s)
func IsInternetUnderHighUsage() (bool, int) {
	cmd := exec.Command("ifstat", "1", "1")
	out, err := cmd.Output()
	if err != nil {
		return false, 0
	}

	out = out[2:]
	re := regexp.MustCompile(`\d+\.\d+`)
	numbers := re.FindAllString(string(out), -1)

	for _, number := range numbers {
		num, err := strconv.ParseFloat(number, 64)
		if err != nil {
			return false, 0
		}
		if num < 500 {
			return false, int(num / 1000)
		} else {
			return true, int(num / 1000)
		}
	}

	return false, 0
}

// IsMemoryUnderHighUsage checks if the memory is being used (false if exceeds 50%)
func IsMemoryUnderHighUsage() (bool, int) {
	out, err := os.ReadFile("/proc/meminfo")
	if err != nil {
		return false, 0
	}

	re := regexp.MustCompile(`(?m)^MemTotal:\s+(\d+)`)
	totalMem := re.FindStringSubmatch(string(out))
	if totalMem == nil {
		return false, 0
	}
	total, err := strconv.ParseUint(totalMem[1], 10, 64)
	if err != nil {
		return false, 0
	}

	re = regexp.MustCompile(`(?m)^MemAvailable:\s+(\d+)`)
	availableMem := re.FindStringSubmatch(string(out))
	if availableMem == nil {
		return false, 0
	}
	available, err := strconv.ParseUint(availableMem[1], 10, 64)
	if err != nil {
		return false, 0
	}

	percent := 100 * (float64(total-available) / float64(total))
	if percent > 50 {
		return true, int(percent + 0.5)
	}
	return false, int(percent + 0.5)
}

// IsCPUUnderHighUsage checks if the CPU is being used (false if exceeds 50%)
func IsCPUUnderHighUsage() (bool, int) {
	cmd := exec.Command("sh", "-c", "top -bn1 | grep \"Cpu(s)\" | awk '{print $2 + $4}'")
	out, err := cmd.Output()
	if err != nil {
		return false, 0
	}

	cpuUsage, err := strconv.ParseFloat(strings.TrimSpace(string(out)), 64)
	if err != nil {
		return false, int(cpuUsage)
	}

	if cpuUsage < 50 {
		return false, int(cpuUsage)
	}

	return true, int(cpuUsage)
}

// GetCPUTemp gets the CPU temperature
func GetCPUTemp() int {
	b, err := os.ReadFile("/sys/class/thermal/thermal_zone0/temp")
	if err != nil {
		return 0
	}

	// the temperature is returned in millidegrees Celsius, so we need to convert it to degrees Celsius
	temp, err := strconv.Atoi(string(b[:len(b)-1]))
	if err != nil {
		return 0
	}
	temp = temp / 1000
	return temp
}

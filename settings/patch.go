package settings

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"slices"
	"strconv"
	"strings"
)

type PatchType int

type Patch struct {
	Type        string   `json:"type"`
	Arguments   []string `json:"arguments"`
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Optional    bool     `json:"optional"`
}

type PatchList struct {
	Version int     `json:"version"`
	Patches []Patch `json:"patches"`
}

type PatchRepo struct {
	PatchList []PatchList `json:"patchList"`
}

var patches []PatchList
var finalPatches []Patch
var flatpaks []string

func LoadPatches() error {
	// f, err := os.Open("/usr/share/vso/patch.json")
	f, err := os.Open("/etc/vso/patch.json")
	if err != nil {
		return err
	}
	defer f.Close()

	var loadedPatches PatchRepo
	err = json.NewDecoder(f).Decode(&loadedPatches)
	if err != nil {
		return err
	}
	patches = loadedPatches.PatchList
	return nil
}

func GetPatches() ([]Patch, error) {
	version := 0

	config, _ := os.LookupEnv("XDG_CONFIG_HOME")
	f, err := os.Open(config + "/vso/patchVersion")
	if err == nil {
		var fileContent []byte
		defer f.Close()
		f.Read(fileContent)
		var err2 error
		version, err2 = strconv.Atoi(strings.Split(string(fileContent), "\n")[0])
		if err2 != nil {
			return []Patch{}, err2
		}
	}

	numberOfPatches := -1
	for i, v := range patches {
		if v.Version <= version {
			numberOfPatches = i
			break
		}
	}

	if numberOfPatches == -1 {
		return formatPatches(patches), nil
	} else {
		return formatPatches(patches[0:numberOfPatches]), nil
	}
}

func formatPatches(patchList []PatchList) []Patch {
	var index int

	for _, v := range patchList {
		for _, w := range v.Patches {
			if w.Type == "flatpak" {
				addFlatpakPatch(w, &index)
			} else {
				finalPatches = append(finalPatches, w)
			}
			index += 1
		}
	}

	return finalPatches
}

func addFlatpakPatch(patch Patch, index *int) {
	if slices.Contains(flatpaks, patch.Arguments[0]) {
		finalPatches = slices.Delete(finalPatches, *index, *index)
		*index -= 1
	}

	if patch.Arguments[0] != "" {
		err := exec.Command("flatpak", "info", patch.Arguments[0]).Run()
		if err != nil {
			patch.Arguments[0] = ""
		}
	}

	err := exec.Command("flatpak", "info", patch.Arguments[1]).Run()
	if err == nil {
		*index -= 1
		return
	}

	flatpaks = append(flatpaks, patch.Arguments[1])
	finalPatches = append(finalPatches, patch)
}

func RunPatches() (bool, error) {
	perfectRun := true

	for _, v := range finalPatches {

		if v.Optional {
			answer := ""
			user_answered := false
			consent := true
			for !user_answered {
				fmt.Printf("Do you want to run patch %s? [Y/n]", v.Name)
				fmt.Scanln(&answer)
				answer = strings.Trim(answer, "\n")
				answer = strings.ToLower(answer)
				switch answer {
				case "", "y", "yes":
					user_answered = true
					consent = true
				case "n", "no":
					user_answered = true
					consent = false
				default:
					fmt.Println("Not a valid answer, please try again")
				}
			}

			if !consent {
				continue
			}
		}

		switch v.Type {
		case "shell":
			exec.Command(v.Arguments[0], v.Arguments[1:len(v.Arguments)]...)
		case "flatpak":
			if v.Arguments[0] != "" {
				err := exec.Command("flatpak", "remove", "-y", v.Arguments[0]).Run()
				if err != nil {
					perfectRun = false
				}
			}
			err := exec.Command("flatpak", "install", "-y", v.Arguments[1]).Run()
			if err != nil {
				perfectRun = false
			}
		}
	}

	if perfectRun {
		config, _ := os.LookupEnv("XDG_CONFIG_HOME")
		f, err := os.Create(config + "/vso/patchVersion")
		if err != nil {
			return false, err
		}
		defer f.Close()
		f.Write([]byte(strconv.Itoa(patches[0].Version)))
	}

	return perfectRun, nil
}

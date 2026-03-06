package core

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
)

// The root JSON object that holds all the patches
type RootPatch struct {
	PatchVersions []PatchVersion `json:"patchVersions"`
}

type PatchVersion struct {
	Version  int    `json:"version"`
	Name     string `json:"name"`
	Patches  []any  `json:"patches"`
	Optional bool   `json:"optional"`
}

// The format that every patch must follow, regardless of type
type Patch struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Type        string `json:"type"`
	Optional    bool   `json:"optional"`
}

type ShellPatch struct {
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Type        string   `json:"type"`
	Optional    bool     `json:"optional"`
	Command     string   `json:"commands"`
	Arguments   []string `json:"arguments"`
}

type FlatpakInstallPatch struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Type        string `json:"type"`
	Optional    bool   `json:"optional"`
	Flatpak     string `json:"flatpak"`
}

type FlatpakRemovePatch struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Type        string `json:"type"`
	Optional    bool   `json:"optional"`
	Flatpak     string `json:"flatpak"`
}

type FlatpakReplacePatch struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Type        string `json:"type"`
	Optional    bool   `json:"optional"`
	OldFlatpak  string `json:"old-flatpak"`
	NewFlatpak  string `json:"new-flatpak"`
}

var versionedPatches []PatchVersion
var requiredPatchList []any
var optionalPatchList []any
var patchListInitialised bool

func GetPatchVersionFilePath() (string, error) {
	currentUser, err := getRealUser()
	if err != nil {
		return "", err
	}

	patchVersionFilePath := fmt.Sprintf("/home/%s/.config/vso/patchVersion", currentUser)
	if _, err := os.Stat(patchVersionFilePath); err != nil {
		return "", nil
	}

	return patchVersionFilePath, nil
}

func LoadPatches() ([]PatchVersion, error) {
	// patchFilePath := "/usr/share/vso/patch.json"
	patchFilePath := "/etc/vso/patch.json"

	f, err := os.Open(patchFilePath)
	if err != nil {
		return []PatchVersion{}, err
	}
	defer f.Close()

	var loadedFile RootPatch
	err = json.NewDecoder(f).Decode(&loadedFile)
	if err != nil {
		return []PatchVersion{}, err
	}

	return loadedFile.PatchVersions, nil
}

func LoadPatchVersion() (int, error) {
	var version int

	patchVersionFilePath, err := GetPatchVersionFilePath()
	if err != nil {
		return 0, err
	}

	// Load the current patch version
	if patchVersionFilePath == "" { // The file does not exist
		version = 0
	} else {
		f, err := os.Open(patchVersionFilePath)
		if err != nil {
			return 0, err
		}

		var content []byte
		defer f.Close()

		_, err = f.Read(content)
		if err != nil {
			return 0, err
		}

		version, err = strconv.Atoi(strings.Split(string(content), "\n")[0])
		if err != nil {
			return 0, err
		}
	}

	return version, nil
}

// GetPatches aggregates all the patches that have to be applied into the final patch lists, and returns the first and last patches to be applied
func GetPatches() (PatchVersion, PatchVersion, error) {
	fullPatchList, err := LoadPatches()
	if err != nil {
		return PatchVersion{}, PatchVersion{}, err
	}

	currentPatchVersion, err := LoadPatchVersion()
	if err != nil {
		return PatchVersion{}, PatchVersion{}, err
	}

	for i, v := range fullPatchList {
		if v.Version <= currentPatchVersion {
			versionedPatches = fullPatchList[0:i]
			break
		}
	}

	for _, v := range fullPatchList {
		for _, w := range v.Patches {
			if !validatePatch(w) {
				continue
			}
			if w.(Patch).Optional {
				optionalPatchList = append(optionalPatchList, w)
			} else {
				requiredPatchList = append(requiredPatchList, w)
			}
		}
	}

	patchListInitialised = true
	return fullPatchList[0], fullPatchList[len(fullPatchList)-1], nil
}

func validatePatch(patch any) bool {
	switch patch.(Patch).Type {
	case "shell":
		return true
	case "flatpak-install":
		return true
	case "flatpak-remove":
		return validateFlatpakRemove(patch.(FlatpakRemovePatch))
	case "flatpak-replace":
		return validateFlatpakReplace(patch.(FlatpakReplacePatch))
	default:
		return false
	}
}

func validateFlatpakRemove(patch FlatpakRemovePatch) bool {
	err := exec.Command("flatpak", "info", patch.Flatpak).Run()
	return err == nil
}

func validateFlatpakReplace(patch FlatpakReplacePatch) bool {
	err := exec.Command("flatpak", "info", patch.OldFlatpak).Run()
	if err != nil {
		return false
	}
	err = exec.Command("flatpak", "info", "patch.NewFlatpak").Run()
	if err == nil {
		return false
	}
	return true
}

func GetOptionalPatches() ([]any, error) {
	if !patchListInitialised {
		return []any{}, fmt.Errorf("Patch list is requested before it is initialised")
	}

	return optionalPatchList, nil
}

func GetRequiredPatches() ([]any, error) {
	if !patchListInitialised {
		return []any{}, fmt.Errorf("Patch list is requested before it is initialised")
	}

	return requiredPatchList, nil
}

//TODO: Restore previous application state in case of failed patch application

func AsyncApplyPatches(infoChannel chan int, errorChannel chan error, optionalPatchConsent ...bool) {
	if len(optionalPatchConsent) != len(optionalPatchList) {
		errorChannel <- fmt.Errorf("Optional Patch Consent list is malformed")
		return
	}

	for _, v := range requiredPatchList {
		infoChannel <- 0

		switch v.(Patch).Type {
		case "shell":
			err := applyShellPatch(v.(ShellPatch))
			if err != nil {
				errorChannel <- err
			}
		case "flatpak-install":
			err := applyFlatpakInstallPatch(v.(FlatpakInstallPatch))
			if err != nil {
				errorChannel <- err
			}
		case "flatpak-remove":
			err := applyFlatpakRemovePatch(v.(FlatpakRemovePatch))
			if err != nil {
				errorChannel <- err
			}
		case "flatpak-replace":
			err := applyFlatpakReplacePatch(v.(FlatpakReplacePatch))
			if err != nil {
				errorChannel <- err
			}
		default:
			errorChannel <- fmt.Errorf("Invalid Patch Type: %s", v.(Patch).Type)
		}

		infoChannel <- 1
	}

	for i, v := range optionalPatchList {
		if !optionalPatchConsent[i] {
			continue
		}

		infoChannel <- 1

		switch v.(Patch).Type {
		case "shell":
			err := applyShellPatch(v.(ShellPatch))
			if err != nil {
				errorChannel <- err
			}
		case "flatpak-install":
			err := applyFlatpakInstallPatch(v.(FlatpakInstallPatch))
			if err != nil {
				errorChannel <- err
			}
		case "flatpak-remove":
			err := applyFlatpakRemovePatch(v.(FlatpakRemovePatch))
			if err != nil {
				errorChannel <- err
			}
		case "flatpak-replace":
			err := applyFlatpakReplacePatch(v.(FlatpakReplacePatch))
			if err != nil {
				errorChannel <- err
			}
		default:
			errorChannel <- fmt.Errorf("Invalid Patch Type: %s", v.(Patch).Type)
		}
		infoChannel <- 1
	}
	errorChannel <- nil
}

func applyShellPatch(patch ShellPatch) error {
	err := exec.Command(patch.Command, patch.Arguments...).Run()
	return err
}

func applyFlatpakInstallPatch(patch FlatpakInstallPatch) error {
	err := exec.Command("flatpak", "install", "-y", patch.Flatpak).Run()
	return err
}

func applyFlatpakRemovePatch(patch FlatpakRemovePatch) error {
	err := exec.Command("flatpak", "remove", "-y", patch.Flatpak).Run()
	return err
}

func applyFlatpakReplacePatch(patch FlatpakReplacePatch) error {
	err := exec.Command("flatpak", "remove", "-y", patch.OldFlatpak).Run()
	if err != nil {
		return err
	}
	err = exec.Command("flatpak", "install", "-y", patch.NewFlatpak).Run()
	return err
}

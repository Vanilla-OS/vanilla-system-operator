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
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/BurntSushi/toml"
	"github.com/buger/jsonparser"
)

type FdroidRepo struct {
	Name           string
	Key            string
	IndexURL       string
	RepoURL        string
	PackageURL     string
	PackageInfoURL string
}

type FdroidPackage struct {
	Name                 string
	RDNSName             string
	Summary              string
	Author               string
	Source               string
	License              string
	InstalledVersionCode int
	Repository           FdroidRepo
	Versions             []byte
}

type NoMatchError struct {
	Search string
	Err    error
}

func (n *NoMatchError) Error() string {
	return fmt.Sprintf("%s not found", n.Search)
}

type PackageInCache struct {
	Name string
	Err  error
}

func (p *PackageInCache) Error() string {
	return fmt.Sprintf("%s already in cache", p.Name)
}

type InstallDeclined struct{}

func (i *InstallDeclined) Error() string {
	return "Installation stopped"
}

var Repositories []FdroidRepo
var Indexes []os.DirEntry
var IndexCacheDir = fmt.Sprintf("%s/.cache/vso/indexes/", os.Getenv("HOME"))
var APKCacheDir = fmt.Sprintf("%s/.cache/vso/apks/", os.Getenv("HOME"))

func GetRepos() error {
	configFiles, err := os.ReadDir("/etc/vso/fdroid.repos.d/")
	if err != nil {
		return err
	}
	for _, config := range configFiles {
		if !strings.Contains(config.Name(), ".toml") {
			continue
		}
		parseRepoFile("/etc/vso/fdroid.repos.d/" + config.Name())
	}
	return nil
}

func parseRepoFile(file string) {
	var repo FdroidRepo
	repoData, err := os.ReadFile(file)
	if err != nil {
		fmt.Printf("err: %v\n", err)
		return
	}

	_, err = toml.Decode(string(repoData), &repo)
	if err != nil {
		_ = fmt.Errorf("ERROR: Failed to parse configuration file %s", file)
		fmt.Printf("err: %v\n", err)
	}
	Repositories = append(Repositories, repo)
}

// Download the index of a repo
// the index argument specifies which index of the Repositories variable to download
// if index is set to -1, then the indexes from all Repositories will be downloaded
func downloadIndex(index int, downTrans string) error {
	if index == -1 {
		for i := 0; i < len(Repositories); i++ {
			err := downloadIndex(i, downTrans)
			if err != nil {
				return err
			}
		}
	} else if len(Repositories)-1 < index {
		return fmt.Errorf("index out of range")
	} else {
		fmt.Printf(downTrans, Repositories[index].Name) // I don't know how to get vso.Trans here so this will have to do
		fmt.Println()
		currTime := time.Now().Unix()
		out, err := os.Create(fmt.Sprintf("%s/index-%s-%d.json", fmt.Sprintf("%s/.cache/vso/indexes/", os.Getenv("HOME")), Repositories[index].Name, currTime))
		if err != nil {
			return err
		}
		defer out.Close()
		resp, err := http.Get(Repositories[index].IndexURL)
		if err != nil {
			return err
		}
		defer resp.Body.Close()
		_, err = io.Copy(out, resp.Body)
		os.Chmod(fmt.Sprintf("%s/index-%s-%d.json", fmt.Sprintf("%s/.cache/vso/indexes/", os.Getenv("HOME")), Repositories[index].Name, currTime), 0777)
		return err
	}
	return nil
}

func SyncIndex(force bool, downTrans string) error {
	_, err := os.Stat(IndexCacheDir)
	if os.IsNotExist(err) {
		err := os.MkdirAll(IndexCacheDir, 0777)
		if err != nil {
			return err
		}
		err = downloadIndex(-1, downTrans)
		return err
	}

	if force {
		err := os.RemoveAll(IndexCacheDir)
		if err != nil {
			return err
		}
		err = os.MkdirAll(IndexCacheDir, 0777)
		if err != nil {
			return err
		}
		err = downloadIndex(-1, downTrans)
		return err
	}

	Indexes, err = os.ReadDir(IndexCacheDir)
	if err != nil {
		return err
	}
	if len(Indexes) <= 0 {
		err := downloadIndex(-1, downTrans)
		return err
	}
	var indexRepos []string
	for _, index := range Indexes {
		fileName := strings.ReplaceAll(index.Name(), ".json", "")
		indexFile := strings.Split(fileName, "-")
		indexRepos = append(indexRepos, indexFile[1])
		timestamp, err := strconv.ParseInt(indexFile[len(indexFile)-1], 10, 64)
		if err != nil {
			fmt.Printf("err: %v\n", err)
		}
		currentTime := time.Now().Unix()
		if currentTime-timestamp >= 604800 {
			fmt.Printf("Index for repo %s outdated! Syncing...\n", indexFile[1])
			err := os.Remove(fmt.Sprintf("%s/%s", IndexCacheDir, index))
			if err != nil {
				return err
			}
			for i, repository := range Repositories {
				if repository.Name == indexFile[1] {
					err := downloadIndex(i, downTrans)
					if err != nil {
						fmt.Printf("Error: %v\n", err)
					}
				}
			}
		}
	}

	for index, repository := range Repositories {
		if !sliceContains(indexRepos, repository.Name) {
			fmt.Printf("Index for repo %s not synced! Syncing now...\n", repository.Name)
			err := downloadIndex(index, downTrans)
			if err != nil {
				fmt.Printf("Error: %v\n", err)
			}
		}
	}
	return nil
}

func SearchIndex(search string, downTrans string) ([]FdroidPackage, error) {
	err := GetRepos()
	if err != nil {
		return nil, err
	}
	err = SyncIndex(false, downTrans)
	if err != nil {
		return nil, err
	}
	var matches []FdroidPackage
	var processed []string
	for _, repository := range Repositories {
		for _, index := range Indexes {
			if strings.Contains(index.Name(), repository.Name) && !sliceContains(processed, repository.Name) {
				processed = append(processed, repository.Name)
				indexContent, err := os.ReadFile(fmt.Sprintf("%s/%s", IndexCacheDir, index.Name()))
				if err != nil {
					return nil, err
				}
				err = jsonparser.ObjectEach(indexContent, func(key []byte, value []byte, dataType jsonparser.ValueType, offset int) error {
					name, err := jsonparser.GetString(value, "metadata", "name", "en-US")
					if err != nil {
						return err
					}
					if strings.Contains(strings.ToLower(name), strings.ToLower(search)) || strings.Contains(strings.ToLower(string(key)), strings.ToLower(search)) {
						summary, err := jsonparser.GetString(value, "metadata", "summary", "en-US")
						if err != nil {
							summary = "No summary found"
						}
						author, err := jsonparser.GetString(value, "metadata", "authorName")
						if err != nil {
							author = "No author found"
						}
						source, err := jsonparser.GetString(value, "metadata", "sourceCode")
						if err != nil {
							source = "No link to source code found"
						}
						license, err := jsonparser.GetString(value, "metadata", "license")
						if err != nil {
							license = "No license found"
						}
						versions, _, _, err := jsonparser.Get(value, "versions")
						if err != nil {
							return err
						}
						match := FdroidPackage{
							Name:       name,
							RDNSName:   string(key),
							Summary:    summary,
							Author:     author,
							Source:     source,
							License:    license,
							Repository: repository,
							Versions:   versions,
						}
						matches = append(matches, match)
					}
					return nil
				}, "packages")
				if err != nil {
					return nil, err
				}
			}
		}
	}
	return matches, nil
}

func GetPackageVersion(pkg FdroidPackage) (string, error) {
	req, err := http.NewRequest("GET", strings.ReplaceAll(pkg.Repository.PackageInfoURL, "%s", pkg.RDNSName), nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("Accept", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	suggestedVersion, _, _, err := jsonparser.Get(bodyBytes, "suggestedVersionCode")
	return string(suggestedVersion), err
}

func FetchPackage(match FdroidPackage) (string, error) {

	version, err := GetPackageVersion(match)
	if err != nil {
		return "", err
	}
	versionAsInt, _ := strconv.ParseInt(version, 10, 0)
	match.InstalledVersionCode = int(versionAsInt)
	apkName := fmt.Sprintf("%s_%s.apk", match.RDNSName, version)

	_, err = os.Stat(fmt.Sprintf("%s/%s", APKCacheDir, apkName))
	if !os.IsNotExist(err) {
		return fmt.Sprintf("%s/%s", APKCacheDir, apkName), &PackageInCache{Name: apkName}
	}
	out, err := os.Create(fmt.Sprintf("%s/%s", APKCacheDir, apkName))
	if err != nil {
		return "", err
	}
	defer out.Close()
	resp, err := http.Get(strings.ReplaceAll(match.Repository.PackageURL, "%s", apkName))
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%s/%s", APKCacheDir, apkName), nil
}

package core

import (
	"fmt"
	"github.com/BurntSushi/toml"
	"github.com/buger/jsonparser"
	"io"
	"net/http"
	"os"
	"slices"
	"strconv"
	"strings"
	"time"
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
	Name       string
	RDNSName   string
	summary    string
	author     string
	source     string
	license    string
	repository FdroidRepo
	versions   []byte
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
func downloadIndex(index int) error {
	if index == -1 {
		fmt.Println("syncing everything")
		for i := 0; i < len(Repositories); i++ {
			err := downloadIndex(i)
			if err != nil {
				return err
			}
		}
	} else if len(Repositories)-1 < index {
		return fmt.Errorf("index out of range")
	} else {
		fmt.Printf("Downloading repo %s\n", Repositories[index].Name)
		currTime := time.Now().Unix()
		out, err := os.Create(fmt.Sprintf("%s/index-%s-%d.json", fmt.Sprintf("%s/.cache/vso/indexes/", os.Getenv("HOME")), Repositories[index].Name, currTime))
		defer out.Close()
		if err != nil {
			return err
		}
		resp, err := http.Get(Repositories[index].IndexURL)
		defer resp.Body.Close()
		if err != nil {
			return err
		}
		_, err = io.Copy(out, resp.Body)
		return err
	}
	return nil
}

func SyncIndex(force bool) error {
	_, err := os.Stat(IndexCacheDir)
	if os.IsNotExist(err) {
		err := os.MkdirAll(IndexCacheDir, 0755)
		if err != nil {
			return err
		}
		err = downloadIndex(-1)
		return err
	}

	if force {
		err := os.RemoveAll(IndexCacheDir)
		if err != nil {
			return err
		}
		err = os.MkdirAll(IndexCacheDir, 07555)
		if err != nil {
			return err
		}
		err = downloadIndex(-1)
		return err
	}

	Indexes, err = os.ReadDir(IndexCacheDir)
	if err != nil {
		return err
	}
	if len(Indexes) <= 0 {
		err := downloadIndex(-1)
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
		if currentTime-timestamp >= 604800 { // TODO: Consider if a different period is more suited (currently 1 week)
			fmt.Printf("Index for repo %s outdated! Syncing...\n", indexFile[1])
			err := os.Remove(fmt.Sprintf("%s/%s", IndexCacheDir, index))
			if err != nil {
				return err
			}
			for i, repository := range Repositories {
				if repository.Name == indexFile[1] {
					err := downloadIndex(i)
					if err != nil {
						fmt.Printf("Error: %v\n", err)
					}
				}
			}
		}
	}

	for index, repository := range Repositories {
		if !slices.Contains(indexRepos, repository.Name) {
			fmt.Printf("Index for repo %s not synced! Syncing now...\n", repository.Name)
			err := downloadIndex(index)
			if err != nil {
				fmt.Printf("Error: %v\n", err)
			}
		}
	}
	return nil
}

func searchIndex(search string) ([]FdroidPackage, error) {
	err := GetRepos()
	if err != nil {
		return nil, err
	}
	err = SyncIndex(false)
	if err != nil {
		return nil, err
	}
	var matches []FdroidPackage
	var processed []string
	for _, repository := range Repositories {
		for _, index := range Indexes {
			if strings.Contains(index.Name(), repository.Name) && !slices.Contains(processed, repository.Name) {
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
							return err
						}
						author, err := jsonparser.GetString(value, "metadata", "authorName")
						if err != nil {
							return err
						}
						source, err := jsonparser.GetString(value, "metadata", "sourceCode")
						if err != nil {
							return err
						}
						license, err := jsonparser.GetString(value, "metadata", "license")
						if err != nil {
							return err
						}
						versions, _, _, err := jsonparser.Get(value, "versions")
						if err != nil {
							fmt.Printf("err")
							return err
						}
						match := FdroidPackage{
							Name:       name,
							RDNSName:   string(key),
							summary:    summary,
							author:     author,
							source:     source,
							license:    license,
							repository: repository,
							versions:   versions,
						}
						matches = append(matches, match)
					}
					return nil
				}, "packages")
			}
		}
	}
	return matches, nil
}

func SearchPackage(search string) error {
	matches, err := searchIndex(search)
	if err != nil {
		return err
	}
	for _, match := range matches {
		fmt.Printf("%s (%s) - %s [%s]\n", match.Name, match.RDNSName, match.summary, match.repository.Name)
	}
	return nil
}

func getPackageVersion(pkg FdroidPackage) (string, error) {
	req, err := http.NewRequest("GET", strings.ReplaceAll(pkg.repository.PackageInfoURL, "%s", pkg.RDNSName), nil)
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

func FetchPackage(installPackage string) (string, error) {
	_, err := os.Stat(APKCacheDir)
	if os.IsNotExist(err) {
		err := os.MkdirAll(APKCacheDir, 0755)
		if err != nil {
			return "", err
		}
	}
	matches, err := searchIndex(installPackage)
	if err != nil {
		return "", err
	}
	if len(matches) == 0 {
		return "", fmt.Errorf("no matches found")
	}
	version, err := getPackageVersion(matches[0])
	if err != nil {
		return "", err
	}
	apkName := fmt.Sprintf("%s_%s.apk", matches[0].RDNSName, version)
	if err != nil {
		return "", err
	}
	_, err = os.Stat(fmt.Sprintf("%s/%s", APKCacheDir, apkName))
	if !os.IsNotExist(err) {
		fmt.Println("APK already exists in cache, not downloading again")
		return fmt.Sprintf("%s/%s", APKCacheDir, apkName), nil
	}
	out, err := os.Create(fmt.Sprintf("%s/%s", APKCacheDir, apkName))
	defer out.Close()
	if err != nil {
		return "", err
	}
	fmt.Printf("Downloading %s", strings.ReplaceAll(matches[0].repository.PackageURL, "%s", apkName))
	resp, err := http.Get(strings.ReplaceAll(matches[0].repository.PackageURL, "%s", apkName))
	defer resp.Body.Close()
	if err != nil {
		return "", err
	}
	_, err = io.Copy(out, resp.Body)
	return fmt.Sprintf("%s/%s", APKCacheDir, apkName), nil
}

package apk

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strings"

	"gitlab.alpinelinux.org/alpine/go/repository"
)

type Context struct {
	client   *http.Client
	indexURL string
}

func New(client *http.Client, indexURL string) Context {
	return Context{
		client:   client,
		indexURL: indexURL,
	}
}

func (c Context) GetApkPackages() (map[string][]*repository.Package, error) {
	req, err := http.NewRequest("GET", c.indexURL, nil)
	if err != nil {
		return nil, err
	}
	resp, err := c.client.Do(req)
	if err != nil {
		return nil, errors.Join(err, fmt.Errorf("failed getting URI %s", c.indexURL))
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("non ok http response for URI %s code: %v", c.indexURL, resp.StatusCode)
	}

	return parseApkIndex(resp.Body)
}

func parseApkIndex(indexData io.ReadCloser) (map[string][]*repository.Package, error) {
	apkIndex, err := repository.IndexFromArchive(indexData)
	if err != nil {
		return nil, errors.Join(err, fmt.Errorf("failed to parse response %v", indexData))
	}

	return getPackagesMap(apkIndex.Packages), nil
}

func getPackagesMap(packages []*repository.Package) map[string][]*repository.Package {
	packageMap := make(map[string][]*repository.Package)
	for _, p := range packages {
		if packageMap[p.Name] == nil {
			packageMap[p.Name] = []*repository.Package{p}
			continue
		}
		packageMap[p.Name] = append(packageMap[p.Name], p)
	}
	return packageMap
}

func WildcardMatchPackageMap(apkPackages map[string][]*repository.Package, wildcardString string) map[string][]*repository.Package {
	packageNames := getPackageNamesFromWildcard(apkPackages, wildcardString)

	return collectPackageVersionsByPackageNames(apkPackages, packageNames)
}

// gets all package names that match the wildcard
// wildcard is only for versions, which means it won't check for packages with a longer suffix
// e.g. the wildcard 'argo-cd-*' will match the package "argo-cd-2.12", but not "argo-cd-2.12-repo-server"
// e.g. the wildcard 'argo-cd-*-repo-server' will match the package "argo-cd-2.12-repo-server"
func getPackageNamesFromWildcard(apkPackages map[string][]*repository.Package, wildcardString string) []string {
	// Check if the pattern contains "*"
	if strings.Contains(wildcardString, "*") {
		// Extract the prefix before "*"
		parts := strings.Split(wildcardString, "*")
		prefix := parts[0]
		suffix := parts[1]

		// Perform prefix match
		matched := []string{}
		var isNumberRegex = regexp.MustCompile(`\d+$`)
		for key := range apkPackages {
			if strings.HasPrefix(key, prefix) {
				if suffix == "" {
					// Match items that end with a number
					if match := isNumberRegex.MatchString(key); match {
						matched = append(matched, key)
					}
				} else {
					// Match items that end with the inferred suffix
					if strings.HasSuffix(key, suffix) {
						matched = append(matched, key)
					}
				}
			}
		}

		fmt.Printf("wildcardString '%s' matched following packages: %s\n", wildcardString, matched)
		return matched
	}
	fmt.Printf("wildcardString '%s' did not match packages '*'\n", wildcardString)
	return []string{wildcardString}
}

func collectPackageVersionsByPackageNames(apkPackages map[string][]*repository.Package, packageNames []string) map[string][]*repository.Package {
	matchedPackages := make(map[string][]*repository.Package)
	for key, packageList := range apkPackages {
		for _, packageName := range packageNames {

			if key == packageName {
				//if strings.HasPrefix(key, packageName) {
				matchedPackages[packageName] = append(matchedPackages[packageName], packageList...)
				for _, p := range packageList {
					fmt.Printf("Got package from apkindex: %s-%s using '%s' as key\n", key, p.Version, key)
				}
			}
		}
	}
	return matchedPackages
}

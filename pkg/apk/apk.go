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

// gets all package names that match the wildcard
// wildcard is only for versions, which means it won't check for packages with a longer suffix
// e.g. the wildcard 'argo-cd-*' will match the package "argo-cd-2.12", but not "argo-cd-2.12-repo-server"
// e.g. the wildcard 'argo-cd-*-repo-server' will match the package "argo-cd-2.12-repo-server"
func FilterPackagesByWildcard(apkPackages map[string][]*repository.Package, wildcardString string) map[string][]*repository.Package {

	matchedPackages := make(map[string][]*repository.Package)

	// Check if the pattern contains "*"
	if strings.Contains(wildcardString, "*") {
		// Extract the prefix before "*"
		parts := strings.Split(wildcardString, "*")
		prefix := parts[0]
		suffix := parts[1]

		// Perform prefix match
		var isNumberRegex = regexp.MustCompile(`\d+$`)
		for packageName, apkPackageRepos := range apkPackages {
			if strings.HasPrefix(packageName, prefix) {
				if suffix == "" {
					// Match items that end with a number
					if match := isNumberRegex.MatchString(packageName); match {
						fmt.Printf("'%s' Matched package with prefix: %s\n", wildcardString, packageName)
						matchedPackages[packageName] = apkPackageRepos
					}
				} else {
					// Match items that end with the inferred suffix
					if strings.HasSuffix(packageName, suffix) {
						fmt.Printf("'%s' Matched package with suffix: %s\n", wildcardString, packageName)
						matchedPackages[packageName] = apkPackageRepos
					}
				}
			}
		}
		return matchedPackages
	} else {
		fmt.Printf("No wildcard detected. Trying '%s' as a direct package-name match\n", wildcardString)
		matchedPackages[wildcardString] = apkPackages[wildcardString]
	}
	fmt.Printf("wildcardString '%s' did not match packages '*'\n", wildcardString)
	return matchedPackages
}

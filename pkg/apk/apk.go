package apk

import (
	"errors"
	"fmt"
	"gitlab.alpinelinux.org/alpine/go/repository"
	"io"
	"net/http"
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

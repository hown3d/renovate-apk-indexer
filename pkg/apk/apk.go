package apk

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"

	"gitlab.alpinelinux.org/alpine/go/repository"
)

type Context struct {
	client    *http.Client
	indexURLs []string
}

func New(client *http.Client, indexURLs []string) Context {
	return Context{
		client:    client,
		indexURLs: indexURLs,
	}
}

func (c Context) GetApkPackages() (map[string][]*repository.Package, error) {
	var packages []*repository.Package

	for _, url := range c.indexURLs {
		req, err := http.NewRequest("GET", url, nil)
		if err != nil {
			return nil, err
		}
		resp, err := c.client.Do(req)
		if err != nil {
			return nil, errors.Join(err, fmt.Errorf("failed getting URI %s", url))
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			return nil, fmt.Errorf("non ok http response for URI %s code: %v", url, resp.StatusCode)
		}

		ps, err := parseApkIndex(resp.Body)
		if err != nil {
			return nil, err
		}

		packages = append(packages, ps...)
	}

	return getPackagesMap(packages), nil
}

func parseApkIndex(indexData io.ReadCloser) ([]*repository.Package, error) {
	apkIndex, err := repository.IndexFromArchive(indexData)
	if err != nil {
		return nil, errors.Join(err, fmt.Errorf("failed to parse response %v", indexData))
	}

	return apkIndex.Packages, nil
}

func getPackagesMap(packages []*repository.Package) map[string][]*repository.Package {
	packageMap := make(map[string][]*repository.Package)
	for _, p := range packages {
		packageMap[p.Name] = append(packageMap[p.Name], p)

		for _, provide := range p.Provides {
			if strings.Contains(provide, ":") {
				continue
			}

			name, _, _ := strings.Cut(provide, "=")
			packageMap[name] = append(packageMap[name], p)
		}
	}
	return packageMap
}

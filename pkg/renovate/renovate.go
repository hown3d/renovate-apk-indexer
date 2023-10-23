package renovate

import (
	"gitlab.alpinelinux.org/alpine/go/repository"
	"time"
)

type Datasource struct {
	Releases        []Release `json:"releases"`
	SourceUrl       string    `json:"sourceUrl"`
	SourceDirectory string    `json:"sourceDirectory"`
	ChangelogUrl    string    `json:"changelogUrl"`
	Homepage        string    `json:"homepage"`
}

type Release struct {
	Version          string    `json:"version"`
	IsDeprecated     bool      `json:"isDeprecated"`
	ReleaseTimestamp time.Time `json:"releaseTimestamp"`
	ChangelogUrl     string    `json:"changelogUrl"`
	SourceUrl        string    `json:"sourceUrl"`
	SourceDirectory  string    `json:"sourceDirectory"`
}

func TransformAPKPackage(apkPackages []*repository.Package) *Datasource {
	releases := make([]Release, 0, len(apkPackages))
	for _, p := range apkPackages {
		releases = append(releases, Release{
			Version:          p.Version,
			IsDeprecated:     false,
			ReleaseTimestamp: p.BuildTime,
			SourceUrl:        p.Origin,
		})
	}
	return &Datasource{
		Releases: releases,
	}
}

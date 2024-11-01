package apk

import (
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"gitlab.alpinelinux.org/alpine/go/repository"
)

type ResultPackages struct {
	PackageNames []string
	VersionList  []string
}

func Test_prefixPackageName(t *testing.T) {
	type args struct {
		prefix string
	}
	tests := []struct {
		name    string
		args    args
		want    map[string][]string
		wantErr assert.ErrorAssertionFunc
	}{
		{
			name: "all packages from package name",
			args: args{
				prefix: "nodejs",
			},
			want: map[string][]string{
				"nodejs": {
					"18.12.1-r0",
					"18.13.0-r0",
				},
			},
			wantErr: assert.NoError,
		},
		{
			name: "all packages from package name with wildcard",
			args: args{
				prefix: "nodejs*",
			},
			want: map[string][]string{
				//should not match 18, as that package does not end with a number
				"nodejs-19": {
					"19.8.1-r0",
					"19.9.0-r0",
				},
				"nodejs-22": {
					"22.7.0-r0",
					"22.8.0-r0",
				},
			},
			wantErr: assert.NoError,
		},
		{
			name: "specific version of package name",
			args: args{
				prefix: "nodejs-22",
			},
			want: map[string][]string{
				"nodejs-22": {
					"22.7.0-r0",
					"22.8.0-r0",
				},
			},
			wantErr: assert.NoError,
		},
		{
			name: "wildcard at the end",
			args: args{
				prefix: "argo-cd-*",
			},
			want: map[string][]string{
				"argo-cd-2.11": {
					"2.11.7-r0",
				},
				"argo-cd-2.12": {
					"2.12.4-r0",
					"2.12.6-r0",
				},
			},
			wantErr: assert.NoError,
		},
		{
			name: "wildcard in the middle",
			args: args{
				prefix: "argo-cd-*-repo-server",
			},
			want: map[string][]string{
				"argo-cd-repo-server": {
					"2.8.0-r1",
				},
				"argo-cd-2.11-repo-server": {
					"2.11.7-r0",
				},
				"argo-cd-2.12-repo-server": {
					"2.12.3-r1",
					"2.12.4-r0",
					"2.12.6-r0",
				},
			},
			wantErr: assert.NoError,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			apkindexFile, err := os.Open("testdata/APKINDEX")
			if err != nil {
				fmt.Println("Error opening file:", err)
				return
			}
			packages, err := parseApkIndexFromTextFile(apkindexFile)
			if err != nil {
				log.Fatalf("error getting apk packages: %s", err)
			}
			fmt.Println("Packages: ", packages)

			got := WildcardMatchPackageMap(packages, tt.args.prefix)
			fmt.Println("Got: ", got)

			versionMap := make(map[string][]string, 0)
			//TODO can't only match for prefix...
			for key, resultPackageList := range got {
				for _, p := range resultPackageList {
					versionMap[key] = append(versionMap[key], p.Version)
				}
			}

			assert.EqualValuesf(t, tt.want, versionMap, "PrefixMatchPackageMap(%v)", tt.args.prefix)
		})
	}
}

func parseApkIndexFromTextFile(indexData io.ReadCloser) (map[string][]*repository.Package, error) {
	apkIndex, err := repository.ParsePackageIndex(indexData)
	if err != nil {
		return nil, errors.Join(err, fmt.Errorf("failed to parse response %v", indexData))
	}

	return getPackagesMap(apkIndex), nil
}

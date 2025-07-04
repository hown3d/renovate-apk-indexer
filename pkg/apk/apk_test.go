package apk

import (
	"embed"
	"testing"

	"github.com/stretchr/testify/assert"
	"gitlab.alpinelinux.org/alpine/go/repository"
)

//go:embed testdata
var testdata embed.FS

func Test_getPackagesMap(t *testing.T) {
	tests := []struct {
		name     string // description of this test case
		packages []*repository.Package
		want     map[string][]*repository.Package
	}{
		{
			name: "packages with provides",
			packages: []*repository.Package{
				{
					Name:    "argocd",
					Version: "3.0",
				},
				{
					Name:    "argocd-2.12",
					Version: "2.12",
					Provides: []string{
						"argocd=2.12",
					},
				},
			},
			want: map[string][]*repository.Package{
				"argocd": {
					{
						Name:    "argocd",
						Version: "3.0",
					},
					{
						Name:    "argocd-2.12",
						Version: "2.12",
						Provides: []string{
							"argocd=2.12",
						},
					},
				},
				"argocd-2.12": {
					{
						Name:    "argocd-2.12",
						Version: "2.12",
						Provides: []string{
							"argocd=2.12",
						},
					},
				},
			},
		},
		{
			name: "packages with provides with cmd prefix",
			packages: []*repository.Package{
				{
					Name:    "argocd",
					Version: "3.0",
				},
				{
					Name:    "argocd-2.12",
					Version: "2.12",
					Provides: []string{
						"cmd:argocd=2.12",
					},
				},
			},
			want: map[string][]*repository.Package{
				"argocd": {
					{
						Name:    "argocd",
						Version: "3.0",
					},
				},
				"argocd-2.12": {
					{
						Name:    "argocd-2.12",
						Version: "2.12",
						Provides: []string{
							"cmd:argocd=2.12",
						},
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := getPackagesMap(tt.packages)
			assert.Equal(t, tt.want, got)
		})
	}
}

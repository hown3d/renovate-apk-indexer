# Renovate APK Index Server

This project aims to create a http server for [renovate](https://github.com/renovatebot/renovate) to query apk packages of the [wolfi project](https://github.com/wolfi-dev/os).

# Usage

For every release a container image is built and stored in the github container registry: `docker run github.com/hown3d/renovate-apk-indexer`

```
$ renovate-apk-indexer -help
Usage of renovate-apk-indexer:
  -apk-index-url string
        url of the apk index to get the package information from (default "https://packages.wolfi.dev/os/x86_64/APKINDEX.tar.gz")
  -update-interval int
        update interval of the apk package index in hours (default 4)
```

## Renovate Gitlab example

Renovate gitlab-job (abbreviated):
```yaml
renovate:
  services:
    - name: ghcr.io/hown3d/renovate-apk-indexer:v0.0.3
      alias: wolfi-apk
  script:
    - renovate
```

Use it in your renovate.json as a custom datasource:
```json
{
  "$schema": "https://docs.renovatebot.com/renovate-schema.json",
  "customDatasources": {
    "wolfi": {
      "defaultRegistryUrlTemplate": "http://wolfi-apk:3000/wildcardVersion/{{packageName}}"
    }
  }
}
```

Usage in code:
```bash
# renovate: datasource=custom.wolfi depName=argo-cd
VERSION=2.13.1-r0
apk update && apk add argo-cd=$VERSION
```

Renovate should now be able to detect updates for the specified dependency.
Note that versioned package names can be detected with the `wildcardVersion` API. For more Details look at [Wildcard Version API](#wildcard-version-api):

```bash
# renovate: datasource=custom.wolfi depName=argo-cd-*
VERSION=2.13.1-r0
apk update && apk add argo-cd=$VERSION

# renovate: datasource=custom.wolfi depName=argo-cd-*-repo-server
VERSION=2.13.1-r0
apk update && apk add argo-cd-repo-server=$VERSION
```

## API

The server provides an endpoint for `/<PACKAGE_NAME>`, which returns the package information of `<PACKAGE_NAME>` in the [format that renovate custom datasources expects](https://docs.renovatebot.com/modules/datasource/custom/)

### Wildcard Version API

An additional endpoint is provided at `/wildcardVersion/<PACKAGE_NAME_WITH_WILDCARD_VERSION>`. The wolfi APK index uses an unusual naming scheme where the version is provided within the package name, it is not easily possible for renovate to track version updates for those packages:

E.g. These are all valid package names in the wolfi APK index. Note that the base name without the number does not necessarily contain the most recent version:
```
argo-cd
argo-cd-compat
argo-cd-repo-server
argo-cd-2.9
argo-cd-2.9-compat
argo-cd-2.9-repo-server
argo-cd-2.10
argo-cd-2.10-compat
argo-cd-2.10-repo-server

nodejs
nodejs-18
nodejs-20
nodejs-22
```

To use the wildcard version endpoint, you can provide a single '*' at the point where the renovate-apk-indexer should expect a version number. Be aware that this will skip the base-package-name if there is no suffix on the wildcard string. This is a technical limitation, see explaination at `/wildcardVersion/argo-cd*`

| Endpoint Request                  | Expected package name searches                                                                                                                                                                                                                      |
|-----------------------------------|-----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| `/wildcardVersion/argo-cd*`         | will only match `argo-cd-<number>`, not `argo-cd` or `argo-cd-<number>-compat`  This is a technical limitation to prevent the regex to match argo-cd-compat. If the suffix of the wildcard is empty it won't consider "non-number" matches for the suffix |
| `/wildcardVersion/argo-cd*compat`   | will only match `argo-cd-<number>-compat` and argo-cd-compat                                                                                                                                                                                          |
| `/wildcardVersion/argo-cd-*-compat` | will only match `argo-cd-<number>-compat` and argo-cd-compat                                                                                                                                                                                          |
| `/wildcardVersion/nodejs*`          | will only match `nodejs-<number>`, not `nodejs`.  This is a technical limitation, see explaination of `/wildcardVersion/argo-cd*`                                                                                                                       |


## APK Index updates

By default, the server updates it's packages every 4 hours.

## Healthchecks

Liveness and Readiness probe endpoints are available at `/livez` and `/readyz`.
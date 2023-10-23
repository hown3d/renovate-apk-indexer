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

## API

The server provides one single endpoint `/<PACKAGE_NAME>`, which returns the package information of `<PACKAGE_NAME>` in the [format that renovate custom datasources expects](https://docs.renovatebot.com/modules/datasource/custom/)

## APK Index updates

By default, the server updates it's packages every 4 hours.
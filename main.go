package main

import (
	"flag"
	"fmt"
	"net/http"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/log"
	"github.com/gofiber/fiber/v2/middleware/healthcheck"
	"github.com/hown3d/renovate-apk-indexer/pkg/apk"
	"github.com/hown3d/renovate-apk-indexer/pkg/renovate"
	"gitlab.alpinelinux.org/alpine/go/repository"
)

const wolfiIndex = "https://packages.wolfi.dev/os/x86_64/APKINDEX.tar.gz"

var (
	updateInterval = flag.Int("update-interval", 4, "update interval of the apk package index in hours")
	apkIndexUrl    = flag.String("apk-index-url", wolfiIndex, "url of the apk index to get the package information from")
)

func main() {
	flag.Parse()

	apkContext := apk.New(http.DefaultClient, *apkIndexUrl)
	apkPackages, err := apkContext.GetApkPackages()
	if err != nil {
		log.Fatalf("error getting apk packages: %s", err)
	}

	ticker := time.NewTicker(time.Duration(*updateInterval) * time.Hour)
	go func() {
		for {
			select {
			case <-ticker.C:
				log.Info("updating apk packages")
				newPackages, err := apkContext.GetApkPackages()
				if err != nil {
					log.Warnf("error updating apk packages: %s", err)
				}
				apkPackages = newPackages
			}
		}
	}()

	app := fiber.New()
	app.Use(healthcheck.New())
	app.Get("/:package", func(c *fiber.Ctx) error {
		packageName := c.Params("package")
		packages, ok := apkPackages[packageName]
		if !ok {
			return fmt.Errorf("%s not found in wolfi apkIndex", packageName)
		}
		datasource := renovate.TransformAPKPackage(packages)
		return c.JSON(datasource)
	})

	app.Get("/wildcardVersion/:package", func(c *fiber.Ctx) error {
		packageName := c.Params("package")
		matchedPackages := apk.WildcardMatchPackageMap(apkPackages, packageName)
		var packageList []*repository.Package
		for _, packages := range matchedPackages {
			packageList = append(packageList, packages...)
		}
		if len(packageList) == 0 {
			return fmt.Errorf("%s not found in wolfi apkIndex", packageName)
		}
		datasource := renovate.TransformAPKPackage(packageList)
		return c.JSON(datasource)
	})

	if err := app.Listen(":3000"); err != nil {
		log.Fatal(err)
	}
}

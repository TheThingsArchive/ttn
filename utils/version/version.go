// Copyright © 2017 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package version

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"runtime"
	"strings"
	"time"

	ttnlog "github.com/TheThingsNetwork/go-utils/log"
	"github.com/kardianos/osext"
	"github.com/spf13/viper"
)

// ErrNotFound indicates that a version was not found
var ErrNotFound = errors.New("Not Found")

// ReleaseHost is where we publish our releases
const ReleaseHost = "ttnreleases.blob.core.windows.net/release"

// Info contains version information
type Info struct {
	Version string
	Commit  string
	Date    time.Time
}

// GetLatestInfo gets information about the latest release for the current branch
func GetLatestInfo() (*Info, error) {
	location := fmt.Sprintf("https://%s/%s/info", ReleaseHost, viper.GetString("gitBranch"))
	resp, err := http.Get(location)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("Status %d was not OK", resp.StatusCode)
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	var info Info
	for _, line := range strings.Split(string(body), "\n") {
		infoLine := strings.SplitAfterN(line, " ", 2)
		if len(infoLine) != 2 {
			continue
		}
		switch strings.TrimSpace(infoLine[0]) {
		case "version":
			info.Version = infoLine[1]
		case "commit":
			info.Commit = infoLine[1]
		case "date":
			if date, err := time.Parse(time.RFC3339, infoLine[1]); err == nil {
				info.Date = date
			}
		}
	}

	return &info, nil
}

// GetLatest gets the latest release binary
func GetLatest(binary string) ([]byte, error) {
	exe := ""
	if runtime.GOARCH == "windows" {
		exe = ".exe"
	}
	filename := fmt.Sprintf("%s-%s-%s%s", binary, runtime.GOOS, runtime.GOARCH, exe)
	location := fmt.Sprintf("https://%s/%s/%s.tar.gz", ReleaseHost, viper.GetString("gitBranch"), filename)
	resp, err := http.Get(location)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	gr, err := gzip.NewReader(bytes.NewBuffer(body))
	if err != nil {
		return nil, err
	}
	archive, err := ioutil.ReadAll(gr)
	if err != nil {
		return nil, err
	}
	tr := tar.NewReader(bytes.NewBuffer(archive))
	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}
		if hdr.Name == filename {
			binary, err := ioutil.ReadAll(tr)
			if err != nil {
				return nil, err
			}
			return binary, nil
		}
	}
	return nil, ErrNotFound
}

// Selfupdate runs a self-update for the current binary
func Selfupdate(ctx ttnlog.Interface, component string) {
	if viper.GetString("gitBranch") == "unknown" {
		ctx.Infof("You are not using an official %s build. Not proceeding with the update", component)
		return
	}

	info, err := GetLatestInfo()
	if err != nil {
		ctx.WithError(err).Fatal("Could not get version information from the server")
	}

	if viper.GetString("gitCommit") == info.Commit {
		ctx.Info("The git commit of the build on the server is the same as yours")
		ctx.Info("Not proceeding with the update")
		return
	}

	if date, err := time.Parse(time.RFC3339, viper.GetString("buildDate")); err == nil {
		if date.Equal(info.Date) {
			ctx.Infof("You have the latest version of %s", component)
			ctx.Info("Nothing to update")
			return
		}
		if date.After(info.Date) {
			ctx.Infof("Your build is %s newer than the build on the server", date.Sub(info.Date))
			ctx.Info("Not proceeding with the update")
			return
		}
		ctx.Infof("The build on the server is %s newer than yours", info.Date.Sub(date))
	}

	ctx.Infof("Downloading the latest %s...", component)
	binary, err := GetLatest(component)
	if err != nil {
		ctx.WithError(err).Fatal("Could not download latest binary")
	}
	filename, err := osext.Executable()
	if err != nil {
		ctx.WithError(err).Fatal("Could not get path to local binary")
	}
	stat, err := os.Stat(filename)
	if err != nil {
		ctx.WithError(err).Fatal("Could not stat local binary")
	}
	ctx.Info("Replacing local binary...")
	if err := ioutil.WriteFile(filename+".new", binary, stat.Mode()); err != nil {
		ctx.WithError(err).Fatal("Could not write new binary to filesystem")
	}
	if err := os.Rename(filename, filename+".old"); err != nil {
		ctx.WithError(err).Fatal("Could not rename binary")
	}
	if err := os.Rename(filename+".new", filename); err != nil {
		ctx.WithError(err).Fatal("Could not rename binary")
	}
	ctx.Infof("Updated %s to the latest version", component)
}

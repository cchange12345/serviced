// Copyright 2015 The Serviced Authors.
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package volume

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/zenoss/glog"
)

func OldDetectDriverType(root string) (DriverType, error) {
	// Check to see if the directory even exists. If not, no driver has been initialized.
	glog.V(2).Infof("Detecting driver type under %s", root)
	if _, err := os.Stat(root); err != nil {
		if os.IsNotExist(err) {
			glog.V(2).Infof("Root does not exist; no driver has been initialized")
			return "", ErrDriverNotInit
		}
		return "", err
	}
	// Check for .devicemapper directory, which unequivocally indicates a devicemapper driver
	if fi, err := os.Stat(filepath.Join(root, ".devicemapper")); !os.IsNotExist(err) && fi != nil && fi.IsDir() {
		glog.V(2).Infof("Found .devicemapper directory; returning devicemapper")
		return DriverTypeDeviceMapper, nil
	}
	// Check if there are any volumes
	fis, err := ioutil.ReadDir(root)
	if err != nil {
		glog.Errorf("Error reading directory: %s", err)
		return "", err
	}
	var names []string
	for _, fi := range fis {
		if fi.IsDir() && fi.Name() != "monitor" {
			names = append(names, fi.Name())
		}
	}
	if len(names) == 0 {
		// No volumes, so essentially no driver
		glog.V(2).Infof("Found no volumes in the root. Assuming no driver has been initialized.")
		return "", ErrDriverNotInit
	}
	glog.V(2).Infof("Found possible volumes under %s: %s", root, names)
	// Check to see if it's a btrfs filesystem
	if IsBtrfsFilesystem(root) {
		var sudoer bool
		if !IsRoot() {
			sudoer = IsSudoer()
			if !sudoer {
				glog.Errorf("Unable to execute btrfs commands, so can't detect driver type")
				return "", ErrInsufficientPermissions
			}
		}
		if _, err := RunBtrFSCmd(sudoer, "subvolume", "show", filepath.Join(root, names[0])); err == nil {
			// It's btrfs
			glog.V(2).Infof("Found btrfs filesystem, and a volume is a subvolume; returning btrfs")
			return DriverTypeBtrFS, nil
		}
	}
	return DriverTypeRsync, nil
}

func DetectDriverType(root string) (DriverType, error) {
	// Check to see if the directory even exists. If not, no driver has been initialized.
	glog.V(2).Infof("Detecting driver type under %s", root)
	if _, err := os.Stat(root); err != nil {
		if os.IsNotExist(err) {
			glog.V(2).Infof("Root does not exist; no driver has been initialized")
			return "", ErrDriverNotInit
		}
		return "", err
	}
	for _, drivertype := range []DriverType{DriverTypeBtrFS, DriverTypeRsync, DriverTypeDeviceMapper} {
		dirname := fmt.Sprintf(".%s", drivertype)
		if fi, err := os.Stat(filepath.Join(root, dirname)); !os.IsNotExist(err) && fi != nil && fi.IsDir() {
			glog.V(2).Infof("Found %s directory; returning %s", dirname, drivertype)
			return drivertype, nil
		}
	}
	return "", ErrDriverNotInit
}
/*
Copyright 2024 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package swap

import (
	"os"
	sysruntime "runtime"
	"sync"

	"k8s.io/klog/v2"
	"k8s.io/mount-utils"
)

var (
	tmpfsNoswapOptionSupported        bool
	tmpfsNoswapOptionAvailabilityOnce sync.Once
)

const TmpfsNoswapOption = "noswap"

func IsTmpfsNoswapOptionSupported(mounter mount.Interface) bool {
	isTmpfsNoswapOptionSupportedHelper := func() bool {
		if sysruntime.GOOS == "windows" {
			return false
		}

		mountDir, err := os.MkdirTemp("", "tmpfs-noswap-test-")
		if err != nil {
			klog.InfoS("error creating dir to test if tmpfs noswap is enabled. Assuming not supported", "mount path", mountDir, "error", err)
			return false
		}

		defer func() {
			err = os.RemoveAll(mountDir)
			if err != nil {
				klog.ErrorS(err, "error removing test tmpfs dir", "mount path", mountDir)
			}
		}()

		err = mounter.MountSensitiveWithoutSystemd("tmpfs", mountDir, "tmpfs", []string{TmpfsNoswapOption}, nil)
		if err != nil {
			klog.InfoS("error mounting tmpfs with the noswap option. Assuming not supported", "error", err)
			return false
		}

		err = mounter.Unmount(mountDir)
		if err != nil {
			klog.ErrorS(err, "error unmounting test tmpfs dir", "mount path", mountDir)
		}

		return true
	}

	tmpfsNoswapOptionAvailabilityOnce.Do(func() {
		tmpfsNoswapOptionSupported = isTmpfsNoswapOptionSupportedHelper()
	})

	return tmpfsNoswapOptionSupported
}

// Copyright © 2018 NAME HERE <EMAIL ADDRESS>
//
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

package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path"
	"syscall"
)

const (
	StatusSuccess = "Success"
	StatusFailure = "Failure"
)

type DriverOutput struct {
	Status       string       `json:"status,omitempty"`
	Message      string       `json:"message,omitempty"`
	Device       string       `json:"device,omitempty"`
	VolumeName   string       `json:"volumeName,omitempty"`
	Attached     bool         `json:"attached,omitempty"`
	Capabilities capabilities `json:"capabilities,omitempty"`
}

type capabilities struct {
	Attach bool `json:"attach"`
}

func output(driverOutput DriverOutput) error {
	data, err := json.Marshal(driverOutput)
	if err != nil {
		return nil
	}
	fmt.Println(string(data))
	return nil
}

func makeOutput(status string, message string) DriverOutput {
	return DriverOutput{Status: status, Message: message}
}

type jsonOptions struct {
	PodNamespace string `json:"kubernetes.io/pod.namespace"`
	PodUid       string `json:"kubernetes.io/pod.uid"`
	PodName      string `json:"kubernetes.io/pod.name"`

	HostPath string `json:"hostPath"`
}

type HostPathPerPodDriver struct {
}

func NewHostPathPerPodDriver() (*HostPathPerPodDriver, error) {
	return &HostPathPerPodDriver{}, nil
}

func (driver *HostPathPerPodDriver) Init() DriverOutput {
	return DriverOutput{
		Status:       StatusSuccess,
		Message:      "Success",
		Capabilities: capabilities{Attach: false},
	}
}

func (driver *HostPathPerPodDriver) Mount(mountDir string, jsonOptRaw string) DriverOutput {
	var jsonOpt jsonOptions

	if err := json.Unmarshal([]byte(jsonOptRaw), &jsonOpt); err != nil {
		return makeOutput(StatusFailure, err.Error())
	}

	hostPath := path.Join(jsonOpt.HostPath, jsonOpt.PodNamespace, jsonOpt.PodName, jsonOpt.PodUid)

	if err := driver.ensureHostPathExists(hostPath); err != nil {
		return makeOutput(StatusFailure, err.Error())
	}

	if err := syscall.Mount(hostPath, mountDir, "none", syscall.MS_BIND, ""); err != nil {
		return makeOutput(StatusFailure, err.Error())
	}

	return DriverOutput{Status: StatusSuccess}
}

func (driver *HostPathPerPodDriver) Unmount(mountDir string) DriverOutput {
	if err := syscall.Unmount(mountDir, syscall.MNT_FORCE); err != nil {
		return makeOutput(StatusFailure, err.Error())
	}
	return DriverOutput{Status: StatusSuccess}
}

func (driver *HostPathPerPodDriver) ensureHostPathExists(p string) error {
	if _, err := os.Lstat(p); err != nil {
		if os.IsNotExist(err) {
			return os.MkdirAll(p, os.ModePerm)
		}
		return err
	}
	return nil
}

func printUsage() {
	fmt.Println(`# This flex volume plugin is like hostPath, but create host directory with pod meta. 
# For example, if you specified the host path like "/root/hostpath", the actual path will be "/root/hostpath/<pod_namespace>/<pod_name>/<pod_uid>".
# Note: Only directory is supported.

Usage: 
    init
    mount <mount dir> <mount device> <json params>"
    unmount <mount dir>"`)
}

func main() {

	driver, err := NewHostPathPerPodDriver()
	if err != nil {
		panic(err)
	}

	if len(os.Args) <= 1 {
		printUsage()
		return
	}

	switch action := os.Args[1]; action {
	case "init":
		output(driver.Init())
	case "mount":
		if len(os.Args) < 3 {
			printUsage()
		}
		output(driver.Mount(os.Args[2], os.Args[3]))
	case "unmount":
		if len(os.Args) < 2 {
			printUsage()
		}
		output(driver.Unmount(os.Args[2]))
	}
}

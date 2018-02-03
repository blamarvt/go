// Copyright 2017 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package msvc

import (
	"fmt"
	"internal/syscall/windows/registry"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// Windows SDK Meta Data
type winSdkVersion struct {
	path    string
	version string
}

// Information required to build via MSVC environment
type MSVCEnvironment struct {
	command        string
	rootPath       string
	executablePath string
	windowsSdk     *winSdkVersion
	includes       []string
	libs           []string
	libpath        []string
	commands       map[string]string
}

// Convert a version string into an ordinal version
func versionOrdinal(version string) string {
	// ISO/IEC 14651:2011
	const maxByte = 1<<8 - 1
	vo := make([]byte, 0, len(version)+8)
	j := -1
	for i := 0; i < len(version); i++ {
		b := version[i]
		if '0' > b || b > '9' {
			vo = append(vo, b)
			j = -1
			continue
		}
		if j == -1 {
			vo = append(vo, 0x00)
			j = len(vo) - 1
		}
		if vo[j] == 1 && vo[j+1] == '0' {
			vo[j+1] = b
			continue
		}
		if vo[j]+1 > maxByte {
			panic("VersionOrdinal: invalid version")
		}
		vo = append(vo, b)
		vo[j]++
	}
	return string(vo)
}

// Get a MSVCEnvironment from a command
// typically this would be a CC command
func FromCommand(command string) (*MSVCEnvironment, error) {
	// Check if we have the full command or just the executable name
	command = strings.Trim(strings.TrimSpace(command), "\"")
	if _, err := os.Stat(command); err != nil {
		fullPath, err := tryLocateCommandPath(command)
		if err != nil {
			return nil, err
		}
		command = strings.TrimSpace(fullPath)
	}
	result := &MSVCEnvironment{command: command}
	result.executablePath = filepath.Dir(command)
	result.rootPath = tryFindBaseDir(result.executablePath)
	result.windowsSdk = getWindowsSDK()
	result.commands = make(map[string]string)
	result.commands[command] = result.GetMSVCCommand(command)
	return result, nil
}

func tryLocateCommandPath(command string) (string, error) {
	cmd := exec.Command("where", command)
	if out, err := cmd.CombinedOutput(); err != nil {
		return "", fmt.Errorf("Could not locate %s command", command)
	} else {
		if _, err := os.Stat(string(out)); err != nil {
		}
		return string(out), nil
	}
}

func compareVersion(v1, v2 string) int {
	v1 = strings.TrimPrefix(v1, "v")
	v2 = strings.TrimPrefix(v2, "v")
	v1 = versionOrdinal(strings.TrimSpace(v1))
	v2 = versionOrdinal(strings.TrimSpace(v2))
	if v1 > v2 {
		return -1
	}
	return 1
}

func getWindowsSDK() *winSdkVersion {
	roots := []registry.Key{
		registry.LOCAL_MACHINE,
		registry.CURRENT_USER,
	}
	paths := []string{
		`SOFTWARE\WOW6432Node\Microsoft\Microsoft SDKs\Windows`,
		`SOFTWARE\Microsoft\Microsoft SDKs\Windows`,
	}
	sdkDirs := make(map[string]*winSdkVersion)
	for _, path := range paths {
		for _, root := range roots {
			k, err := registry.OpenKey(root, path, registry.QUERY_VALUE|registry.ENUMERATE_SUB_KEYS)
			if err == nil {
				defer k.Close()
				subs, err := k.ReadSubKeyNames(-1)
				if err == nil {
					for _, sub := range subs {
						kSub, err := registry.OpenKey(root, path+"\\"+sub, registry.QUERY_VALUE|registry.ENUMERATE_SUB_KEYS)
						defer kSub.Close()
						if err == nil {
							installFolder, _, err := kSub.GetStringValue("InstallationFolder")
							if err == nil {
								productVersion, _, err := kSub.GetStringValue("ProductVersion")
								if err == nil {
									sdkDirs[sub] = &winSdkVersion{path: installFolder, version: productVersion}
								}
							}
						}
					}
				}
			}
		}
	}
	largestVersion := ""
	for version, _ := range sdkDirs {
		if largestVersion == "" {
			largestVersion = version
		} else {
			if compareVersion(largestVersion, version) == 1 {
				largestVersion = version
			}
		}
	}
	if largestVersion == "" {
		return nil
	}
	return sdkDirs[largestVersion]
}

func tryFindBaseDir(command string) string {
	dir := filepath.Dir(command)
	for dir != "." && filepath.Dir(dir) != dir {
		files, err := ioutil.ReadDir(dir)
		if err != nil {
			return ""
		}
		needed := []string{"lib", "include", "atlmfc"}
		matches := 0
		for _, file := range files {
			if file.IsDir() {
				for _, msvcDir := range needed {
					if msvcDir == strings.ToLower(file.Name()) {
						matches += 1
					}
				}
			}
		}
		if matches == len(needed) {
			return dir
		}
		dir = filepath.Dir(dir)
	}
	return ""
}

func (msvc *MSVCEnvironment) LocateIncludes(arch string) ([]string, error) {
	if len(msvc.includes) > 1 {
		return msvc.includes, nil
	}
	var includes []string
	winSdk := msvc.windowsSdk
	if winSdk != nil {
		var versionBasePath string
		files, err := filepath.Glob(fmt.Sprintf("%sInclude\\%s*", winSdk.path, winSdk.version))
		if err == nil {
			for _, file := range files {
				if _, err := os.Stat(file + "\\um\\Windows.h"); err == nil {
					versionBasePath = file
				}
			}
		}
		includePaths := []string{"shared", "um", "winrt", "ucrt"}
		for _, includePath := range includePaths {
			path := fmt.Sprintf("%s\\%s", versionBasePath, includePath)
			if _, err := os.Stat(path); err == nil {
				includes = append(includes, path)
			} else {
				path = fmt.Sprintf("%sInclude\\%s", winSdk.path, includePath)
				if _, err := os.Stat(path); err == nil {
					includes = append(includes, path)
				}
			}
		}
	}
	includes = append(includes, fmt.Sprintf("%s\\ATLMFC\\include", msvc.rootPath))
	includes = append(includes, fmt.Sprintf("%s\\include", msvc.rootPath))
	msvc.includes = includes
	return includes, nil
}

func (msvc *MSVCEnvironment) LocateLibs(arch string) ([]string, error) {
	if len(msvc.libs) > 1 {
		return msvc.libs, nil
	}
	var libs []string
	winSdk := msvc.windowsSdk
	if winSdk != nil {
		var versionBasePath string
		files, err := filepath.Glob(fmt.Sprintf("%sLib\\%s*", winSdk.path, winSdk.version))
		if err == nil {
			for _, file := range files {
				if _, err := os.Stat(file + "\\um\\" + arch + "\\kernel32.Lib"); err == nil {
					versionBasePath = file
				}
			}
		}
		libPaths := []string{"ucrt", "um"}
		for _, libPath := range libPaths {
			path := fmt.Sprintf("%s\\%s\\%s", versionBasePath, libPath, arch)
			if _, err := os.Stat(path); err == nil {
				libs = append(libs, path)
			} else {
				path = fmt.Sprintf("%slib\\%s\\%s", winSdk.path, libPath, arch)
				if _, err := os.Stat(path); err == nil {
					libs = append(libs, path)
				}
			}
		}
	}
	libs = append(libs, fmt.Sprintf("%s\\ATLMFC\\lib\\%s", msvc.rootPath, arch))
	libs = append(libs, fmt.Sprintf("%s\\lib\\%s", msvc.rootPath, arch))
	msvc.libs = libs
	return libs, nil
}

func (msvc *MSVCEnvironment) LocateLibPaths(arch string) ([]string, error) {
	if len(msvc.libpath) > 1 {
		return msvc.libpath, nil
	}
	var libs []string
	libs = append(libs, fmt.Sprintf("%s\\ATLMFC\\lib\\%s", msvc.rootPath, arch))
	libs = append(libs, fmt.Sprintf("%s\\lib\\%s", msvc.rootPath, arch))
	msvc.libpath = libs
	return libs, nil
}

func (msvc *MSVCEnvironment) GetMSVCCommand(command string) string {
	command = strings.Trim(strings.TrimSpace(command), "\"")
	cmd, ok := msvc.commands[command]
	if !ok {
		if _, err := os.Stat(command); err != nil {
			msvc.commands[command] = msvc.executablePath + "\\" + command
			return msvc.executablePath + "\\" + command
		}
		msvc.commands[command] = command
		return command
	}
	return cmd
}

// Copyright 2017 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build !windows

package msvc

import (
	"fmt"
)

type MSVCEnvironment struct {
}

func FromCommand(command string) (*MSVCEnvironment, error) {
	return nil, fmt.Errorf("MSVC not available on non windows OSes")
}

func (msvc *MSVCEnvironment) LocateIncludes(arch string) ([]string, error) {
	return nil, fmt.Errorf("MSVC not available on non windows OSes")
}

func (msvc *MSVCEnvironment) LocateLibs(arch string) ([]string, error) {
	return nil, fmt.Errorf("MSVC not available on non windows OSes")
}

func (msvc *MSVCEnvironment) LocateLibPaths(arch string) ([]string, error) {
	return nil, fmt.Errorf("MSVC not available on non windows OSes")
}

func (msvc *MSVCEnvironment) GetMSVCCommand(command string) string {
	return nil, fmt.Errorf("MSVC not available on non windows OSes")
}

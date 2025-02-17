// Copyright 2018 The Operator-SDK Authors
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

package projutil

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	homedir "github.com/mitchellh/go-homedir"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

const (
	GoPathEnv  = "GOPATH"
	GoFlagsEnv = "GOFLAGS"
	GoModEnv   = "GO111MODULE"
	SrcDir     = "src"

	fsep            = string(filepath.Separator)
	mainFile        = "cmd" + fsep + "manager" + fsep + "main.go"
	buildDockerfile = "build" + fsep + "Dockerfile"
	rolesDir        = "roles"
	helmChartsDir   = "helm-charts"
	goModFile       = "go.mod"
	gopkgTOMLFile   = "Gopkg.toml"
)

// OperatorType - the type of operator
type OperatorType = string

const (
	// OperatorTypeGo - golang type of operator.
	OperatorTypeGo OperatorType = "go"
	// OperatorTypeAnsible - ansible type of operator.
	OperatorTypeAnsible OperatorType = "ansible"
	// OperatorTypeHelm - helm type of operator.
	OperatorTypeHelm OperatorType = "helm"
	// OperatorTypeUnknown - unknown type of operator.
	OperatorTypeUnknown OperatorType = "unknown"
)

type ErrUnknownOperatorType struct {
	Type string
}

func (e ErrUnknownOperatorType) Error() string {
	if e.Type == "" {
		return "unknown operator type"
	}
	return fmt.Sprintf(`unknown operator type "%v"`, e.Type)
}

type ErrUnknownInputOperatorType struct {
	Type string
}

func (e ErrUnknownInputOperatorType) Error() string {
	if e.Type == "" {
		return "unknown input operator type"
	}
	return fmt.Sprintf(`unknown input operator type "%v"`, e.Type)
}

type ErrUnknownOutputOperatorType struct {
	Type string
}

func (e ErrUnknownOutputOperatorType) Error() string {
	if e.Type == "" {
		return "unknown output operator type"
	}
	return fmt.Sprintf(`unknown output operator type "%v"`, e.Type)
}

type DepManagerType string

const (
	DepManagerGoMod DepManagerType = "modules"
	DepManagerDep   DepManagerType = "dep"
)

type ErrInvalidDepManager string

func (e ErrInvalidDepManager) Error() string {
	return fmt.Sprintf(`"%s" is not a valid dep manager; dep manager must be one of ["%v", "%v"]`, e, DepManagerDep, DepManagerGoMod)
}

var ErrNoDepManager = fmt.Errorf(`no valid dependency manager file found; dep manager must be one of ["%v", "%v"]`, DepManagerDep, DepManagerGoMod)

func GetDepManagerType() (DepManagerType, error) {
	if IsDepManagerDep() {
		return DepManagerDep, nil
	} else if IsDepManagerGoMod() {
		return DepManagerGoMod, nil
	}
	return "", ErrNoDepManager
}

func IsDepManagerDep() bool {
	_, err := os.Stat(gopkgTOMLFile)
	return err == nil || os.IsExist(err)
}

func IsDepManagerGoMod() bool {
	_, err := os.Stat(goModFile)
	return err == nil || os.IsExist(err)
}

// MustInProjectRoot checks if the current dir is the project root and returns
// the current repo's import path, ex github.com/example-inc/app-operator
func MustInProjectRoot() {
	// If the current directory has a "build/dockerfile", then it is safe to say
	// we are at the project root.
	if _, err := os.Stat(buildDockerfile); err != nil {
		if os.IsNotExist(err) {
			log.Fatalf("Must run command in project root dir: project structure requires %s", buildDockerfile)
		}
		log.Fatalf("Error while checking if current directory is the project root: (%v)", err)
	}
}

func CheckGoProjectCmd(cmd *cobra.Command) error {
	if IsOperatorGo() {
		return nil
	}
	return fmt.Errorf("'%s' can only be run for Go operators; %s does not exist.", cmd.CommandPath(), mainFile)
}

func MustGetwd() string {
	wd, err := os.Getwd()
	if err != nil {
		log.Fatalf("Failed to get working directory: (%v)", err)
	}
	return wd
}

func getHomeDir() (string, error) {
	hd, err := homedir.Dir()
	if err != nil {
		return "", err
	}
	return homedir.Expand(hd)
}

// CheckAndGetProjectGoPkg checks if this project's repository path is rooted
// under $GOPATH and returns the current directory's import path,
// e.g: "github.com/example-inc/app-operator"
func CheckAndGetProjectGoPkg() string {
	gopath := MustSetGopath(MustGetGopath())
	goSrc := filepath.Join(gopath, SrcDir)
	wd := MustGetwd()
	pathedPkg := strings.Replace(wd, goSrc, "", 1)
	// Make sure package only contains the "/" separator and no others, and
	// trim any leading/trailing "/".
	return strings.Trim(filepath.ToSlash(pathedPkg), "/")
}

// GetOperatorType returns type of operator is in cwd.
// This function should be called after verifying the user is in project root.
func GetOperatorType() OperatorType {
	switch {
	case IsOperatorGo():
		return OperatorTypeGo
	case IsOperatorAnsible():
		return OperatorTypeAnsible
	case IsOperatorHelm():
		return OperatorTypeHelm
	}
	return OperatorTypeUnknown
}

func IsOperatorGo() bool {
	_, err := os.Stat(mainFile)
	return err == nil
}

func IsOperatorAnsible() bool {
	stat, err := os.Stat(rolesDir)
	return err == nil && stat.IsDir()
}

func IsOperatorHelm() bool {
	stat, err := os.Stat(helmChartsDir)
	return err == nil && stat.IsDir()
}

// MustGetGopath gets GOPATH and ensures it is set and non-empty. If GOPATH
// is not set or empty, MustGetGopath exits.
func MustGetGopath() string {
	gopath, ok := os.LookupEnv(GoPathEnv)
	if !ok || len(gopath) == 0 {
		log.Fatal("GOPATH env not set")
	}
	return gopath
}

// MustSetGopath sets GOPATH=currentGopath after processing a path list,
// if any, then returns the set path. If GOPATH cannot be set, MustSetGopath
// exits.
func MustSetGopath(currentGopath string) string {
	var (
		newGopath   string
		cwdInGopath bool
		wd          = MustGetwd()
	)
	for _, newGopath = range strings.Split(currentGopath, ":") {
		if strings.HasPrefix(filepath.Dir(wd), newGopath) {
			cwdInGopath = true
			break
		}
	}
	if !cwdInGopath {
		log.Fatalf("Project not in $GOPATH")
	}
	if err := os.Setenv(GoPathEnv, newGopath); err != nil {
		log.Fatal(err)
	}
	return newGopath
}

var flagRe = regexp.MustCompile("(.* )?-v(.* )?")

// SetGoVerbose sets GOFLAGS="${GOFLAGS} -v" if GOFLAGS does not
// already contain "-v" to make "go" command output verbose.
func SetGoVerbose() error {
	gf, ok := os.LookupEnv(GoFlagsEnv)
	if !ok || len(gf) == 0 {
		return os.Setenv(GoFlagsEnv, "-v")
	}
	if !flagRe.MatchString(gf) {
		return os.Setenv(GoFlagsEnv, gf+" -v")
	}
	return nil
}

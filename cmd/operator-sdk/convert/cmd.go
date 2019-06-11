// Copyright 2018 Shishir Jakati
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
package convert

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/operator-framework/operator-sdk/internal/pkg/scaffold"
	"github.com/operator-framework/operator-sdk/internal/pkg/scaffold/ansible"
	"github.com/operator-framework/operator-sdk/internal/pkg/scaffold/input"
	"github.com/operator-framework/operator-sdk/internal/util/projutil"

	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func NewCmd() *cobra.Command {
	newCmd := &cobra.Command{
		Use:   "convert <project-name>",
		Short: "Converts an existing operator to an operator of a different kind",
		Long: `The convert command ports an existing operator application and generates an accurate directory layout based on the source operator. project-name should match the project name in the source operator.
        The name of the new operator is the same.
        `,
		RunE: convertFuncPlaceholder,
	}

	newCmd.Flags().StringVar(&inputOperatorType, "input-type", "", "Type of existing operator (choices: \"ansible\" or \"helm\"")
	newCmd.Flags().StringVar(&outputOperatorType, "output-type", "", "Type of desired operator (choices: \"ansible\" or \"helm\"")
	newCmd.Flags().BoolVar(&generatePlaybook, "generate-playbook", false, "Generate a playbook skeleton. (Only used for --outputType ansible)")
	newCmd.Flags().StringVar(&helmChartRef, "helm-chart", "", "Initialize helm operator with existing helm chart (<URL>, <repo>/<name>, or local path)")
	newCmd.Flags().StringVar(&helmChartVersion, "helm-chart-version", "", "Specific version of the helm chart (default is latest version)")
	newCmd.Flags().StringVar(&helmChartRepo, "helm-chart-repo", "", "Chart repository URL for the requested helm chart")

	return newCmd
}

func convertFuncPlaceholder(cmd *cobra.Command, args []string) error {

	fmt.Printf("Conversion Function Placeholder\n")
	return nil
}

var (
	inputOperatorType  string
	inputOperatorPath  string
	outputOperatorType string
	projectName        string
	generatePlaybook   bool
	helmChartRef       string
	helmChartVersion   string
	helmChartRepo      string
)

func convertFunc(cmd *cobra.Command, args []string) error {
	if err := parse(cmd, args); err != nil {
		return err
	}
	inputMustExist()
	if err := verifyFlags(); err != nil {
		return err
	}

	log.Infof("Converting Existing Operator of Type %s and Creating New Opeator %s of Type %s", strings.Title(inputOperatorType), strings.Title(projectName), strings.Title(outputOperatorType))

	switch inputOperatorType {
	case projutil.OperatorTypeHelm:
		r, err := getResourceFromHelmChart()
		if err != nil {
			return err
		}

		if outputOperatorType == projutil.OperatorTypeAnsible {
			if err := doAnsibleScaffoldWithResource(r); err != nil {
				return fmt.Errorf("Converting to Ansible Failed")
			}
		}
	case projutil.OperatorTypeAnsible:
		// TODO: not implemented yet
	default:
		// TODO: not implemented yet
	}

	log.Info("Project Conversion Complete")
	return nil
}

func parse(cmd *cobra.Command, args []string) error {
	// checks that all program variables are being inputted using flags
	if len(args) > 1 {
		return fmt.Errorf("command requires exactly one argument; please use flags for options")
	}
	projectName = args[0]
	if len(projectName) == 0 {
		return fmt.Errorf("Parse Error Placeholder")
	}
	return nil
}

func verifyFlags() error {
	if inputOperatorType != projutil.OperatorTypeAnsible && inputOperatorType != projutil.OperatorTypeHelm {
		return errors.Wrap(projutil.ErrUnknownInputOperatorType{Type: inputOperatorType}, "--input-operator-type can only be of type `helm` or `ansible`")
	}
	if outputOperatorType != projutil.OperatorTypeAnsible && outputOperatorType != projutil.OperatorTypeHelm {
		return errors.Wrap(projutil.ErrUnknownOutputOperatorType{Type: outputOperatorType}, "--output-operator-type can only be ot type `helm` or `ansible`")
	}
	if outputOperatorType != projutil.OperatorTypeAnsible && generatePlaybook {
		return fmt.Errorf("value of --generate-playbook can only be used with --output-type `ansible`")
	}
	if inputOperatorType == outputOperatorType {
		return fmt.Errorf("value of --input-type must be different than --output-type")
	}

	if len(helmChartRef) != 0 {
		if inputOperatorType != projutil.OperatorTypeHelm {
			return fmt.Errorf("--helm-chart can only be used with --input-type `helm`")
		}
	} else if len(helmChartRepo) != 0 {
		return fmt.Errorf("value of --helm-chart-repo can only be used with --input-type=helm and --helm-chart")
	} else if len(helmChartVersion) != 0 {
		return fmt.Errorf("value of --helm-chart-version can only be used with --type=helm and --helm-chart")
	}
	return nil
}

func inputMustExist() {
	// first check if the directory is a working directory; will fail if directory is not a working directory
	projutil.MustGetwd()
	// then check if the directory matches the file structure of the scaffold
	operatorType := projutil.GetOperatorType()
	if err := operatorType != inputOperatorType; err {
		log.Fatal("Input Operator Type Not Same as Implied Operator Type")
	}
}

// checks if the given output project exists under the current directory
// it exits with error when the output project exists
func outputMustBeNewProject() {
	newProjectName := outputOperatorType + "-" + projectName
	fp := filepath.Join(projutil.MustGetwd())
	stat, err := os.Stat(fp)
	if err != nil && os.IsNotExist(err) {
		return
	}
	if err != nil {
		log.Fatalf("Failed to determine if project (%v) exists", newProjectName)
	}
	if stat.IsDir() {
		log.Fatalf("Project (%v) in (%v) path already exists. Please use a different project name or delete the existing one", newProjectName, fp)
	}
}

func getResourceFromHelmChart() (*scaffold.Resource, error) {
	// placeholder
	return nil, nil
}

func doAnsibleScaffoldWithResource(r *scaffold.Resource) error {
	cfg := &input.Config{
		AbsProjectPath: filepath.Join(projutil.MustGetwd(), projectName),
		ProjectName:    projectName,
	}

	roleFiles := ansible.RolesFiles{Resource: *r}
	roleTemplates := ansible.RolesTemplates{Resource: *r}

	s := &scaffold.Scaffold{}
	err := s.Execute(cfg,
		&scaffold.ServiceAccount{},
		&scaffold.Role{},
		&scaffold.RoleBinding{},
		&scaffold.CRD{Resource: r},
		&scaffold.CR{Resource: r},
		&ansible.BuildDockerfile{GeneratePlaybook: generatePlaybook},
		&ansible.RolesReadme{Resource: *r},
		&ansible.RolesMetaMain{Resource: *r},
		&roleFiles,
		&roleTemplates,
		&ansible.RolesVarsMain{Resource: *r},
		&ansible.MoleculeTestLocalPlaybook{Resource: *r},
		&ansible.RolesDefaultsMain{Resource: *r},
		&ansible.RolesTasksMain{Resource: *r},
		&ansible.MoleculeDefaultMolecule{},
		&ansible.BuildTestFrameworkDockerfile{},
		&ansible.MoleculeTestClusterMolecule{},
		&ansible.MoleculeDefaultPrepare{},
		&ansible.MoleculeDefaultPlaybook{
			GeneratePlaybook: generatePlaybook,
			Resource:         *r,
		},
		&ansible.BuildTestFrameworkAnsibleTestScript{},
		&ansible.MoleculeDefaultAsserts{},
		&ansible.MoleculeTestClusterPlaybook{Resource: *r},
		&ansible.RolesHandlersMain{Resource: *r},
		&ansible.Watches{
			GeneratePlaybook: generatePlaybook,
			Resource:         *r,
		},
		&ansible.DeployOperator{},
		&ansible.Travis{},
		&ansible.MoleculeTestLocalMolecule{},
		&ansible.MoleculeTestLocalPrepare{Resource: *r},
	)
	if err != nil {
		return err
	}

	return nil

}

func execProjCmd(cmd string, args ...string) error {
	dc := exec.Command(cmd, args...)
	dc.Dir = filepath.Join(projutil.MustGetwd(), projectName)
	return projutil.ExecCmd(dc)
}

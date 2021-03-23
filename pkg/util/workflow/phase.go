/*
Copyright 2018 The Kubernetes Authors.

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

package workflow

import (
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

// Phase provides an implementation of a workflow phase that allows
// creation of new phases by simply instantiating a variable of this type.
type Phase struct { //工作流一个阶段
	// name of the phase.
	// Phase name should be unique among peer phases (phases belonging to
	// the same workflow or phases belonging to the same parent phase).
	Name string

	// Aliases returns the aliases for the phase.
	Aliases []string

	// Short description of the phase.
	Short string

	// Long returns the long description of the phase.
	Long string

	// Example returns the example for the phase.
	Example string

	// Hidden define if the phase should be hidden in the workflow help.
	// e.g. PrintFilesIfDryRunning phase in the kubeadm init workflow is candidate for being hidden to the users
	Hidden bool //是否在工作流中隐藏该阶段

	// Phases defines a nested, ordered sequence of phases.
	Phases []Phase //定义一个嵌套有序的阶段序列

	// RunAllSiblings allows to assign to a phase the responsibility to
	// run all the sibling phases //允许把本阶段分配给其他的同级阶段
	// Nb. phase marked as RunAllSiblings can not have Run functions //标记为RunAllSiblings的阶段不能有Run函数
	RunAllSiblings bool //跳过本阶段

	// Run defines a function implementing the phase action.
	// It is recommended to implent type assertion, e.g. using golang type switch,
	// for validating the RunData type.
	Run func(data RunData) error //这个阶段的Running函数

	// RunIf define a function that implements a condition that should be checked
	// before executing the phase action.
	// If this function return nil, the phase action is always executed.
	RunIf func(data RunData) (bool, error) //这个阶段执行前的check函数

	// InheritFlags defines the list of flags that the cobra command generated for this phase should Inherit
	// from local flags defined in the parent command / or additional flags defined in the phase runner.
	// If the values is not set or empty, no flags will be assigned to the command
	// Nb. global flags are automatically inherited by nested cobra command
	InheritFlags []string //集继承的Flags的列表

	// LocalFlags defines the list of flags that should be assigned to the cobra command generated
	// for this phase.
	// Nb. if two or phases have the same local flags, please consider using local flags in the parent command
	// or additional flags defined in the phase runner.
	LocalFlags *pflag.FlagSet //这个阶段的Flags的列表

	// ArgsValidator defines the positional arg function to be used for validating args for this phase
	// If not set a phase will adopt the args of the top level command.
	ArgsValidator cobra.PositionalArgs //用于验证此阶段参数的函数
}

// AppendPhase adds the given phase to the nested, ordered sequence of phases.
func (t *Phase) AppendPhase(phase Phase) {
	t.Phases = append(t.Phases, phase)
}

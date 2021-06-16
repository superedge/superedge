/*
Copyright 2020 The SuperEdge Authors.

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

package module

type Module interface {
	Name() string
	Start(mode string)
	CleanUp()
}

var Modules map[string]Module

func Register(m Module) {
	Modules[m.Name()] = m
}

func InitModules(mode string) {
	Modules = make(map[string]Module)
}

func GetModules() map[string]Module {
	return Modules
}

#!/bin/bash

#Copyright 2022 The SuperEdge Authors.
#
#Licensed under the Apache License, Version 2.0 (the "License");
#you may not use this file except in compliance with the License.
#You may obtain a copy of the License at
#
#    http://www.apache.org/licenses/LICENSE-2.0
#
#Unless required by applicable law or agreed to in writing, software
#distributed under the License is distributed on an "AS IS" BASIS,
#WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
#See the License for the specific language governing permissions and
#limitations under the License.

set -o errexit
set -o nounset

GO_VERSION=$(go version)

if [[ -z $(echo "${GO_VERSION[2]}" | grep -E 'go1.14|go1.15|go1.16|go1.17') ]]; then
  echo "Unknown go version '${GO_VERSION[2]}', skipping gofmt."
  exit 1
fi

find_files() {
  find . -not \( \
      \( \
        -wholename './output' \
        -o -wholename './_output' \
        -o -wholename './docs' \
        -o -wholename './hack' \
        -o -wholename './.git' \
        -o -wholename '*/third_party/*' \
        -o -wholename '*/Godeps/*' \
        -o -wholename '*/examples/*' \
        -o -wholename '*/sample/*' \
        -o -wholename '*/build/*' \
        -o -wholename '*/CHANGELOG/*' \
      \) -prune \
    \) -name '*.go'
}

GOFMT="gofmt -s"
bad_files=$(find_files | xargs $GOFMT -l)
if [[ -n "${bad_files}" ]]; then
  echo "!!! '$GOFMT' needs to be run on the following files: "
  echo "${bad_files}"
  exit 1
fi

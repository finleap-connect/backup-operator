#!/bin/bash
# Copyright 2022 BackupOperator Authors
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#      http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

set -e

if [[ "$GITHUB_REF" =~ refs/tags ]]; then
    VERSION=${GITHUB_REF##*/}
elif [ ! -z "$GITHUB_RUN_ID" ]; then
    VERSION="0.0.0-$(echo $GITHUB_SHA | cut -c 1-6)"
else
    VERSION="local-$(git rev-parse HEAD)"
fi

echo $VERSION

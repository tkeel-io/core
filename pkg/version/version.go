/*
 * Copyright (C) 2019 Yunify, Inc.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this work except in compliance with the License.
 * You may obtain a copy of the License in the LICENSE file, or at:
 *
 * http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package version

import (
	"fmt"
	"runtime"
)

//GitCommit The git commit that was compiled. This will be filled in by the compiler.
var GitCommit string

//Version The main version number that is being run at the moment.
const Version = "0.1.1"

//BuildDate The build datetime at the moment.
var BuildDate = ""

//GoVersion The go compiler version.
var GoVersion = runtime.Version()

//OsArch The system info.
var OsArch = fmt.Sprintf("%s %s", runtime.GOOS, runtime.GOARCH)

func Show() {
	fmt.Println("Build Date:", BuildDate)
	fmt.Println("Git Commit:", GitCommit)
	fmt.Println("Version:", Version)
	fmt.Println("Go Version:", GoVersion)
	fmt.Println("OS / Arch:", OsArch)
}

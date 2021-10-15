/*
Copyright © 2021 NAME HERE <EMAIL ADDRESS>

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
package cmd

import (
	"fmt"
	"os"
	"runtime"

	"github.com/spf13/cobra"
)

// Version The main version number that is being run at the moment.
const Version = "0.1.1"

var (
	// GitCommit The git commit that was compiled. This will be filled in by the compiler.
	GitCommit string

	// GoVersion The go compiler version.
	GoVersion = runtime.Version()

	// OSArch The system info.
	OSArch = runtime.GOOS + " " + runtime.GOARCH
)

// rootCmd represents the base command when called without any subcommands.
var rootCmd = &cobra.Command{
	Use:   "core",
	Short: "core is the database for digital twins.",
	// Uncomment the following line if your bare application
	// has an action associated with it:
	//	Run: func(cmd *cobra.Command, args []string) { },
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	rootCmd.Version = Version
	setVersion()
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.

	// Cobra also supports local flags, which will only run
	// when this action is called directly.
	// rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle").
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file path")
}

func setVersion() {
	template := fmt.Sprintln("Core Version", Version)
	template += fmt.Sprintln("Git Commit:", GitCommit)
	template += fmt.Sprintln("Go Version", GoVersion)
	template += fmt.Sprintln("OS / Arch", OSArch)
	rootCmd.SetVersionTemplate(template)
}

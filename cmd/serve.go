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
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/tkeel-io/core/pkg/bootstrap"
	"github.com/tkeel-io/core/pkg/config"

	"github.com/spf13/cobra"
	"github.com/tkeel-io/core/pkg/logger"
)

var cfgFile string
var log = logger.NewLogger("kcore.commands")

// serveCmd represents the serve command.
var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "core serve",
	Example: `
# Run with default configurations:
  core serve
# Run with configuration:
  core serve --config configfile
  `,
	Args: cobra.MinimumNArgs(0),
	PreRun: func(cmd *cobra.Command, args []string) {
		// viper.BindPFlag("config", cmd.Flags().Lookup("config"))
	},
	Run: func(cmd *cobra.Command, args []string) {
		log.Info("start kcore...")

		// configure logger default.
		logger.SetDefaultAppID(config.GetConfig().Server.AppID)
		logger.SetDefaultLevel(config.GetConfig().Logger.Level)
		logger.SetDefaultJSONOutput(config.GetConfig().Logger.OutputJSON)

		stopCh := make(chan struct{}, 1)
		ctx, cancel := context.WithCancel(context.Background())

		server := bootstrap.NewServer(ctx, config.GetConfig())
		go func() {
			if err := server.Run(); nil != err {
				log.Errorf("start service failed, error: %s", err.Error())
				stopCh <- struct{}{}
			}
		}()

		signalChan := make(chan os.Signal, 1)
		signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)

		select {
		case <-stopCh:
			cancel()
			server.Close()
			os.Exit(0)
		case <-signalChan:

			log.Infof("KCore Exited.")

			cancel()
			server.Close()
			os.Exit(0)
		}
	},
}

func init() {
	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// serveCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// serveCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")

	rootCmd.AddCommand(serveCmd)
	cobra.OnInitialize(func() {
		config.InitConfig(cfgFile)
	})
}

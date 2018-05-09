// Copyright © 2018 NAME HERE <EMAIL ADDRESS>
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

package cmd

import (
	"github.com/overdrive3000/justforfunc32/contenedor/level"
	"github.com/spf13/cobra"
)

var container level.Container

// runCmd represents the run command
var runCmd = &cobra.Command{
	Use:   level.USAGE,
	Short: level.SHORT,
	Long:  level.LONG,
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		err := level.Run(args, container)
		return err
	},
	SilenceErrors: true,
	SilenceUsage:  true,
}

func init() {
	rootCmd.AddCommand(runCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// runCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// runCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
	runCmd.Flags().StringSliceVarP(&container.Env, "env", "e", nil, "Environment variables (PORT=80)")
	runCmd.Flags().StringVarP(&container.ImageName, "image-name", "i", "ubuntu", "Image name")
	runCmd.Flags().StringVarP(&container.ImageDir, "image-dir", "", "/workshop/images", "Images directory")
	runCmd.Flags().StringVarP(&container.ContainerDir, "container-dir", "", "/workshop/containers", "Containers directory")
}

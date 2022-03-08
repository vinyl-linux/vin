/*
Copyright © 2021 James Condron <james@zero-internet.org.uk>
All rights reserved.

Redistribution and use in source and binary forms, with or without
modification, are permitted provided that the following conditions are met:

1. Redistributions of source code must retain the above copyright notice,
   this list of conditions and the following disclaimer.

2. Redistributions in binary form must reproduce the above copyright notice,
   this list of conditions and the following disclaimer in the documentation
   and/or other materials provided with the distribution.

3. Neither the name of the copyright holder nor the names of its contributors
   may be used to endorse or promote products derived from this software
   without specific prior written permission.

THIS SOFTWARE IS PROVIDED BY THE COPYRIGHT HOLDERS AND CONTRIBUTORS "AS IS"
AND ANY EXPRESS OR IMPLIED WARRANTIES, INCLUDING, BUT NOT LIMITED TO, THE
IMPLIED WARRANTIES OF MERCHANTABILITY AND FITNESS FOR A PARTICULAR PURPOSE
ARE DISCLAIMED. IN NO EVENT SHALL THE COPYRIGHT HOLDER OR CONTRIBUTORS BE
LIABLE FOR ANY DIRECT, INDIRECT, INCIDENTAL, SPECIAL, EXEMPLARY, OR
CONSEQUENTIAL DAMAGES (INCLUDING, BUT NOT LIMITED TO, PROCUREMENT OF
SUBSTITUTE GOODS OR SERVICES; LOSS OF USE, DATA, OR PROFITS; OR BUSINESS
INTERRUPTION) HOWEVER CAUSED AND ON ANY THEORY OF LIABILITY, WHETHER IN
CONTRACT, STRICT LIABILITY, OR TORT (INCLUDING NEGLIGENCE OR OTHERWISE)
ARISING IN ANY WAY OUT OF THE USE OF THIS SOFTWARE, EVEN IF ADVISED OF THE
POSSIBILITY OF SUCH DAMAGE.
*/
package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// Flag values
var (
	version string
	force   bool
)

// installCmd represents the install command
var installCmd = &cobra.Command{
	Use:   "install [package name] | [package name] [package name]",
	Short: "install package(s)",
	Long:  `install package(s)`,
	Args: func(cmd *cobra.Command, args []string) error {
		argCount := len(args)

		if argCount == 0 {
			cmd.Usage()

			return fmt.Errorf("missing package(s)")
		}

		if argCount > 1 && version != "" {
			cmd.Usage()

			return fmt.Errorf("setting version with multiple packages makes no sense")
		}

		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) (err error) {
		client, err := newClient(socketAddr)

		err = errWrap("connecting to vin", err)
		if err != nil {
			return err
		}

		for _, pkg := range args {
			err = client.install(pkg, version, force)
			if err != nil {
				break
			}
		}

		return errWrap("installation", err)
	},
}

func errWrap(msg string, err error) error {
	if err != nil {
		err = fmt.Errorf("%s: %w", msg, err)
	}

	return err
}

func init() {
	rootCmd.AddCommand(installCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// installCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// installCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
	installCmd.Flags().StringVarP(&version, "version", "v", "latest", `Version constraint to install. This command allows strict versions, such as "1.2.3", or loose versions such as ">=1.20, <1.3.5"`)
	installCmd.Flags().BoolVarP(&force, "force", "f", false, "Force installation, even when targets are marked as installed")
}

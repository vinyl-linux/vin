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
	"math"
	"runtime"
	"time"

	"github.com/spf13/cobra"
	"github.com/vinyl-linux/vin/config"
)

// adviseCmd represents the advise command
var adviseCmd = &cobra.Command{
	Use:   "advise",
	Short: "generate a vin config.toml for this machine",
	Long:  "generate a vin config.toml for this machine",
	Run: func(cmd *cobra.Command, args []string) {
		c := config.Config{}

		c.ConfigureFlags = "--prefix=/ --enable-openssl"
		c.MakeOpts = fmt.Sprintf("-j%d", int(math.Max(float64(runtime.NumCPU()-1), 1.0)))

		c.CFlags = "-D_FORTIFY_SOURCE=2 -fasynchronous-unwind-tables -fexceptions -fpie -Wl,-pie -fpic -shared -fstack-clash-protection -fstack-protector-strong -mcet -fcf-protection -O2 -pipe -Wall -Werror=format-security -Werror=implicit-function-declaration -Wl,-z,defs"
		c.CXXFlags = c.CFlags

		fmt.Fprintln(cmd.OutOrStdout(), "# vin config file")
		fmt.Fprintln(cmd.OutOrStdout(), "#")
		fmt.Fprintf(cmd.OutOrStdout(), "# generated at %s\n", time.Now())
		fmt.Fprintln(cmd.OutOrStdout())
		fmt.Fprint(cmd.OutOrStdout(), c.String())
	},
}

func init() {
	rootCmd.AddCommand(adviseCmd)
}

// Copyright 2020 The Cockroach Authors.
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package main

import (
	"io/ioutil"
	"log"
	"os"
	"os/exec"

	"github.com/cockroachdb/errors"
	"github.com/spf13/cobra"
)

var devCmd = &cobra.Command{
	Use:     "dev [command] (flags)",
	Short:   "Dev is the general-purpose dev tool for folks working on cockroachdb/cockroach.",
	Version: "v0.0",
	Long: `
Dev is the general-purpose dev tool for folks working cockroachdb/cockroach. It
lets engineers do a few things:

- build various binaries (cockroach, optgen, ...)
- run arbitrary tests (unit tests, logic tests, ...)
- run tests under arbitrary configurations (under stress, using race builds, ...)
- generate code (bazel files, protobufs, ...)

...and much more.

(PS: Almost none of the above is implemented yet, haha.)
`,
	// Disable automatic printing of usage information whenever an error
	// occurs. We presume that most errors will not the result of bad command
	// invocation; they'll be due to legitimate build/test errors. Printing out
	// the usage information in these cases obscures the real cause of the
	// error. Commands should manually print usage information when the error
	// is, in fact, a result of a bad invocation, e.g. too many arguments.
	SilenceUsage: true,
	// Disable automatic printing of the error. We want to also print
	// details and hints, which cobra does not do for us. Instead
	// we do the printing in the command implementation.
	SilenceErrors: true,
}

var (
	remoteCacheAddr string
	debugLogger     = log.New(ioutil.Discard, "DEBUG: ", 0)
)

func init() {
	log.SetFlags(0)
	log.SetPrefix("")

	cmds := []*cobra.Command{
		benchCmd,
		buildCmd,
		generateCmd,
		lintCmd,
		testCmd,
	}

	// Add all the shared flags.
	var debugVar bool
	for _, cmd := range cmds {
		cmd.Flags().BoolVar(&debugVar, "debug", false, "enable debug logging for dev itself")
		// This points to the grpc endpoint of a running `buchr/bazel-remote`
		// instance. We're tying ourselves to the one implementation, but that
		// seems fine for now. It seems mature, and has (very experimental)
		// support for the  Remote Asset API, which helps speed things up when
		// the cache sits across the network boundary.
		cmd.Flags().StringVar(&remoteCacheAddr, "remote-cache", "", "remote caching grpc endpoint to use")
	}
	for _, cmd := range cmds {
		cmd.PreRun = func(cmd *cobra.Command, args []string) {
			if debugVar {
				debugLogger.SetOutput(os.Stderr)
			}
		}
	}

	devCmd.AddCommand(cmds...)

	// Hide the `help` sub-command.
	devCmd.SetHelpCommand(&cobra.Command{
		Use:    "noop-help",
		Hidden: true,
	})
}

func runDev() error {
	_, err := exec.LookPath("bazel")
	if err != nil {
		return errors.New("bazel not found in $PATH")
	}

	if err := devCmd.Execute(); err != nil {
		return err
	}
	return nil
}

func main() {
	if err := runDev(); err != nil {
		log.Printf("ERROR: %v", err)
		os.Exit(1)
	}
}

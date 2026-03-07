package main

import (
	"fmt"
	"log"
	"os"
	"runtime/debug"
	"strings"

	"github.com/simcap/sfuzz"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "sfuzz",
	Short: "Simple fuzzer to harness a resilient JSON API",
	Run:   func(cmd *cobra.Command, args []string) {},
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		log.Fatal(err)
	}
}

var (
	fuzzFilenameFlag string
)

func init() {
	rootCmd.AddCommand(versionCmd, runCmd)

	runCmd.Flags().StringVarP(&fuzzFilenameFlag, "fuzzfile", "f", "", "Fuzz file containing request on each line")
}

var logger = sfuzz.NewConsoleLogger(os.Stdout)

var runCmd = &cobra.Command{
	Use:   "run",
	Short: "Launch a fuzzer run on given entries",
	RunE: func(cmd *cobra.Command, args []string) error {
		fuzzFile, err := os.Open(fuzzFilenameFlag)
		if err != nil {
			return err
		}
		requests, err := sfuzz.Parse(fuzzFile)
		if err != nil {
			return err
		}

		var targets int
		for _, r := range requests {
			targets = targets + len(r.ParsedKeywords)
		}
		logger.Info(fmt.Sprintf("%d requests parsed; %d targets to be fuzzed", len(requests), targets))

		runner := sfuzz.NewRunner(
			sfuzz.WithLogger(logger),
			sfuzz.WithSelector(func(sfuzz.FuzzKeyword) sfuzz.Generator {
				return sfuzz.CounterGenerator(5)
			}),
		)

		runner.Run(cmd.Context(), requests)
		return nil
	},
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print version and build info",
	RunE: func(cmd *cobra.Command, args []string) error {
		info, ok := debug.ReadBuildInfo()
		if !ok {
			fmt.Println("cannot read build info")
			return nil
		}
		var out strings.Builder
		out.WriteString(info.Main.Version)
		for _, setting := range info.Settings {
			if setting.Key == "vcs.revision" {
				out.WriteString(fmt.Sprintf(", sha: %s", setting.Value))
			}
		}
		out.WriteString(fmt.Sprintf(", built with: %s", info.GoVersion))
		fmt.Println(out.String())
		return nil
	},
}

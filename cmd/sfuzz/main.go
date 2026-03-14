package main

import (
	"errors"
	"fmt"
	"log"
	"os"
	"runtime/debug"
	"strings"

	"github.com/simcap/sfuzz"
	"github.com/simcap/sfuzz/spec"
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
	htmlOutputFlag           bool
	maxRequestsPerSecondFlag uint
)

func init() {
	rootCmd.AddCommand(versionCmd, runCmd, genCmd)

	runCmd.Flags().BoolVar(&htmlOutputFlag, "html", false, "Output as single HTML page")
	runCmd.Flags().UintVar(&maxRequestsPerSecondFlag, "rps", 100, "Max requests sent per second allowed")
}

var logger = sfuzz.NewConsoleLogger(os.Stdout)

var runCmd = &cobra.Command{
	Use:   "run",
	Short: "Launch a fuzzer run on given requests",
	PreRunE: func(cmd *cobra.Command, args []string) error {
		if len(args) < 1 {
			return errors.New("required at least one argument")
		}
		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		fuzzFile, err := os.Open(args[0])
		if err != nil {
			return err
		}
		requests, err := sfuzz.Parse(fuzzFile)
		if err != nil {
			return err
		}

		report := sfuzz.NewReport(requests)

		runner := sfuzz.NewRunner(
			report,
			sfuzz.WithLogger(logger),
			sfuzz.WithMaxRPS(maxRequestsPerSecondFlag),
		)
		runner.Run(cmd.Context())

		output, err := sfuzz.NewHTMLOutput(report)
		if err != nil {
			return err
		}
		f, err := os.Create("sfuzz-report.html")
		if err != nil {
			return err
		}
		defer f.Close()

		if err := output.Write(f); err != nil {
			return err
		}

		logger.Info(fmt.Sprintf("report generated at %s", f.Name()))
		return nil
	},
}

var genCmd = &cobra.Command{
	Use:     "gen",
	Aliases: []string{"g", "generate"},
	Short:   "Generate fuzz file from a Open API specification (>= 3.0)",
	PreRunE: func(cmd *cobra.Command, args []string) error {
		if len(args) < 1 {
			return errors.New("required at least one argument")
		}
		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		f, err := os.Open(args[0])
		if err != nil {
			return err
		}
		api, err := spec.NewOAPISpec(f)
		if err != nil {
			return err
		}
		return api.GenerateFuzzFile(os.Stdout)
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

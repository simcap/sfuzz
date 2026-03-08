package main

import (
	"fmt"
	"log"
	"os"
	"runtime/debug"
	"strings"
	"time"

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
	htmlOutputFlag   bool
)

func init() {
	rootCmd.AddCommand(versionCmd, runCmd)

	runCmd.Flags().StringVarP(&fuzzFilenameFlag, "fuzzfile", "f", "", "Fuzz file containing request on each line")
	runCmd.Flags().BoolVar(&htmlOutputFlag, "html", false, "Output as single HTML page")
}

var logger = sfuzz.NewConsoleLogger(os.Stdout)

var runCmd = &cobra.Command{
	Use:   "run",
	Short: "Launch a fuzzer run on given requests",
	RunE: func(cmd *cobra.Command, args []string) error {
		fuzzFile, err := os.Open(fuzzFilenameFlag)
		if err != nil {
			return err
		}
		requests, err := sfuzz.Parse(fuzzFile)
		if err != nil {
			return err
		}

		report := sfuzz.NewReport(requests)

		runner := sfuzz.NewRunner(report, sfuzz.WithLogger(logger))
		runner.Run(cmd.Context())

		output, err := sfuzz.NewHTMLOutput(report)
		if err != nil {
			return err
		}
		f, err := os.Create(fmt.Sprintf("sfuzz-report-%d.html", time.Now().Unix()))
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

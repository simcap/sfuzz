package main

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
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
	htmlOutputFlag, showProgressFlag              bool
	fuzzFilepathFlag, genericWordlistFilepathFlag string
	openAPIURLFlag, openAPIFilepathFlag           string
	numWordlistFilepathFlag                       string
	uidWordlistFilepathFlag                       string
	requestURLFlag, serverHostFlag                string
	headersFilepathFlag                           string
	verbFlag                                      string
	maxRequestsPerSecondFlag                      uint
)

func init() {
	rootCmd.AddCommand(versionCmd, fuzzCmd, genCmd)

	genCmd.Flags().StringVarP(&openAPIFilepathFlag, "oapi-file", "f", "", "Filepath of OpenAPI specification")
	genCmd.Flags().StringVarP(&openAPIURLFlag, "oapi-url", "u", "", "URL endpoint for OpenAPI specification")
	genCmd.Flags().StringVarP(&serverHostFlag, "server", "s", "", "Server (i.e. host) of endpoint. Ex: http://localhost:9000")
	genCmd.MarkFlagsMutuallyExclusive("oapi-file", "oapi-url")
	genCmd.MarkFlagsOneRequired("oapi-file", "oapi-url")
	fuzzCmd.Flags().BoolVar(&htmlOutputFlag, "html", false, "Output as single HTML page")
	fuzzCmd.Flags().UintVar(&maxRequestsPerSecondFlag, "rps", 100, "Max requests sent per second allowed")
	fuzzCmd.Flags().StringVarP(&fuzzFilepathFlag, "fuzzfile", "f", "", "Filepath of fuzz file")
	fuzzCmd.Flags().StringVar(&requestURLFlag, "url", "", "Single request URL to fuzz")
	fuzzCmd.Flags().StringVar(&verbFlag, "method", "GET", "HTTP method for single URL to fuzz on")
	fuzzCmd.Flags().StringVar(&headersFilepathFlag, "headers", "", "File with an HTTP header on each line to add to sent requests.")
	fuzzCmd.Flags().StringVar(&genericWordlistFilepathFlag, "wordlist-str", "", "Wordlist to inject as fuzz values in FUZZSTR")
	fuzzCmd.Flags().StringVar(&numWordlistFilepathFlag, "wordlist-num", "", "Wordlist to inject as fuzz values in FUZZNUM")
	fuzzCmd.Flags().StringVar(&uidWordlistFilepathFlag, "wordlist-uid", "", "Wordlist to inject as fuzz values in FUZZUID")
	fuzzCmd.Flags().BoolVarP(&showProgressFlag, "progress", "p", false, "Show oneliner progress info instead of logs")
	fuzzCmd.MarkFlagsRequiredTogether("method", "url")
	fuzzCmd.MarkFlagsMutuallyExclusive("fuzzfile", "url")
}

var logger = sfuzz.NewConsoleLogger(os.Stdout)

var fuzzCmd = &cobra.Command{
	Use:   "fuzz",
	Short: "Launch a fuzz run on given fuzz-file or a single given request",
	RunE: func(cmd *cobra.Command, args []string) error {
		var requests []sfuzz.Request
		if fuzzFilepathFlag != "" {
			fuzzFile, err := os.Open(fuzzFilepathFlag)
			if err != nil {
				return err
			}
			requests, err = sfuzz.Parse(fuzzFile)
			if err != nil {
				return err
			}
		} else {
			parsed, err := sfuzz.Parse(strings.NewReader(requestURLFlag))
			if err != nil {
				return err
			}
			if len(parsed) != 1 {
				return fmt.Errorf("expected only one requests to parse, got %d", len(requests))
			}
			single, err := parsed[0].AutoGenerateKeywords()
			if err != nil {
				return fmt.Errorf("cannot autogenerate fuzz keywords: %w", err)
			}
			single.Verb = verbFlag
			requests = append(requests, single)
		}

		if len(requests) == 0 {
			return fmt.Errorf("no requests found to fuzz")
		}

		var options []sfuzz.Option
		if headersFilepathFlag != "" {
			headers, err := parseHeadersFile(headersFilepathFlag)
			if err != nil {
				return err
			}
			options = append(options, sfuzz.WithHeaders(headers))
		}

		option, err := sfuzz.WithWordlist(map[sfuzz.Kind]string{
			sfuzz.GenericString: genericWordlistFilepathFlag,
			sfuzz.Numeral:       numWordlistFilepathFlag,
			sfuzz.UniversalID:   uidWordlistFilepathFlag,
		})
		if err != nil {
			return err
		}
		options = append(options, option)
		options = append(options, sfuzz.WithLogger(logger))
		options = append(options, sfuzz.WithProgress(showProgressFlag))
		options = append(options, sfuzz.WithMaxRPS(maxRequestsPerSecondFlag))

		runner := sfuzz.NewRunner(requests, options...)
		runner.Run(cmd.Context())

		output, err := sfuzz.NewHTMLOutput(runner.Results())
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
	Short:   "Generate fuzz file from Open API specification (>= 3.0) file or URL",
	RunE: func(cmd *cobra.Command, args []string) (err error) {
		var content io.ReadCloser
		if filename := openAPIFilepathFlag; filename != "" {
			content, err = os.Open(filename)
			if err != nil {
				return err
			}
		} else if openAPIURLFlag != "" {
			resp, err := http.Get(openAPIURLFlag)
			if err != nil {
				return err
			}
			content = resp.Body
		}
		api, err := spec.NewOAPISpec(content, spec.WithServer(serverHostFlag))
		if err != nil {
			return err
		}
		defer content.Close()

		if len(api.Server()) == 0 {
			return errors.New("no server(s) in Open API spec, specify one with --server")
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

func parseHeadersFile(filename string) (http.Header, error) {
	f, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	header := make(http.Header)
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Text()
		parts := strings.SplitN(line, ":", 2)
		if len(parts) != 2 {
			continue
		}
		header.Set(parts[0], parts[1])
	}
	return header, scanner.Err()
}

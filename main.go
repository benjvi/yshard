package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

const OUTPUT_YAML_EXT = ".yml"

func main() {
	rootCmd := buildRootCmd()
	if err := rootCmd.Execute(); err != nil {
		// TODO: clean up all print and log statements
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func buildRootCmd() *cobra.Command {
	var rootCmd = &cobra.Command{
		Use:   "yshard",
		Short: "yshard is a tool for sharding multi-document YAML files.",
		Long: `
yshard is a tool for sharding multi-document YAML files.
		
It uses a jq "groupby" query to categorize documents based on the value of the specified path.
		
It reads YAML from stdin, and writes output to the specified directory

e.g.:
kubectl get po -o yaml | gojq --from-yaml --output-yaml ".items[]" | yshard -g ".metadata.namespace" -o pods-by-namespace

`,
		Run: run,
	}
	//rootCmd.PersistentFlags().StringP("config", "c", "", "set a custom config file")
	rootCmd.PersistentFlags().StringP("output-dir", "o", "", "directory to place the output files [REQUIRED]")
	rootCmd.MarkPersistentFlagRequired("output-dir")
	rootCmd.PersistentFlags().StringP("groupby-path", "g", "", "path to the element of the document to group by (in jq syntax) [REQUIRED]")
	rootCmd.MarkPersistentFlagRequired("groupby-path")
	return rootCmd
}

func run(cmd *cobra.Command, args []string) {
	// first retrieve all the associated flag values

	// TODO: not implemented
	/*configFile, err := cmd.Flags().GetString("config")
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}*/

	outputDir, err := cmd.Flags().GetString("output-dir")
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	} else if outputDir == "" {
		fmt.Println("--output-dir value cannot be empty")
	}

	groupbyPath, err := cmd.Flags().GetString("groupby-path")
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	} else if groupbyPath == "" {
		fmt.Println("--groupby-path value cannot be empty")
	}

	// exit if there's no data piped in
	ensureDataOnPipe()

	// read the data from stdin and convert from YAML to Go objects so Gojq can query it
	inputData := readYamlFromStdin()
	fmt.Printf("Read %d YAML docs from stdin\n", len(inputData))

	groupedData, ungroupedData := groupYamlDocsByPathValues(groupbyPath, inputData)

	outputGroupedDataToDir(groupedData, outputDir)
	outputUngroupedDataToDir(ungroupedData, outputDir)

	// TODO: check for orphaned files in the output (.yml files that we didn't write to)
	// we can't assume that the folder is empty to start with
	// so if we run yshard multiple times on changing input, the output can get polluted
}

func ensureDataOnPipe() {
	// Check if stdin is in non-blocking mode
	stat, err := os.Stdin.Stat()
	if err != nil {
		fmt.Println("Error getting stdin status:", err)
		os.Exit(1)
	}

	if (stat.Mode() & os.ModeCharDevice) != 0 {
		// TODO: if no data piped to program, a file has to be specified
		fmt.Println("No data on stdin, check that data is correctly piped as input")
		os.Exit(0)
	}
}

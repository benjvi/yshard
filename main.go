package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

const OUTPUT_YAML_EXT = ".yml"

// Logger is a generic interface for logging
type Logger interface {
	Info(msg string, fields ...zap.Field)
	Error(msg string, fields ...zap.Field)
	Fatal(msg string, fields ...zap.Field)
	Sync() error
}

func main() {
	// Create a logger that writes console-encoded logs to stderr
	logger := buildLogger()

	rootCmd := buildRootCmd(logger)
	if err := rootCmd.Execute(); err != nil {
		logger.Fatal(fmt.Sprintf("error in cobra root command argument parsing or execution: %q", err.Error()))
	}
}

func encodeTime(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
	enc.AppendString(t.Format("2006-01-02 15:04:05"))
}

func buildLogger() Logger {
	// Define a custom encoder function for timestamp

	// Create an EncoderConfig with custom encoder function
	cfg := zap.NewProductionEncoderConfig()
	cfg.EncodeTime = encodeTime
	cfg.ConsoleSeparator = " | "

	// Create a Logger with ConsoleEncoder and custom EncoderConfig
	logger := zap.New(zapcore.NewCore(
		zapcore.NewConsoleEncoder(cfg),
		os.Stderr,
		zap.DebugLevel, // minimum level that will be logged
	))
	return logger
}

func buildRootCmd(logger Logger) *cobra.Command {
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
	rootCmd.SetContext(context.WithValue(context.Background(), "logger", logger))
	return rootCmd
}

func run(cmd *cobra.Command, args []string) {
	logger := cmd.Context().Value("logger").(Logger)
	// first retrieve all the associated flag values

	// TODO: not implemented
	/*configFile, err := cmd.Flags().GetString("config")
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}*/

	outputDir, err := cmd.Flags().GetString("output-dir")
	if err != nil {
		logger.Fatal("required arg \"--output-dir\" not found")
	} else if outputDir == "" {
		logger.Fatal("--output-dir value cannot be empty")
	}

	groupbyPath, err := cmd.Flags().GetString("groupby-path")
	if err != nil {
		logger.Fatal("required arg \"--groupby-path\" not found")
	} else if groupbyPath == "" {
		logger.Fatal("--groupby-path value cannot be empty")
	}

	// exit if there's no data piped in
	err = ensureDataOnPipe()
	if err != nil {
		logger.Fatal(err.Error())
	}

	// read the data from stdin and convert from YAML to Go objects so Gojq can query it
	inputData := readYamlFromStdin(logger)
	logger.Info(fmt.Sprintf("Read %d YAML docs from stdin\n", len(inputData)))

	groupedData, ungroupedData := groupYamlDocsByPathValues(logger, groupbyPath, inputData)

	outputGroupedDataToDir(logger, groupedData, outputDir)
	outputUngroupedDataToDir(logger, ungroupedData, outputDir)

	// TODO: check for orphaned files in the output (.yml files that we didn't write to)
	// we can't assume that the folder is empty to start with
	// so if we run yshard multiple times on changing input, the output can get polluted
}

func ensureDataOnPipe() error {
	// Check if stdin is in non-blocking mode
	stat, err := os.Stdin.Stat()
	if err != nil {
		return fmt.Errorf("Error getting stdin status: %s", err.Error())
	}

	if (stat.Mode() & os.ModeCharDevice) != 0 {
		// TODO: if no data piped to program, a file has to be specified
		return fmt.Errorf("No data on stdin, check that data is correctly piped as input")
	}
	return nil
}

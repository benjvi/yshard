package main

import (
	"fmt"
	"log"
	"os"

	"github.com/itchyny/gojq"
	"github.com/spf13/cobra"
)

func main() {
	rootCmd := buildRootCmd()
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func buildRootCmd() *cobra.Command {
	var rootCmd = &cobra.Command{
		Use:   "yshard",
		Short: "yshard is a tool for processing JSON data",
		Long:  `yshard is a tool for processing JSON data.`,
		Run:   run,
	}
	rootCmd.PersistentFlags().StringP("config", "c", "", "Set a custom config file")
	rootCmd.PersistentFlags().StringP("output-dir", "o", "", "Directory to place the output files")
	rootCmd.MarkPersistentFlagRequired("output-dir")
	rootCmd.PersistentFlags().StringP("groupby-path", "g", "", "Path to the element of the document to group by (in jq syntax)")
	rootCmd.MarkPersistentFlagRequired("groupby-path")
	return rootCmd
}

func run(cmd *cobra.Command, args []string) {
	configFile, err := cmd.Flags().GetString("config")
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	outputDir, err := cmd.Flags().GetString("output-dir")
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	groupbyPath, err := cmd.Flags().GetString("groupby-path")
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	fmt.Printf("Arguments: %+v\n", cmd)

	fmt.Printf("configFile: %s, outputDir: %s, groupbyPath: %s\n", configFile, outputDir, groupbyPath)
	query, err := gojq.Parse(groupbyPath)
	if err != nil {
		log.Fatalln(err)
	}
	input := map[string]any{"foo": []any{1, 2, 3}}
	iter := query.Run(input) // or query.RunWithContext
	for {
		v, ok := iter.Next()
		if !ok {
			break
		}
		if err, ok := v.(error); ok {
			log.Fatalln(err)
		}
		fmt.Printf("%#v\n", v)
	}
}

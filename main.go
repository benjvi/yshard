package main

import (
	"fmt"
	"io"
	"log"
	"os"

	"github.com/itchyny/gojq"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
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

	fmt.Printf("configFile: %s, outputDir: %s, groupbyPath: %s\n", configFile, outputDir, groupbyPath)

	// TODO: construct the groupby query
	//let jq_groupby = format!("[ .[] |  select({groupby})] | [group_by({groupby})[] | {{ (.[0] | {groupby}): . }}] | add", groupby=groupby_path);
	// TODO: get items where the groupby path is not present
	// let jq_ungrouped = format!("[.[] |  select({groupby} | not)]", groupby=groupby_path);

	query, err := gojq.Parse(groupbyPath)
	if err != nil {
		log.Fatalln(err)
	}

	/*data, err := ioutil.ReadAll(os.Stdin)
	if err != nil {
		panic(err)
	}
	textContent := string(data)*/

	//TODO: load YAML from file

	fmt.Printf("reading data from stdin...")
	decoder := yaml.NewDecoder(os.Stdin)

	objects := make([]interface{}, 0)
	for {
		var obj map[string]interface{}
		err := decoder.Decode(&obj)
		if err == io.EOF {
			break
		} else if err != nil {
			fmt.Println("Error decoding YAML:", err)
			os.Exit(1)
		}
		fmt.Println("Decoded object:", obj)
		objects = append(objects, obj)
	}

	iter := query.Run(objects) // or query.RunWithContext
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

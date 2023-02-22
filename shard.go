package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
)

// TODO: use shard type where its being passed around
type shard struct {
	name     string
	yamlDocs []interface{}
}

func groupYamlDocsByPathValues(groupbyPath string, inputData []interface{}) (map[string]interface{}, []interface{}) {
	groupbyQuery := fmt.Sprintf("[ .[] |  select(%[1]s)] | [group_by(%[1]s)[] | { (.[0] | %[1]s): . }] | add", groupbyPath)
	fmt.Printf("Grouping YAML docs using the query: %q\n", groupbyQuery)
	groupedData := runGojqQuery(groupbyQuery, inputData)
	//fmt.Printf("Grouped data: %#v\n", groupedData)

	// groupby query doesn't include documents where the groupby path is missing or has no value
	ungroupedQuery := fmt.Sprintf("[.[] |  select(%[1]s | not)]", groupbyPath)
	fmt.Printf("Retrieving ungrouped YAML docs using the query: %q\n", ungroupedQuery)
	ungroupedData := runGojqQuery(ungroupedQuery, inputData)
	// fmt.Printf("Ungrouped data: %#v\n", ungroupedData)

	ungroupedDataOut := ungroupedData.([]interface{})      // an array of YAML docs
	groupedDataOut := groupedData.(map[string]interface{}) // maps shard name to an array of YAML docs

	return groupedDataOut, ungroupedDataOut
}

/*
Writes the grouped data to the output directory
*/
func outputGroupedDataToDir(groupedData map[string]interface{}, outputDir string) {
	for groupby_key, shardVal := range groupedData {
		groupedDocs := shardVal.([]interface{})
		if len(groupedDocs) > 0 {
			outputSingleShardToFile(groupedDocs, groupby_key, outputDir)
		}

	}
	// TODO: return details of the files created
}

func outputUngroupedDataToDir(ungroupedDocs []interface{}, outputDir string) {
	if len(ungroupedDocs) > 0 {
		outputSingleShardToFile(ungroupedDocs, "__ungrouped__", outputDir)
	}
	// TODO: return details of the created file
}

func outputSingleShardToFile(docs []interface{}, shardName string, outputDir string) {

	multiYamlDoc, err := multiDocYAMLToString(docs)
	if err != nil {
		log.Fatalln(err)
	}

	//fmt.Printf(multiYamlDoc)

	//output the doc to a file {{groupby_key}}.yml in the outputDir
	// need to sanitize the key before using it as a filename
	filename := filepath.Clean(shardName)
	outPath := filepath.Join(outputDir, filename+OUTPUT_YAML_EXT)

	// TODO: make sure that directory exists - if not, create it
	// TODO: check if this overwrites existing files (it should)
	if err := os.WriteFile(outPath, []byte(multiYamlDoc), 0666); err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Wrote %d YAML docs to file %q\n", len(docs), outPath)

}

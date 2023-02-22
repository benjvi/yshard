package main

import (
	"bytes"
	"fmt"
	"io"
	"os"

	"gopkg.in/yaml.v3"
)

func readYamlFromStdin() []interface{} {
	decoder := yaml.NewDecoder(os.Stdin)

	inputData := make([]interface{}, 0)
	for {
		var obj map[string]interface{}
		err := decoder.Decode(&obj)
		if err == io.EOF {
			break
		} else if err != nil {
			fmt.Println("Error decoding YAML:", err)
			os.Exit(1)
		}
		//fmt.Println("Decoded object:", obj)
		inputData = append(inputData, obj)
	}
	return inputData
}

func multiDocYAMLToString(docs []interface{}) (string, error) {
	var buf bytes.Buffer
	enc := yaml.NewEncoder(&buf)

	errs := make([]error, 0)
	for _, obj := range docs {
		//fmt.Printf("Encoding YAML doc #%d\n", i)
		err := enc.Encode(obj)
		if err != nil {
			errs = append(errs, err)
		}
	}
	if len(errs) != 0 {
		return "", fmt.Errorf("errors in encoding docs: %v", errs)
	}

	err := enc.Close()
	if err != nil {
		return "", err
	}

	return buf.String(), nil
}

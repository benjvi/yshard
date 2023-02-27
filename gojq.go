package main

import (
	"fmt"

	"github.com/itchyny/gojq"
)

func runGojqQuery(logger Logger, queryStr string, objects []interface{}) interface{} {
	query, err := gojq.Parse(queryStr)
	if err != nil {
		logger.Fatal(err.Error())
	}

	output := make([]interface{}, 0)

	iter := query.Run(objects) // or query.RunWithContext
	i := 0
	for {
		if i > 0 {
			// gojq says it can result many result objects but in our case, we always have one result object
			// later logic might not work properly if we have to merge multiple results
			logger.Fatal(fmt.Sprintf("unexpected query output with more than one result"))
		}
		v, ok := iter.Next()
		if !ok {
			break
		}
		// TODO: write a test that triggers this error handling
		if err, ok := v.(error); ok {
			logger.Fatal(err.Error())
		}
		//fmt.Printf("object: %#v\n", v)
		output = append(output, v)
	}
	return output[0]
}

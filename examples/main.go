package main

import (
    "encoding/json"
    "fmt"
    "io/ioutil"
    "os"
    "strings"

    "github.com/cloudputation/terrafactor"
)

// Configurable indentation level (number of spaces)
const indentLevel = 4

func main() {
    if len(os.Args) < 3 {
        fmt.Println("Usage: go run main.go <json_file> <operationTag>")
        os.Exit(1)
    }

    jsonFile      := os.Args[1]
    operationTag  := os.Args[2]

    // Create a string with the specified number of spaces for indentation
    indentStr := strings.Repeat(" ", indentLevel)

    // Read the JSON file
    dataBytes, err := ioutil.ReadFile(jsonFile)
    if err != nil {
        fmt.Println("Error reading JSON file:", err)
        os.Exit(1)
    }

    // Parse the JSON data
    var data interface{}
    err = json.Unmarshal(dataBytes, &data)
    if err != nil {
        fmt.Println("Error parsing JSON:", err)
        os.Exit(1)
    }

    // Use terrafactor to pretty-print the data with the specified operationTag and dynamic indentation
    err = terrafactor.Print(data, operationTag, indentStr, os.Stdout)
    if err != nil {
        fmt.Println("Error printing data:", err)
        os.Exit(1)
    }
}

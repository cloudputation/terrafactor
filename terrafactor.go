package terrafactor

import (
    "fmt"
    "io"
    "sort"
    "strings"
)

// ANSI color codes for terminal output
const (
    colorReset = "\033[0m"
    colorRed   = "\033[31m"
    colorGreen = "\033[32m"
)

// Print converts JSON data into Terraform-like output.
// It prepends colored plus or minus signs based on the operationTag ("create" or "destroy").
// It also allows controlling the indentation style.
func Print(data interface{}, operationTag string, indentStr string, w io.Writer) error {
    // Determine the prefix based on the operationTag
    var prefix string
    switch operationTag {
    case "create":
        prefix = colorGreen + "+  " + colorReset
    case "destroy":
        prefix = colorRed + "-  " + colorReset
    default:
        error := fmt.Errorf("invalid operation: %s. Supported operations are 'create' or 'destroy'", operationTag)
        return error
    }

    return printData(data, 0, prefix, indentStr, w)
}

// printData recursively traverses the JSON data and prints it in a formatted style.
func printData(data interface{}, indent int, prefix string, indentStr string, w io.Writer) error {
    // Length of the visible prefix (excluding ANSI color codes)
    visiblePrefix := "+  " // Assuming "+  " or "-  "
    prefixLength := len(visiblePrefix)

    // Base indentation in spaces
    baseIndent := indent * len(indentStr)

    // Function to calculate indentation for lines with and without prefix
    getIndent := func(hasPrefix bool) string {
        adjustedIndent := baseIndent
        if hasPrefix {
            adjustedIndent = baseIndent - prefixLength
            if adjustedIndent < 0 {
                adjustedIndent = 0
            }
        }
        return strings.Repeat(" ", adjustedIndent)
    }

    switch v := data.(type) {
    case map[string]interface{}:
        // Sort keys for consistent output
        keys := make([]string, 0, len(v))
        for k := range v {
            keys = append(keys, k)
        }
        sort.Strings(keys)

        for _, key := range keys {
            value := v[key]
            // Check if the value is a nested structure
            switch value.(type) {
            case map[string]interface{}:
                // Indentation without prefix
                indents := getIndent(false)
                // Print the object with curly braces without the prefix
                fmt.Fprintf(w, "%s%s {\n", indents, key)
                // Recursive call for nested map
                printData(value, indent+1, prefix, indentStr, w)
                fmt.Fprintf(w, "%s}\n", indents)
            case []interface{}:
                // Indentation without prefix
                indents := getIndent(false)
                // Print the array with square brackets without the prefix
                fmt.Fprintf(w, "%s%s [\n", indents, key)
                // Recursive call for array elements
                printArray(value.([]interface{}), indent+1, prefix, indentStr, w)
                fmt.Fprintf(w, "%s]\n", indents)
            default:
                // Indentation with prefix
                indents := getIndent(prefix != "")
                // Replace nil with "(Unknown Value)"
                var valStr string
                if value == nil {
                    valStr = "(Unknown Value)"
                } else {
                    valStr = fmt.Sprintf("%v", value)
                }
                // Primitive values with the prefix
                fmt.Fprintf(w, "%s%s%s = %s\n", indents, prefix, key, valStr)
            }
        }
    case []interface{}:
        // Array of primitives or nested structures
        printArray(v, indent, prefix, indentStr, w)
    default:
        // Indentation with prefix
        indents := getIndent(true)
        // Replace nil with "(Unknown Value)"
        var valStr string
        if v == nil {
            valStr = "(Unknown Value)"
        } else {
            valStr = fmt.Sprintf("%v", v)
        }
        // Print primitive values directly with the prefix
        fmt.Fprintf(w, "%s%s%v\n", indents, prefix, valStr)
    }

    return nil
}

// printArray handles the specific case where we are iterating through arrays.
func printArray(data []interface{}, indent int, prefix string, indentStr string, w io.Writer) {
    // Base indentation in spaces
    baseIndent := indent * len(indentStr)

    // Function to calculate indentation
    getIndent := func() string {
        return strings.Repeat(" ", baseIndent)
    }

    for _, elem := range data {
        switch v := elem.(type) {
        case map[string]interface{}:
            // Check if the map has only one key (e.g., "subscriber")
            if len(v) == 1 {
                for key, value := range v {
                    if nestedMap, ok := value.(map[string]interface{}); ok {
                        // Indentation without prefix
                        indents := getIndent()
                        // Print the key without additional braces or prefix
                        fmt.Fprintf(w, "%s%s {\n", indents, key)
                        // Recursive call for nested map
                        printData(nestedMap, indent+1, prefix, indentStr, w)
                        fmt.Fprintf(w, "%s}\n", indents)
                    }
                }
            } else {
                // Each object within an array uses its own braces without prefix
                indents := getIndent()
                fmt.Fprintf(w, "%s{\n", indents)
                // Recursive call for nested map
                printData(v, indent+1, prefix, indentStr, w)
                fmt.Fprintf(w, "%s}\n", indents)
            }
        case []interface{}:
            // Indentation without prefix
            indents := getIndent()
            // Nested arrays use square brackets without prefix
            fmt.Fprintf(w, "%s[\n", indents)
            // Recursive call for nested arrays
            printArray(v, indent+1, prefix, indentStr, w)
            fmt.Fprintf(w, "%s]\n", indents)
        default:
            // Indentation with prefix
            indents := getIndent()
            // Replace nil with "(Unknown Value)"
            var valStr string
            if elem == nil {
                valStr = "(Unknown Value)"
            } else {
                valStr = fmt.Sprintf("%v", elem)
            }
            // For primitive values, just print them with the prefix
            fmt.Fprintf(w, "%s%s%s\n", indents, prefix, valStr)
        }
    }
}

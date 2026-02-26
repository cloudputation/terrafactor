package terrafactor

import (
	"bufio"
	"fmt"
	"io"
	"sort"
	"strings"
)

const (
	colorReset  = "\033[0m"
	colorRed    = "\033[31m"
	colorGreen  = "\033[32m"
	colorYellow = "\033[33m"
	indentWidth = 16 // spaces per indent level used by all diff rendering
)

// Print converts JSON data into Terraform-like output.
// It prepends colored plus or minus signs based on the operationTag ("create" or "destroy").
// It also allows controlling the indentation style.
func Print(data interface{}, operationTag string, indentStr string, w io.Writer) error {
	var prefix string
	switch operationTag {
	case "create":
		prefix = colorGreen + "+  " + colorReset
	case "destroy":
		prefix = colorRed + "-  " + colorReset
	default:
		return fmt.Errorf("invalid operation: %s. Supported operations are 'create' or 'destroy'", operationTag)
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
			// Check if the value is a nested structure
			switch value := v[key].(type) {
			case map[string]interface{}:
				indents := getIndent(false)
				fmt.Fprintf(w, "%s%s {\n", indents, key)
				printData(value, indent+1, prefix, indentStr, w)
				fmt.Fprintf(w, "%s}\n", indents)
			case []interface{}:
				indents := getIndent(false)
				fmt.Fprintf(w, "%s%s [\n", indents, key)
				printArray(value, indent+1, prefix, indentStr, w)
				fmt.Fprintf(w, "%s]\n", indents)
			default:
				indents := getIndent(prefix != "")
				var valStr string
				if value == nil {
					valStr = "(Unknown Value)"
				} else {
					valStr = fmt.Sprintf("%v", value)
				}
				fmt.Fprintf(w, "%s%s%s = %s\n", indents, prefix, key, valStr)
			}
		}
	case []interface{}:
		printArray(v, indent, prefix, indentStr, w)
	default:
		indents := getIndent(true)
		var valStr string
		if v == nil {
			valStr = "(Unknown Value)"
		} else {
			valStr = fmt.Sprintf("%v", v)
		}
		fmt.Fprintf(w, "%s%s%v\n", indents, prefix, valStr)
	}

	return nil
}

// printArray handles the specific case where we are iterating through arrays.
func printArray(data []interface{}, indent int, prefix string, indentStr string, w io.Writer) {
	baseIndent := indent * len(indentStr)

	getIndent := func() string {
		return strings.Repeat(" ", baseIndent)
	}

	for _, elem := range data {
		switch v := elem.(type) {
		case map[string]interface{}:
			if len(v) == 1 {
				for key, value := range v {
					if nestedMap, ok := value.(map[string]interface{}); ok {
						indents := getIndent()
						fmt.Fprintf(w, "%s%s {\n", indents, key)
						printData(nestedMap, indent+1, prefix, indentStr, w)
						fmt.Fprintf(w, "%s}\n", indents)
					}
				}
			} else {
				indents := getIndent()
				fmt.Fprintf(w, "%s{\n", indents)
				printData(v, indent+1, prefix, indentStr, w)
				fmt.Fprintf(w, "%s}\n", indents)
			}
		case []interface{}:
			indents := getIndent()
			fmt.Fprintf(w, "%s[\n", indents)
			printArray(v, indent+1, prefix, indentStr, w)
			fmt.Fprintf(w, "%s]\n", indents)
		default:
			indents := getIndent()
			var valStr string
			if elem == nil {
				valStr = "(Unknown Value)"
			} else {
				valStr = fmt.Sprintf("%v", elem)
			}
			fmt.Fprintf(w, "%s%s%s\n", indents, prefix, valStr)
		}
	}
}

// fieldConstraint returns a human-readable constraint label for a ResourceField.
func fieldConstraint(fields []ResourceField, name string) string {
	for _, f := range fields {
		if f.Name == name {
			switch {
			case f.Computed && !f.Optional:
				return "(computed)"
			case f.Required:
				return "(required)"
			case f.Optional && f.Computed:
				return "(optional, computed)"
			case f.Optional:
				return "(optional)"
			case f.Sensitive:
				return "(sensitive)"
			}
		}
	}
	return "(computed)"
}

// PrintProviderDiff writes a Terraform-style diff for a provider to w.
// prev is a map of resourceName → []fieldName from the saved state (empty on first run).
// next is the list of ResourceSpecs parsed from the current OpenAPI spec.
func PrintProviderDiff(providerName string, prev map[string][]string, next []ResourceSpec, w io.Writer) {
	if prev == nil {
		prev = map[string][]string{}
	}

	nextMap := make(map[string]*ResourceSpec, len(next))
	for i := range next {
		nextMap[next[i].ResourceName] = &next[i]
	}

	allNames := make(map[string]struct{})
	for n := range prev {
		allNames[n] = struct{}{}
	}
	for n := range nextMap {
		allNames[n] = struct{}{}
	}
	sortedNames := make([]string, 0, len(allNames))
	for n := range allNames {
		sortedNames = append(sortedNames, n)
	}
	sort.Strings(sortedNames)

	providerHasChanges := len(prev) == 0
	if !providerHasChanges {
		for _, name := range sortedNames {
			_, inPrev := prev[name]
			_, inNext := nextMap[name]
			if !inPrev || !inNext {
				providerHasChanges = true
				break
			}
			if diffHasFieldChanges(prev[name], nextMap[name].Fields) {
				providerHasChanges = true
				break
			}
		}
	}

	providerSymbol := " "
	providerColor := colorReset
	if providerHasChanges {
		providerSymbol = "~"
		providerColor = colorYellow
	}

	fmt.Fprintf(w, "%s%s%s provider %s {\n\n", providerColor, providerSymbol, colorReset, providerName)

	for _, name := range sortedNames {
		_, inPrev := prev[name]
		rSpec, inNext := nextMap[name]

		switch {
		case !inPrev && inNext:
			printResourceDiff(w, "+", colorGreen, name, nil, rSpec.Fields)
		case inPrev && !inNext:
			printResourceDiff(w, "-", colorRed, name, prev[name], nil)
		default:
			prevFields := prev[name]
			if diffHasFieldChanges(prevFields, rSpec.Fields) {
				printResourceDiff(w, "~", colorYellow, name, prevFields, rSpec.Fields)
			} else {
				printResourceDiff(w, " ", colorReset, name, prevFields, rSpec.Fields)
			}
		}
	}

	fmt.Fprintf(w, "%s%s%s }\n\n", providerColor, providerSymbol, colorReset)
}

// printResourceDiff writes one resource block with per-field +/~/- symbols.
func printResourceDiff(w io.Writer, symbol, color, name string, prevFieldNames []string, nextFields []ResourceField) {
	lvl1 := strings.Repeat(" ", indentWidth)
	lvl2 := strings.Repeat(" ", indentWidth*2)

	if symbol == "~" {
		fmt.Fprintf(w, "%s%s~%s resource \"%s\" => \"%s\" {\n", lvl1, colorYellow, colorReset, name, name)
	} else {
		fmt.Fprintf(w, "%s%s%s%s resource \"%s\" {\n", lvl1, color, symbol, colorReset, name)
	}

	prevSet := make(map[string]struct{}, len(prevFieldNames))
	for _, f := range prevFieldNames {
		prevSet[f] = struct{}{}
	}
	nextSet := make(map[string]struct{}, len(nextFields))
	for _, f := range nextFields {
		nextSet[f.Name] = struct{}{}
	}

	allFields := make(map[string]struct{})
	for f := range prevSet {
		allFields[f] = struct{}{}
	}
	for f := range nextSet {
		allFields[f] = struct{}{}
	}
	sorted := make([]string, 0, len(allFields))
	for f := range allFields {
		sorted = append(sorted, f)
	}
	sort.Strings(sorted)

	maxLen := 0
	for _, f := range sorted {
		if len(f) > maxLen {
			maxLen = len(f)
		}
	}

	for _, fieldName := range sorted {
		_, inPrev := prevSet[fieldName]
		_, inNext := nextSet[fieldName]
		pad := strings.Repeat(" ", maxLen-len(fieldName))

		switch {
		case !inPrev && inNext:
			constraint := fieldConstraint(nextFields, fieldName)
			fmt.Fprintf(w, "%s%s+%s   %s%s = %s %s(will be created)%s\n", lvl2, colorGreen, colorReset, fieldName, pad, constraint, colorGreen, colorReset)
		case inPrev && !inNext:
			fmt.Fprintf(w, "%s%s-%s   %s%s %s(will be destroyed)%s\n", lvl2, colorRed, colorReset, fieldName, pad, colorRed, colorReset)
		default:
			constraint := fieldConstraint(nextFields, fieldName)
			fmt.Fprintf(w, "%s    %s%s = %s %s(will be modified)%s\n", lvl2, fieldName, pad, constraint, colorYellow, colorReset)
		}
	}

	fmt.Fprintf(w, "%s%s%s%s }\n\n", lvl1, color, symbol, colorReset)
}

// PromptApproval prints a diff summary and asks for "yes" confirmation.
// Returns true only if the user types exactly "yes".
func PromptApproval(prev map[string][]string, next []ResourceSpec, r io.Reader, w io.Writer) bool {
	nextMap := make(map[string]*ResourceSpec, len(next))
	for i := range next {
		nextMap[next[i].ResourceName] = &next[i]
	}

	allNames := make(map[string]struct{})
	for n := range prev {
		allNames[n] = struct{}{}
	}
	for n := range nextMap {
		allNames[n] = struct{}{}
	}

	toAdd, toDestroy, toChange := 0, 0, 0
	for name := range allNames {
		_, inPrev := prev[name]
		_, inNext := nextMap[name]
		switch {
		case !inPrev && inNext:
			toAdd++
		case inPrev && !inNext:
			toDestroy++
		default:
			if diffHasFieldChanges(prev[name], nextMap[name].Fields) {
				toChange++
			}
		}
	}

	fmt.Fprintf(w, "  %d to add, %d to destroy, %d to change\n\n", toAdd, toDestroy, toChange)
	fmt.Fprintf(w, "Do you want to perform these actions?\n")
	fmt.Fprintf(w, "  Only 'yes' will be accepted to approve.\n\n")
	fmt.Fprintf(w, "  Enter a value: ")

	scanner := bufio.NewScanner(r)
	scanner.Scan()
	input := strings.TrimSpace(scanner.Text())
	fmt.Fprintln(w)

	if input == "yes" {
		return true
	}

	fmt.Fprintf(w, "  Apply cancelled.\n\n")
	return false
}

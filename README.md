# TerraFactor
Add Terraform functionalities to your software


## Description
Provide Terraform like output functionality that you can integrate to your software.
Takes JSON objects and outputs them in a Terraform-like format. It processes the structure, applying colored prefixes (`+` for creation, `-` for destruction) to indicate changes, and formats the output with customizable indentation.

## Module installation
- Install module: `go get github.com/cloudputation/terrafactor`
- Import the module in application code: `import "github.com/cloudputation/terrafactor"`

## Terraform Output Usage
operationTag should be set to either `create` or `destroy`.


## Error Handling
Not providing a value for `operationTag` will return the following error: `invalid operation: <operation>. Supported operations are 'create' or 'destroy'`

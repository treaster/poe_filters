Implement a command line tool for generating/compiling Path of Exile filter files.

Author the filter in a mostly-normal style, but the compiler's input format
additionally supports:
- Variable declarations and references
- Reusable, parameterizable style blocks
- Multiple BaseType or Prophecy lines in one block

See the example for details on the syntax.

Use the following to compile:
```go run treaster/applications/poe_filter/main.go --input=example.input --output=example.filter```

Move the resulting file to your Path of Exile filters directory.

package lib

import (
	"fmt"
	"regexp"
	"sort"
	"strings"
)

var alphaRegex *regexp.Regexp
var canonRegex *regexp.Regexp
var splitRegex *regexp.Regexp
var varRegex *regexp.Regexp

func init() {
	alphaRegex = regexp.MustCompile("[A-Za-z]+")
	canonRegex = regexp.MustCompile("[ ]+")
	splitRegex = regexp.MustCompile(`("[^"]*")|([^ "]*)`)
	varRegex = regexp.MustCompile(`\[\[[^\[\]]+\]\]`)
}

func SplitLine(line string) []string {
	parts := splitRegex.FindAllString(line, -1)
	var output []string
	for _, part := range parts {
		output = append(output, strings.Trim(part, `" `))
	}
	return output
}

func Canonicalize(input string) string {
	input = strings.TrimSpace(input)
	input = canonRegex.ReplaceAllString(input, " ")
	return input
}

func ParseLine(line string) (string, []string) {
	line = Canonicalize(line)

	parts := strings.SplitN(line, " ", 2)

	keyword := ""
	args := []string{}
	switch len(parts) {
	case 0:
	case 1:
		keyword = parts[0]
	case 2:
		keyword = parts[0]
		args = SplitLine(parts[1])
	default:
		panic("unexpected extra count in ParseLine")
	}

	return keyword, args
}

func FormatLine(key string, args []string) string {
	var formattedArgs []string
	for _, arg := range args {
		// Wrap an arg in quotes if it contains spaces
		if len(arg) > 0 && arg[0] != '#' && alphaRegex.MatchString(arg) && strings.Contains(arg, " ") {
			arg = fmt.Sprintf(`"%s"`, arg)
		}
		formattedArgs = append(formattedArgs, arg)
	}
	return fmt.Sprintf("    %s %s", key, strings.Join(formattedArgs, " "))
}

func Compile(input string) (string, error) {
	lines := strings.Split(input, "\n")

	styleMap := map[string]styleDefinition{}
	var outputLines []string

	varMap := map[string]string{}

LOOP:
	for i := 0; i < len(lines); i++ {
		line := lines[i]

		keyword, _ := ParseLine(line)
		switch keyword {
		case "":
			continue LOOP
		case "#":
			outputLines = append(outputLines, line)
		case "Show":
			fallthrough
		case "Hide":
			processShowOrHide(lines, &i, styleMap, &outputLines)
		case "DefineStyle":
			processDefineStyle(lines, &i, styleMap)
		case "DefineVar":
			processDefineVar(lines, &i, varMap)
		default:
			panic(fmt.Sprintf("unexpected keyword %q on line %d: %s", keyword, i, lines[i]))
		}
	}

	result := strings.Join(outputLines, "\n")

	for key, value := range varMap {
		result = strings.Replace(result, fmt.Sprintf("[[%s]]", key), value, -1)
	}

	return result, nil
}

func processShowOrHide(lines []string, i *int, styleDefinitions map[string]styleDefinition, outputLines *[]string) {
	showOrHide := lines[*i]
	*i++

	var baseTypeKey string
	baseTypesSet := map[string]bool{}
	otherLines := map[string][]string{}

LOOP:
	for *i = *i; *i < len(lines); *i++ {
		line := lines[*i]

		keyword, rest := ParseLine(line)
		switch keyword {
		case "":
			continue LOOP
		case "#":
			*outputLines = append(*outputLines, line)
		case "Show":
			fallthrough
		case "Hide":
			fallthrough
		case "DefineStyle":
			*i--
			break LOOP
		case "BaseType":
			fallthrough
		case "Prophecy":
			if baseTypeKey != "" && baseTypeKey != keyword {
				panic(fmt.Sprintf("conflicting filters %q and %q on the same rule", baseTypeKey, keyword))
			}
			baseTypeKey = keyword
			for _, bt := range rest {
				_, hasBaseType := baseTypesSet[bt]
				if hasBaseType {
					panic(fmt.Sprintf("duplicate %q specified: %q", baseTypeKey, bt))
				}
				baseTypesSet[bt] = true
			}
		case "UseStyle":
			// Parse the UseStyle name + args
			var styleName string
			var styleArgValues []string
			switch len(rest) {
			case 0:
				panic("UseStyle missing names. expected 'UseStyle [name1] [name2] ...")
			default:
				styleName = rest[0]
				styleArgValues = rest[1:]
			}

			// Look up the style name in styleDefinitions
			styleDef, hasStyle := styleDefinitions[styleName]
			if !hasStyle {
				panic(fmt.Sprintf("style %q referenced but not predefined", styleName))
			}

			// Compute the arg assignments for this UseStyle invocation
			if len(styleDef.args) != len(styleArgValues) {
				panic(fmt.Sprintf("Style %q expected %d values, but received %d values", styleName, len(styleDef.args), len(styleArgValues)))
			}

			// Build a map of style argName->value
			styleArgAssigns := map[string]string{}
			for i := 0; i < len(styleDef.args); i++ {
				arg := fmt.Sprintf("[[%s]]", styleDef.args[i])
				value := styleArgValues[i]
				styleArgAssigns[arg] = value
			}

			// Build an ordered list of filter keywords used by the style
			var styleKeys []string
			for keyword, _ := range styleDef.lines {
				styleKeys = append(styleKeys, keyword)
			}
			sort.Strings(styleKeys)

			// Iterate through the style lines in keyword order
			// Replace each element the is a var reference with the
			// corresponding assigned value.
			for _, keyword := range styleKeys {
				line := styleDef.lines[keyword]
				var finalLine []string
				for j := 0; j < len(line); j++ {
					item := line[j]
					value, hasItem := styleArgAssigns[item]
					if hasItem {
						item = value
					}
					finalLine = append(finalLine, item)
				}
				otherLines[keyword] = finalLine
			}

		default:
			otherLines[keyword] = rest
		}
	}

	var otherKeys []string
	for key, _ := range otherLines {
		otherKeys = append(otherKeys, key)
	}
	sort.Strings(otherKeys)

	var baseTypes []string
	for key, _ := range baseTypesSet {
		baseTypes = append(baseTypes, key)
	}
	sort.Strings(baseTypes)

	if len(baseTypes) == 0 {
		*outputLines = append(*outputLines, showOrHide)
		for _, key := range otherKeys {
			*outputLines = append(*outputLines, FormatLine(key, otherLines[key]))
		}
		*outputLines = append(*outputLines, "")
	} else {
		for _, baseType := range baseTypes {
			*outputLines = append(*outputLines, showOrHide)
			*outputLines = append(*outputLines, FormatLine(baseTypeKey, []string{baseType}))

			for _, key := range otherKeys {
				*outputLines = append(*outputLines, FormatLine(key, otherLines[key]))
			}
			*outputLines = append(*outputLines, "")
		}
	}
}

type styleDefinition struct {
	args  []string
	lines map[string][]string // keyword: args
}

func processDefineStyle(lines []string, i *int, styleMap map[string]styleDefinition) {
	_, nameArr := ParseLine(lines[*i])
	var name string
	var args []string
	switch len(nameArr) {
	case 0:
		panic("DefineStyle missing name. expected 'DefineStyle [name]")
	default:
		name = nameArr[0]
		args = nameArr[1:]
	}
	*i++

	styleDef := styleDefinition{
		args:  args,
		lines: map[string][]string{},
	}
LOOP:
	for *i = *i; *i < len(lines); *i++ {
		keyword, rest := ParseLine(lines[*i])
		switch keyword {
		case "":
			continue LOOP
		case "Show":
			fallthrough
		case "Hide":
			fallthrough
		case "DefineStyle":
			*i--
			break LOOP
		default:
			rest = append(rest, fmt.Sprintf("# Style %q", name))
			styleDef.lines[keyword] = rest
		}
	}

	styleMap[name] = styleDef
}

func processDefineVar(lines []string, i *int, varMap map[string]string) {
	_, args := ParseLine(lines[*i])
	var name string
	var value string
	switch len(args) {
	case 0:
		fallthrough
	case 1:
		panic("DefineVar bad usage. expected 'DefineVar [name] [value]")
	case 2:
		name = args[0]
		value = args[1]
	case 3:
		panic(fmt.Sprintf("DefineVar more than two args. Line is: %q", lines[*i]))
	}

	varMap[name] = value
}

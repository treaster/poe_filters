package lib_test

import (
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"treaster/applications/poe_filter/lib"
)

func TestCanonicalize(t *testing.T) {
	testCases := []struct {
		input    string
		expected string
	}{
		{"input", "input"},
		{"   input   ", "input"},
		{"input input", "input input"},
		{"   input input   ", "input input"},
		{"   input    input   ", "input input"},
	}

	for _, testCase := range testCases {
		require.Equal(t, testCase.expected, lib.Canonicalize(testCase.input))
	}
}

func TestSplitLine(t *testing.T) {
	testCases := []struct {
		input    string
		expected []string
	}{
		{`input`, []string{"input"}},
		{`"input"`, []string{"input"}},
		{`foo bar baz`, []string{"foo", "bar", "baz"}},
		{`"foo" bar "baz"`, []string{"foo", "bar", "baz"}},
		{`foo "bar" baz`, []string{"foo", "bar", "baz"}},
		{`"foo" "bar" "baz"`, []string{"foo", "bar", "baz"}},
		{`"foo foo foo" "bar bar bar" "baz baz baz"`, []string{"foo foo foo", "bar bar bar", "baz baz baz"}},
		{`"foo foo foo" bar bar bar "baz baz baz"`, []string{"foo foo foo", "bar", "bar", "bar", "baz baz baz"}},
	}

	for testI, testCase := range testCases {
		require.Equal(t, testCase.expected, lib.SplitLine(testCase.input), fmt.Sprintf("Test %d", testI))
	}
}

func TestParseLine(t *testing.T) {
	testCases := []struct {
		input           string
		expectedKeyword string
		expectedArgs    []string
	}{
		{`input`, "input", []string{}},
		{`input foo bar baz`, "input", []string{"foo", "bar", "baz"}},
		{`input "foo" "bar" "baz"`, "input", []string{"foo", "bar", "baz"}},
		{`input foo "bar" baz "quoz"`, "input", []string{"foo", "bar", "baz", "quoz"}},
	}

	for _, testCase := range testCases {
		keyword, args := lib.ParseLine(testCase.input)
		require.Equal(t, testCase.expectedKeyword, keyword)
		require.Equal(t, testCase.expectedArgs, args)
	}
}

func TestFormatLine(t *testing.T) {
	testCases := []struct {
		inputKey  string
		inputArgs []string
		expected  string
	}{
		{
			"keyword",
			[]string{"arg"},
			"    keyword arg",
		},
	}

	for _, testCase := range testCases {
		require.Equal(t, testCase.expected, lib.FormatLine(testCase.inputKey, testCase.inputArgs))
	}
}

func TestSpec(t *testing.T) {
	testCases := []struct {
		inputSpec      string
		expectedOutput string
	}{
		// basic echo of a BaseType rule
		{
			`
Show
    BaseType "Ancient Shard"
    SetBackgroundColor 150 150 0`,
			`
Show
    BaseType "Ancient Shard"
    SetBackgroundColor 150 150 0`,
		},

		// expand BaseType rule into multiple final entries
		{
			`
Show
    BaseType "Ancient Shard" "Exalted Orb"
    SetBackgroundColor 150 150 0`,
			`
Show
    BaseType "Ancient Shard"
    SetBackgroundColor 150 150 0

Show
    BaseType "Exalted Orb"
    SetBackgroundColor 150 150 0`,
		},

		// rule with no BaseType should be echoed directly
		{
			`
Show
    Class Currency
    SetBackgroundColor 150 150 0`,
			`
Show
    Class Currency
    SetBackgroundColor 150 150 0`,
		},

		// fill in one style
		{
			`
DefineStyle Valuable
    SetBackgroundColor 150 150 0

Show
    BaseType "Ancient Shard"
	UseStyle Valuable`,
			`
Show
    BaseType "Ancient Shard"
    SetBackgroundColor 150 150 0 # Style "Valuable"`,
		},

		// fill in multiple styles on one rule, separate UseStyle entries
		{
			`
DefineStyle Valuable
    SetBackgroundColor 150 150 0
    MinimapIcon 1 Green Circle

DefineStyle Chromatic
    SetBorderColor 0 255 0
    PlayAlertSound 7 100

Show
    BaseType "Ancient Shard"
	UseStyle Valuable
	UseStyle Chromatic
	`,
			`
Show
    BaseType "Ancient Shard"
    MinimapIcon 1 Green Circle # Style "Valuable"
    PlayAlertSound 7 100 # Style "Chromatic"
    SetBackgroundColor 150 150 0 # Style "Valuable"
    SetBorderColor 0 255 0 # Style "Chromatic"
`,
		},

		// fill in multiple styles on multiple rules
		{
			`
DefineStyle Valuable
    SetBackgroundColor 150 150 0
    MinimapIcon 1 Green Circle

DefineStyle Chromatic
    SetBorderColor 0 255 0
    PlayAlertSound 7 100

Show
    BaseType "Ancient Shard"
	UseStyle Valuable
	UseStyle Chromatic

Show
    BaseType "Exalted Orb"
	UseStyle Valuable
	UseStyle Chromatic
	`,
			`
Show
    BaseType "Ancient Shard"
    MinimapIcon 1 Green Circle # Style "Valuable"
    PlayAlertSound 7 100 # Style "Chromatic"
    SetBackgroundColor 150 150 0 # Style "Valuable"
    SetBorderColor 0 255 0 # Style "Chromatic"

Show
    BaseType "Exalted Orb"
    MinimapIcon 1 Green Circle # Style "Valuable"
    PlayAlertSound 7 100 # Style "Chromatic"
    SetBackgroundColor 150 150 0 # Style "Valuable"
    SetBorderColor 0 255 0 # Style "Chromatic"
`,
		},

		// define and reference variables
		{
			`
DefineVar CurrencyShape Square
Show
    BaseType "Ancient Shard"
	MinimapIcon 1 Red [[CurrencyShape]]`,
			`
Show
    BaseType "Ancient Shard"
    MinimapIcon 1 Red Square`,
		},

		// pass args to style
		{
			`
DefineStyle Valuable A B
	MinimapIcon 1 [[A]] [[B]]

Show
    BaseType "Ancient Shard"
	UseStyle Valuable Red Square`,
			`
Show
    BaseType "Ancient Shard"
    MinimapIcon 1 Red Square # Style "Valuable"`,
		},

		// pass numeric/color args to style
		{
			`
DefineStyle Valuable BGColor TColor
	SetBackgroundColor [[BGColor]]
	SetTextColor [[TColor]]

Show
    BaseType "Ancient Shard"
	UseStyle Valuable "1 1 1" "2 2 2"`,
			`
Show
    BaseType "Ancient Shard"
    SetBackgroundColor 1 1 1 # Style "Valuable"
    SetTextColor 2 2 2 # Style "Valuable"`,
		},

		// TODO(treaster): Add tests for handling comments
		// TODO(treaster): Add tests for handling Prophecy keyword
	}
	for testI, testCase := range testCases {
		output, err := lib.Compile(testCase.inputSpec)
		require.NoError(t, err)

		require.Equal(t, strings.TrimSpace(testCase.expectedOutput), strings.TrimSpace(output), fmt.Sprintf("Test %d", testI))
	}
}

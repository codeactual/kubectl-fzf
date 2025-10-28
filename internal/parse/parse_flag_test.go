package parse

import "testing"

func TestUnmanagedArgs(t *testing.T) {
	cmdArgs := [][]string{
		{"-t"},
		{"-i"},
		{"--field-selector"},
		{"--selector"},
	}
	for _, args := range cmdArgs {
		r := CheckFlagManaged(args)
		if r.String() != FlagUnmanaged.String() {
			t.Errorf("CheckFlagManaged(%q) = %s, want %s", args, r.String(), FlagUnmanaged.String())
		}
	}
}

type flagTest struct {
	flag   []string
	result FlagCompletion
}

func TestManagedArgs(t *testing.T) {
	cmdArgs := []flagTest{
		{[]string{"--selector="}, FlagLabel},
		{[]string{"--field-selector", ""}, FlagFieldSelector},
		{[]string{"--field-selector="}, FlagFieldSelector},
		{[]string{"--all-namespaces", ""}, FlagNone},
		{[]string{"-t", ""}, FlagNone},
		{[]string{"-i", ""}, FlagNone},
		{[]string{"-ti", ""}, FlagNone},
		{[]string{"-it", ""}, FlagNone},
		{[]string{"-n"}, FlagNamespace},
		{[]string{"-n="}, FlagNamespace},
		{[]string{"-n", " "}, FlagNamespace},
		{[]string{"--namespace", ""}, FlagNamespace},
	}
	for _, args := range cmdArgs {
		r := CheckFlagManaged(args.flag)
		if r.String() != args.result.String() {
			t.Errorf("CheckFlagManaged(%q) = %s, want %s", args.flag, r.String(), args.result.String())
		}
	}
}

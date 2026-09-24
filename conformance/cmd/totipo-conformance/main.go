package main

import (
	"errors"
	"fmt"
	"os"
	"strings"
	"totipo/conformance/internal/corpus"
	"totipo/conformance/internal/runner"
)

func main() {
	root := ""
	filter := ""
	args := os.Args[1:]
	for len(args) > 0 {
		a := args[0]
		args = args[1:]
		if a == "--filter" {
			if len(args) == 0 {
				fatal("--filter requires a prefix")
			}
			filter = args[0]
			args = args[1:]
		} else if strings.HasPrefix(a, "--filter=") {
			filter = strings.TrimPrefix(a, "--filter=")
		} else if root == "" && !strings.HasPrefix(a, "-") {
			root = a
		} else {
			fatal("usage: totipo-conformance [--filter PREFIX] VECTOR_DIRECTORY")
		}
	}
	if root == "" {
		fatal("usage: totipo-conformance [--filter PREFIX] VECTOR_DIRECTORY")
	}
	cases, e := corpus.Load(root)
	if e != nil {
		fatal(e.Error())
	}
	pass, fail, blocked := 0, 0, 0
	coverage := map[string]int{}
	for _, c := range cases {
		coverage[strings.Split(c.ID, "/")[1]]++
		if !strings.HasPrefix(c.ID, filter) {
			continue
		}
		if e := runner.Run(c); e != nil {
			if errors.Is(e, runner.ErrBlocked) {
				blocked++
				fmt.Printf("BLOCKED %s\n     %s\n", c.ID, e)
				continue
			}
			fail++
			fmt.Printf("FAIL %s\n     %s\n     source: %s (%s)\n", c.ID, e, c.Provenance.Source, c.Provenance.Locator)
		} else {
			pass++
			fmt.Println("PASS", c.ID)
		}
	}
	for _, category := range []string{"ed25519", "lifecycle", "transitions", "recovery", "presentation"} {
		prefix := "v0/" + category + "/"
		if coverage[category] == 0 && (filter == "" || strings.HasPrefix(prefix, filter) || strings.HasPrefix(filter, prefix)) {
			blocked++
			fmt.Printf("BLOCKED %srequired-corpus\n     No cases supplied for required category\n", prefix)
		}
	}
	if pass+fail+blocked == 0 {
		fatal("filter matched no cases")
	}
	fmt.Printf("PASS %d / FAIL %d / BLOCKED %d\n", pass, fail, blocked)
	if fail > 0 {
		os.Exit(1)
	}
	if blocked > 0 {
		os.Exit(2)
	}
}
func fatal(s string) { fmt.Fprintln(os.Stderr, s); os.Exit(1) }

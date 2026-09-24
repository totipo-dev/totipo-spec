package main

import (
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

	"totipo/conformance/internal/corpus"
	"totipo/conformance/internal/requirements"
	"totipo/conformance/internal/runner"
)

func main() { os.Exit(run(os.Args[1:], os.Stdout, os.Stderr)) }

const usage = "usage: totipo-conformance [--requirements FILE] [--filter PREFIX] [--verify-only] [VECTOR_DIRECTORY]"

func run(args []string, out, errOut io.Writer) int {
	fatal := func(err any) int { fmt.Fprintln(errOut, err); return 1 }
	root, filter, profilePath := "", "", ""
	filtered, verifyOnly := false, false
	for len(args) > 0 {
		a := args[0]
		args = args[1:]
		switch {
		case a == "--filter" || a == "--requirements":
			if len(args) == 0 || strings.HasPrefix(args[0], "--") {
				return fatal(a + " requires a value")
			}
			if a == "--filter" {
				filter, filtered = args[0], true
			} else {
				profilePath = args[0]
			}
			args = args[1:]
		case strings.HasPrefix(a, "--filter="):
			filter, filtered = strings.TrimPrefix(a, "--filter="), true
		case strings.HasPrefix(a, "--requirements="):
			profilePath = strings.TrimPrefix(a, "--requirements=")
			if profilePath == "" {
				return fatal("--requirements requires a value")
			}
		case a == "--verify-only":
			verifyOnly = true
		case root == "" && !strings.HasPrefix(a, "-"):
			root = a
		default:
			return fatal(usage)
		}
	}
	if profilePath == "" && (root == "" || verifyOnly) || verifyOnly && filtered {
		return fatal(usage)
	}
	var cases []corpus.Case
	var profile *requirements.Profile
	var err error
	if profilePath != "" {
		profile, err = requirements.Load(profilePath)
		if err != nil {
			return fatal(err)
		}
		repo, err := requirements.RepositoryRoot(profilePath)
		if err != nil {
			return fatal(err)
		}
		cases, err = profile.LoadCases(repo, root)
		if err != nil {
			return fatal(err)
		}
		if verifyOnly {
			fmt.Fprintf(out, "VALID %s: integrity and %d required IDs verified; cases not executed\n", profile.Profile, len(cases))
			return 0
		}
		if filtered {
			fmt.Fprintf(out, "PARTIAL %s: diagnostic filter %q; not a complete profile-conformance run\n", profile.Profile, filter)
		} else {
			fmt.Fprintf(out, "Requirements profile %s: %d required cases\n", profile.Profile, len(cases))
		}
	} else {
		cases, err = corpus.Load(root)
		if err != nil {
			return fatal(err)
		}
	}
	counts := requirements.Counts{}
	coverage := map[string]int{}
	for _, c := range cases {
		coverage[strings.Split(c.ID, "/")[1]]++
		if !strings.HasPrefix(c.ID, filter) {
			continue
		}
		if err := runner.Run(c); err != nil {
			if errors.Is(err, runner.ErrBlocked) {
				counts.Blocked++
				fmt.Fprintf(out, "BLOCKED %s\n     %s\n", c.ID, err)
				continue
			}
			counts.Fail++
			fmt.Fprintf(out, "FAIL %s\n     %s\n     source: %s (%s)\n", c.ID, err, c.Provenance.Source, c.Provenance.Locator)
		} else {
			counts.Pass++
			fmt.Fprintln(out, "PASS", c.ID)
		}
	}
	// Current-corpus development mode retains its original coverage gates.
	// Profile coverage comes from the frozen ID set, not future CLI policy.
	if profile == nil {
		for _, category := range []string{"ed25519", "lifecycle", "transitions", "recovery", "presentation"} {
			prefix := "v0/" + category + "/"
			if coverage[category] == 0 && (filter == "" || strings.HasPrefix(prefix, filter) || strings.HasPrefix(filter, prefix)) {
				counts.Blocked++
				fmt.Fprintf(out, "BLOCKED %srequired-corpus\n     No cases supplied for required category\n", prefix)
			}
		}
	}
	if counts.Pass+counts.Fail+counts.Blocked == 0 {
		return fatal("filter matched no cases")
	}
	fmt.Fprintf(out, "PASS %d / FAIL %d / BLOCKED %d\n", counts.Pass, counts.Fail, counts.Blocked)
	if counts.Fail > 0 {
		return 1
	}
	if counts.Blocked > 0 {
		return 2
	}
	if profile != nil && !filtered {
		if counts != profile.Requirements.Expected {
			return fatal("profile result counts do not match expected counts")
		}
		fmt.Fprintln(out, "CONFORMANT", profile.Profile)
	}
	return 0
}

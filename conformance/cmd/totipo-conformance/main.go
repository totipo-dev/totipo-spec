package main

import (
	"flag"
	"fmt"
	"os"
	"totipo/conformance/internal/vectors"
)

func main() {
	root := flag.String("root", ".", "repository root")
	verify := flag.Bool("verify-only", false, "validate structure, checksums and moving requirements without executing cases")
	capability := flag.String("capability", "", "include conditional advisory-history cases alongside baseline")
	flag.Parse()
	if flag.NArg() != 0 {
		fmt.Fprintln(os.Stderr, "unexpected arguments")
		os.Exit(2)
	}
	m, cases, e := vectors.Read(*root)
	if e == nil {
		e = vectors.VerifyProfile(*root, m)
	}
	if e != nil {
		fmt.Fprintln(os.Stderr, "FAIL:", e)
		os.Exit(1)
	}
	var capabilities []string
	if *capability != "" {
		capabilities = []string{*capability}
	}
	selected, e := vectors.Select(m, cases, capabilities)
	if e != nil {
		fmt.Fprintln(os.Stderr, "FAIL:", e)
		os.Exit(1)
	}
	labels := map[string]string{}
	for _, entry := range m.Cases {
		labels[entry.ID] = entry.Applicability.Label()
	}
	if *verify {
		fmt.Printf("PASS: manifest, profile and %d case files verified\n", len(cases))
		return
	}
	failures := 0
	baseline, conditional := 0, 0
	for _, c := range selected {
		if labels[c.ID] == "baseline" {
			baseline++
		} else {
			conditional++
		}
		if e := vectors.RunWithCapabilities(c, capabilities); e != nil {
			fmt.Fprintf(os.Stderr, "FAIL %s [%s]: %v\n", c.ID, labels[c.ID], e)
			failures++
		}
	}
	if failures > 0 {
		fmt.Fprintf(os.Stderr, "FAIL: %d/%d cases\n", failures, len(selected))
		os.Exit(1)
	}
	fmt.Printf("PASS: baseline moving-pre-rc: %d cases\n", baseline)
	if len(capabilities) > 0 {
		fmt.Printf("PASS: capability=advisory-history: %d conditional cases\n", conditional)
		fmt.Printf("PASS: full reference corpus: %d cases\n", len(selected))
	}
}

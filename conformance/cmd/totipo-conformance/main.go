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
	if *verify {
		fmt.Printf("PASS: manifest, profile and %d case files verified\n", len(cases))
		return
	}
	failures := 0
	for _, c := range cases {
		if e := vectors.Run(c); e != nil {
			fmt.Fprintf(os.Stderr, "FAIL %s: %v\n", c.ID, e)
			failures++
		}
	}
	if failures > 0 {
		fmt.Fprintf(os.Stderr, "FAIL: %d/%d cases\n", failures, len(cases))
		os.Exit(1)
	}
	fmt.Printf("PASS: %d Totipo v1/r9 conformance cases\n", len(cases))
}

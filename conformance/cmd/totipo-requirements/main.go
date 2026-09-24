// Explicit maintenance command. Normal verification never invokes this tool.
package main

import (
	"flag"
	"fmt"
	"io"
	"os"

	"totipo/conformance/internal/requirements"
)

func main() { os.Exit(run(os.Args[1:], os.Stdout, os.Stderr)) }

func run(args []string, out, errOut io.Writer) int {
	flags := flag.NewFlagSet("totipo-requirements", flag.ContinueOnError)
	flags.SetOutput(errOut)
	name := flags.String("profile", "", "new profile name, e.g. v0-rc2")
	root := flags.String("repo-root", ".", "repository root")
	write := flags.Bool("write", false, "write new profile and frozen manifest; never overwrite")
	if err := flags.Parse(args); err != nil {
		return 1
	}
	if *name == "" || flags.NArg() != 0 {
		fmt.Fprintln(errOut, "usage: totipo-requirements --profile NAME [--repo-root DIR] [--write]")
		return 1
	}
	p, manifest, err := requirements.Candidate(*root, *name)
	if err == nil && *write {
		err = requirements.WriteCandidate(*root, p, manifest)
		if err == nil {
			fmt.Fprintf(out, "Created requirements/%s.json and requirements/%s.manifest.sha256; review before release\n", *name, *name)
			return 0
		}
	}
	if err == nil {
		var b []byte
		b, err = p.JSON()
		if err == nil {
			_, err = out.Write(b)
		}
	}
	if err != nil {
		fmt.Fprintln(errOut, err)
		return 1
	}
	return 0
}

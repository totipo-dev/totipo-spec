// Only hashes existing artifacts; never writes vector expectations.
package main

import (
	"crypto/sha256"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

func main() {
	write := flag.Bool("write-manifest", false, "explicit maintenance")
	flag.Parse()
	if !*write {
		panic("requires --write-manifest")
	}
	root := "vectors/v0"
	var names []string
	e := filepath.WalkDir(root, func(p string, d os.DirEntry, e error) error {
		if e != nil {
			return e
		}
		if !d.IsDir() && strings.HasSuffix(p, ".json") {
			names = append(names, p)
		}
		return nil
	})
	if e != nil {
		panic(e)
	}
	sort.Strings(names)
	var out strings.Builder
	for _, p := range names {
		b, e := os.ReadFile(p)
		if e != nil {
			panic(e)
		}
		fmt.Fprintf(&out, "%x  %s\n", sha256.Sum256(b), filepath.ToSlash(strings.TrimPrefix(p, root+"/")))
	}
	if e := os.WriteFile(root+"/manifest.sha256", []byte(out.String()), 0644); e != nil {
		panic(e)
	}
}

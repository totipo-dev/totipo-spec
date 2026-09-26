package vectors

import (
	"encoding/hex"
	"fmt"
	"reflect"
	"sort"
	"strings"
	"totipo/conformance/internal/cryptov1"
	"totipo/conformance/internal/graph"
	"totipo/conformance/internal/object"
	"totipo/conformance/internal/storage"
)

// Resolve references only to existing hash-verified envelope cases. Storage
// cases cannot recurse or supply alternate unchecked fixture paths.
func resolveStorage(cases []Case) error {
	index := map[string]Case{}
	for _, c := range cases {
		index[c.ID] = c
	}
	for i := range cases {
		c := &cases[i]
		if c.Storage == nil {
			continue
		}
		c.fixtures = map[string]Case{}
		paths := map[string]bool{}
		for _, entry := range c.Storage.Entries {
			if paths[entry.Path] {
				return fmt.Errorf("%s duplicate environment path", c.ID)
			}
			paths[entry.Path] = true
			switch entry.Kind {
			case "regular", "directory", "symlink", "fifo", "socket", "device":
			default:
				return fmt.Errorf("invalid entry kind")
			}
			if entry.Size < 0 || entry.Size > 4096 {
				return fmt.Errorf("invalid environment size")
			}
			if entry.Fixture != "" {
				f, ok := index[entry.Fixture]
				if !ok || f.Crypto == nil || (f.Operation != "crypto" && f.Operation != "dispatch") || f.Root != c.Storage.Root || entry.Data != "" || entry.Size != 0 {
					return fmt.Errorf("%s invalid envelope fixture reference %s", c.ID, entry.Fixture)
				}
				c.fixtures[entry.Fixture] = f
			} else if entry.Data != "" && entry.Size != 0 {
				return fmt.Errorf("mixed entry content")
			}
		}
	}
	return nil
}
func runStorage(c Case) error {
	x := c.Storage
	root, e := unhex(x.Root)
	if e != nil {
		return e
	}
	keys, e := cryptov1.Derive(root)
	if e != nil {
		return e
	}
	entries := []storage.Entry{}
	for _, v := range x.Entries {
		entry := v
		entries = append(entries, storage.Entry{Path: entry.Path, Kind: entry.Kind, Read: func() ([]byte, error) {
			if entry.Fixture != "" {
				f, ok := c.fixtures[entry.Fixture]
				if !ok {
					return nil, fmt.Errorf("unresolved fixture")
				}
				return unhex(f.Crypto.Object)
			}
			if entry.Size != 0 {
				return make([]byte, entry.Size), nil
			}
			return unhex(entry.Data)
		}})
	}
	observations, e := storage.Scan(x.NamespaceKind, entries, keys)
	if e != nil {
		return e
	}
	state := graph.New()
	got := StorageExpected{Observations: []StorageObservation{}, Learned: []string{}, Unscoped: []string{}}
	for _, obs := range observations {
		got.Observations = append(got.Observations, StorageObservation{Path: obs.Path, Class: obs.Class})
		if obs.Class != object.Supported && obs.Class != object.Opaque && obs.Class != object.Unscoped {
			continue
		}
		id := strings.TrimPrefix(obs.Path, storage.Namespace+"/")
		n := graph.Node{ID: id, Class: obs.Class, Digest: Hash(obs.Semantic)}
		var value *graph.Value
		if o := obs.Object; o != nil {
			n.Version = o.Version
			n.Identity = hex.EncodeToString(o.Identity)
			n.Author = hex.EncodeToString(o.Author)
			n.AuthorTime = o.AuthorTime
			n.PublicKey = hex.EncodeToString(o.PublicKey)
			n.Type = "TOKEN"
			if o.Type == object.Device {
				n.Type = "DEVICE"
			}
			for _, p := range o.Parents {
				n.Parents = append(n.Parents, hex.EncodeToString(p))
			}
			if obs.Class == object.Supported {
				status := "LIVE"
				if o.Status == 2 {
					status = "TOMBSTONE"
				}
				value = &graph.Value{Status: status, Issuer: o.Issuer, Account: o.Account, Algorithm: o.Algorithm, Digits: o.Digits, Period: o.Period, Secret: hex.EncodeToString(o.Secret), DisplayName: o.DisplayName, Verified: o.Type == object.Device && keys.Provenance(*o, nil) == "VERIFIED"}
			}
		}
		if e := state.Learn(n, value, true); e != nil {
			return e
		}
	}
	for id, n := range state.Nodes {
		got.Learned = append(got.Learned, id)
		if n.Class == object.Unscoped {
			got.Unscoped = append(got.Unscoped, id)
		}
	}
	sort.Strings(got.Learned)
	sort.Strings(got.Unscoped)
	got.Authoritative = state.Authoritative()
	q := x.Query
	got.View = state.Evaluate(q.Identity, q.Candidate, q.Device)
	if !reflect.DeepEqual(got, x.Want) {
		return fmt.Errorf("storage observation/state mismatch: got %+v want %+v", got, x.Want)
	}
	return nil
}

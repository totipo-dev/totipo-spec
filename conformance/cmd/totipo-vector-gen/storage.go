package main

import (
	"os"
	"path/filepath"
	"sort"
	"strings"
	"totipo/conformance/internal/graph"
	"totipo/conformance/internal/object"
	"totipo/conformance/internal/vectors"
)

func fixture(id string) vectors.Case {
	b, e := os.ReadFile(filepath.Join(rootDir, "vectors", casePath(id)))
	must(e)
	var c vectors.Case
	must(vectors.Decode(b, &c))
	return c
}
func storageCases() {
	supported := fixture("v1.crypto.token-root.001")
	opaque := fixture("v1.crypto.future-token-opaque.001")
	device := fixture("v1.crypto.future-device-opaque.001")
	sid, oid, did := supported.Crypto.ObjectID, opaque.Crypto.ObjectID, device.Crypto.ObjectID
	entry := func(namespace string, f vectors.Case) vectors.StorageEntry {
		return vectors.StorageEntry{Path: namespace + "/" + f.Crypto.ObjectID, Kind: "regular", Fixture: f.ID}
	}
	base := entry("objects-v1", supported)
	// Expected results below follow sections 4, 12, 15 and 37 directly. They are
	// never obtained by running storage.Scan or graph.State.Evaluate.
	normal := graph.Result{Heads: []string{sid}, ValueState: "UNAMBIGUOUS", Ordinary: true, Author: true, Candidate: true, CandidateWarning: true}
	query := vectors.Query{Identity: hx(supported.Input.Identity), Candidate: sid}
	want := func(observations []vectors.StorageObservation, learned []string, view graph.Result) vectors.StorageExpected {
		sort.Strings(learned)
		return vectors.StorageExpected{Observations: observations, Learned: learned, Unscoped: []string{}, Authoritative: true, View: view}
	}
	obs := func(e vectors.StorageEntry, class string) vectors.StorageObservation {
		return vectors.StorageObservation{Path: e.Path, Class: class}
	}
	addCase := func(id, notes string, entries []vectors.StorageEntry, q vectors.Query, expected vectors.StorageExpected) {
		add(vectors.Case{ID: id, Operation: "storage", Expected: "PASS", Storage: &vectors.StorageCase{Notes: notes, Root: hx(root), NamespaceKind: "directory", Entries: entries, Query: q, Want: expected}}, "semantic", "4", "12.4–12.7", "15", "37", "50")
	}
	devEntry := entry("objects-v1", device)
	view := normal
	view.Presentation = "OPAQUE"
	q := query
	q.Device = hx(device.Future.Routing.Identity)
	addCase("v1.storage.objects-v1.001", "Discover supported TOKEN and routable future DEVICE only in the exact family namespace. Opaque DEVICE presentation does not block token use.", []vectors.StorageEntry{base, devEntry}, q, want([]vectors.StorageObservation{obs(base, object.Supported), obs(devEntry, object.Opaque)}, []string{sid, did}, view))
	ignored := []vectors.StorageEntry{base,
		{Path: "objects-v1/" + strings.ToUpper(oid), Kind: "regular", Fixture: opaque.ID},
		{Path: "objects-v1/" + oid + ".tmp", Kind: "regular", Fixture: opaque.ID},
		{Path: "objects-v1/" + oid + ".sync-conflict", Kind: "regular", Fixture: opaque.ID},
		{Path: "objects-v1/nested/" + oid, Kind: "regular", Fixture: opaque.ID},
		{Path: "objects-v1/../objects-v1/" + oid, Kind: "regular", Fixture: opaque.ID},
	}
	for i, kind := range []string{"directory", "symlink", "fifo", "socket", "device"} {
		ignored = append(ignored, vectors.StorageEntry{Path: "objects-v1/" + strings.Repeat(string(rune('b'+i)), 64), Kind: kind, Fixture: opaque.ID})
	}
	addCase("v1.storage.nonobject-name-ignored.001", "Noncanonical names, nested paths, symlinks and special entries never enter semantic processing. Entry kinds describe un-followed filesystem observations.", ignored, query, want([]vectors.StorageObservation{obs(base, object.Supported)}, []string{sid}, normal))
	wrong := []vectors.StorageEntry{base}
	observations := []vectors.StorageObservation{obs(base, object.Supported)}
	for i, n := range []int{0, 1023, 1025} {
		e := vectors.StorageEntry{Path: "objects-v1/" + strings.Repeat(string(rune('b'+i)), 64), Kind: "regular", Size: n}
		wrong = append(wrong, e)
		observations = append(observations, obs(e, "INVALID_STORAGE"))
	}
	addCase("v1.storage.wrong-size-not-opaque.001", "Wrong-sized direct candidates are invalid storage observations, never persisted opaque evidence. Existing supported state remains ordinarily and explicitly usable.", wrong, query, want(observations, []string{sid}, normal))
	siblings := []vectors.StorageEntry{base, entry("objects-v999", opaque), entry("objects-v2", supported), {Path: "objects-v999/nested/arbitrary", Kind: "regular", Data: "ffff"}, {Path: "unrelated/file", Kind: "regular", Size: 2048}}
	addCase("v1.storage.unknown-sibling-ignored.001", "Even valid-looking or cryptographically valid bytes in unknown sibling namespaces are ignored. They do not affect graph heads/readiness or add warnings; the explicit-candidate warning is the same as for the supported baseline alone.", siblings, query, want([]vectors.StorageObservation{obs(base, object.Supported)}, []string{sid}, normal))
	richer := vectors.StorageEntry{Path: "objects-v2/future-state", Kind: "regular", Size: 2048}
	shadow := entry("objects-v1", opaque)
	view = graph.Result{Heads: []string{sid, oid}, ValueState: "VALUE_INCOMPLETE_OPAQUE", Candidate: true, CandidateWarning: true}
	sort.Strings(view.Heads)
	addCase("v1.future.family-compat-shadow.001", "The 2048-byte sibling is illustrative richer future-family material, not a v2 format definition. Only the authenticated objects-v1 compatibility assertion enters the graph; its concurrent opaque TOKEN scopes degradation to this token and retains explicit supported candidate use.", []vectors.StorageEntry{base, richer, shadow}, query, want([]vectors.StorageObservation{obs(base, object.Supported), obs(shadow, object.Opaque)}, []string{sid, oid}, view))
	addCase("v1.future.family-no-shadow.001", "Without an objects-v1 compatibility assertion the sibling yields no authenticated future-state evidence and no warning/block. This means that future family is not providing rolling-upgrade compatibility to v1 for that state.", []vectors.StorageEntry{base, richer}, query, want([]vectors.StorageObservation{obs(base, object.Supported)}, []string{sid}, normal))
}

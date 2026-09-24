package main

import (
	"fmt"
	"sort"
)

type E struct {
	Op       string `json:"op"`
	Args     M      `json:"args"`
	Expected M      `json:"expected"`
}
type MC struct {
	ID          string `json:"id"`
	Kind        string `json:"kind"`
	Description string `json:"description"`
	Provenance  P      `json:"provenance"`
	Input       M      `json:"input"`
	Expected    M      `json:"expected"`
	Trace       []E    `json:"trace"`
}

func event(op string, args M, result string, checks M) E {
	if args == nil {
		args = M{}
	}
	want := M{"result": result}
	for k, v := range checks {
		want[k] = v
	}
	return E{op, args, want}
}
func pending(id, scope string) E {
	return event("observe_pending", M{"id": id, "scope": scope}, "AWAITING_DURABILITY", nil)
}
func persist(key string, checks M) E { return event("persist", M{"key": key}, "DURABLE", checks) }
func valid(id, scope, parents string) E {
	return event("valid", M{"id": id, "scope": scope, "parents": parents}, "OBSERVED", nil)
}
func remember(scope, heads string) E {
	return event("remember", M{"scope": scope, "heads": heads}, "AWAITING_DURABILITY", nil)
}
func memoryFixtures() {
	var cases []MC
	add := func(id, sections, binding string, trace ...E) {
		cases = append(cases, MC{"v0/recovery/" + id, "memory", id, P{"review/phase2/REVIEW.md", "spec-derived-reviewed", "r36 " + sections + "; " + id}, M{"binding": binding}, M{"disposition": "VERIFIED"}, trace})
	}
	crash := event("crash", nil, "RESTARTED", nil)
	add("pending-durability-and-locality", "56,63", "v", event("observe_pending", M{"id": "p", "scope": "token:t"}, "AWAITING_DURABILITY", M{"pending:token:t:p": "false", "writer:token:t": "false", "writer:token:u": "true", "staged_count": "1"}), persist("pending:token:t:p", M{"pending:token:t:p": "true", "writer:token:t": "false"}), event("path_failure", M{"id": "p"}, "UNAVAILABLE", M{"pending:token:t:p": "true"}), event("crash", nil, "RESTARTED", M{"pending:token:t:p": "true", "writer:token:t": "false", "writer:token:u": "true"}))
	add("pending-explicit-abandonment", "56.1,63", "v", pending("p", "token:t"), persist("pending:token:t:p", nil), event("abandon", M{"id": "p", "scope": "token:t"}, "AWAITING_DURABILITY", M{"writer:token:t": "false"}), persist("abandon:token:t:p", M{"writer:token:t": "true", "pending:token:t:p": "true", "complete:token:t": "false"}), event("acknowledge", M{"id": "p", "scope": "token:t"}, "AWAITING_DURABILITY", nil), persist("ack:token:t:p", M{"acknowledged:token:t:p": "true"}), crash, event("abandon", M{"id": "missing", "scope": "token:t"}, "UNKNOWN_PENDING", M{"abandoned:token:t:p": "true", "pending:token:t:p": "true", "writer:token:t": "true"}), pending("q", "token:t"), persist("pending:token:t:q", M{"writer:token:t": "false"}))
	add("abandonment-bound-to-vault-and-token", "56.1", "v", pending("p", "token:t"), persist("pending:token:t:p", nil), event("abandon", M{"id": "p", "scope": "token:u"}, "UNKNOWN_PENDING", nil), event("abandon", M{"id": "p", "scope": "token:t", "binding": "wrong"}, "BINDING_MISMATCH", M{"writer:token:t": "false"}))
	add("pending-reentry-handoff", "56.1,63.1", "v", pending("p", "token:t"), persist("pending:token:t:p", nil), event("abandon", M{"id": "p", "scope": "token:t"}, "AWAITING_DURABILITY", nil), persist("abandon:token:t:p", nil), valid("r", "token:t", ""), valid("p", "token:t", "r"), valid("during", "token:t", "r"), event("discard_pending", M{"id": "p", "scope": "token:t"}, "REPLACEMENT_NOT_DURABLE", M{"heads:token:t": "during,p", "pending:token:t:p": "true"}), remember("token:t", "during,p"), persist("heads:token:t", M{"pending:token:t:p": "true"}), event("discard_pending", M{"id": "p", "scope": "token:t"}, "COMPACTED", M{"pending:token:t:p": "false", "abandoned:token:t:p": "false", "writer:token:t": "true", "heads:token:t": "during,p"}))
	add("path-failure-is-not-terminal", "56.2", "v", pending("p", "token:t"), persist("pending:token:t:p", nil), event("path_failure", M{"id": "p", "reason": "wrong-length"}, "UNAVAILABLE", nil), event("path_failure", M{"id": "p", "reason": "AEAD"}, "UNAVAILABLE", nil), event("terminal", M{"id": "p", "scope": "token:t", "identity_grounded": "false"}, "NOT_IDENTITY_GROUNDED", nil), event("discard_pending", M{"id": "p", "scope": "token:t"}, "REPLACEMENT_NOT_DURABLE", M{"pending:token:t:p": "true", "writer:token:t": "false"}))
	add("identity-grounded-terminal-handoff", "56.2,63.1", "v", pending("p", "token:t"), persist("pending:token:t:p", nil), event("terminal", M{"id": "p", "scope": "token:t", "identity_grounded": "true"}, "AWAITING_DURABILITY", nil), event("discard_pending", M{"id": "p", "scope": "token:t"}, "REPLACEMENT_NOT_DURABLE", nil), persist("terminal:token:t:p", nil), event("discard_pending", M{"id": "p", "scope": "token:t"}, "COMPACTED", M{"writer:token:t": "true"}))
	future := func(id string) E {
		return event("observe_future", M{"id": id, "version": "1"}, "AWAITING_DURABILITY", nil)
	}
	add("future-durable-bypass-is-incomplete", "59,63", "v", future("f"), persist("future:f", M{"future:f": "true", "writer:token:t": "false", "writer:token:u": "false"}), event("lose", M{"id": "f"}, "UNAVAILABLE", nil), crash, event("bypass_future", M{"id": "wrong"}, "UNKNOWN_FUTURE", nil), event("bypass_future", M{"id": "f"}, "AWAITING_DURABILITY", M{"writer:token:t": "false"}), persist("bypass:f", M{"writer:token:t": "true", "future:f": "true", "complete:token:t": "false"}), crash, future("g"), persist("future:g", M{"writer:token:t": "false", "bypass:f": "true"}))
	add("future-compatible-durable-handoff", "59,63.1", "v", future("f"), persist("future:f", nil), event("future_handoff", M{"id": "f", "compatible": "false"}, "HANDOFF_REJECTED", nil), event("future_handoff", M{"id": "f", "compatible": "true"}, "AWAITING_DURABILITY", nil), event("clear_future", M{"id": "f"}, "REPLACEMENT_NOT_DURABLE", nil), persist("handoff:f", nil), event("clear_future", M{"id": "f"}, "COMPACTED", M{"future:f": "false", "writer:token:t": "true", "complete:token:t": "true"}))
	base := func() []E {
		return []E{valid("x", "token:t", ""), remember("token:t", "x"), persist("heads:token:t", nil)}
	}
	add("known-history-loss-empty", "64", "v", append(base(), event("lose", M{"id": "x"}, "UNAVAILABLE", M{"degraded:token:t": "true", "writer:token:t": "false", "complete:token:t": "false", "writer:token:u": "true"}), pending("p", "token:t"), persist("pending:token:t:p", nil), event("abandon", M{"id": "p", "scope": "token:t"}, "DEGRADED_KNOWN_HISTORY_LOSS", nil), event("compact_head", M{"id": "x", "scope": "token:t"}, "REPLACEMENT_NOT_DURABLE", M{"remembered:token:t": "x"}))...)
	add("known-history-unrelated-and-restored", "64", "v", append(base(), event("lose", M{"id": "x"}, "UNAVAILABLE", nil), valid("other", "token:t", ""), event("remember", M{"scope": "token:t", "heads": "other"}, "DEGRADED_KNOWN_HISTORY_LOSS", M{"degraded:token:t": "true"}), event("restore", M{"id": "x"}, "OBSERVED", M{"degraded:token:t": "false"}), remember("token:t", "other,x"), persist("heads:token:t", M{"writer:token:t": "true"}))...)
	add("remembered-descendant-no-regression", "63,64", "v", append(base(), valid("h", "token:t", "x"), remember("token:t", "h"), persist("heads:token:t", M{"degraded:token:t": "false", "writer:token:t": "true"}), event("compact_head", M{"id": "x", "scope": "token:t"}, "COMPACTED", M{"remembered:token:t": "h"}))...)
	add("remembered-fork-reunion", "64", "v", valid("a", "token:t", ""), valid("b", "token:t", ""), remember("token:t", "a,b"), persist("heads:token:t", nil), valid("h", "token:t", "a,b"), remember("token:t", "h"), persist("heads:token:t", M{"degraded:token:t": "false"}), event("compact_head", M{"id": "a", "scope": "token:t"}, "COMPACTED", nil), event("compact_head", M{"id": "b", "scope": "token:t"}, "COMPACTED", M{"remembered:token:t": "h"}))
	add("crash-before-head-durability", "63.1", "v", append(base(), valid("h", "token:t", "x"), remember("token:t", "h"), event("compact_head", M{"id": "x", "scope": "token:t"}, "REPLACEMENT_NOT_DURABLE", nil), event("crash", nil, "RESTARTED", M{"remembered:token:t": "x", "writer:token:t": "false"}))...)
	add("crash-after-head-durability", "63.1", "v", append(base(), valid("h", "token:t", "x"), remember("token:t", "h"), persist("heads:token:t", nil), event("crash", nil, "RESTARTED", M{"remembered:token:t": "h,x", "writer:token:t": "true"}), event("compact_head", M{"id": "x", "scope": "token:t"}, "COMPACTED", M{"remembered:token:t": "h"}))...)
	add("non-atomic-batch-partial-flush", "63.1", "v", pending("p", "token:t"), pending("q", "token:u"), persist("pending:token:t:p", nil), persist("pending:token:u:q", nil), valid("p", "token:t", ""), valid("q", "token:u", ""), remember("token:t", "p"), remember("token:u", "q"), persist("heads:token:t", nil), event("discard_pending", M{"id": "p", "scope": "token:t"}, "COMPACTED", nil), event("discard_pending", M{"id": "q", "scope": "token:u"}, "REPLACEMENT_NOT_DURABLE", nil), event("crash", nil, "RESTARTED", M{"pending:token:u:q": "true", "remembered:token:u": "", "remembered:token:t": "p", "writer:token:u": "false", "writer:token:t": "true"}))
	add("device-pending-abandonment-reentry", "49.1,63.1", "v", pending("p", "device:k"), persist("pending:device:k:p", M{"writer:device:k": "false", "writer:token:t": "true"}), event("abandon", M{"id": "p", "scope": "device:k"}, "AWAITING_DURABILITY", nil), persist("abandon:device:k:p", nil), event("crash", nil, "RESTARTED", M{"writer:device:k": "true", "pending:device:k:p": "true"}), valid("r", "device:k", ""), valid("p", "device:k", "r"), valid("during", "device:k", "r"), remember("device:k", "during,p"), persist("heads:device:k", nil), event("discard_pending", M{"id": "p", "scope": "device:k"}, "COMPACTED", M{"heads:device:k": "during,p", "writer:token:t": "true"}))
	add("device-remembered-replacement-and-rollback", "49.1,63.1", "v", valid("r", "device:k", ""), remember("device:k", "r"), persist("heads:device:k", nil), valid("h", "device:k", "r"), remember("device:k", "h"), event("compact_head", M{"id": "r", "scope": "device:k"}, "REPLACEMENT_NOT_DURABLE", nil), persist("heads:device:k", nil), event("compact_head", M{"id": "r", "scope": "device:k"}, "COMPACTED", nil), event("lose", M{"id": "h"}, "UNAVAILABLE", M{"degraded:device:k": "true", "writer:device:k": "false", "writer:token:t": "true"}))
	confirm := event("confirm", M{"scope": "token:t"}, "CONFIRMED", nil)
	add("confirmation-token-context-change", "39,71", "v", confirm, event("context", M{"scope": "token:t", "value": "new-field-heads"}, "OBSERVED", nil), event("publish", M{"scope": "token:t", "id": "write"}, "CONFIRMATION_STALE", M{"published:write": "false"}))
	add("confirmation-pending-change", "39", "v", confirm, pending("p", "token:t"), persist("pending:token:t:p", nil), event("abandon", M{"id": "p", "scope": "token:t"}, "AWAITING_DURABILITY", nil), persist("abandon:token:t:p", nil), event("publish", M{"scope": "token:t", "id": "write"}, "CONFIRMATION_STALE", nil))
	add("confirmation-future-change-return", "39", "v", confirm, future("f"), persist("future:f", nil), event("future_handoff", M{"id": "f", "compatible": "true"}, "AWAITING_DURABILITY", nil), persist("handoff:f", nil), event("clear_future", M{"id": "f"}, "COMPACTED", M{"future:f": "false", "writer:token:t": "true"}), event("publish", M{"scope": "token:t", "id": "write"}, "CONFIRMATION_STALE", nil))
	add("confirmation-lost-on-restart", "39", "v", confirm, crash, event("publish", M{"scope": "token:t", "id": "write"}, "CONFIRMATION_STALE", nil))
	add("observation-before-linearization", "71", "v", confirm, future("f"), event("publish", M{"scope": "token:t", "id": "write"}, "WRITER_BLOCKED", M{"published:write": "false"}))
	add("observation-after-linearization", "39,71", "v", confirm, event("publish", M{"scope": "token:t", "id": "write"}, "LINEARIZED", M{"published:write": "true"}), future("f"), persist("future:f", M{"published:write": "true", "writer:token:t": "false"}))
	begin := event("begin", M{"binding": "a"}, "AWAITING_DURABILITY", nil)
	install := event("install", M{"binding": "a"}, "INSTALLED", nil)
	add("establishment-racing-cas", "8.1,10.1", "", begin, event("begin", M{"binding": "b"}, "CAS_CONFLICT", nil), event("install", M{"binding": "a"}, "BINDING_NOT_DURABLE", nil), persist("establishment", M{"establishment": "PENDING", "open_allowed": "false"}), event("begin", M{"binding": "b"}, "CAS_CONFLICT", nil), install, event("establish", M{"binding": "b"}, "BINDING_MISMATCH", nil), event("establish", M{"binding": "a"}, "AWAITING_DURABILITY", M{"open_allowed": "false"}), persist("established", M{"establishment": "ESTABLISHED", "binding": "a", "open_allowed": "true"}), event("begin", M{"binding": "b"}, "CAS_CONFLICT", nil))
	add("establishment-crash-before-install", "8.1,10.1", "", begin, persist("establishment", nil), event("crash", nil, "RESTARTED", M{"binding": "a", "establishment": "PENDING", "installed": "", "open_allowed": "false"}), event("install", M{"binding": "b"}, "BINDING_NOT_DURABLE", nil), install)
	add("establishment-crash-after-install", "10.1", "", begin, persist("establishment", nil), install, event("establish", M{"binding": "a"}, "AWAITING_DURABILITY", nil), event("crash", nil, "RESTARTED", M{"establishment": "PENDING", "installed": "a", "open_allowed": "false"}), event("establish", M{"binding": "a"}, "AWAITING_DURABILITY", nil), persist("established", M{"open_allowed": "true"}))
	add("establishment-conflicting-existing-vault", "8.1,10.1", "", begin, persist("establishment", nil), event("existing_vault", M{"binding": "b"}, "OBSERVED", nil), event("install", M{"binding": "a"}, "DESTINATION_EXISTS", nil), event("establish", M{"binding": "a"}, "BINDING_MISMATCH", M{"open_allowed": "false", "binding": "a"}))
	trace := []E{}
	heads := ""
	for n := 0; n < 32; n++ {
		id := fmt.Sprintf("h%02d", n)
		trace = append(trace, valid(id, "token:t", ""))
		if heads != "" {
			heads += ","
		}
		heads += id
	}
	trace = append(trace, remember("token:t", heads), persist("heads:token:t", M{"capacity:token:t": "ENCODABLE", "head_count:token:t": "32", "writer:token:t": "true"}), valid("h32", "token:t", ""), remember("token:t", heads+",h32"), persist("heads:token:t", M{"capacity:token:t": "CAPACITY_BLOCKED", "head_count:token:t": "33", "writer:token:t": "false"}), event("migrate", M{"scope": "token:t", "binding": "new-vault", "new_token": "token:fresh", "old_signer": "device-a", "new_signer": "device-new", "selected_values": "LIVE,chosen-issuer,chosen-account,chosen-credential"}, "MIGRATED_COPY", M{"head_count:token:t": "33", "migration:token": "token:fresh", "migration:root_parents": "0", "migration:continuity": "false", "migration:retirement": "false", "migration:erasure": "false"}))
	add("capacity-32-33-and-migration", "65,65.1", "v", trace...)
	sort.Slice(cases, func(i, j int) bool { return cases[i].ID < cases[j].ID })
	write("vectors/v0/recovery/phase2.json", struct {
		Schema int  `json:"schema"`
		Cases  []MC `json:"cases"`
	}{1, cases})
	fmt.Println("authored", len(cases), "memory traces")
}

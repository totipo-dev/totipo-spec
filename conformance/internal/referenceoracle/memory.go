// The memory reviewer uses a flat durable-fact ledger and reconstructs graph
// reachability by transitive closure. It shares event DTOs, not state or helpers,
// with the primary memory simulator.
package referenceoracle

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
	"totipo/conformance/internal/securitymemory"
)

type memoryReview struct {
	facts     map[string]string
	staged    map[string]map[string]string
	objects   map[string]map[string]string
	absent    map[string]bool
	context   map[string]string
	limited   map[string]bool
	confirmed map[string]string
	epoch     int
	published map[string]bool
}

func newMemoryReview(binding string) *memoryReview {
	phase := "UNESTABLISHED"
	if binding != "" {
		phase = "ESTABLISHED"
	}
	return &memoryReview{facts: map[string]string{"binding": binding, "establishment": phase, "installed": ""}, staged: map[string]map[string]string{}, objects: map[string]map[string]string{}, absent: map[string]bool{}, context: map[string]string{}, limited: map[string]bool{}, confirmed: map[string]string{}, published: map[string]bool{}}
}
func split(s string) []string {
	if s == "" {
		return []string{}
	}
	return strings.Split(s, ",")
}
func (r *memoryReview) relation() map[string]map[string]bool {
	ids := []string{}
	rel := map[string]map[string]bool{}
	for id, node := range r.objects {
		ids = append(ids, id)
		rel[id] = map[string]bool{id: true}
		for _, p := range split(node["parents"]) {
			rel[id][p] = true
		}
	}
	for _, k := range ids {
		for _, i := range ids {
			if rel[i][k] {
				for _, j := range ids {
					rel[i][j] = rel[i][j] || rel[k][j]
				}
			}
		}
	}
	return rel
}
func (r *memoryReview) current(scope string) []string {
	ready := map[string]bool{}
	for change := true; change; {
		change = false
		for id, node := range r.objects {
			if ready[id] || r.absent[id] {
				continue
			}
			ok := true
			for _, p := range split(node["parents"]) {
				ok = ok && ready[p]
			}
			if ok {
				ready[id] = true
				change = true
			}
		}
	}
	rel := r.relation()
	heads := []string{}
	for id, node := range r.objects {
		if node["scope"] != scope || !ready[id] {
			continue
		}
		maximal := true
		for other, v := range r.objects {
			if other != id && ready[other] && v["scope"] == scope && rel[other][id] {
				maximal = false
			}
		}
		if maximal {
			heads = append(heads, id)
		}
	}
	sort.Strings(heads)
	return heads
}
func (r *memoryReview) recordIDs(prefix string) []string {
	out := []string{}
	for k := range r.facts {
		if strings.HasPrefix(k, prefix) {
			out = append(out, strings.TrimPrefix(k, prefix))
		}
	}
	sort.Strings(out)
	return out
}
func (r *memoryReview) regression(scope string) bool {
	rel := r.relation()
	for _, old := range r.recordIDs("head:" + scope + ":") {
		covered := false
		for _, h := range r.current(scope) {
			covered = covered || rel[h][old]
		}
		if !covered {
			return true
		}
	}
	return false
}
func (r *memoryReview) writer(scope string) bool {
	if r.facts["establishment"] != "ESTABLISHED" || r.limited[scope] || r.regression(scope) || len(r.current(scope)) > 32 {
		return false
	}
	for _, id := range r.recordIDs("future:") {
		if r.facts["bypass:"+id] == "" {
			return false
		}
	}
	for _, id := range r.recordIDs("pending:" + scope + ":") {
		if r.facts["abandon:"+scope+":"+id] == "" {
			return false
		}
	}
	for _, s := range r.staged {
		if s["type"] == "future" || s["type"] == "establishment" || s["scope"] == scope {
			return false
		}
	}
	for _, h := range r.current(scope) {
		if r.facts["head:"+scope+":"+h] == "" {
			return false
		}
	}
	return true
}
func (r *memoryReview) complete(scope string) bool {
	if r.limited[scope] || r.regression(scope) || len(r.recordIDs("future:")) > 0 || len(r.recordIDs("pending:"+scope+":")) > 0 {
		return false
	}
	for _, s := range r.staged {
		if s["type"] == "future" || s["scope"] == scope {
			return false
		}
	}
	return true
}
func (r *memoryReview) stamp(scope, status string) string {
	return fmt.Sprintf("%s|%s|%s|%s|%s|%s|%d", scope, status, r.context[scope], strings.Join(r.current(scope), ","), strings.Join(r.recordIDs("pending:"+scope+":"), ","), strings.Join(r.recordIDs("abandon:"+scope+":"), ","), r.epoch)
}
func (r *memoryReview) step(op string, a map[string]string) string {
	id, scope := a["id"], a["scope"]
	pair := scope + ":" + id
	binding := a["binding"]
	if binding != "" && binding != r.facts["binding"] {
		switch op {
		case "begin", "install", "establish", "migrate", "existing_vault":
		default:
			return "BINDING_MISMATCH"
		}
	}
	stage := func(k, typ string) string {
		v := map[string]string{}
		for key, value := range a {
			v[key] = value
		}
		v["type"] = typ
		r.staged[k] = v
		return "AWAITING_DURABILITY"
	}
	switch op {
	case "valid":
		if id == "" || scope == "" {
			return "INVALID_EVENT"
		}
		if old, ok := r.objects[id]; ok && (old["scope"] != scope || old["parents"] != a["parents"]) {
			return "IMMUTABLE_ID_CONFLICT"
		}
		r.objects[id] = map[string]string{"scope": scope, "parents": a["parents"]}
		delete(r.absent, id)
		return "OBSERVED"
	case "lose", "path_failure":
		r.absent[id] = true
		return "UNAVAILABLE"
	case "restore":
		if r.objects[id] == nil {
			return "UNKNOWN_ID"
		}
		delete(r.absent, id)
		return "OBSERVED"
	case "observe_pending":
		if id == "" || scope == "" {
			return "INVALID_EVENT"
		}
		return stage("pending:"+pair, "pending")
	case "observe_future":
		if id == "" || a["version"] == "" || a["version"] == "0" {
			return "INVALID_EVENT"
		}
		r.epoch++
		return stage("future:"+id, "future")
	case "abandon":
		if r.facts["pending:"+pair] == "" {
			return "UNKNOWN_PENDING"
		}
		if r.regression(scope) {
			return "DEGRADED_KNOWN_HISTORY_LOSS"
		}
		return stage("abandon:"+pair, "abandon")
	case "acknowledge":
		if r.facts["pending:"+pair] == "" || r.facts["abandon:"+pair] == "" {
			return "UNKNOWN_ABANDONMENT"
		}
		return stage("ack:"+pair, "ack")
	case "bypass_future":
		if r.facts["future:"+id] == "" {
			return "UNKNOWN_FUTURE"
		}
		return stage("bypass:"+id, "bypass")
	case "remember":
		hs := split(a["heads"])
		sort.Strings(hs)
		if strings.Join(hs, ",") != strings.Join(r.current(scope), ",") {
			return "INCOMPLETE_FRONTIER"
		}
		if r.regression(scope) {
			return "DEGRADED_KNOWN_HISTORY_LOSS"
		}
		return stage("heads:"+scope, "heads")
	case "persist":
		key := a["key"]
		s, ok := r.staged[key]
		if !ok {
			return "NOT_STAGED"
		}
		typ := s["type"]
		spair := s["scope"] + ":" + s["id"]
		if (typ == "abandon" || typ == "ack") && r.facts["pending:"+spair] == "" {
			return "UNKNOWN_PENDING"
		}
		if typ == "bypass" && r.facts["future:"+s["id"]] == "" {
			return "UNKNOWN_FUTURE"
		}
		if typ == "established" && (r.facts["establishment"] != "PENDING" || r.facts["installed"] != r.facts["binding"]) {
			return "BINDING_MISMATCH"
		}
		switch typ {
		case "heads":
			for _, h := range split(s["heads"]) {
				r.facts["head:"+s["scope"]+":"+h] = "true"
			}
		case "establishment":
			r.facts["binding"] = s["binding"]
			r.facts["establishment"] = "PENDING"
		case "established":
			r.facts["establishment"] = "ESTABLISHED"
		default:
			r.facts[key] = "true"
			if typ == "future" {
				r.facts[key] = s["version"]
			}
			if typ == "future" || typ == "bypass" {
				r.epoch++
			}
		}
		delete(r.staged, key)
		return "DURABLE"
	case "compact_head":
		if r.facts["head:"+pair] == "" {
			return "UNKNOWN_HEAD"
		}
		rel := r.relation()
		for _, h := range r.recordIDs("head:" + scope + ":") {
			if h != id && rel[h][id] {
				delete(r.facts, "head:"+pair)
				return "COMPACTED"
			}
		}
		return "REPLACEMENT_NOT_DURABLE"
	case "discard_pending":
		if r.facts["pending:"+pair] == "" {
			return "UNKNOWN_PENDING"
		}
		covered := r.facts["terminal:"+pair] != ""
		rel := r.relation()
		for _, h := range r.recordIDs("head:" + scope + ":") {
			covered = covered || (r.objects[id]["scope"] == scope && rel[h][id])
		}
		if !covered {
			return "REPLACEMENT_NOT_DURABLE"
		}
		for _, prefix := range []string{"pending:", "abandon:", "ack:"} {
			delete(r.facts, prefix+pair)
		}
		return "COMPACTED"
	case "terminal":
		if a["identity_grounded"] != "true" {
			return "NOT_IDENTITY_GROUNDED"
		}
		return stage("terminal:"+pair, "terminal")
	case "future_handoff":
		if r.facts["future:"+id] == "" || a["compatible"] != "true" {
			return "HANDOFF_REJECTED"
		}
		return stage("handoff:"+id, "handoff")
	case "clear_future":
		if r.facts["handoff:"+id] == "" {
			return "REPLACEMENT_NOT_DURABLE"
		}
		delete(r.facts, "future:"+id)
		delete(r.facts, "bypass:"+id)
		r.epoch++
		return "COMPACTED"
	case "resource_limit":
		r.limited[scope] = a["active"] == "true"
		return "UPDATED"
	case "context":
		r.context[scope] = a["value"]
		return "OBSERVED"
	case "confirm":
		if !r.writer(scope) {
			return "WRITER_BLOCKED"
		}
		status := a["status"]
		if status == "" {
			status = "LIVE"
		}
		if status != "LIVE" && status != "TOMBSTONE" {
			return "INVALID_EVENT"
		}
		r.confirmed[scope] = r.stamp(scope, status)
		return "CONFIRMED"
	case "publish":
		if !r.writer(scope) {
			return "WRITER_BLOCKED"
		}
		status := a["status"]
		if status == "" {
			status = "LIVE"
		}
		if r.confirmed[scope] != r.stamp(scope, status) {
			return "CONFIRMATION_STALE"
		}
		r.published[id] = true
		delete(r.confirmed, scope)
		return "LINEARIZED"
	case "begin":
		if binding == "" {
			return "INVALID_EVENT"
		}
		if r.facts["establishment"] != "UNESTABLISHED" || r.staged["establishment"] != nil {
			return "CAS_CONFLICT"
		}
		return stage("establishment", "establishment")
	case "install":
		if r.facts["establishment"] != "PENDING" || binding != r.facts["binding"] {
			return "BINDING_NOT_DURABLE"
		}
		if r.facts["installed"] != "" {
			return "DESTINATION_EXISTS"
		}
		r.facts["installed"] = binding
		return "INSTALLED"
	case "existing_vault":
		r.facts["installed"] = binding
		return "OBSERVED"
	case "establish":
		if r.facts["establishment"] != "PENDING" || binding != r.facts["binding"] || r.facts["installed"] != binding {
			return "BINDING_MISMATCH"
		}
		return stage("established", "established")
	case "crash":
		r.staged = map[string]map[string]string{}
		r.confirmed = map[string]string{}
		return "RESTARTED"
	case "migrate":
		values := split(a["selected_values"])
		if binding == "" || binding == r.facts["binding"] || a["new_token"] == "" || a["new_token"] == scope || len(values) != 4 || values[0] != "LIVE" || a["new_signer"] == "" || a["new_signer"] == a["old_signer"] {
			return "MIGRATION_REJECTED"
		}
		for _, o := range r.objects {
			if o["scope"] == a["new_token"] || o["scope"] == "device:"+a["new_signer"] {
				return "MIGRATION_REJECTED"
			}
		}
		for k, v := range map[string]string{"binding": binding, "token": a["new_token"], "signer": a["new_signer"], "root_parents": "0", "selected_values": a["selected_values"], "continuity": "false", "retirement": "false", "erasure": "false"} {
			r.facts["migration:"+k] = v
		}
		return "MIGRATED_COPY"
	}
	return "UNKNOWN_OPERATION"
}
func (r *memoryReview) query(q string) string {
	b := strconv.FormatBool
	switch q {
	case "binding", "establishment", "installed":
		return r.facts[q]
	case "open_allowed":
		return b(r.facts["establishment"] == "ESTABLISHED")
	case "staged_count":
		return strconv.Itoa(len(r.staged))
	}
	part := strings.SplitN(q, ":", 2)
	if len(part) != 2 {
		return "UNKNOWN_QUERY"
	}
	kind, id := part[0], part[1]
	switch kind {
	case "pending", "future", "bypass":
		return b(r.facts[q] != "")
	case "abandoned":
		return b(r.facts["abandon:"+id] != "")
	case "acknowledged":
		return b(r.facts["ack:"+id] != "")
	case "published":
		return b(r.published[id])
	case "writer":
		return b(r.writer(id))
	case "complete":
		return b(r.complete(id))
	case "degraded":
		return b(r.regression(id))
	case "heads":
		return strings.Join(r.current(id), ",")
	case "head_count":
		return strconv.Itoa(len(r.current(id)))
	case "remembered":
		return strings.Join(r.recordIDs("head:"+id+":"), ",")
	case "capacity":
		if len(r.current(id)) > 32 {
			return "CAPACITY_BLOCKED"
		}
		return "ENCODABLE"
	case "migration":
		return r.facts[q]
	}
	return "UNKNOWN_QUERY"
}
func ReviewMemory(binding string, events []securitymemory.Event) error {
	r := newMemoryReview(binding)
	for i, e := range events {
		got := r.step(e.Op, e.Args)
		if got != e.Expected["result"] {
			return fmt.Errorf("reference memory event %d %s: expected %s, got %s", i, e.Op, e.Expected["result"], got)
		}
		for q, want := range e.Expected {
			if q == "result" {
				continue
			}
			if got := r.query(q); got != want {
				return fmt.Errorf("reference memory event %d %s: %s expected %s, got %s", i, e.Op, q, want, got)
			}
		}
	}
	return nil
}

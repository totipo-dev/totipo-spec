// Package securitymemory is an abstract durable-store state machine for r36.
// Object authentication/validation is an explicit precondition of "valid" and
// identity-grounded terminal events. Symbolic IDs are not wire identifiers.
package securitymemory

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
	"totipo/conformance/internal/model"
)

type Event struct {
	Op       string            `json:"op"`
	Args     map[string]string `json:"args"`
	Expected map[string]string `json:"expected"`
}
type object struct {
	scope   string
	parents []string
}
type item struct {
	kind, id, scope, value string
	heads                  []string
}
type confirmation struct {
	fingerprint string
	epoch       uint64
	status      string
}
type Memory struct {
	Binding, Establishment, Installed string
	objects                           map[string]object
	available                         map[string]bool
	pending, abandoned, acknowledged  map[string]string
	future, bypass, terminal, handoff map[string]string
	remembered                        map[string]map[string]bool
	staged                            map[string]item
	context                           map[string]string
	incomplete                        map[string]bool
	confirmations                     map[string]confirmation
	published                         map[string]bool
	epoch                             uint64
	migration                         map[string]string
	MigrationRoot                     *model.Update
}

func New(binding string) *Memory {
	state := "UNESTABLISHED"
	if binding != "" {
		state = "ESTABLISHED"
	}
	return &Memory{Binding: binding, Establishment: state, objects: map[string]object{}, available: map[string]bool{}, pending: map[string]string{}, abandoned: map[string]string{}, acknowledged: map[string]string{}, future: map[string]string{}, bypass: map[string]string{}, terminal: map[string]string{}, handoff: map[string]string{}, remembered: map[string]map[string]bool{}, staged: map[string]item{}, context: map[string]string{}, incomplete: map[string]bool{}, confirmations: map[string]confirmation{}, published: map[string]bool{}, migration: map[string]string{}}
}
func list(s string) []string {
	if s == "" {
		return []string{}
	}
	return strings.Split(s, ",")
}
func sorted(m map[string]bool) []string {
	out := []string{}
	for k, v := range m {
		if v {
			out = append(out, k)
		}
	}
	sort.Strings(out)
	return out
}
func key(scope, id string) string { return scope + ":" + id }
func (m *Memory) ancestor(a, b string) bool {
	todo := []string{b}
	seen := map[string]bool{}
	for len(todo) > 0 {
		x := todo[0]
		todo = todo[1:]
		if x == a {
			return true
		}
		if seen[x] {
			continue
		}
		seen[x] = true
		todo = append(todo, m.objects[x].parents...)
	}
	return false
}
func (m *Memory) reconstructible(id string, path map[string]bool) bool {
	if !m.available[id] || path[id] {
		return false
	}
	o, ok := m.objects[id]
	if !ok {
		return false
	}
	path[id] = true
	defer delete(path, id)
	for _, p := range o.parents {
		if !m.reconstructible(p, path) {
			return false
		}
	}
	return true
}
func (m *Memory) heads(scope string) []string {
	ids := map[string]bool{}
	for id, o := range m.objects {
		if o.scope == scope && m.reconstructible(id, map[string]bool{}) {
			ids[id] = true
		}
	}
	for a := range ids {
		for b := range ids {
			if a != b && m.ancestor(a, b) {
				delete(ids, a)
				break
			}
		}
	}
	return sorted(ids)
}
func (m *Memory) degraded(scope string) bool {
	for x := range m.remembered[scope] {
		covered := false
		for _, h := range m.heads(scope) {
			covered = covered || m.ancestor(x, h)
		}
		if !covered {
			return true
		}
	}
	return false
}
func (m *Memory) blockingPending(scope string) bool {
	for k, s := range m.pending {
		if s == scope && m.abandoned[k] != scope {
			return true
		}
	}
	return false
}
func (m *Memory) complete(scope string) bool {
	if m.incomplete[scope] || m.degraded(scope) || len(m.future) > 0 {
		return false
	}
	for _, s := range m.pending {
		if s == scope {
			return false
		}
	}
	for _, i := range m.staged {
		if i.kind == "future" || i.scope == scope {
			return false
		}
	}
	return true
}
func (m *Memory) Writer(scope string) bool {
	if m.Establishment != "ESTABLISHED" || m.degraded(scope) || m.incomplete[scope] || m.blockingPending(scope) || len(m.heads(scope)) > 32 {
		return false
	}
	for id := range m.future {
		if m.bypass[id] == "" {
			return false
		}
	}
	for _, i := range m.staged {
		if i.kind == "future" || i.scope == scope || i.kind == "establishment" {
			return false
		}
	}
	for _, h := range m.heads(scope) {
		if !m.remembered[scope][h] {
			return false
		}
	}
	return true
}
func (m *Memory) fingerprint(scope string) string {
	parts := []string{scope, m.context[scope], strings.Join(m.heads(scope), ",")}
	for k, s := range m.pending {
		if s == scope {
			parts = append(parts, "pending:"+k)
		}
	}
	for k, s := range m.abandoned {
		if s == scope {
			parts = append(parts, "abandon:"+k)
		}
	}
	sort.Strings(parts)
	return strings.Join(parts, "|")
}

// Apply returns an outcome; rejected operations do not silently alter evidence.
func (m *Memory) Apply(op string, a map[string]string) string {
	id, scope := a["id"], a["scope"]
	k := key(scope, id)
	if binding := a["binding"]; binding != "" && op != "begin" && op != "install" && op != "establish" && op != "migrate" && op != "existing_vault" && binding != m.Binding {
		return "BINDING_MISMATCH"
	}
	switch op {
	case "valid":
		if id == "" || scope == "" {
			return "INVALID_EVENT"
		}
		if old, ok := m.objects[id]; ok && (old.scope != scope || strings.Join(old.parents, ",") != a["parents"]) {
			return "IMMUTABLE_ID_CONFLICT"
		}
		m.objects[id] = object{scope, list(a["parents"])}
		m.available[id] = true
		return "OBSERVED"
	case "path_failure", "lose":
		m.available[id] = false
		return "UNAVAILABLE"
	case "restore":
		if _, ok := m.objects[id]; !ok {
			return "UNKNOWN_ID"
		}
		m.available[id] = true
		return "OBSERVED"
	case "observe_pending":
		if id == "" || scope == "" {
			return "INVALID_EVENT"
		}
		m.staged["pending:"+k] = item{kind: "pending", id: id, scope: scope}
		return "AWAITING_DURABILITY"
	case "observe_future":
		if id == "" || a["version"] == "" || a["version"] == "0" {
			return "INVALID_EVENT"
		}
		m.staged["future:"+id] = item{kind: "future", id: id, value: a["version"]}
		m.epoch++
		return "AWAITING_DURABILITY"
	case "abandon":
		if m.pending[k] != scope {
			return "UNKNOWN_PENDING"
		}
		if m.degraded(scope) {
			return "DEGRADED_KNOWN_HISTORY_LOSS"
		}
		m.staged["abandon:"+k] = item{kind: "abandon", id: id, scope: scope}
		return "AWAITING_DURABILITY"
	case "acknowledge":
		if m.abandoned[k] != scope || m.pending[k] != scope {
			return "UNKNOWN_ABANDONMENT"
		}
		m.staged["ack:"+k] = item{kind: "ack", id: id, scope: scope}
		return "AWAITING_DURABILITY"
	case "bypass_future":
		if m.future[id] == "" {
			return "UNKNOWN_FUTURE"
		}
		m.staged["bypass:"+id] = item{kind: "bypass", id: id}
		return "AWAITING_DURABILITY"
	case "remember":
		heads := list(a["heads"])
		sortedHeads := append([]string{}, heads...)
		sort.Strings(sortedHeads)
		if strings.Join(sortedHeads, ",") != strings.Join(m.heads(scope), ",") {
			return "INCOMPLETE_FRONTIER"
		}
		if m.degraded(scope) {
			return "DEGRADED_KNOWN_HISTORY_LOSS"
		}
		m.staged["heads:"+scope] = item{kind: "heads", scope: scope, heads: heads}
		return "AWAITING_DURABILITY"
	case "persist":
		i, ok := m.staged[a["key"]]
		if !ok {
			return "NOT_STAGED"
		}
		ik := key(i.scope, i.id)
		if (i.kind == "abandon" || i.kind == "ack") && m.pending[ik] != i.scope {
			return "UNKNOWN_PENDING"
		}
		if i.kind == "bypass" && m.future[i.id] == "" {
			return "UNKNOWN_FUTURE"
		}
		if i.kind == "established" && (m.Establishment != "PENDING" || m.Installed != m.Binding) {
			return "BINDING_MISMATCH"
		}
		switch i.kind {
		case "pending":
			m.pending[ik] = i.scope
		case "abandon":
			m.abandoned[ik] = i.scope
		case "ack":
			m.acknowledged[ik] = i.scope
		case "future":
			m.future[i.id] = i.value
			m.epoch++
		case "bypass":
			m.bypass[i.id] = "true"
			m.epoch++
		case "heads":
			if m.remembered[i.scope] == nil {
				m.remembered[i.scope] = map[string]bool{}
			}
			for _, h := range i.heads {
				m.remembered[i.scope][h] = true
			}
		case "terminal":
			m.terminal[ik] = i.scope
		case "handoff":
			m.handoff[i.id] = "true"
		case "establishment":
			m.Binding = i.value
			m.Establishment = "PENDING"
		case "established":
			m.Establishment = "ESTABLISHED"
		}
		delete(m.staged, a["key"])
		return "DURABLE"
	case "compact_head":
		if !m.remembered[scope][id] {
			return "UNKNOWN_HEAD"
		}
		for h := range m.remembered[scope] {
			if h != id && m.ancestor(id, h) {
				delete(m.remembered[scope], id)
				return "COMPACTED"
			}
		}
		return "REPLACEMENT_NOT_DURABLE"
	case "discard_pending":
		if m.pending[k] != scope {
			return "UNKNOWN_PENDING"
		}
		covered := m.terminal[k] == scope
		for h := range m.remembered[scope] {
			covered = covered || (m.objects[id].scope == scope && m.ancestor(id, h))
		}
		if !covered {
			return "REPLACEMENT_NOT_DURABLE"
		}
		delete(m.pending, k)
		delete(m.abandoned, k)
		delete(m.acknowledged, k)
		return "COMPACTED"
	case "terminal":
		if a["identity_grounded"] != "true" {
			return "NOT_IDENTITY_GROUNDED"
		}
		m.staged["terminal:"+k] = item{kind: "terminal", id: id, scope: scope}
		return "AWAITING_DURABILITY"
	case "future_handoff":
		if m.future[id] == "" || a["compatible"] != "true" {
			return "HANDOFF_REJECTED"
		}
		m.staged["handoff:"+id] = item{kind: "handoff", id: id}
		return "AWAITING_DURABILITY"
	case "clear_future":
		if m.handoff[id] == "" {
			return "REPLACEMENT_NOT_DURABLE"
		}
		delete(m.future, id)
		delete(m.bypass, id)
		m.epoch++
		return "COMPACTED"
	case "resource_limit":
		m.incomplete[scope] = a["active"] == "true"
		return "UPDATED"
	case "context":
		m.context[scope] = a["value"]
		return "OBSERVED"
	case "confirm":
		if !m.Writer(scope) {
			return "WRITER_BLOCKED"
		}
		status := a["status"]
		if status == "" {
			status = "LIVE"
		}
		if status != "LIVE" && status != "TOMBSTONE" {
			return "INVALID_EVENT"
		}
		m.confirmations[scope] = confirmation{m.fingerprint(scope), m.epoch, status}
		return "CONFIRMED"
	case "publish":
		c, ok := m.confirmations[scope]
		if !m.Writer(scope) {
			return "WRITER_BLOCKED"
		}
		status := a["status"]
		if status == "" {
			status = "LIVE"
		}
		if !ok || c.epoch != m.epoch || c.fingerprint != m.fingerprint(scope) || c.status != status {
			return "CONFIRMATION_STALE"
		}
		m.published[id] = true
		delete(m.confirmations, scope)
		return "LINEARIZED"
	case "begin":
		binding := a["binding"]
		if binding == "" {
			return "INVALID_EVENT"
		}
		if m.Establishment != "UNESTABLISHED" || m.staged["establishment"].kind != "" {
			return "CAS_CONFLICT"
		}
		m.staged["establishment"] = item{kind: "establishment", value: binding}
		return "AWAITING_DURABILITY"
	case "install":
		if m.Establishment != "PENDING" || m.Binding != a["binding"] {
			return "BINDING_NOT_DURABLE"
		}
		if m.Installed != "" {
			return "DESTINATION_EXISTS"
		}
		m.Installed = a["binding"]
		return "INSTALLED"
	case "existing_vault":
		m.Installed = a["binding"]
		return "OBSERVED"
	case "establish":
		if m.Establishment != "PENDING" || a["binding"] != m.Binding || m.Installed != m.Binding {
			return "BINDING_MISMATCH"
		}
		m.staged["established"] = item{kind: "established"}
		return "AWAITING_DURABILITY"
	case "crash":
		m.staged = map[string]item{}
		m.confirmations = map[string]confirmation{}
		return "RESTARTED"
	case "migrate":
		values := list(a["selected_values"])
		if a["binding"] == "" || a["binding"] == m.Binding || a["new_token"] == "" || a["new_token"] == scope || len(values) != 4 || values[0] != "LIVE" || a["new_signer"] == "" || a["new_signer"] == a["old_signer"] {
			return "MIGRATION_REJECTED"
		}
		for _, o := range m.objects {
			if o.scope == a["new_token"] || o.scope == "device:"+a["new_signer"] {
				return "MIGRATION_REJECTED"
			}
		}
		m.MigrationRoot = &model.Update{ID: "new-root", Token: a["new_token"], Parents: []string{}, Fields: map[string]string{"STATUS": "LIVE", "ISSUER": values[1], "ACCOUNT": values[2], "CREDENTIAL": values[3]}}
		m.migration = map[string]string{"binding": a["binding"], "token": a["new_token"], "signer": a["new_signer"], "root_parents": "0", "selected_values": a["selected_values"], "continuity": "false", "retirement": "false", "erasure": "false"}
		return "MIGRATED_COPY"
	default:
		return "UNKNOWN_OPERATION"
	}
}

// Query exposes exact durable facts and writer observations for fixture assertions.
func (m *Memory) Query(q string) string {
	b := func(v bool) string { return strconv.FormatBool(v) }
	switch q {
	case "binding":
		return m.Binding
	case "establishment":
		return m.Establishment
	case "installed":
		return m.Installed
	case "open_allowed":
		return b(m.Establishment == "ESTABLISHED")
	case "staged_count":
		return strconv.Itoa(len(m.staged))
	}
	parts := strings.SplitN(q, ":", 2)
	if len(parts) != 2 {
		return "UNKNOWN_QUERY"
	}
	kind, id := parts[0], parts[1]
	switch kind {
	case "pending":
		return b(m.pending[id] != "")
	case "abandoned":
		return b(m.abandoned[id] != "")
	case "acknowledged":
		return b(m.acknowledged[id] != "")
	case "future":
		return b(m.future[id] != "")
	case "bypass":
		return b(m.bypass[id] != "")
	case "published":
		return b(m.published[id])
	case "writer":
		return b(m.Writer(id))
	case "complete":
		return b(m.complete(id))
	case "degraded":
		return b(m.degraded(id))
	case "heads":
		return strings.Join(m.heads(id), ",")
	case "head_count":
		return strconv.Itoa(len(m.heads(id)))
	case "remembered":
		return strings.Join(sorted(m.remembered[id]), ",")
	case "capacity":
		if len(m.heads(id)) > 32 {
			return "CAPACITY_BLOCKED"
		}
		return "ENCODABLE"
	case "migration":
		return m.migration[id]
	}
	return "UNKNOWN_QUERY"
}
func Run(binding string, events []Event) error {
	m := New(binding)
	for i, e := range events {
		got := m.Apply(e.Op, e.Args)
		want, ok := e.Expected["result"]
		if !ok || got != want {
			return fmt.Errorf("event %d %s: expected result %s, actual %s", i, e.Op, want, got)
		}
		for query, want := range e.Expected {
			if query == "result" {
				continue
			}
			if got := m.Query(query); got != want {
				return fmt.Errorf("event %d %s: %s expected %s, actual %s", i, e.Op, query, want, got)
			}
		}
	}
	return nil
}

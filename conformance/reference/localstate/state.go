// Package localstate contains only local evidence, never synchronized secrets.
package localstate

import "totipo/conformance/reference/fault"

type State struct {
	Location      string
	Binding       string
	Establishment string
	Generation    uint64
	Heads         map[string][]string
	Pending       map[string]string // ID -> token:<id> or device:<public key>
	Abandoned     map[string]bool
	Acknowledged  map[string]bool
	Future        map[string]byte
	Bypass        map[string]bool
	Terminal      map[string]bool
}

func Empty() State {
	return State{Heads: map[string][]string{}, Pending: map[string]string{}, Abandoned: map[string]bool{}, Acknowledged: map[string]bool{}, Future: map[string]byte{}, Bypass: map[string]bool{}, Terminal: map[string]bool{}}
}

type Transaction interface {
	State() *State
	Commit(operation string) error
}

// With serializes the entire callback across threads/processes; Commit may be
// called repeatedly for explicitly non-atomic durability handoffs.
type Store interface {
	With(func(Transaction) error) error
}
type Options struct{ Hook fault.Hook }

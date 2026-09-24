package corpus

import (
	"fmt"
	"strings"
)

// Trace schema validation is about artifact shape, not protocol outcomes.
func validateTrace(e TraceEvent) error {
	required, optional := "", "binding"
	switch e.Op {
	case "valid":
		required = "id scope parents"
	case "lose", "restore":
		required = "id"
	case "path_failure":
		required = "id"
		optional += " reason"
	case "observe_pending", "abandon", "acknowledge", "compact_head", "discard_pending":
		required = "id scope"
	case "observe_future":
		required = "id version"
	case "bypass_future", "clear_future":
		required = "id"
	case "remember":
		required = "scope heads"
	case "persist":
		required = "key"
	case "terminal":
		required = "id scope identity_grounded"
	case "future_handoff":
		required = "id compatible"
	case "resource_limit":
		required = "scope active"
	case "context":
		required = "scope value"
	case "confirm":
		required = "scope"
		optional += " status"
	case "publish":
		required = "scope id"
		optional += " status"
	case "begin", "install", "existing_vault", "establish":
		required = "binding"
		optional = ""
	case "crash":
		optional = ""
	case "migrate":
		required = "scope binding new_token old_signer new_signer selected_values"
		optional = ""
	default:
		return fmt.Errorf("unknown trace operation %s", e.Op)
	}
	if err := fields(e.Args, required, optional); err != nil {
		return err
	}
	if id, ok := e.Args["id"]; ok && (id == "" || strings.ContainsAny(id, ":,")) {
		return fmt.Errorf("invalid symbolic event ID")
	}
	for k := range e.Expected {
		if k == "result" {
			continue
		}
		switch k {
		case "binding", "establishment", "installed", "open_allowed", "staged_count":
			continue
		}
		parts := strings.SplitN(k, ":", 2)
		if len(parts) != 2 || parts[1] == "" {
			return fmt.Errorf("unknown checkpoint query %s", k)
		}
		switch parts[0] {
		case "pending", "abandoned", "acknowledged", "future", "bypass", "published", "writer", "complete", "degraded", "heads", "head_count", "remembered", "capacity", "migration":
		default:
			return fmt.Errorf("unknown checkpoint query %s", k)
		}
	}
	return nil
}

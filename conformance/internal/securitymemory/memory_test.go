package securitymemory

import (
	"testing"
	"totipo/conformance/internal/model"
)

func TestEstablishmentRechecksCanonicalBinding(t *testing.T) {
	m := New("")
	steps := []Event{{Op: "begin", Args: map[string]string{"binding": "a"}}, {Op: "persist", Args: map[string]string{"key": "establishment"}}, {Op: "install", Args: map[string]string{"binding": "a"}}, {Op: "establish", Args: map[string]string{"binding": "a"}}, {Op: "existing_vault", Args: map[string]string{"binding": "b"}}}
	for _, e := range steps {
		m.Apply(e.Op, e.Args)
	}
	if got := m.Apply("persist", map[string]string{"key": "established"}); got != "BINDING_MISMATCH" || m.Query("open_allowed") != "false" {
		t.Fatal("stale establishment became ordinary state")
	}
}
func TestConfirmationBindsDesiredStatus(t *testing.T) {
	m := New("v")
	m.Apply("confirm", map[string]string{"scope": "token:t", "status": "LIVE"})
	if got := m.Apply("publish", map[string]string{"scope": "token:t", "status": "TOMBSTONE", "id": "x"}); got != "CONFIRMATION_STALE" {
		t.Fatal(got)
	}
}
func TestMigrationMaterializesNewRoot(t *testing.T) {
	m := New("v")
	args := map[string]string{"scope": "token:old", "new_token": "token:new", "binding": "fresh", "old_signer": "old-key", "new_signer": "new-key", "selected_values": "LIVE,issuer,account,credential"}
	if got := m.Apply("migrate", args); got != "MIGRATED_COPY" {
		t.Fatal(got)
	}
	v, e := model.Evaluate([]model.Update{*m.MigrationRoot})
	if e != nil || v.Validation["new-root"] != "FULLY_VALID" || len(m.MigrationRoot.Parents) != 0 || m.MigrationRoot.Token == "token:old" {
		t.Fatal("invalid destination root")
	}
	args["new_signer"] = "old-key"
	if m.Apply("migrate", args) != "MIGRATION_REJECTED" {
		t.Fatal("reused source device key")
	}
	args["new_signer"] = "new-key"
	args["selected_values"] = "TOMBSTONE,i,a,c"
	if m.Apply("migrate", args) != "MIGRATION_REJECTED" {
		t.Fatal("copied tombstone into migration")
	}
}

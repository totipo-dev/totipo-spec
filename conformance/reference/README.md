# Linux reference integration

This is Layer C: real encrypted bytes, real local files, and process restart tests. Layer A (`vectors/v0/`) remains the 388 reviewed, language-neutral cases. Layer B (`internal/`) remains the portable byte/crypto and semantic/reference models. Nothing observed on Linux is promoted into normative vectors.

The public integration packages are `client`, `storage`, `localstate`, and `fault`. Linux adapters use build constraints; there is no macOS or Windows storage adapter. Portable interfaces/client code compile on those platforms, but that is not platform-storage evidence.

## Using the package

Create a private vault directory and a **separate, trusted 0700 local-state directory**. The local directory must be outside the synchronized namespace; use one stable directory for each configured vault. Open `storage.Open(vaultPath)` and `localstate.Open(localPath)`, then `client.New(objects, security, canonicalConfiguredLocation)`. The location string must be stable and unique (use the canonical absolute configuration path, not a changing alias).

- `Create(password)` generates a root and bootstrap, durably pins PENDING, installs VAULT without replacement, reopens it, and durably completes establishment.
- `Open(password, device)` validates and pins an existing root or finishes matching PENDING recovery. A supplied `Device` is bound to its original root. Omitting it generates a fresh signing identity after establishment.
- `Scan()` re-reads and authenticates all candidates and returns dispositions, token/presentation state, degradation, durable pending/future evidence, and aggregate completeness. A returned error means the view must not be used as accepted state.
- `Prepare(scope, fields)` records the accepted context and session/generation fingerprint. Scope is `token:<64-hex-token-id>` or `device:<64-hex-public-key>`. Fields use the Phase 2 model names; credentials are hex of the complete canonical nested TLV. `Operation.Confirm()` stands for explicit human confirmation of the prepared lifecycle status and context; there is no UI here.
- `Publish(operation)` checks freshness and all writer gates while holding the process lock, constructs/signs/encrypts the canonical update, publishes it, reads it back through validation, and persists acceptance. **An error is never an acknowledgement**, even if the returned ID or immutable file already exists.
- `CreateToken(fields)` uses a fresh CSPRNG token ID. `NewCredential()` generates a 20-byte secret in a complete SHA-1/6-digit/30-second credential. `AssessUse(token)` applies the existing Phase 2 ordinary-use policy separately from consensus validity.
- `Recovery(exactIDs, "abandon" | "acknowledge" | "bypass")` persists explicitly selected decisions. Acknowledgement is not abandonment; neither abandonment nor bypass erases observations. New IDs need new decisions.
- `Rewrap(newPassword)` only wraps the established unchanged root. `Migrate(destination, password, selectedFields)` materializes locally selected complete LIVE values in a new vault with a new device and token. It writes nothing to the source and makes no continuity, retirement, or erasure claim.

The reference caller owns device-key storage. `Device()` returns sensitive signing material for a trusted caller; no plaintext signing key or root is written into the security database. A product needs a protected device-key store and credential UI. The API is not a promise of concurrent mutation safety for a single Client or its configuration: callers serialize session setup/options/confirmation. Separate configured Client instances/processes coordinate through the state-store lock.

## Ingestion and semantic integration

The Linux reader anchors directory handles, opens candidate inodes without following symlinks, checks regular-file type before a normal read, and reads at most the bound plus one byte. The existing `objectcrypto.Open` enforces file length, AEAD, deterministic envelope, recomputed OBJECT_ID, and common-prefix dispatch. The v0 parser and strict signature verifier then run before IDs/fields enter the models. Future tails stay opaque.

Token evaluation runs through both the primary model and independent reference oracle and requires exact agreement. DEVICE_UPDATE uses the existing presentation evaluator, with identity-grounded invalid/wrong-kind parents represented as invalid dependencies. Missing/untrustworthy dependencies remain unavailable. Every scan re-reads bytes, so a same-size/same-mtime change is reconsidered. Invalid garbage does not itself acquire pending or regression authority.

Pending and future evidence is committed before a scan completes or an excluding frontier is accepted. Accepted token and presentation frontiers are persisted next. Missing remembered heads are retained conservatively and surface degradation. A separate handoff commit removes pending records only after accepted-frontier persistence, or atomically records terminal invalidity with removal. Mutable-path failure is never terminal-invalidity proof.

An optional `Client.Cache` stores retained encrypted, validated immutable objects in a trusted store outside synchronization. Restart revalidates those bytes through the same pipeline. All other derived state is rebuilt; there is no persistent semantic cache.

The default candidate budget is 4096, configurable with `MaxCandidates`. Exceeding it returns an incomplete-validation error before accepting a partial result. It is an implementation limit, not a protocol limit. It does not bound directory enumeration or total model CPU; this deliberately literal implementation is not a hardened hostile-input resource scheduler.

Future evidence remains visible after bypass and even after disappearance. Aggregate `View.Complete` remains false. Pending abandonment/future bypass can permit selected writes, while `AssessUse` conservatively refuses ordinary current-token use with unresolved applicable evidence.

## Durable local state

The adapter uses a small JSON snapshot with **no new database dependency**. JSON is a local implementation format, never a protocol/vector format. It contains the configured location, exact r36 HMAC root binding, establishment state, monotonic generation, remembered token/presentation heads, pending observations, abandonment/acknowledgement, future versions/bypass, and terminal-handoff facts.

Each `Store.With` opens and exclusively `flock`s a stable `lock` inode, reloads the current snapshot, and holds the lock across the callback. Independent opens coordinate processes and goroutines; the lock file is never replaced. All cooperating clients for a configured vault must use the same lock domain. A callback may commit multiple times to make ordering explicit.

A commit writes a random exclusive 0600 temporary file, writes the complete snapshot, fsyncs the file, renames it over `state.json`, and fsyncs the directory. The snapshot is one atomic replacement; a shared flush is not being mistaken for a transaction. Readers finish the directory fsync of a recovered visible snapshot before relying on it. The 16 MiB snapshot bound is checked before writing; capacity exhaustion fails without discarding old evidence.

The directory must be private and integrity-protected from the sync attacker. No synchronization provider, hostile same-user local process, root, or backup tool may alter it during normal operation. It contains no root, password, plaintext credential, or private signing key, but IDs and recovery metadata are still sensitive local metadata. Disk encryption/backup protection remain deployment responsibilities.

**Restoring an old local snapshot loses later evidence.** The rollback integration test demonstrates this explicitly. There is no hardware monotonic counter, trusted remote witness, or continuity guarantee across local database restoration/deletion. Opening an old snapshot cannot detect an external rollback that removed all evidence of its successor. A product must disclose/reset its assurance accordingly.

## Linux primitives and assumptions

Supported evidence is Linux with procfs, regular local files, working advisory locks, and filesystem support for the following primitives:

| Property | Primitive/assumption |
|---|---|
| Component safety | Walk absolute directory components with `openat(O_DIRECTORY | O_NOFOLLOW | O_CLOEXEC)`; keep handles pinned |
| Safe regular-file reads | `openat(O_PATH | O_NOFOLLOW)`, `fstat` regular type, reopen the pinned inode through `/proc/self/fd`, bounded read |
| Temporary creation | CSPRNG-named exclusive `mkdirat` staging directory (0700), then `openat(O_CREAT | O_EXCL | O_NOFOLLOW)` candidate (0600) |
| Immutable/initial VAULT install | `renameat2(RENAME_NOREPLACE)` from same-filesystem staging directory |
| Same-root VAULT replacement | `renameat2(..., 0)` after the client validates the established root |
| Local-state replacement | `renameat` of complete fsynced snapshot |
| Required barriers | File `fsync`, then final directory `fsync`; failures propagate |
| Cross-process ordering | Exclusive `flock` on the stable local lock inode |

The Linux manual documents [no-replace and replacement rename semantics](https://man7.org/linux/man-pages/man2/rename.2.html), [O_PATH/no-follow and descriptor-based access](https://man7.org/linux/man-pages/man2/open.2.html), [file versus directory fsync](https://man7.org/linux/man-pages/man2/fsync.2.html), and [flock’s advisory locking semantics](https://man7.org/linux/man-pages/man2/flock.2.html). Unsupported primitives fail; there is no overwrite/copy fallback.

Temporary bytes are read back and compared to the validated input before installation. Final bytes are checked again before storage success. Existing immutable bytes must be ordinary, identical, and fsynced before idempotent success; conflicting/unsafe entries are never overwritten. Only client APIs validate crypto/root identity; raw storage methods intentionally know only names, file types, bounds, and durability.

Anchored handles do not follow later parent/path replacement. Operations target the acquired directory identity; they do not jump to an attacker-supplied replacement. Symlinks, FIFOs, directories, and sockets are rejected without opening their targets or waiting on them. Staging-name and candidate-name symlink collisions fail safely. Another malicious local process with the same UID/root and access to private staging/state is outside the trust model. Filesystem providers that do not honor these Linux local semantics are unsupported.

Interrupted operations may leave noncanonical staging directories/temporary files. Discovery ignores them; no unsafe automatic garbage collection is implemented. The source staging directory need not survive for canonical recovery; final directory durability and already-fsynced file data are the required publication barriers.

## Publication linearization

All scan acceptance, recovery decisions, preparation snapshots, and publication run under the same local state lock. Publication rescans and checks establishment, accepted context, pending/bypass policy, regression, work budget, capacity, session, generation, and confirmation fingerprint while holding it. The linearization point is the successful final gate check entering the locked install sequence. Acceptance of a competing observation cannot interleave before installation. File bytes arriving after that check are later observations; the post-publication scan can accept them and expose a new conflict/block.

The fingerprint includes scope, desired fields/status, token/field heads, lifecycle witnesses, recovery evidence, and future context. A monotonic durable generation stales confirmations even across a future-context change that later looks equivalent. A random session identity rejects confirmations after restart. Global serialization/freshness is conservative: unrelated changes can require a retry, and different tokens are serialized even though the protocol permits more concurrency.

## Testing and limits

Integration tests are Linux-only and separately named `TestObjectIngest...`, `TestCrash...`, `TestProcess...`, `TestFilesystem...`, `TestMigration...`, and related policy/reconstruction tests. The Phase 3 report maps these to requested integration categories.

Fault hooks identify operations and stages (`before-write`, `during-write`, `before-file-sync`, `after-file-sync`, `after-install`, `after-directory-sync`, plus publication handoffs). Crash helpers call `os.Exit(86)` without defers, after which the parent opens fresh adapters/clients and checks the filesystem/database. Real process races use release barriers and bounded subprocess timeouts. Error-only hooks are used for hostile-entry injection, not as crash evidence.

Process death preserves the running kernel and its cache. Tests distinguish calls completed before death but cannot infer persistence after actual power loss from a visible unsynced rename. The final suite uses ext4, and reports those assumptions without claiming a power-loss proof. macOS/Windows adapters, remote CI evidence, production secret/UI integration, network-filesystem durability, and external security audit remain outside demonstrated coverage.

Reproduce from the repository root:

```sh
mkdir -m 700 -p .phase3-test-tmp
TMPDIR="$PWD/.phase3-test-tmp" go -C conformance test -count=1 -timeout=180s ./...
TMPDIR="$PWD/.phase3-test-tmp" go -C conformance test -race -count=1 -timeout=300s ./...
go run ./conformance/cmd/totipo-conformance ./vectors/v0
```

A sandbox that prohibits Unix sockets needs permission for the hostile-socket test. CI workflow files separate portable matrix, Linux integration artifacts, and race-detector jobs; no remote execution is implied by their presence.

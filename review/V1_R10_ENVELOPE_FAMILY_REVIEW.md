# Totipo v1/r10 — Envelope-Family Namespace Review

## Change

r10 separates two concepts that r9 still partially conflated:

```text
semantic version:
    OBJECT_VERSION inside the v1 envelope family

envelope/storage family:
    objects-v1/ and its fixed filename/size/crypto/routing contract
```

The v1 family remains:

```text
objects-v1/
64-hex OBJECT_ID filenames
1024-byte object files
v1-family object crypto
frozen routable prefix
```

A future semantic version can remain in this family and be `OPAQUE_ROUTABLE`.

A future design that needs larger/variable-size objects or a fundamentally different layout becomes a new envelope family in a separate namespace.

## Why unknown sibling namespaces are ignored

A storage provider is hostile.

Therefore:

```text
objects-v2/
objects-v999/
```

cannot themselves be security evidence.

Otherwise the provider could manufacture future-version warnings or authoritative-operation denial merely by creating a directory.

v1 security-relevant future evidence remains authenticated content in `objects-v1/`.

## Rolling compatibility contract

A future family that claims rolling-upgrade interoperability with v1 maintains a compatibility projection in `objects-v1/`.

The compatibility assertion may be:

- an ordinary supported v1 object if the future state is exactly v1-representable; or
- an opaque-routable future semantic version inside the v1 family.

The old v1 client does not need to understand or locate the real future-family object.

Thus a future larger object can coexist conceptually as:

```text
objects-v1/<compatibility assertion>   # 1024 B, old clients route this
objects-v2/<future-family state>       # arbitrary v2 representation
```

## Wrong-size review finding

A wrong-size file inside `objects-v1/` remains invalid storage evidence.

It does not become `OPAQUE_UNSCOPED`, because the client has not authenticated any v1-family envelope.

This avoids allowing unauthenticated junk to create a durable protocol block.

The review concern is handled instead by explicitly defining a larger future representation as another family and requiring a v1-family compatibility assertion when rolling compatibility is claimed.

## Semantic impact

No TOKEN/DEVICE semantic byte layout changes from r9.

No crypto-domain changes.

No routing-prefix changes.

No object-size change.

The storage path changes from:

```text
objects/
```

to:

```text
objects-v1/
```

and future-family coexistence semantics are now explicit.

## Verdict

The namespace split is simpler than forcing all future Totipo representations to retain a 1024-byte physical object format.

It also preserves the core philosophy:

> Semantic uncertainty degrades capabilities; unauthenticated filesystem names do not become trusted protocol facts.

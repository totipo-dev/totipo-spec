#!/usr/bin/env python3
"""Cross-consume Java R31Review pinned vectors with the independent Python r31 TLV/envelope validator.

This intentionally does NOT verify Ed25519/HMAC/HKDF/AES-GCM. It verifies that bytes emitted
by the Java implementation are accepted and interpreted identically at the canonical TLV and
authenticated-envelope plaintext layers by the independent Python implementation.
"""
from __future__ import annotations
import importlib.util
import sys
from pathlib import Path

HERE = Path(__file__).resolve().parent
PY_IMPL = HERE / "totp-vault-v0-r31-tlv-corpus.py"
VECTORS = HERE / "r31-object-vectors.txt"

spec = importlib.util.spec_from_file_location("r31_corpus", PY_IMPL)
mod = importlib.util.module_from_spec(spec)
sys.modules[spec.name] = mod
assert spec.loader is not None
spec.loader.exec_module(mod)


def load_vectors(path: Path):
    sections = {}
    current = None
    for raw in path.read_text(encoding="utf-8").splitlines():
        line = raw.strip()
        if not line or line.startswith("#"):
            continue
        if line.startswith("[") and line.endswith("]"):
            current = line[1:-1]
            if current in sections:
                raise AssertionError(f"duplicate section {current}")
            sections[current] = {}
            continue
        if current is None or "=" not in line:
            raise AssertionError(f"malformed vector line: {raw!r}")
        k, v = line.split("=", 1)
        if k in sections[current]:
            raise AssertionError(f"duplicate field {current}.{k}")
        try:
            sections[current][k] = bytes.fromhex(v)
        except ValueError as e:
            raise AssertionError(f"non-hex field {current}.{k}") from e
    return sections


def check_fixture(name: str, f: dict[str, bytes]) -> list[str]:
    checks = []
    signed = f["signed_plaintext"]
    unsigned = f["unsigned"]
    signature = f["signature"]
    envelope = f["envelope_plaintext"]

    expected_type = "TOKEN_UPDATE" if name == "TOKEN_UPDATE" else "DEVICE_UPDATE"
    got = mod.validate_semantic(signed, form="signed")
    assert got == expected_type, (name, got)
    checks.append("Python accepts Java signed_plaintext as canonical signed " + expected_type)

    got_u = mod.validate_semantic(unsigned, form="unsigned")
    assert got_u == expected_type, (name, got_u)
    checks.append("Python accepts Java unsigned as canonical unsigned " + expected_type)

    items = mod.parse_tlvs(signed)
    assert items and items[-1].tag == mod.T_SIGNATURE
    assert items[-1].value == signature
    assert items[-1].end == len(signed)
    assert signed[:items[-1].start] == unsigned
    assert len(signed) - len(unsigned) == 4 + 64
    checks.append("Java unsigned equals signed bytes with complete terminal FF01/64-byte TLV omitted")

    recovered = mod.validate_envelope_plaintext(envelope)
    assert recovered == signed
    assert int.from_bytes(envelope[:2], "big") == len(signed)
    assert envelope[2:2+len(signed)] == signed
    assert not any(envelope[2+len(signed):])
    checks.append("Python accepts Java envelope_plaintext and extracts exact signed_plaintext with zero padding")

    # Internal fixture relationships that require no cryptographic implementation.
    assert len(f["object_id"]) == 32
    assert len(f["nonce"]) == 12
    assert f["nonce"] == f["object_id"][:12]
    assert f["aad"] == b"TOTP-Vault/v0/object" + f["object_id"]
    assert len(f["ciphertext"]) == mod.ENCRYPTION_PLAINTEXT_LENGTH
    assert len(f["tag"]) == 16
    assert len(f["ciphertext"] + f["tag"]) == mod.OBJECT_FILE_LENGTH
    checks.append("Java fixture nonce/AAD/ciphertext/tag framing relationships match r31")

    # Signature input must contain the exact Java unsigned bytes as its tail.
    literal = (b"TOTP-Vault/v0/token-update" if name == "TOKEN_UPDATE"
               else b"TOTP-Vault/v0/device-update")
    sig_input = f["signature_input"]
    assert sig_input.startswith(literal)
    assert sig_input.endswith(unsigned)
    assert len(sig_input) == len(literal) + 32 + len(unsigned)
    checks.append("Java signature_input uses correct type domain literal and exact canonical unsigned bytes")

    return checks


def main():
    vecs = load_vectors(VECTORS)
    assert set(vecs) == {"TOKEN_UPDATE", "DEVICE_UPDATE"}, set(vecs)
    all_checks = []
    for name in ("TOKEN_UPDATE", "DEVICE_UPDATE"):
        checks = check_fixture(name, vecs[name])
        all_checks.extend((name, c) for c in checks)
    print(f"PASS: {len(all_checks)} cross-implementation checks")
    for name, c in all_checks:
        print(f"  {name}: {c}")


if __name__ == "__main__":
    main()

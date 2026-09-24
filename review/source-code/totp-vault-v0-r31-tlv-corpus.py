#!/usr/bin/env python3
"""Generate and verify the TOTP Vault v0 r31 non-cryptographic byte corpus.

Scope:
- canonical semantic TLV framing and object grammar (r31 §69)
- canonical unsigned form (signed object with final FF01 TLV omitted)
- authenticated-envelope *plaintext* length/padding framing (r31 §14/§69.6), without AEAD

Out of scope:
- Ed25519 verification (SIGNATURE bytes are deterministic opaque placeholders)
- OBJECT_ID/HMAC, HKDF, AES-GCM
- history-dependent semantic validity (ancestry, lifecycle, token creation semantics)
"""
from __future__ import annotations

import json
from dataclasses import dataclass
from pathlib import Path
from typing import Iterable, Sequence

REVISION = 31
OUT = Path(__file__).with_name("totp-vault-v0-r31-tlv-corpus.json")

# Frozen r31 registry.
OBJECT_VERSION = 0
TYPE_TOKEN_UPDATE = 1
TYPE_DEVICE_UPDATE = 2
T_OBJECT_VERSION = 0x0001
T_OBJECT_TYPE = 0x0002
T_DEVICE_PUBLIC_KEY = 0x0003
T_CONTEXT_PARENT_COUNT = 0x0004
T_CONTEXT_PARENT_UPDATE_ID = 0x0005
T_TOKEN_ID = 0x0101
T_STATUS = 0x0102
T_ISSUER = 0x0103
T_ACCOUNT = 0x0104
T_CREDENTIAL = 0x0105
T_DISPLAY_NAME = 0x0201
T_ALGORITHM = 0x0301
T_DIGITS = 0x0302
T_PERIOD = 0x0303
T_SECRET_BYTES = 0x0304
T_SIGNATURE = 0xFF01

STATUS_LIVE = 1
STATUS_TOMBSTONE = 2
ALG_SHA1 = 1
ALG_SHA256 = 2
ALG_SHA512 = 3

MAX_PARENTS = 32
MAX_STRING_BYTES = 256
MAX_SECRET_BYTES = 128
SEMANTIC_LENGTH_U16BE_WIDTH = 2
OBJECT_FILE_LENGTH = 2048
GCM_TAG_WIDTH = 16
ENCRYPTION_PLAINTEXT_LENGTH = OBJECT_FILE_LENGTH - GCM_TAG_WIDTH  # 2032
MAX_SEMANTIC_LENGTH = ENCRYPTION_PLAINTEXT_LENGTH - SEMANTIC_LENGTH_U16BE_WIDTH  # 2030

PK = bytes(range(0x20, 0x40))
TOKEN_ID = bytes(range(0x80, 0xA0))
SIG = bytes(range(0x40, 0x80))  # opaque structural placeholder, NOT asserted to be Ed25519-valid


class Reject(ValueError):
    def __init__(self, code: str):
        super().__init__(code)
        self.code = code


def u8(n: int) -> bytes:
    return n.to_bytes(1, "big")


def u16(n: int) -> bytes:
    return n.to_bytes(2, "big")


def u32(n: int) -> bytes:
    return n.to_bytes(4, "big")


def tlv(tag: int, value: bytes) -> bytes:
    if not 0 <= tag <= 0xFFFF or len(value) > 0xFFFF:
        raise ValueError("TLV range")
    return u16(tag) + u16(len(value)) + value


def parents(n: int) -> list[bytes]:
    # Strictly increasing 32-byte raw IDs.
    return [i.to_bytes(32, "big") for i in range(1, n + 1)]


def cred(*, algorithm: int = ALG_SHA1, digits: int = 6, period: int = 30, secret: bytes = b"s" * 20) -> bytes:
    return b"".join((
        tlv(T_ALGORITHM, u8(algorithm)),
        tlv(T_DIGITS, u8(digits)),
        tlv(T_PERIOD, u32(period)),
        tlv(T_SECRET_BYTES, secret),
    ))


def common(object_type: int, ps: Sequence[bytes]) -> list[bytes]:
    return [
        tlv(T_OBJECT_VERSION, b"\x00"),
        tlv(T_OBJECT_TYPE, u8(object_type)),
        tlv(T_DEVICE_PUBLIC_KEY, PK),
        tlv(T_CONTEXT_PARENT_COUNT, u16(len(ps))),
        *[tlv(T_CONTEXT_PARENT_UPDATE_ID, p) for p in ps],
    ]


def token(*, ps: Sequence[bytes] = (), status: int | None = None,
          issuer: bytes | None = None, account: bytes | None = None,
          credential: bytes | None = None, signed: bool = True) -> bytes:
    parts = common(TYPE_TOKEN_UPDATE, ps)
    parts.append(tlv(T_TOKEN_ID, TOKEN_ID))
    if status is not None:
        parts.append(tlv(T_STATUS, u8(status)))
    if issuer is not None:
        parts.append(tlv(T_ISSUER, issuer))
    if account is not None:
        parts.append(tlv(T_ACCOUNT, account))
    if credential is not None:
        parts.append(tlv(T_CREDENTIAL, credential))
    if signed:
        parts.append(tlv(T_SIGNATURE, SIG))
    return b"".join(parts)


def device(*, ps: Sequence[bytes] = (), name: bytes = b"Laptop Alice", signed: bool = True) -> bytes:
    parts = common(TYPE_DEVICE_UPDATE, ps)
    parts.append(tlv(T_DISPLAY_NAME, name))
    if signed:
        parts.append(tlv(T_SIGNATURE, SIG))
    return b"".join(parts)


@dataclass(frozen=True)
class Item:
    tag: int
    value: bytes
    start: int
    end: int


def parse_tlvs(data: bytes) -> list[Item]:
    pos = 0
    out: list[Item] = []
    while pos < len(data):
        if len(data) - pos < 4:
            raise Reject("truncated_tlv_header")
        start = pos
        tag = int.from_bytes(data[pos:pos+2], "big")
        n = int.from_bytes(data[pos+2:pos+4], "big")
        pos += 4
        if n > len(data) - pos:
            raise Reject("truncated_tlv_value")
        value = data[pos:pos+n]
        pos += n
        out.append(Item(tag, value, start, pos))
    return out


def fixed(v: bytes, n: int, code: str) -> None:
    if len(v) != n:
        raise Reject(code)


def valid_utf8(v: bytes) -> None:
    if len(v) > MAX_STRING_BYTES:
        raise Reject("string_too_long")
    try:
        v.decode("utf-8", errors="strict")
    except UnicodeDecodeError:
        raise Reject("invalid_utf8")


def validate_order(items: Sequence[Item]) -> None:
    prev = -1
    for it in items:
        if it.tag < prev:
            raise Reject("tag_order")
        if it.tag == prev and it.tag != T_CONTEXT_PARENT_UPDATE_ID:
            raise Reject("duplicate_tag")
        prev = it.tag


def validate_credential(data: bytes) -> None:
    items = parse_tlvs(data)
    if [x.tag for x in items] != [T_ALGORITHM, T_DIGITS, T_PERIOD, T_SECRET_BYTES]:
        # This covers missing, duplicate, unknown, and out-of-order nested tags.
        raise Reject("credential_tag_sequence")
    alg, digits, period, secret = (x.value for x in items)
    fixed(alg, 1, "algorithm_width")
    if alg[0] not in (ALG_SHA1, ALG_SHA256, ALG_SHA512):
        raise Reject("algorithm_value")
    fixed(digits, 1, "digits_width")
    if digits[0] not in (6, 7, 8):
        raise Reject("digits_value")
    fixed(period, 4, "period_width")
    if int.from_bytes(period, "big") == 0:
        raise Reject("period_value")
    if not 1 <= len(secret) <= MAX_SECRET_BYTES:
        raise Reject("secret_length")


def validate_semantic(data: bytes, *, form: str) -> str:
    if len(data) > MAX_SEMANTIC_LENGTH:
        raise Reject("semantic_too_long")
    items = parse_tlvs(data)
    validate_order(items)
    if len(items) < 4:
        raise Reject("missing_common_fields")
    # Required common prefix and exact widths.
    if items[0].tag != T_OBJECT_VERSION:
        raise Reject("object_version_not_first")
    fixed(items[0].value, 1, "object_version_width")
    if items[0].value != b"\x00":
        raise Reject("not_v0")  # corpus does not model authenticated-future classification
    if items[1].tag != T_OBJECT_TYPE:
        raise Reject("object_type_not_second")
    fixed(items[1].value, 1, "object_type_width")
    obj_type = items[1].value[0]
    if obj_type not in (TYPE_TOKEN_UPDATE, TYPE_DEVICE_UPDATE):
        raise Reject("object_type_value")
    if items[2].tag != T_DEVICE_PUBLIC_KEY:
        raise Reject("missing_device_public_key")
    fixed(items[2].value, 32, "device_public_key_width")
    if items[3].tag != T_CONTEXT_PARENT_COUNT:
        raise Reject("missing_parent_count")
    fixed(items[3].value, 2, "parent_count_width")
    count = int.from_bytes(items[3].value, "big")
    if count > MAX_PARENTS:
        raise Reject("parent_count_range")

    # Count must be followed immediately by exactly count parent TLVs.
    idx = 4
    ps: list[bytes] = []
    while idx < len(items) and items[idx].tag == T_CONTEXT_PARENT_UPDATE_ID:
        fixed(items[idx].value, 32, "parent_width")
        ps.append(items[idx].value)
        idx += 1
    if len(ps) != count:
        raise Reject("parent_count_mismatch")
    if any(a >= b for a, b in zip(ps, ps[1:])):
        raise Reject("parent_order_or_duplicate")

    # Signature presence is form-specific and FF01 must be terminal by ordering/grammar.
    sigs = [i for i, x in enumerate(items) if x.tag == T_SIGNATURE]
    if form == "signed":
        if sigs != [len(items) - 1]:
            raise Reject("signature_presence_or_position")
        fixed(items[-1].value, 64, "signature_width")
        body_end = len(items) - 1
    elif form == "unsigned":
        if sigs:
            raise Reject("signature_forbidden_unsigned")
        body_end = len(items)
    else:
        raise ValueError(form)

    body = items[idx:body_end]
    tags = [x.tag for x in body]
    assigned = {
        T_TOKEN_ID, T_STATUS, T_ISSUER, T_ACCOUNT, T_CREDENTIAL,
        T_DISPLAY_NAME, T_SIGNATURE,
    }
    # Unknown top-level tags, including other FFxx, are invalid.
    for x in body:
        if x.tag not in assigned:
            raise Reject("unknown_top_level_tag")

    if obj_type == TYPE_TOKEN_UPDATE:
        if not body or body[0].tag != T_TOKEN_ID:
            raise Reject("token_id_required")
        fixed(body[0].value, 32, "token_id_width")
        allowed = {T_TOKEN_ID, T_STATUS, T_ISSUER, T_ACCOUNT, T_CREDENTIAL}
        if any(x.tag not in allowed for x in body):
            raise Reject("forbidden_token_tag")
        semantic = body[1:]
        if not semantic:
            raise Reject("token_field_required")
        for x in semantic:
            if x.tag == T_STATUS:
                fixed(x.value, 1, "status_width")
                if x.value[0] not in (STATUS_LIVE, STATUS_TOMBSTONE):
                    raise Reject("status_value")
            elif x.tag in (T_ISSUER, T_ACCOUNT):
                valid_utf8(x.value)
            elif x.tag == T_CREDENTIAL:
                validate_credential(x.value)
        return "TOKEN_UPDATE"
    else:
        if tags != [T_DISPLAY_NAME]:
            raise Reject("device_body_sequence")
        valid_utf8(body[0].value)
        return "DEVICE_UPDATE"


def make_envelope(p: bytes) -> bytes:
    if len(p) > MAX_SEMANTIC_LENGTH:
        raise ValueError("too long")
    return u16(len(p)) + p + bytes(MAX_SEMANTIC_LENGTH - len(p))


def validate_envelope_plaintext(data: bytes) -> bytes:
    if len(data) != ENCRYPTION_PLAINTEXT_LENGTH:
        raise Reject("envelope_plaintext_length")
    n = int.from_bytes(data[:2], "big")
    if n > MAX_SEMANTIC_LENGTH:
        raise Reject("semantic_length_range")
    p = data[2:2+n]
    padding = data[2+n:]
    if any(padding):
        raise Reject("nonzero_padding")
    return p


def mutate_replace_tlv(data: bytes, tag: int, occurrence: int, replacement: bytes) -> bytes:
    items = parse_tlvs(data)
    seen = 0
    for it in items:
        if it.tag == tag:
            if seen == occurrence:
                return data[:it.start] + replacement + data[it.end:]
            seen += 1
    raise KeyError((tag, occurrence))


def mutate_remove_tlv(data: bytes, tag: int, occurrence: int = 0) -> bytes:
    items = parse_tlvs(data)
    seen = 0
    for it in items:
        if it.tag == tag:
            if seen == occurrence:
                return data[:it.start] + data[it.end:]
            seen += 1
    raise KeyError((tag, occurrence))


def semantic_case(cid: str, name: str, data: bytes, *, form: str, expect: str,
                  mutation: str = "", source: str = "", note: str = "") -> dict:
    return {
        "id": cid,
        "kind": "semantic_tlv",
        "name": name,
        "form": form,
        "expect": expect,
        "semantic_length": len(data),
        "hex": data.hex(),
        **({"source_case": source} if source else {}),
        **({"mutation": mutation} if mutation else {}),
        **({"note": note} if note else {}),
    }


def envelope_case(cid: str, name: str, data: bytes, *, expect: str, mutation: str = "", note: str = "", semantic_parse_expect: str | None = None) -> dict:
    return {
        "id": cid,
        "kind": "envelope_plaintext",
        "name": name,
        "expect": expect,
        "length": len(data),
        "hex": data.hex(),
        **({"mutation": mutation} if mutation else {}),
        **({"note": note} if note else {}),
        **({"semantic_parse_expect": semantic_parse_expect} if semantic_parse_expect else {}),
    }


def file_length_case(cid: str, name: str, n: int, *, expect: str, note: str = "") -> dict:
    return {
        "id": cid,
        "kind": "object_file_length_gate",
        "name": name,
        "expect": expect,
        "length": n,
        "hex": (b"\x00" * n).hex(),
        **({"note": note} if note else {}),
    }


def build_cases() -> list[dict]:
    cases: list[dict] = []
    add = cases.append

    # Positive signed/unsigned semantic TLV vectors.
    p_creation = token(status=STATUS_LIVE, issuer=b"Example Corp", account=b"alice@example.test",
                       credential=cred(secret=b"S" * 20))
    u_creation = token(status=STATUS_LIVE, issuer=b"Example Corp", account=b"alice@example.test",
                       credential=cred(secret=b"S" * 20), signed=False)
    add(semantic_case("P001", "TOKEN_UPDATE typical creation-shaped signed form", p_creation,
                      form="signed", expect="valid", note="Structural only; history-dependent token-creation validity is out of scope."))
    add(semantic_case("P002", "TOKEN_UPDATE exact canonical unsigned form", u_creation,
                      form="unsigned", expect="valid", source="P001",
                      note="Must equal P001 with the complete final FF01 TLV removed and no other byte changed."))
    add(semantic_case("P003", "TOKEN_UPDATE issuer-only edit, one parent", token(ps=parents(1), issuer=b"Example Corp"),
                      form="signed", expect="valid"))
    add(semantic_case("P004", "TOKEN_UPDATE status-only, two parents", token(ps=parents(2), status=STATUS_TOMBSTONE),
                      form="signed", expect="valid"))
    add(semantic_case("P005", "TOKEN_UPDATE credential min bounds", token(ps=parents(1), credential=cred(algorithm=ALG_SHA1, digits=6, period=1, secret=b"x")),
                      form="signed", expect="valid"))
    add(semantic_case("P006", "TOKEN_UPDATE credential max scalar/secret bounds", token(ps=parents(1), credential=cred(algorithm=ALG_SHA512, digits=8, period=0xFFFFFFFF, secret=b"x" * 128)),
                      form="signed", expect="valid"))
    add(semantic_case("P007", "TOKEN_UPDATE empty issuer and account are representable", token(ps=parents(1), issuer=b"", account=b""),
                      form="signed", expect="valid"))
    add(semantic_case("P008", "TOKEN_UPDATE 256-byte UTF-8 issuer (64 four-byte code points)", token(ps=parents(1), issuer=("😀" * 64).encode("utf-8")),
                      form="signed", expect="valid", note="Pins byte-counted, not code-point-counted string limit."))
    add(semantic_case("P009", "TOKEN_UPDATE 32-parent canonical frontier", token(ps=parents(32), status=STATUS_LIVE),
                      form="signed", expect="valid"))
    add(semantic_case("P015", "TOKEN_UPDATE 256-byte ACCOUNT boundary", token(ps=parents(1), account=b"A"*256),
                      form="signed", expect="valid"))
    add(semantic_case("P016", "TOKEN_UPDATE SHA-256 / 7-digit credential enum coverage", token(ps=parents(1), credential=cred(algorithm=ALG_SHA256, digits=7, period=30, secret=b"x"*32)),
                      form="signed", expect="valid"))
    p_max = token(ps=parents(32), status=STATUS_LIVE, issuer=b"I"*256, account=b"A"*256,
                  credential=cred(algorithm=ALG_SHA512, digits=8, period=0xFFFFFFFF, secret=b"S"*128))
    add(semantic_case("P010", "maximum valid TOKEN_UPDATE under r31 registry", p_max,
                      form="signed", expect="valid", note="Expected semantic length 1987."))
    add(semantic_case("P011", "DEVICE_UPDATE root signed form", device(name=b"Laptop Alice"),
                      form="signed", expect="valid"))
    add(semantic_case("P012", "DEVICE_UPDATE exact canonical unsigned form", device(name=b"Laptop Alice", signed=False),
                      form="unsigned", expect="valid", source="P011"))
    add(semantic_case("P013", "DEVICE_UPDATE empty display name is representable", device(name=b""),
                      form="signed", expect="valid"))
    p_devmax = device(ps=parents(32), name=b"D"*256)
    add(semantic_case("P014", "maximum valid DEVICE_UPDATE under r31 registry", p_devmax,
                      form="signed", expect="valid", note="Expected semantic length 1532."))

    # Base cases used for mutations.
    base = p_creation
    base_status = token(ps=parents(2), status=STATUS_LIVE)
    base_device = device(name=b"Laptop Alice")
    base_cred = token(ps=parents(1), credential=cred())

    # Primitive framing failures.
    add(semantic_case("N001", "truncated TLV header", base + b"\x00", form="signed", expect="reject",
                      source="P001", mutation="append one byte after complete TLV sequence"))
    b = bytearray(base); b[2:4] = b"\x00\x02"  # OBJECT_VERSION claims 2 bytes; consumes next tag byte and destroys framing.
    add(semantic_case("N002", "truncated/misaligned value caused by wrong first length", bytes(b), form="signed", expect="reject",
                      source="P001", mutation="OBJECT_VERSION length 1 -> 2 without inserting a byte"))
    add(semantic_case("N003", "semantic plaintext exceeds 2030-byte generic envelope maximum", base + b"\x00"*(2031-len(base)),
                      form="signed", expect="reject", source="P001", mutation="pad raw semantic input to 2031 bytes"))

    # Common field width/order/value failures.
    add(semantic_case("N004", "OBJECT_VERSION wrong width", mutate_replace_tlv(base, T_OBJECT_VERSION, 0, tlv(T_OBJECT_VERSION, b"\x00\x00")),
                      form="signed", expect="reject", source="P001", mutation="OBJECT_VERSION u8 -> 2 bytes"))
    add(semantic_case("N005", "OBJECT_TYPE unknown v0 value", mutate_replace_tlv(base, T_OBJECT_TYPE, 0, tlv(T_OBJECT_TYPE, b"\x03")),
                      form="signed", expect="reject", source="P001", mutation="OBJECT_TYPE 01 -> 03"))
    add(semantic_case("N046", "OBJECT_TYPE wrong width", mutate_replace_tlv(base, T_OBJECT_TYPE, 0, tlv(T_OBJECT_TYPE, b"\x00\x01")),
                      form="signed", expect="reject", source="P001", mutation="OBJECT_TYPE u8 -> 2 bytes"))
    add(semantic_case("N006", "DEVICE_PUBLIC_KEY wrong width", mutate_replace_tlv(base, T_DEVICE_PUBLIC_KEY, 0, tlv(T_DEVICE_PUBLIC_KEY, PK[:-1])),
                      form="signed", expect="reject", source="P001", mutation="32 -> 31 bytes"))
    add(semantic_case("N007", "CONTEXT_PARENT_COUNT wrong width", mutate_replace_tlv(base, T_CONTEXT_PARENT_COUNT, 0, tlv(T_CONTEXT_PARENT_COUNT, b"\x00")),
                      form="signed", expect="reject", source="P001", mutation="u16be -> one byte"))
    # Move TOKEN_ID before parent count, making numeric order invalid.
    its = parse_tlvs(base)
    tid = next(x for x in its if x.tag == T_TOKEN_ID)
    cnt = next(x for x in its if x.tag == T_CONTEXT_PARENT_COUNT)
    reordered = base[:cnt.start] + base[tid.start:tid.end] + base[cnt.start:tid.start] + base[tid.end:]
    add(semantic_case("N008", "top-level numeric tag order violation", reordered, form="signed", expect="reject",
                      source="P001", mutation="move 0101 TOKEN_ID before 0004 parent count"))
    dup_status = mutate_replace_tlv(base_status, T_SIGNATURE, 0, tlv(T_STATUS, b"\x01") + tlv(T_SIGNATURE, SIG))
    add(semantic_case("N009", "duplicate singleton STATUS", dup_status, form="signed", expect="reject",
                      source="P004", mutation="insert second 0102 STATUS before signature"))
    unknown = mutate_replace_tlv(base, T_SIGNATURE, 0, tlv(0x0106, b"x") + tlv(T_SIGNATURE, SIG))
    add(semantic_case("N010", "unknown v0 top-level tag", unknown, form="signed", expect="reject",
                      source="P001", mutation="insert unassigned 0106"))
    unknown_ff = mutate_replace_tlv(base, T_SIGNATURE, 0, tlv(0xFF00, b"x") + tlv(T_SIGNATURE, SIG))
    add(semantic_case("N011", "unassigned FFxx authentication tag", unknown_ff, form="signed", expect="reject",
                      source="P001", mutation="insert unassigned FF00 before FF01"))

    # Parent repetition/count/order boundaries.
    c33 = common(TYPE_TOKEN_UPDATE, parents(32))
    # parent count says 33 while only 32 follow; range rejection precedes mismatch.
    bad33 = b"".join([c33[0], c33[1], c33[2], tlv(T_CONTEXT_PARENT_COUNT, u16(33)), *c33[4:], tlv(T_TOKEN_ID, TOKEN_ID), tlv(T_STATUS, b"\x01"), tlv(T_SIGNATURE, SIG)])
    add(semantic_case("N012", "parent count 33 exceeds v0 maximum", bad33, form="signed", expect="reject",
                      mutation="CONTEXT_PARENT_COUNT=33"))
    mismatch = mutate_replace_tlv(base_status, T_CONTEXT_PARENT_COUNT, 0, tlv(T_CONTEXT_PARENT_COUNT, u16(1)))
    add(semantic_case("N013", "parent count/repetition mismatch", mismatch, form="signed", expect="reject",
                      source="P004", mutation="count 2 -> 1 while two parent TLVs remain"))
    # Remove one of two parents but keep count=2.
    add(semantic_case("N014", "parent count missing repeated parent", mutate_remove_tlv(base_status, T_CONTEXT_PARENT_UPDATE_ID, 1),
                      form="signed", expect="reject", source="P004", mutation="remove second parent but keep count=2"))
    add(semantic_case("N015", "parent ID wrong width", mutate_replace_tlv(base_status, T_CONTEXT_PARENT_UPDATE_ID, 0, tlv(T_CONTEXT_PARENT_UPDATE_ID, parents(1)[0][:-1])),
                      form="signed", expect="reject", source="P004", mutation="first parent 32 -> 31 bytes"))
    its = parse_tlvs(base_status)
    parent_items = [x for x in its if x.tag == T_CONTEXT_PARENT_UPDATE_ID]
    unsorted = base_status[:parent_items[0].start] + base_status[parent_items[1].start:parent_items[1].end] + base_status[parent_items[0].start:parent_items[0].end] + base_status[parent_items[1].end:]
    add(semantic_case("N016", "parents not raw-byte sorted", unsorted, form="signed", expect="reject",
                      source="P004", mutation="swap two canonical parent TLVs"))
    duplicate_parent = mutate_replace_tlv(base_status, T_CONTEXT_PARENT_UPDATE_ID, 1, tlv(T_CONTEXT_PARENT_UPDATE_ID, parents(2)[0]))
    add(semantic_case("N017", "duplicate parent ID", duplicate_parent, form="signed", expect="reject",
                      source="P004", mutation="second parent value -> first parent value"))

    # TOKEN_UPDATE body failures.
    no_fields = token(ps=parents(1), issuer=b"x")
    no_fields = mutate_remove_tlv(no_fields, T_ISSUER)
    add(semantic_case("N018", "TOKEN_UPDATE has no semantic field assertion", no_fields, form="signed", expect="reject",
                      mutation="only TOKEN_ID remains after common prefix"))
    add(semantic_case("N019", "TOKEN_ID wrong width", mutate_replace_tlv(base, T_TOKEN_ID, 0, tlv(T_TOKEN_ID, TOKEN_ID[:-1])),
                      form="signed", expect="reject", source="P001", mutation="32 -> 31 bytes"))
    add(semantic_case("N020", "STATUS wrong width", mutate_replace_tlv(base_status, T_STATUS, 0, tlv(T_STATUS, b"\x00\x01")),
                      form="signed", expect="reject", source="P004", mutation="u8 -> 2 bytes"))
    add(semantic_case("N021", "STATUS unknown value", mutate_replace_tlv(base_status, T_STATUS, 0, tlv(T_STATUS, b"\x03")),
                      form="signed", expect="reject", source="P004", mutation="01 -> 03"))
    add(semantic_case("N022", "ISSUER invalid UTF-8", mutate_replace_tlv(base, T_ISSUER, 0, tlv(T_ISSUER, b"\xC3\x28")),
                      form="signed", expect="reject", source="P001", mutation="replace issuer with invalid UTF-8 sequence C3 28"))
    add(semantic_case("N023", "ISSUER 257 bytes exceeds 256-byte limit", token(ps=parents(1), issuer=b"I"*257),
                      form="signed", expect="reject", mutation="issuer 256 -> 257 encoded bytes"))
    add(semantic_case("N044", "ACCOUNT invalid UTF-8", token(ps=parents(1), account=b"\xC3\x28"),
                      form="signed", expect="reject", mutation="account replaced with invalid UTF-8 sequence C3 28"))
    add(semantic_case("N045", "ACCOUNT 257 bytes exceeds 256-byte limit", token(ps=parents(1), account=b"A"*257),
                      form="signed", expect="reject", mutation="account 256 -> 257 encoded bytes"))
    forbidden_display = mutate_replace_tlv(base, T_SIGNATURE, 0, tlv(T_DISPLAY_NAME, b"x") + tlv(T_SIGNATURE, SIG))
    add(semantic_case("N024", "DEVICE_UPDATE-only tag in TOKEN_UPDATE", forbidden_display, form="signed", expect="reject",
                      source="P001", mutation="insert 0201 DISPLAY_NAME"))

    # DEVICE_UPDATE body failures.
    add(semantic_case("N025", "DISPLAY_NAME missing from DEVICE_UPDATE", mutate_remove_tlv(base_device, T_DISPLAY_NAME),
                      form="signed", expect="reject", source="P011", mutation="remove required 0201"))
    forbidden_token = mutate_replace_tlv(base_device, T_DISPLAY_NAME, 0, tlv(T_TOKEN_ID, TOKEN_ID))
    add(semantic_case("N026", "TOKEN_UPDATE-only tag in DEVICE_UPDATE", forbidden_token, form="signed", expect="reject",
                      source="P011", mutation="replace 0201 with 0101 TOKEN_ID"))
    add(semantic_case("N027", "DISPLAY_NAME invalid UTF-8", mutate_replace_tlv(base_device, T_DISPLAY_NAME, 0, tlv(T_DISPLAY_NAME, b"\x80")),
                      form="signed", expect="reject", source="P011", mutation="replace name with lone continuation byte"))
    add(semantic_case("N028", "DISPLAY_NAME 257 bytes exceeds limit", device(name=b"D"*257),
                      form="signed", expect="reject", mutation="256 -> 257 bytes"))

    # Nested CREDENTIAL failures.
    citems = parse_tlvs(cred())
    swapped = citems[1]
    swapped_cred = tlv(T_DIGITS, citems[1].value) + tlv(T_ALGORITHM, citems[0].value) + b"".join(tlv(x.tag,x.value) for x in citems[2:])
    add(semantic_case("N029", "nested credential tags out of order", mutate_replace_tlv(base_cred, T_CREDENTIAL, 0, tlv(T_CREDENTIAL, swapped_cred)),
                      form="signed", expect="reject", source="P005", mutation="swap 0301/0302 order"))
    missing_secret = b"".join(tlv(x.tag,x.value) for x in citems[:-1])
    add(semantic_case("N030", "nested credential missing SECRET_BYTES", mutate_replace_tlv(base_cred, T_CREDENTIAL, 0, tlv(T_CREDENTIAL, missing_secret)),
                      form="signed", expect="reject", mutation="remove 0304"))
    dup_digits = b"".join((tlv(T_ALGORITHM,b"\x01"), tlv(T_DIGITS,b"\x06"), tlv(T_DIGITS,b"\x06"), tlv(T_PERIOD,u32(30)), tlv(T_SECRET_BYTES,b"x")))
    add(semantic_case("N031", "nested credential duplicate tag", mutate_replace_tlv(base_cred, T_CREDENTIAL, 0, tlv(T_CREDENTIAL, dup_digits)),
                      form="signed", expect="reject", mutation="duplicate 0302"))
    unknown_nested = b"".join((tlv(T_ALGORITHM,b"\x01"), tlv(T_DIGITS,b"\x06"), tlv(T_PERIOD,u32(30)), tlv(0x0305,b"x"), tlv(T_SECRET_BYTES,b"x")))
    add(semantic_case("N032", "nested credential unknown tag", mutate_replace_tlv(base_cred, T_CREDENTIAL, 0, tlv(T_CREDENTIAL, unknown_nested)),
                      form="signed", expect="reject", mutation="insert 0305"))
    bad_alg_width = cred(); bad_alg_width = mutate_replace_tlv(bad_alg_width, T_ALGORITHM, 0, tlv(T_ALGORITHM,b"\x00\x01"))
    add(semantic_case("N033", "ALGORITHM wrong width", mutate_replace_tlv(base_cred, T_CREDENTIAL, 0, tlv(T_CREDENTIAL,bad_alg_width)),
                      form="signed", expect="reject", mutation="0301 u8 -> 2 bytes"))
    bad_alg = cred(algorithm=4)
    add(semantic_case("N034", "ALGORITHM unknown enum", mutate_replace_tlv(base_cred, T_CREDENTIAL, 0, tlv(T_CREDENTIAL,bad_alg)),
                      form="signed", expect="reject", mutation="algorithm=04"))
    bad_digits_width = b"".join((tlv(T_ALGORITHM,b"\x01"), tlv(T_DIGITS,b"\x00\x06"), tlv(T_PERIOD,u32(30)), tlv(T_SECRET_BYTES,b"x")))
    add(semantic_case("N047", "DIGITS wrong width", mutate_replace_tlv(base_cred, T_CREDENTIAL, 0, tlv(T_CREDENTIAL,bad_digits_width)),
                      form="signed", expect="reject", mutation="0302 u8 -> 2 bytes"))
    bad_digits = cred(digits=9)
    add(semantic_case("N035", "DIGITS invalid enum", mutate_replace_tlv(base_cred, T_CREDENTIAL, 0, tlv(T_CREDENTIAL,bad_digits)),
                      form="signed", expect="reject", mutation="digits=09"))
    zero_period = cred(period=0)
    add(semantic_case("N036", "PERIOD zero", mutate_replace_tlv(base_cred, T_CREDENTIAL, 0, tlv(T_CREDENTIAL,zero_period)),
                      form="signed", expect="reject", mutation="period=00000000"))
    # PERIOD wrong width built manually.
    bad_period_width = b"".join((tlv(T_ALGORITHM,b"\x01"), tlv(T_DIGITS,b"\x06"), tlv(T_PERIOD,b"\x00\x00\x1e"), tlv(T_SECRET_BYTES,b"x")))
    add(semantic_case("N037", "PERIOD wrong width", mutate_replace_tlv(base_cred, T_CREDENTIAL, 0, tlv(T_CREDENTIAL,bad_period_width)),
                      form="signed", expect="reject", mutation="u32be -> 3 bytes"))
    empty_secret = cred(secret=b"")
    add(semantic_case("N038", "SECRET_BYTES empty", mutate_replace_tlv(base_cred, T_CREDENTIAL, 0, tlv(T_CREDENTIAL,empty_secret)),
                      form="signed", expect="reject", mutation="secret length 1+ -> 0"))
    long_secret = cred(secret=b"x"*129)
    add(semantic_case("N039", "SECRET_BYTES 129 exceeds maximum", mutate_replace_tlv(base_cred, T_CREDENTIAL, 0, tlv(T_CREDENTIAL,long_secret)),
                      form="signed", expect="reject", mutation="secret 128 -> 129 bytes"))
    nested_trailing = cred() + b"\x00"
    add(semantic_case("N048", "nested CREDENTIAL trailing byte", mutate_replace_tlv(base_cred, T_CREDENTIAL, 0, tlv(T_CREDENTIAL,nested_trailing)),
                      form="signed", expect="reject", mutation="append one byte after complete nested TLV sequence"))

    # Signature trailer / unsigned form failures.
    add(semantic_case("N040", "signed form missing signature", mutate_remove_tlv(base, T_SIGNATURE),
                      form="signed", expect="reject", source="P001", mutation="remove final FF01"))
    add(semantic_case("N041", "SIGNATURE wrong width", mutate_replace_tlv(base, T_SIGNATURE, 0, tlv(T_SIGNATURE, SIG[:-1])),
                      form="signed", expect="reject", source="P001", mutation="64 -> 63 bytes"))
    add(semantic_case("N042", "unsigned form contains signature", base, form="unsigned", expect="reject",
                      source="P001", mutation="parse signed bytes as canonical unsigned form"))
    # Put signature before CREDENTIAL: numeric order fails and signature isn't terminal.
    its = parse_tlvs(base)
    sig = next(x for x in its if x.tag == T_SIGNATURE)
    cred_item = next(x for x in its if x.tag == T_CREDENTIAL)
    sig_early = base[:cred_item.start] + base[sig.start:sig.end] + base[cred_item.start:sig.start]
    add(semantic_case("N043", "SIGNATURE not final", sig_early, form="signed", expect="reject",
                      source="P001", mutation="move FF01 before 0105"))

    # Envelope plaintext cases (no AEAD): exact 2032 bytes, u16be semantic length, all-zero remainder.
    env_creation = make_envelope(p_creation)
    add(envelope_case("E001", "canonical envelope plaintext for P001", env_creation, expect="valid",
                      note="No AEAD in this corpus; validates only authenticated plaintext framing.", semantic_parse_expect="valid_signed"))
    env_max = make_envelope(p_max)
    add(envelope_case("E002", "canonical envelope plaintext for 1987-byte maximum TOKEN_UPDATE", env_max, expect="valid", semantic_parse_expect="valid_signed"))
    p2030 = b"z" * 2030
    add(envelope_case("E003", "generic envelope maximum SEMANTIC_LENGTH_U16BE=2030", make_envelope(p2030), expect="valid",
                      note="Framing-valid envelope payload; not a valid semantic TLV object.", semantic_parse_expect="reject"))
    # 0 and 1 are envelope framing-valid but are semantic TLV negatives; keep that distinction explicit.
    add(envelope_case("E004", "SEMANTIC_LENGTH_U16BE=0 with all-zero padding", make_envelope(b""), expect="valid",
                      note="Envelope framing-valid; empty P must fail semantic parsing.", semantic_parse_expect="reject"))
    add(envelope_case("E005", "SEMANTIC_LENGTH_U16BE=1 with zero padding", make_envelope(b"\x00"), expect="valid",
                      note="Envelope framing-valid; one-byte P must fail semantic parsing.", semantic_parse_expect="reject"))
    badlen = u16(2031) + bytes(2030)  # exactly 2032 total, but declared semantic length exceeds maximum.
    add(envelope_case("E006", "SEMANTIC_LENGTH_U16BE=2031 rejected", badlen, expect="reject", mutation="length prefix 2031"))
    nonzero = bytearray(env_creation); nonzero[-1] = 1
    add(envelope_case("E007", "non-zero deterministic padding rejected", bytes(nonzero), expect="reject", mutation="last padding byte 00 -> 01"))
    add(envelope_case("E008", "short envelope plaintext rejected", env_creation[:-1], expect="reject", mutation="2032 -> 2031 bytes"))
    add(envelope_case("E009", "long envelope plaintext rejected", env_creation + b"\x00", expect="reject", mutation="2032 -> 2033 bytes"))

    # Pre-AEAD object-file length gate. Bytes are deliberately meaningless: only length is tested here.
    add(file_length_case("F001", "exact 2048-byte object file passes pre-AEAD length gate", 2048, expect="valid",
                         note="Does not imply AEAD authentication or semantic validity."))
    add(file_length_case("F002", "2047-byte object file rejected before AEAD", 2047, expect="reject"))
    add(file_length_case("F003", "2049-byte object file rejected before AEAD", 2049, expect="reject"))

    return cases


def verify_case(c: dict) -> tuple[bool, str]:
    data = bytes.fromhex(c["hex"])
    try:
        if c["kind"] == "semantic_tlv":
            validate_semantic(data, form=c["form"])
        elif c["kind"] == "envelope_plaintext":
            p = validate_envelope_plaintext(data)
            if c.get("semantic_parse_expect"):
                sem_ok = True
                try:
                    validate_semantic(p, form="signed")
                except Reject:
                    sem_ok = False
                want = c["semantic_parse_expect"] == "valid_signed"
                if sem_ok != want:
                    raise Reject("semantic_parse_relation")
        elif c["kind"] == "object_file_length_gate":
            if len(data) != OBJECT_FILE_LENGTH:
                raise Reject("object_file_length")
        else:
            raise AssertionError(c["kind"])
        accepted = True
        detail = "accepted"
    except Reject as e:
        accepted = False
        detail = e.code
    expected = c["expect"] == "valid"
    return accepted == expected, detail


def relations(cases: Sequence[dict]) -> list[dict]:
    by_id = {c["id"]: c for c in cases}
    rels = []
    for signed_id, unsigned_id in (("P001", "P002"), ("P011", "P012")):
        s = bytes.fromhex(by_id[signed_id]["hex"])
        u = bytes.fromhex(by_id[unsigned_id]["hex"])
        items = parse_tlvs(s)
        sig = items[-1]
        ok = sig.tag == T_SIGNATURE and s[:sig.start] == u and sig.end == len(s)
        if not ok:
            raise AssertionError(f"unsigned relation failed {signed_id}/{unsigned_id}")
        rels.append({
            "kind": "unsigned_is_signed_minus_final_signature_tlv",
            "signed": signed_id,
            "unsigned": unsigned_id,
            "signature_tlv_hex": s[sig.start:sig.end].hex(),
        })
    return rels


def main() -> None:
    cases = build_cases()
    # Hard r31 census assertions.
    by_id = {c["id"]: c for c in cases}
    assert by_id["P001"]["semantic_length"] == 245
    assert by_id["P010"]["semantic_length"] == 1987
    assert by_id["P014"]["semantic_length"] == 1532

    failed = []
    outcomes = {}
    for c in cases:
        ok, detail = verify_case(c)
        outcomes[c["id"]] = detail
        if not ok:
            failed.append((c["id"], c["expect"], detail))
    if failed:
        raise SystemExit("self-check failures: " + repr(failed))

    doc = {
        "corpus": "TOTP Vault v0 r31 canonical TLV and envelope-plaintext corpus",
        "revision": REVISION,
        "normative_basis": ["r31 §14", "r31 §69", "r31 §70", "r31 §74"],
        "scope": {
            "included": [
                "primitive u16be TLV framing",
                "v0 signed and canonical unsigned object grammar",
                "fixed integer widths/enums",
                "parent count/repetition/raw-byte ordering",
                "UTF-8 byte limits",
                "nested CREDENTIAL grammar",
                "FF01 signature structural position/width only",
                "2032-byte deterministic envelope plaintext length/padding framing",
                "2048-byte pre-AEAD object-file length gate",
            ],
            "excluded": [
                "Ed25519 signature verification",
                "OBJECT_ID/HMAC",
                "HKDF/per-object key derivation",
                "AES-GCM",
                "history-dependent semantic validity and lifecycle rules",
            ],
        },
        "constants": {
            "tlv_tag_width": 2,
            "tlv_length_width": 2,
            "byte_order": "big-endian",
            "max_parents": MAX_PARENTS,
            "max_string_bytes": MAX_STRING_BYTES,
            "max_secret_bytes": MAX_SECRET_BYTES,
            "semantic_length_field": "SEMANTIC_LENGTH_U16BE",
            "semantic_length_width": SEMANTIC_LENGTH_U16BE_WIDTH,
            "encryption_plaintext_length": ENCRYPTION_PLAINTEXT_LENGTH,
            "max_semantic_length": MAX_SEMANTIC_LENGTH,
            "object_file_length_after_gcm_tag": OBJECT_FILE_LENGTH,
            "signature_placeholder_hex": SIG.hex(),
            "signature_placeholder_note": "opaque 64-byte structural placeholder; no Ed25519 validity is asserted",
        },
        "relations": relations(cases),
        "cases": cases,
        "generator_self_check_outcomes": outcomes,
    }
    OUT.write_text(json.dumps(doc, indent=2, sort_keys=False) + "\n", encoding="utf-8")

    pos = sum(1 for c in cases if c["expect"] == "valid")
    neg = len(cases) - pos
    print(f"wrote {OUT}")
    print(f"cases: {len(cases)} total = {pos} valid + {neg} reject")
    print(f"semantic TLV: {sum(c['kind']=='semantic_tlv' for c in cases)}")
    print(f"envelope plaintext: {sum(c['kind']=='envelope_plaintext' for c in cases)}")
    print(f"object-file length gate: {sum(c['kind']=='object_file_length_gate' for c in cases)}")
    print(f"P001={by_id['P001']['semantic_length']} P010={by_id['P010']['semantic_length']} P014={by_id['P014']['semantic_length']}")


if __name__ == "__main__":
    main()

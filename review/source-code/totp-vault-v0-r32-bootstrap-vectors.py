from __future__ import annotations

from dataclasses import dataclass
from pathlib import Path
import hashlib, hmac

from argon2.low_level import hash_secret_raw, Type
from cryptography.hazmat.primitives.ciphers.aead import AESGCM
from cryptography.exceptions import InvalidTag

MAGIC = b"TOTP-VAULT"
VERSION = 0
VAULT_LEN = 87
HEADER_LEN = 39
PASSWORD_MAX_BYTES = 1024

ARGON2_MEMORY_KIB = 65536
ARGON2_ITERATIONS = 3
ARGON2_PARALLELISM = 4
ARGON2_HASH_LEN = 32
ARGON2_VERSION = 0x13


def argon2id(password: bytes, salt: bytes) -> bytes:
    if len(password) > PASSWORD_MAX_BYTES:
        raise ValueError("password length")
    if len(salt) != 16:
        raise ValueError("salt length")
    return hash_secret_raw(
        secret=password,
        salt=salt,
        time_cost=ARGON2_ITERATIONS,
        memory_cost=ARGON2_MEMORY_KIB,
        parallelism=ARGON2_PARALLELISM,
        hash_len=ARGON2_HASH_LEN,
        type=Type.ID,
        version=ARGON2_VERSION,
    )


def header(salt: bytes, nonce: bytes) -> bytes:
    if len(salt) != 16 or len(nonce) != 12:
        raise ValueError("field width")
    h = MAGIC + bytes([VERSION]) + salt + nonce
    assert len(h) == HEADER_LEN
    return h


def encode(password: bytes, root: bytes, salt: bytes, nonce: bytes) -> dict[str, bytes]:
    if len(root) != 32:
        raise ValueError("root length")
    h = header(salt, nonce)
    k = argon2id(password, salt)
    out = AESGCM(k).encrypt(nonce, root, h)
    assert len(out) == 48
    ct, tag = out[:32], out[32:]
    file = h + ct + tag
    assert len(file) == VAULT_LEN
    return {"header": h, "k_wrap": k, "wrapped_root": ct, "wrap_tag": tag, "vault_file": file}


def parse_and_unwrap(vault: bytes, password: bytes, *, counter: list[int] | None = None) -> bytes:
    # Normative processing order: reject framing/version/password bound before KDF.
    if len(vault) != VAULT_LEN:
        raise ValueError("vault length")
    if vault[:10] != MAGIC:
        raise ValueError("magic")
    if vault[10] != VERSION:
        raise ValueError("bootstrap version")
    if len(password) > PASSWORD_MAX_BYTES:
        raise ValueError("password length")
    salt = vault[11:27]
    nonce = vault[27:39]
    aad = vault[:39]
    if counter is not None:
        counter[0] += 1
    k = argon2id(password, salt)
    try:
        return AESGCM(k).decrypt(nonce, vault[39:], aad)
    except InvalidTag as e:
        raise ValueError("unwrap authentication") from e


def hx(b: bytes) -> str:
    return b.hex()

ROOT = bytes(range(32))
CASES = [
    ("INITIAL", b"correct horse battery staple", bytes(range(0x10, 0x20)), bytes(range(0xA0, 0xAC))),
    ("REWRAP", "Tr0ub4dor&3-\N{SNOWMAN}".encode("utf-8"), bytes(range(0x20, 0x30)), bytes(range(0xB0, 0xBC))),
]

lines = [
    "# Public deterministic TOTP Vault v0 r32 bootstrap vectors; NOT real secrets.",
    "magic_ascii=TOTP-VAULT",
    f"magic={hx(MAGIC)}",
    f"bootstrap_version={VERSION}",
    f"vault_length={VAULT_LEN}",
    f"header_aad_length={HEADER_LEN}",
    f"root={hx(ROOT)}",
    "argon2_type=id",
    "argon2_version=0x13",
    "argon2_memory_kib=65536",
    "argon2_iterations=3",
    "argon2_parallelism=4",
    "argon2_output_bytes=32",
]

for name, password, salt, nonce in CASES:
    v = encode(password, ROOT, salt, nonce)
    assert parse_and_unwrap(v["vault_file"], password) == ROOT
    lines += [
        "",
        f"[{name}]",
        f"password_utf8={hx(password)}",
        f"argon2_salt={hx(salt)}",
        f"wrap_nonce={hx(nonce)}",
        f"header_aad={hx(v['header'])}",
        f"k_wrap={hx(v['k_wrap'])}",
        f"wrapped_root={hx(v['wrapped_root'])}",
        f"wrap_tag={hx(v['wrap_tag'])}",
        f"vault_file={hx(v['vault_file'])}",
    ]

# Same root, different wrapping material must produce different bootstrap bytes.
initial = encode(CASES[0][1], ROOT, CASES[0][2], CASES[0][3])
rewrap = encode(CASES[1][1], ROOT, CASES[1][2], CASES[1][3])
assert initial["vault_file"] != rewrap["vault_file"]

# Pre-KDF rejection ordering.
for bad in (initial["vault_file"][:-1], initial["vault_file"] + b"\0"):
    c = [0]
    try: parse_and_unwrap(bad, CASES[0][1], counter=c)
    except ValueError as e: assert str(e) == "vault length" and c[0] == 0
    else: raise AssertionError("bad length accepted")

bad_magic = bytearray(initial["vault_file"]); bad_magic[0] ^= 1
c=[0]
try: parse_and_unwrap(bytes(bad_magic), CASES[0][1], counter=c)
except ValueError as e: assert str(e)=="magic" and c[0]==0
else: raise AssertionError("bad magic accepted")

bad_version = bytearray(initial["vault_file"]); bad_version[10] = 1
c=[0]
try: parse_and_unwrap(bytes(bad_version), CASES[0][1], counter=c)
except ValueError as e: assert str(e)=="bootstrap version" and c[0]==0
else: raise AssertionError("bad version accepted")

c=[0]
try: parse_and_unwrap(initial["vault_file"], b"x"*1025, counter=c)
except ValueError as e: assert str(e)=="password length" and c[0]==0
else: raise AssertionError("overlong password accepted")

# Authenticated/KDF-sensitive mutations fail unwrap rather than parse framing.
for off in (11, 27, 39, 86):
    b=bytearray(initial["vault_file"]); b[off]^=1
    c=[0]
    try: parse_and_unwrap(bytes(b), CASES[0][1], counter=c)
    except ValueError as e: assert str(e)=="unwrap authentication" and c[0]==1
    else: raise AssertionError(f"mutation at {off} accepted")

try: parse_and_unwrap(initial["vault_file"], b"wrong password")
except ValueError as e: assert str(e)=="unwrap authentication"
else: raise AssertionError("wrong password accepted")

Path('/mnt/data/totp-vault-v0-r32-bootstrap-vectors.txt').write_text("\n".join(lines)+"\n")
print("r32 bootstrap vectors PASS")
print("initial k_wrap", initial["k_wrap"].hex())
print("initial vault", initial["vault_file"].hex())

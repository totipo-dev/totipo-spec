#!/usr/bin/env python3
"""Validate this repository's JSON schema subset without third-party packages.

This is a schema checker for the checked-in contracts, not a general JSON Schema
implementation. Unknown keywords fail closed so schema extensions need review.
Cryptographic hashes and requirements pins are checked by the Go consumer.
"""
import json
import re
from pathlib import Path


class Invalid(ValueError):
    pass


def pairs(items):
    out = {}
    for key, value in items:
        if key in out:
            raise Invalid("duplicate JSON member")
        out[key] = value
    return out


def read(path):
    return json.loads(path.read_text(encoding="utf-8"), object_pairs_hook=pairs)


def validate(value, schema, path="$"):
    allowed = {"$schema", "$id", "title", "description", "type", "const", "enum",
               "properties", "required", "additionalProperties", "items", "minItems",
               "maxItems", "minimum", "maximum", "pattern", "minLength", "maxLength",
               "allOf", "if", "then", "else"}
    if set(schema) - allowed:
        raise Invalid(f"{path}: unsupported schema keyword")
    def require(ok, rule):
        if not ok:
            raise Invalid(f"{path}: {rule}")
    if "type" in schema:
        types = schema["type"]
        types = [types] if isinstance(types, str) else types
        def matches(t):
            return {"object": type(value) is dict, "array": type(value) is list,
                    "string": type(value) is str, "integer": type(value) is int,
                    "boolean": type(value) is bool, "null": value is None}[t]
        require(any(matches(t) for t in types), "type")
    if "const" in schema:
        require(value == schema["const"], "const")
    if "enum" in schema:
        require(value in schema["enum"], "enum")
    if isinstance(value, dict):
        require(all(k in value for k in schema.get("required", [])), "required member")
        props = schema.get("properties", {})
        if schema.get("additionalProperties") is False:
            require(not (set(value) - set(props)), "unknown member")
        for key, subschema in props.items():
            if key in value:
                validate(value[key], subschema, f"{path}.{key}")
    if isinstance(value, list):
        require(len(value) >= schema.get("minItems", 0), "minItems")
        require(len(value) <= schema.get("maxItems", len(value)), "maxItems")
        for i, item in enumerate(value):
            validate(item, schema.get("items", {}), f"{path}[{i}]")
    if isinstance(value, str):
        if "pattern" in schema:
            require(re.search(schema["pattern"], value) is not None, "pattern")
        require(len(value) >= schema.get("minLength", 0), "minLength")
        require(len(value) <= schema.get("maxLength", len(value)), "maxLength")
    if type(value) is int:
        require(value >= schema.get("minimum", value), "minimum")
        require(value <= schema.get("maximum", value), "maximum")
    for subschema in schema.get("allOf", []):
        validate(value, subschema, path)
    if "if" in schema:
        try:
            validate(value, schema["if"], path)
        except Invalid:
            validate(value, schema.get("else", {}), path)
        else:
            validate(value, schema.get("then", {}), path)


def main():
    base = Path(__file__).resolve().parent.parent / "vectors"
    manifest = read(base / "manifest.json")
    validate(manifest, read(base / "manifest.schema.json"))
    schema = read(base / "case.schema.json")
    for entry in manifest["cases"]:
        path = (base / entry["path"]).resolve()
        if not path.is_relative_to(base):
            raise Invalid("case path escapes vector directory")
        try:
            validate(read(path), schema)
        except Invalid as error:
            raise Invalid(f"{entry['id']}: {error}") from error
    print(f"PASS: JSON schemas for manifest and {len(manifest['cases'])} cases")


if __name__ == "__main__":
    main()

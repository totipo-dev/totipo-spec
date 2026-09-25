# Requirements Profiles

Do not freeze a release requirements profile yet.

During vector development, the agent may create:

```text
requirements/v1-pre-rc.json
```

once real vector case IDs exist.

`v1-pre-rc.json` is explicitly moving and may add/remove cases while the protocol evidence is being completed.

Create a frozen release profile only when:

- the normative vector bytes are frozen;
- stable case IDs are frozen;
- the Go reference/conformance consumer passes;
- the first real/live Totipo implementation has independently consumed the portable vector set;
- remaining review has no wire/semantic blocker.

At that point create the release-candidate profile (for example `v1-rc1.json`) and never silently expand its required case set afterward.

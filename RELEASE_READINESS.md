# Mongo Developer V1.5.0 Release Readiness

Status: **manual-test candidate — not yet declared released**.

## Source gates

- 22/22 standalone frontend smoke tests PASS.
- App.vue JavaScript syntax PASS.
- 20 Go source files parse PASS.
- gofmt PASS.
- Backup archive/store unit tests: 4/4 PASS in an isolated stdlib-only test module.
- Windows metadata JSON/BOM/version consistency PASS.
- New Go explain/write-stage unit tests are included for the Windows release gate.

## Required Windows gate

From the exact V1.5.0 project root:

```powershell
cd E:\projects\Mongo-Developer-v1.5.0
Set-ExecutionPolicy -Scope Process -ExecutionPolicy Bypass
.\BUILD-INSTALLER.cmd
```

Require both:

```text
ALL RELEASE-GATE TESTS PASSED
BUILD SUCCEEDED.
```

Then install `dist\MongoDeveloper-Setup-1.5.0.exe` and complete `V1.5.0-MANUAL-TEST.md`.

## Release rule

Feature freeze is active after V1.5.0. Only release-blocking defects should be changed before publication. Any code change after a failed manual test should increment the patch version and repeat the complete release gate/manual checklist.

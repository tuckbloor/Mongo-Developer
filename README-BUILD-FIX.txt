MONGO DEVELOPER BUILD OVERLAY V2
================================

Extract THIS ZIP directly into:

    E:\projects\Mongo-Developer-v1.1.1

The ZIP intentionally starts with:
    BUILD-INSTALLER.cmd
    scripts\
    frontend\

so it overwrites the project files at the correct level.

Fixes
-----
- Correct overlay folder structure.
- Stops immediately when Go/npm/Wails returns a non-zero exit code.
- Detects NSIS under Program Files (x86) automatically.
- Detects Wails even when GOPATH\bin is not on PATH.
- Prevents a raw app executable being renamed as an installer.
- Repairs the known incomplete Rolldown optional native dependency install.
- Enforces Node.js >= 22.12.0 before the frontend build starts.
- Runs go mod tidy + go mod verify before tests.
- Adds BUILD-INSTALLER.cmd so no manual PowerShell execution-policy command is needed.

After extraction
----------------
Open PowerShell in the project root and run:

    .\BUILD-INSTALLER.cmd 1.1.1

If Node.js is too old, upgrade Node.js LTS:

    winget upgrade -e --id OpenJS.NodeJS.LTS

If winget says it is not installed:

    winget install -e --id OpenJS.NodeJS.LTS

Then CLOSE PowerShell, open a fresh PowerShell window, return to the project
folder, and run BUILD-INSTALLER.cmd again.

Expected final output:
    dist\MongoDeveloper-Setup-1.1.1.exe
    dist\MongoDeveloper-Setup-1.1.1.exe.sha256

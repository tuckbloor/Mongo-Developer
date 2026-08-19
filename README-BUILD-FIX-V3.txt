MONGO DEVELOPER BUILD OVERLAY V3
================================

Extract this ZIP directly into:
    E:\projects\Mongo-Developer-v1.1.1

Choose Replace files.

Why V3 exists
-------------
The machine's global npm installation is crashing before it can install any
project dependency. V3 no longer depends on the global npm installation.

Instead, the build:
- Downloads the official Node.js 22.22.0 Windows x64 ZIP from nodejs.org.
- Downloads the official SHASUMS256.txt.
- Verifies the archive SHA-256 before extraction.
- Installs Node/npm only under this project's .tools directory.
- Prepends that pinned local toolchain to PATH for the build.
- Deletes the failed old frontend node_modules/package-lock once.
- Runs npm install, tests, Vite build, Wails build, and NSIS packaging.
- Stops immediately at the first real error.

Run
---
    .\BUILD-INSTALLER.cmd 1.1.1

Expected result
---------------
    dist\MongoDeveloper-Setup-1.1.1.exe
    dist\MongoDeveloper-Setup-1.1.1.exe.sha256

This does NOT uninstall or modify the machine-wide Node.js/npm installation.

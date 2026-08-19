MONGO DEVELOPER BUILD OVERLAY V4
================================

Extract this ZIP directly into:

    E:\projects\Mongo-Developer-v1.1.1

Choose Replace files.

V4 fixes the npm failure in V3
-------------------------------
V3 used a private Node/npm executable, but npm was still using the user's
machine-wide cache. The reported log path proved that.

V4:
- Uses the pinned Node.js 22.22.0 toolchain from V3.
- Uses a project-only npm cache under .tools\npm-cache.
- Uses legacy-peer-deps to bypass npm/Arborist's current edgesOut peer-tree crash.
- Makes @rolldown/binding-win32-x64-msvc 1.2.4 an explicit dependency.
- Verifies the Windows Rolldown binding exists after npm install.
- Builds and tests the frontend exactly once.
- Calls `wails build -s`, so Wails skips its own frontend install/build pass.
- Stops on the first non-zero native command.
- Produces SHA-256 alongside the installer.

Run
---
    .\BUILD-INSTALLER.cmd 1.1.1

Expected output
---------------
    dist\MongoDeveloper-Setup-1.1.1.exe
    dist\MongoDeveloper-Setup-1.1.1.exe.sha256

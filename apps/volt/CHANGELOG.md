# Changelog

All notable changes to the `volt` tool (`apps/volt`, tags `volt/vX.Y.Z`) are listed here. The format follows [Keep a Changelog](https://keepachangelog.com/en/1.1.0/). Releases before 0.3.3 are described on their GitHub release pages.

## [0.3.3] - 2026-11-13

### Fixed

- `volt build` and `volt release` stamp every binary as a clean build again. Archives used to be written into `dist/` between platform builds, so with `dist/` inside a checkout that does not ignore it, `go build` saw untracked files and stamped every binary after the first with `vcs.modified=true`. Archives and `checksums.txt` are now staged in a temporary directory and moved into `dist/` after the last compile. Pinned by `TestRunKeepsVCSStampClean`.

[0.3.3]: https://github.com/ubgo/voltkit/releases/tag/volt/v0.3.3

# SYNC

- Sync files periodically using 1-URL command to start. 
- Container image provided at [ghcr.io/yosebyte/sync](https://ghcr.io/yosebyte/sync).

## Usage

```
sync "cmd://?<ivl=hour>&<src=dir>&<dst=dir>"

sync "cmd://?ivl=12&src=/path/to/source&dst=/path/to/target"
```
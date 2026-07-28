# files paths and namespace support

> **Prerequisite:** read [`../SKILL.md`](../SKILL.md) first.

Every files path is:

```text
<fileType>/<extend>/<subPath...>
```

| Segment | Meaning |
|---|---|
| `fileType` | Storage class: `drive`, `cache`, `sync`, `external`, `awss3`, `dropbox`, `google`, `tencent`, `share`, `internal` |
| `extend` | Volume / repo / account. Drive: `Home`, `Data`, or `Common` exactly. Cache/external: node name. Sync: Seafile repo id. Cloud: account key |
| `subPath` | Path inside `extend`; leading `/` is implicit |

Examples: `drive/Home/`, `drive/Home/Documents/report.pdf`, `drive/Common/huggingface/`, `sync/<repo_id>/notes/`, `awss3/<account>/<bucket>/key.txt`.

## Namespace support by verb

| Verb | Supported namespaces |
|---|---|
| `ls` / `cat` / `download` / `rm` / `rename` | `drive`, `cache`, `sync`, `external`, `awss3`, `google`, `dropbox`, `tencent` |
| `edit` | `drive`, `sync`, `cache`, `external` only |
| `mkdir` | `drive`, `cache`, `sync`, `external`, `awss3`, `google`, `dropbox`, `tencent` |
| `cp` / `mv` | same as `mkdir` |
| `upload` | `drive/Home`, `drive/Data`, `drive/Common`, `sync/<repo_id>`, `cache/<node>`, `external/<node>/<volume>`, `awss3`, `google`, `dropbox`; `tencent` rejected |
| `chown` | `drive/Home`, `drive/Data`, `drive/Common`, `cache/<node>` |
| `share internal` | `drive`, `sync`, `external`, `cache`; `drive/Common` refused |
| `share smb` | `drive`, `external`, `cache`; sync/cloud refused; `drive/Common` refused |
| `share public` | `drive` only; this is the only share flavor `drive/Common` allows |
| `compress` / `extract` / `archive entries` / `archive cat` | `drive/Home`, `drive/Data`, `drive/Common`, `cache/<node>`, `external/<node>/<volume>` only |
| `smb` / `nfs` | keyed by `<node>` + remote target, not frontend paths |
| `task` | keyed by `<node>` + `task_id`, not frontend paths |
| `repos` | Sync (Seafile) library catalog, not frontend paths |

`drive/Common`, archive verbs, and `nfs` need Olares >= 1.12.6.

## Trailing slash convention

Whether a path ends with `/` is meaningful:

| Form | Meaning |
|---|---|
| `drive/Home/Foo/` | Directory intent |
| `drive/Home/Foo` | File intent |

- `rm drive/Home/Foo/` requires `-r`.
- `upload <local> drive/Home/Documents/` uploads into the directory; `upload <local> drive/Home/Documents/file.pdf` writes that exact path.
- `cp` / `mv` destinations must end with `/` because they are drop-into-directory operations; use `rename` for in-place basename changes.
- `cp -r drive/Home/old/` requires `-r` because the source trailing slash declares a directory.

> Client-side hard constraints (the 5 numbered quirks — `external` virtual layer, protected `drive/Home`, `cache` node-picker, and the two POST/GET quirks) live in the parent [`../SKILL.md`](../SKILL.md).

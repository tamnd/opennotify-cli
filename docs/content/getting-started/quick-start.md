---
title: "Quick start"
description: "Fetch your first record with opennotify."
weight: 30
---

Once `opennotify` is on your `PATH`, fetch a page. The argument is the path
of the page on opennotify.com (everything after the host), or a full URL:

```bash
opennotify page <path>
```

By default you get an aligned table. Ask for JSON when you want to pipe it:

```bash
$ opennotify page <path> -o json
[
  {
    "id": "<path>",
    "url": "https://opennotify.com/<path>",
    "title": "<path>",
    "body": "..."
  }
]
```

## Shape the output

The same flags work on every command:

```bash
opennotify page <path> --fields id,url        # keep only these columns
opennotify page <path> --template '{{.Body}}' # just the body text
opennotify page <path> -o jsonl | jq .url     # one object per line, into jq
```

`-o` takes `table`, `json`, `jsonl`, `csv`, `tsv`, `url`, or `raw`. Left to
`auto`, it prints a table to a terminal and JSONL into a pipe, so the same
command reads well by hand and parses cleanly downstream. See
[output formats](/reference/output/) for the full contract.

## Follow the links

`links` lists the pages a page links to, and each one is a path you can fetch in
turn:

```bash
opennotify links <path> -n 10                 # the first ten links
opennotify links <path> -o url                # just the URLs
opennotify links <path> -o url | head -3 | xargs -n1 opennotify page
```

## Serve it instead

The same operations are available over HTTP and to agents over MCP:

```bash
opennotify serve --addr :7777 &
curl -s 'localhost:7777/v1/page/<path>'          # NDJSON, one record per line
opennotify mcp                                # MCP over stdio: page, links
```

## What to build next

This scaffold ships one example type, `page`, wired end to end so the whole
chain works today. To make it really about opennotify, model the records you
care about in `opennotify/` and declare their operations in
`opennotify/domain.go`. Each one you add shows up as a command here, a route
under `serve`, and a tool under `mcp`, with no extra wiring. The
[guides](/guides/) cover the common jobs.

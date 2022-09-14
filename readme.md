# Readme

Search, aggregate, backup your browsing history from the command line.

## Project status

Just started. Currently it extracts and stores all your browsing history in SQLite. It does not yet support search. That's next!

## Why?

I created [BrowserParrot][] to have GUI access to all my browsing history with a quick fuzzy search. This worked out well, but the stack chosen at the time (Clojure/JVM) turned out not to be ideal for the problem.

In this iteration if switched to Go, which can provide:

- Lower memory usage
- Quick startup time
- Smaller binary
- More consistent deployments

### Is this a rewrite of BrowserParrot?

Not currently. For now the focus is on acheiving desired UX from the command line. To be a real BrowserParrot alternative we'd need a GUI. However, I've been investigating [Wails](https://wails.io/) for a separate project and quite like it. Since this repo uses Go we'd be in a good position to wrap the functionality in a UI using Wails.

## Importing from [BrowserParrot][]

Import URLs from BrowserParrot:

```sh
browser-gopher browserparrot
```

Same as above, but with a custom DB path:

```sh
browser-gopher browserparrot --db-path ~/.config/uncloud/persistory.db
```

(This may be useful if you tried out [Uncloud](https://www.uncloud.gg/) and have a browserparrot-like database somewhere else on your system)

[browserparrot]: (https://www.browserparrot.com/)

## Todo / Wishlist

- search (yeah, need to add this)
  - actions: open, copy, etc
- a TUI for searching and filtering for a more GUI-like experience
- full text indexing
  - ideally with more sophisticated extraction mechanisms than previous

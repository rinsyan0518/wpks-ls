# wpks-ls

`wpks-ls` is a Language Server Protocol (LSP) implementation for Ruby projects that use [Packwerk](https://github.com/Shopify/packwerk).  
It provides diagnostics for Packwerk violations, helping you maintain modular boundaries in large Ruby codebases.  
This server is designed to be used with editors that support LSP, such as Neovim.

## Installation

### Requirements

- Go 1.20 or later (recommended: Go 1.24.4)
- Ruby project using Packwerk
- packwerk 3.0.0 or later

### Build

Clone this repository and build the binary:

```sh
git clone https://github.com/rinsyan0518/wpks-ls.git
cd wpks-ls
go build -o bin/wpks-ls ./cmd/wpks-ls
```

Or use the provided Taskfile (requires [go-task](https://taskfile.dev/)):

```sh
task build
```

The binary will be available at `bin/wpks-ls`.

### Install with go install

You can also install the binary directly using Go (Go 1.17+):

```sh
go install github.com/rinsyan0518/wpks-ls/cmd/wpks-ls@latest
```

This will place the `wpks-ls` binary in your `$GOPATH/bin` or `$GOBIN` directory.

## How to use with Neovim

You can use `wpks-ls` as a language server in Neovim with plugins like [nvim-lspconfig](https://github.com/neovim/nvim-lspconfig).

### Example configuration (Lua)

Add this to your Neovim config (e.g., `init.lua`):

```lua
vim.lsp.config['wpks-ls'] = {
  cmd = { '/path/to/wpks-ls/bin/wpks-ls' },
  filetypes = { 'ruby' },
  root_markers = { 'Gemfile', '.git' },
}

vim.lsp.enable('wpks-ls')
```

- Replace `/path/to/wpks-ls/bin/wpks-ls` with the actual path to your built binary.
- Make sure your Ruby project has a `packwerk.yml` at the root.

## Fallback Order

When running diagnostics, `wpks-ls` tries the following commands in order until one succeeds:

1. **pks**  
   ```
   pks -e check -- <file>
   ```
2. **bundle exec packwerk**  
   ```
   bundle exec packwerk check -- <file>
   ```
3. **packwerk**  
   ```
   packwerk check -- <file>
   ```

If a command is not found, it falls back to the next. If none succeed, no diagnostics are returned.

## Commands Executed

Depending on your environment, `wpks-ls` will execute one of the following commands to check for Packwerk violations:

- `pks -e check -- <file>`
- `bundle exec packwerk check -- <file>`
- `packwerk check -- <file>`

Make sure at least one of these commands is available in your project or system.

## License

See [LICENSE](LICENSE) for details.

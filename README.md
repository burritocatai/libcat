# LibCat

A terminal-based book tracking application built with Go and Bubble Tea.

Track your book library, reading progress, and discover new books using the Open Library API.

## Features

- Beautiful TUI interface for browsing your library
- Search and add books using the Open Library API (no API key required)
- Track reading progress with page numbers
- CLI commands for quick actions
- Persistent local storage (JSON)
- Reading status tracking (unread, reading, finished)

## Installation

```bash
go install github.com/libcat/libcat@latest
```

Or build from source:

```bash
git clone https://github.com/libcat/libcat.git
cd libcat
go build -o libcat .
```

## Usage

### TUI Mode

Run without arguments to start the interactive TUI:

```bash
libcat
```

**TUI Controls:**
- `↑/↓` or `j/k`: Navigate through books
- `Enter`: View book details
- `a` or `/`: Search and add new books
- `u`: Update current page
- `d`: Delete selected book
- `q` or `Esc`: Back/Quit

### CLI Commands

#### Add a book

By ISBN:
```bash
libcat add --isbn 9780345339683
```

By search:
```bash
libcat add --search "The Hobbit"
libcat add -s "George Orwell 1984"
```

#### List your library

```bash
libcat list
libcat ls
libcat list --status reading
libcat list --status finished
```

#### Update reading progress

Interactive mode:
```bash
libcat update
```

Direct update:
```bash
libcat update "The Hobbit" 150
libcat update 9780345339683 42
```

## Data Storage

Your library is stored in `~/.libcat/library.json`.

## API

LibCat uses the [Open Library API](https://openlibrary.org/developers/api) which is free and requires no authentication.

## Tech Stack

- [Bubble Tea](https://github.com/charmbracelet/bubbletea) - TUI framework
- [Lip Gloss](https://github.com/charmbracelet/lipgloss) - Styling
- [Bubbles](https://github.com/charmbracelet/bubbles) - TUI components
- [Cobra](https://github.com/spf13/cobra) - CLI framework

## License

MIT

## Spotify-Tui

<img width="1024" height="670" alt="image" src="https://github.com/user-attachments/assets/88363e31-b2b1-447a-8f24-7d78c68f6730" />



A terminal-based Spotify client built in Go with the [Bubble Tea](https://github.com/charmbracelet/bubbletea) framework.

## Features

- Browse and view your Spotify playlists
- View tracks within playlists
- Keyboard-driven navigation
- Persistent OAuth token storage

## Prerequisites

- Go 1.25+
- A [Spotify Developer Application](https://developer.spotify.com/dashboard) with Client ID and Client Secret

## Setup

1. Clone the repository

2. Create a `.env` file at the project root:

```
SPOTIFY_CLIENT_ID=your_client_id
SPOTIFY_CLIENT_SECRET=your_client_secret
```

3. Run the application:

```bash
make run
```

On first launch, the browser will open to authenticate with Spotify. After authorization, the token is cached at `~/.spotify-tui/token.json`.

## Controls

| Key | Action |
|---|---|
| `esc` | Logout - deletes user client + token |
| `Tab` | Cycle focus between panels |
| `↑` / `↓` or `j` / `k` | Navigate list items |
| `Enter` | Select item (playlist) |
| `←` / `→` or `h` / `l` | Navigate tabs (when focused) |
| `q` | Quit |
| `X` | Remove from playlist (when focused) |
| `/` | Navigate to search |
| `Q` | Toggle Queued Songs |
| `S` | Toggle Shuffle |
| `L` | Toggle Looping |


## Architecture


```
cmd/spotify-tui/     Entry point
internal/
  client/            Spotify API and OAuth auth clients
  entities/          Domain models (Track, Playlist, etc.)
  service/           Business logic layer
  repository/        Token persistence
  view/              UI components (Bubble Tea)
```


The codebase follows a strict layered architecture — each layer only knows about the layer below it, and dependencies are injected via interfaces defined at the consumer rather than the producer. This means each service declares only the methods it actually needs from the Spotify client, keeping things easy to test and extend.

## Data Storage

User tokens are stored securely in the system keychain and are removed on logout. Only non-sensitive configuration (the Spotify client details) is persisted locally in:

$HOME/.spotify-tui/.config.json

No secrets or sensitive data are stored in plaintext or within the application.

### Component messaging

Bubble Tea's default pattern of passing messages up through a nested model tree doesn't scale well for multi-panel UIs. Instead, the view layer uses a lightweight pub/sub message bus built on top of Bubble Tea's own `tea.Cmd` primitives. Components subscribe to message types they care about, and publishing a message returns a `tea.Cmd` that feeds back into Bubble Tea's update loop — triggering rerenders without any component needing to know about its siblings.

### Auth

OAuth2 flow with CSRF state validation and a local callback server. On subsequent launches the cached token is loaded from disk and refreshed automatically if expired — a `savingTokenSource` wrapper persists any newly refreshed tokens back to disk so restarts are seamless.

## Stack

- **UI**: [Bubble Tea](https://github.com/charmbracelet/bubbletea) + [Bubbles](https://github.com/charmbracelet/bubbles) + [Lip Gloss](https://github.com/charmbracelet/lipgloss)
- **Auth**: OAuth2 with automatic token refresh
- **API**: Custom HTTP client layer with interface-driven dependency injection

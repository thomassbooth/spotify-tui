## Spotify-Tui

<img width="954" height="682" alt="image" src="https://github.com/user-attachments/assets/f1a9bc92-7ad8-42a4-86da-6cdc872c4186" />


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
go run ./cmd/spotify-tui
```

On first launch, the browser will open to authenticate with Spotify. After authorization, the token is cached at `~/.spotify-tui/token.json`.

## Controls

| Key | Action |
|---|---|
| `Tab` / `Shift+Tab` | Cycle focus between panels |
| `↑` / `↓` or `j` / `k` | Navigate list items |
| `Enter` | Select item (playlist) |
| `←` / `→` or `h` / `l` | Navigate tabs (when focused) |
| `1` / `2` / `3` | Jump to tab (when focused) |
| `q` | Quit |

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

The UI follows a component-driven pattern with a pub/sub message bus for inter-component communication.

## Stack

- **UI**: [Bubble Tea](https://github.com/charmbracelet/bubbletea) + [Bubbles](https://github.com/charmbracelet/bubbles) + [Lip Gloss](https://github.com/charmbracelet/lipgloss)
- **Auth**: OAuth2 PKCE flow
- **API**: Custom api layer to interact with Spotifys public API

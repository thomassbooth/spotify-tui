# spotify-tui

Terminal interface to interact with the spotify api.

Built in go using Bubbletea and Lipgloss.

Architecture:

- Domain driven design layered structure

- Clients: Auth, Spotify

- Services: Interacts with clients and repositories

- Entities: Handles the apps version of the data for things like Tracks, Playlists etc.

- Repositories: Token storage

- Views: Handles all our view components

View Architecture:

- Component driven, due to there being a static navigation, playlist and playbar and a main window that changes.

- Observer style event pattern for components to interact with one another via messages.

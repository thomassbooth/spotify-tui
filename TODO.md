# TODO

## Bugs
- [X] Fix `ToggleRepeat` — params need to go in query string not request body - is this needed? idk
- [X] Fix `VolumeToggle` - make it work!
- [X] Fix `TogglePlay` — `if true` dead code, needs to check actual playback state - removed, uneeded as we have pause and resume.
- [X] Remove duplicate service methods (`Next`/`NextTrack`, `Pause`/`PausePlayback`, `Previous`/`PreviousTrack`)

## Features
- [X] Device picker — show available Spotify devices when no active device found on play, pass `device_id` as query param on the initial play call - done but, it just gets the current computer and adds it to first request - maybe i can add a devices section in the future.
- [X] Search — wire up actual Spotify search API call, currently just a UI skeleton
- [ ] Pagination — playlists and tracks both capped at Spotify's default 20, need to handle `next` cursor from responses
- [ ] First run setup flow — interactive terminal UI to guide through Spotify app creation and config instead of requiring a manual `.env`
- [x] Remove from playlist — `x` or `backspace` on a track while viewing a playlist, calls `DELETE /playlists/{id}/tracks`
- [ ] Add to playlist — `a` on a track shows a playlist picker, calls `POST /playlists/{id}/tracks`

## Architecture
- [X] Move playlists fetch in `NewSidebar` to an async `tea.Cmd` — currently blocks render on startup
- [X] Replace `time.Sleep` in playbar commands with a delayed `tea.Tick` poll
- [ ] Tidy up main.go - too much logic, it should literally just start the app not handle client calls for auth!

## Error handling
- [ ] Surface errors to the user in the TUI — currently swallowed with a TODO comment in `page.go`
- [ ] Wire up logging service

## Testing
- [ ] Write tests for `PlaylistService`
- [ ] Write tests for `PlaybackService`
- [ ] Write tests for auth flow

## Tooling
- [ ] Add GitHub Actions CI — `go test ./...` and `golangci-lint` on push
- [ ] Add GoReleaser when project is release ready
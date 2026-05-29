package service

import (
	"context"
	"errors"
	"testing"

	"github.com/thomassbooth/spotify-tui/internal/client/spotify/response"
	"github.com/thomassbooth/spotify-tui/internal/entities"
)

// --- Mock ---

type mockPlaylistClient struct {
	playlistsResp     *response.GetPlaylistsResponse
	playlistsErr      error
	playlistItemsResp *response.GetPlaylistItemsResponse
	playlistItemsErr  error
	deleteErr         error
	searchResults     *entities.SearchResults
	searchErr         error

	// Capture calls
	capturedPlaylistID string
	capturedTrackURI   string
	capturedQuery      string
}

func (m *mockPlaylistClient) GetPlaylists(_ context.Context) (*response.GetPlaylistsResponse, error) {
	return m.playlistsResp, m.playlistsErr
}

func (m *mockPlaylistClient) GetPlaylistItems(_ context.Context, id string) (*response.GetPlaylistItemsResponse, error) {
	m.capturedPlaylistID = id
	return m.playlistItemsResp, m.playlistItemsErr
}

func (m *mockPlaylistClient) DeletePlaylistItem(_ context.Context, playlistID string, trackURI string) error {
	m.capturedPlaylistID = playlistID
	m.capturedTrackURI = trackURI
	return m.deleteErr
}

func (m *mockPlaylistClient) GetSearchResults(_ context.Context, query string) (*entities.SearchResults, error) {
	m.capturedQuery = query
	return m.searchResults, m.searchErr
}

func newPlaylistService(client *mockPlaylistClient) *PlaylistService {
	svc := NewPlaylistService(client)
	return &svc
}

// --- GetPlaylists ---

func TestGetPlaylists_MapsFieldsCorrectly(t *testing.T) {
	mock := &mockPlaylistClient{
		playlistsResp: &response.GetPlaylistsResponse{
			Items: []response.PlaylistItem{
				{
					ID:          "pl1",
					Name:        "My Playlist",
					Description: "A great playlist",
					URI:         "spotify:playlist:pl1",
					Type:        "playlist",
					Owner:       response.Owner{DisplayName: "thomas"},
					Tracks:      response.PlaylistTracks{Total: 42},
					Images:      []response.Image{{URL: "https://example.com/img.jpg"}},
				},
			},
		},
	}
	svc := newPlaylistService(mock)

	playlists, err := svc.GetPlaylists()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(playlists) != 1 {
		t.Fatalf("expected 1 playlist, got %d", len(playlists))
	}

	p := playlists[0]
	if p.ID != "pl1" {
		t.Errorf("expected ID pl1, got %q", p.ID)
	}
	if p.Name != "My Playlist" {
		t.Errorf("expected Name 'My Playlist', got %q", p.Name)
	}
	if p.Description != "A great playlist" {
		t.Errorf("expected Description 'A great playlist', got %q", p.Description)
	}
	if p.URI != "spotify:playlist:pl1" {
		t.Errorf("expected URI spotify:playlist:pl1, got %q", p.URI)
	}
	if p.Type != "playlist" {
		t.Errorf("expected Type playlist, got %q", p.Type)
	}
	if p.OwnerName != "thomas" {
		t.Errorf("expected OwnerName thomas, got %q", p.OwnerName)
	}
	if p.TrackCount != 42 {
		t.Errorf("expected TrackCount 42, got %d", p.TrackCount)
	}
	if p.ImageURL != "https://example.com/img.jpg" {
		t.Errorf("expected ImageURL https://example.com/img.jpg, got %q", p.ImageURL)
	}
}

func TestGetPlaylists_PicksFirstImage(t *testing.T) {
	mock := &mockPlaylistClient{
		playlistsResp: &response.GetPlaylistsResponse{
			Items: []response.PlaylistItem{
				{
					Images: []response.Image{
						{URL: "https://example.com/large.jpg"},
						{URL: "https://example.com/small.jpg"},
					},
				},
			},
		},
	}
	svc := newPlaylistService(mock)

	playlists, err := svc.GetPlaylists()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if playlists[0].ImageURL != "https://example.com/large.jpg" {
		t.Errorf("expected first image URL, got %q", playlists[0].ImageURL)
	}
}

func TestGetPlaylists_NoImages_LeavesImageURLEmpty(t *testing.T) {
	mock := &mockPlaylistClient{
		playlistsResp: &response.GetPlaylistsResponse{
			Items: []response.PlaylistItem{
				{Images: []response.Image{}},
			},
		},
	}
	svc := newPlaylistService(mock)

	playlists, err := svc.GetPlaylists()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if playlists[0].ImageURL != "" {
		t.Errorf("expected empty ImageURL, got %q", playlists[0].ImageURL)
	}
}

func TestGetPlaylists_EmptyItems_ReturnsEmptySlice(t *testing.T) {
	mock := &mockPlaylistClient{
		playlistsResp: &response.GetPlaylistsResponse{Items: []response.PlaylistItem{}},
	}
	svc := newPlaylistService(mock)

	playlists, err := svc.GetPlaylists()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(playlists) != 0 {
		t.Errorf("expected empty slice, got %d items", len(playlists))
	}
}

func TestGetPlaylists_PropagatesError(t *testing.T) {
	mock := &mockPlaylistClient{playlistsErr: errors.New("api error")}
	svc := newPlaylistService(mock)

	_, err := svc.GetPlaylists()
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

// --- GetPlaylistTracks ---

func TestGetPlaylistTracks_MapsFieldsCorrectly(t *testing.T) {
	mock := &mockPlaylistClient{
		playlistItemsResp: &response.GetPlaylistItemsResponse{
			Items: []response.PlaylistTrackItem{
				{
					Track: response.Track{
						ID:         "track1",
						Name:       "Song One",
						DurationMs: 210000,
						Artists: []response.Artist{
							{ID: "artist1", Name: "Artist One", URI: "spotify:artist:artist1"},
						},
						Album: response.Album{ID: "album1", Name: "Album One"},
					},
				},
			},
		},
	}
	svc := newPlaylistService(mock)

	tracks, err := svc.GetPlaylistTracks("pl1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(tracks) != 1 {
		t.Fatalf("expected 1 track, got %d", len(tracks))
	}

	tr := tracks[0]
	if tr.ID != "track1" {
		t.Errorf("expected ID track1, got %q", tr.ID)
	}
	if tr.Name != "Song One" {
		t.Errorf("expected Name 'Song One', got %q", tr.Name)
	}
	if tr.DurationMs != 210000 {
		t.Errorf("expected DurationMs 210000, got %d", tr.DurationMs)
	}
	if tr.Album.ID != "album1" {
		t.Errorf("expected Album.ID album1, got %q", tr.Album.ID)
	}
	if tr.Album.Name != "Album One" {
		t.Errorf("expected Album.Name 'Album One', got %q", tr.Album.Name)
	}
	if len(tr.Artists) != 1 {
		t.Fatalf("expected 1 artist, got %d", len(tr.Artists))
	}
	if tr.Artists[0].ID != "artist1" || tr.Artists[0].Name != "Artist One" || tr.Artists[0].URI != "spotify:artist:artist1" {
		t.Errorf("unexpected artist: %+v", tr.Artists[0])
	}
}

func TestGetPlaylistTracks_MultipleArtists(t *testing.T) {
	mock := &mockPlaylistClient{
		playlistItemsResp: &response.GetPlaylistItemsResponse{
			Items: []response.PlaylistTrackItem{
				{
					Track: response.Track{
						Artists: []response.Artist{
							{ID: "a1", Name: "Artist A"},
							{ID: "a2", Name: "Artist B"},
						},
					},
				},
			},
		},
	}
	svc := newPlaylistService(mock)

	tracks, err := svc.GetPlaylistTracks("pl1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(tracks[0].Artists) != 2 {
		t.Errorf("expected 2 artists, got %d", len(tracks[0].Artists))
	}
}

func TestGetPlaylistTracks_PassesIDToClient(t *testing.T) {
	mock := &mockPlaylistClient{
		playlistItemsResp: &response.GetPlaylistItemsResponse{},
	}
	svc := newPlaylistService(mock)

	_, _ = svc.GetPlaylistTracks("my-playlist-id")
	if mock.capturedPlaylistID != "my-playlist-id" {
		t.Errorf("expected playlist ID my-playlist-id, got %q", mock.capturedPlaylistID)
	}
}

func TestGetPlaylistTracks_EmptyItems_ReturnsEmptySlice(t *testing.T) {
	mock := &mockPlaylistClient{
		playlistItemsResp: &response.GetPlaylistItemsResponse{Items: []response.PlaylistTrackItem{}},
	}
	svc := newPlaylistService(mock)

	tracks, err := svc.GetPlaylistTracks("pl1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(tracks) != 0 {
		t.Errorf("expected empty slice, got %d items", len(tracks))
	}
}

func TestGetPlaylistTracks_PropagatesError(t *testing.T) {
	mock := &mockPlaylistClient{playlistItemsErr: errors.New("not found")}
	svc := newPlaylistService(mock)

	_, err := svc.GetPlaylistTracks("pl1")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

// --- RemoveFromPlaylist ---

func TestRemoveFromPlaylist_BuildsCorrectURI(t *testing.T) {
	mock := &mockPlaylistClient{}
	svc := newPlaylistService(mock)

	err := svc.RemoveFromPlaylist("pl1", "track123")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if mock.capturedTrackURI != "spotify:track:track123" {
		t.Errorf("expected URI spotify:track:track123, got %q", mock.capturedTrackURI)
	}
}

func TestRemoveFromPlaylist_PassesPlaylistID(t *testing.T) {
	mock := &mockPlaylistClient{}
	svc := newPlaylistService(mock)

	_ = svc.RemoveFromPlaylist("pl-abc", "track123")
	if mock.capturedPlaylistID != "pl-abc" {
		t.Errorf("expected playlist ID pl-abc, got %q", mock.capturedPlaylistID)
	}
}

func TestRemoveFromPlaylist_PropagatesError(t *testing.T) {
	mock := &mockPlaylistClient{deleteErr: errors.New("delete failed")}
	svc := newPlaylistService(mock)

	err := svc.RemoveFromPlaylist("pl1", "track1")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

// --- GetSearchResults ---

func TestGetSearchResults_ReturnsResults(t *testing.T) {
	expected := &entities.SearchResults{}
	mock := &mockPlaylistClient{searchResults: expected}
	svc := newPlaylistService(mock)

	got, err := svc.GetSearchResults("radiohead")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != expected {
		t.Errorf("expected %v, got %v", expected, got)
	}
}

func TestGetSearchResults_PassesQueryToClient(t *testing.T) {
	mock := &mockPlaylistClient{searchResults: &entities.SearchResults{}}
	svc := newPlaylistService(mock)

	_, _ = svc.GetSearchResults("my query")
	if mock.capturedQuery != "my query" {
		t.Errorf("expected query 'my query', got %q", mock.capturedQuery)
	}
}

func TestGetSearchResults_PropagatesError(t *testing.T) {
	mock := &mockPlaylistClient{searchErr: errors.New("search error")}
	svc := newPlaylistService(mock)

	_, err := svc.GetSearchResults("anything")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}
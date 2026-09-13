package main

import (
	"context"
	"sync"
	"time"

	"plexcord/internal/config"
	"plexcord/internal/discord"
	"plexcord/internal/history"
	"plexcord/internal/platform"
	"plexcord/internal/plex"
	"plexcord/internal/retry"
	"plexcord/internal/updater"
	"plexcord/internal/version"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// Shared test doubles for the interfaces App depends on. Having a fake for
// every collaborator is the practical proof that the dependency inversion is
// real: the whole binding surface can be driven with no Plex server, no
// Discord socket, no keychain, no system tray and no Wails window.

// ----------------------------------------------------------------------------
// Desktop (Wails runtime)
// ----------------------------------------------------------------------------

// fakeDesktop records every window/process/browser call.
type fakeDesktop struct {
	mu          sync.Mutex
	shown       int
	hidden      int
	minimised   int
	unminimised int
	quits       int
	alwaysOnTop []bool
	sizes       [][2]int
	centered    int
	openedURLs  []string

	// isMinimised is what IsMinimised reports.
	isMinimised bool
	// screens and screensErr drive Screens.
	screens    []runtime.Screen
	screensErr error
}

func (d *fakeDesktop) Show(context.Context) {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.shown++
}

func (d *fakeDesktop) Hide(context.Context) {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.hidden++
}

func (d *fakeDesktop) Minimise(context.Context) {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.minimised++
}

func (d *fakeDesktop) Unminimise(context.Context) {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.unminimised++
}

func (d *fakeDesktop) IsMinimised(context.Context) bool {
	d.mu.Lock()
	defer d.mu.Unlock()
	return d.isMinimised
}

func (d *fakeDesktop) SetAlwaysOnTop(_ context.Context, onTop bool) {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.alwaysOnTop = append(d.alwaysOnTop, onTop)
}

func (d *fakeDesktop) SetSize(_ context.Context, w, h int) {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.sizes = append(d.sizes, [2]int{w, h})
}

func (d *fakeDesktop) Center(context.Context) {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.centered++
}

func (d *fakeDesktop) Quit(context.Context) {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.quits++
}

func (d *fakeDesktop) OpenURL(_ context.Context, url string) {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.openedURLs = append(d.openedURLs, url)
}

func (d *fakeDesktop) Screens(context.Context) ([]runtime.Screen, error) {
	d.mu.Lock()
	defer d.mu.Unlock()
	return d.screens, d.screensErr
}

func (d *fakeDesktop) counts() (shown, hidden, quits int) {
	d.mu.Lock()
	defer d.mu.Unlock()
	return d.shown, d.hidden, d.quits
}

var _ Desktop = (*fakeDesktop)(nil)

// ----------------------------------------------------------------------------
// Platform integration
// ----------------------------------------------------------------------------

// fakeTray records tray lifecycle and the notices pushed to it.
type fakeTray struct {
	mu      sync.Mutex
	started int
	stopped int
	notices []platform.UpdateNotice
}

func (t *fakeTray) Start() {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.started++
}

func (t *fakeTray) Stop() {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.stopped++
}

func (t *fakeTray) SetUpdateNotice(n platform.UpdateNotice) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.notices = append(t.notices, n)
}

func (t *fakeTray) lastNotice() (platform.UpdateNotice, bool) {
	t.mu.Lock()
	defer t.mu.Unlock()
	if len(t.notices) == 0 {
		return platform.UpdateNotice{}, false
	}
	return t.notices[len(t.notices)-1], true
}

var _ TrayController = (*fakeTray)(nil)

// fakeAutoStart records launch-on-login registration without touching the
// registry, a LaunchAgent plist, or an XDG desktop file.
type fakeAutoStart struct {
	enabled  bool
	ensured  int
	setCalls []bool
	setErr   error
}

func (f *fakeAutoStart) IsEnabled() bool { return f.enabled }

func (f *fakeAutoStart) SetEnabled(enabled bool) error {
	f.setCalls = append(f.setCalls, enabled)
	if f.setErr != nil {
		return f.setErr
	}
	f.enabled = enabled
	return nil
}

func (f *fakeAutoStart) Disable() error { return f.SetEnabled(false) }

func (f *fakeAutoStart) EnsureRegistered() error {
	f.ensured++
	return nil
}

var _ AutoStartController = (*fakeAutoStart)(nil)

// ----------------------------------------------------------------------------
// Storage
// ----------------------------------------------------------------------------

// fakeTokenStore is an in-memory credential store.
type fakeTokenStore struct {
	token   string
	getErr  error
	setErr  error
	deleted bool
}

func (f *fakeTokenStore) Get() (string, error) { return f.token, f.getErr }

func (f *fakeTokenStore) Set(token string) error {
	if f.setErr != nil {
		return f.setErr
	}
	f.token = token
	return nil
}

func (f *fakeTokenStore) Delete() error {
	f.deleted = true
	f.token = ""
	return nil
}

var _ TokenStore = (*fakeTokenStore)(nil)

// fakeConfigGateway serves a config from memory instead of from disk.
type fakeConfigGateway struct {
	cfg           *config.Config
	loadErr       error
	deleted       bool
	setupComplete bool
	dir           string
}

func (f *fakeConfigGateway) Load() (*config.Config, error) {
	if f.loadErr != nil {
		return nil, f.loadErr
	}
	if f.cfg == nil {
		return config.DefaultConfig(), nil
	}
	return f.cfg, nil
}

func (f *fakeConfigGateway) Delete() error {
	f.deleted = true
	return nil
}

func (f *fakeConfigGateway) IsSetupComplete() bool { return f.setupComplete }
func (f *fakeConfigGateway) ConfigDir() string     { return f.dir }

var _ ConfigGateway = (*fakeConfigGateway)(nil)

// fakeHistory keeps listening history in memory.
type fakeHistory struct {
	mu      sync.Mutex
	entries []history.Entry
	cleared int
}

func (h *fakeHistory) Add(e history.Entry) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.entries = append(h.entries, e)
}

func (h *fakeHistory) GetRecent(limit int) []history.Entry {
	h.mu.Lock()
	defer h.mu.Unlock()
	if limit <= 0 || limit > len(h.entries) {
		limit = len(h.entries)
	}
	out := make([]history.Entry, limit)
	copy(out, h.entries[:limit])
	return out
}

func (h *fakeHistory) GetStats() history.Stats {
	h.mu.Lock()
	defer h.mu.Unlock()
	return history.Stats{TotalTracks: len(h.entries)}
}

func (h *fakeHistory) Clear() {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.entries = nil
	h.cleared++
}

var _ HistoryStore = (*fakeHistory)(nil)

// ----------------------------------------------------------------------------
// Plex
// ----------------------------------------------------------------------------

// fakePlexAPI answers the four Plex calls from canned data.
type fakePlexAPI struct {
	validation *plex.ValidationResult
	validErr   error
	users      []plex.PlexUser
	usersErr   error
	music      []plex.MusicSession
	musicErr   error
	media      []plex.MediaSession
}

func (f *fakePlexAPI) ValidateConnection() (*plex.ValidationResult, error) {
	if f.validErr != nil {
		return nil, f.validErr
	}
	if f.validation != nil {
		return f.validation, nil
	}
	return &plex.ValidationResult{Success: true, ServerName: "Fake Plex"}, nil
}

func (f *fakePlexAPI) GetUsers() ([]plex.PlexUser, error) { return f.users, f.usersErr }

func (f *fakePlexAPI) GetMusicSessions(string) ([]plex.MusicSession, error) {
	return f.music, f.musicErr
}

func (f *fakePlexAPI) GetMediaSessions(string, []string) ([]plex.MediaSession, error) {
	return f.media, nil
}

var _ PlexAPI = (*fakePlexAPI)(nil)

// fakeDiscoverer returns canned servers instead of doing GDM multicast.
type fakeDiscoverer struct {
	servers []plex.Server
	err     error
}

func (f *fakeDiscoverer) Discover(time.Duration) ([]plex.Server, error) {
	return f.servers, f.err
}

var _ ServerDiscoverer = (*fakeDiscoverer)(nil)

// ----------------------------------------------------------------------------
// Updates
// ----------------------------------------------------------------------------

// fakeUpdater stands in for the background update checker.
type fakeUpdater struct {
	mu        sync.Mutex
	status    updater.Status
	started   int
	stopped   int
	listener  func(updater.Status)
	checkInfo *version.UpdateInfo
	checkErr  error
}

func (u *fakeUpdater) StartChecker(context.Context) {
	u.mu.Lock()
	defer u.mu.Unlock()
	u.started++
}

func (u *fakeUpdater) StopChecker() {
	u.mu.Lock()
	defer u.mu.Unlock()
	u.stopped++
}

func (u *fakeUpdater) GetStatus() updater.Status {
	u.mu.Lock()
	defer u.mu.Unlock()
	return u.status
}

func (u *fakeUpdater) OnStatusChange(fn func(updater.Status)) {
	u.mu.Lock()
	u.listener = fn
	status := u.status
	u.mu.Unlock()
	if fn != nil {
		fn(status)
	}
}

// publish sets a new status and notifies the registered listener, mimicking a
// real lifecycle transition.
func (u *fakeUpdater) publish(status updater.Status) {
	u.mu.Lock()
	u.status = status
	fn := u.listener
	u.mu.Unlock()
	if fn != nil {
		fn(status)
	}
}

func (u *fakeUpdater) Check() (*version.UpdateInfo, error) { return u.checkInfo, u.checkErr }

func (u *fakeUpdater) StartDownload(context.Context, bool) (*version.UpdateInfo, error) {
	return u.checkInfo, u.checkErr
}

var _ UpdateService = (*fakeUpdater)(nil)

// ----------------------------------------------------------------------------
// Discord
// ----------------------------------------------------------------------------

// recordingPresence captures what PlexCord publishes to Discord.
type recordingPresence struct {
	mu           sync.Mutex
	connected    bool
	clientID     string
	connectErr   error
	connectCalls []string
	updates      []discord.Playback
	lastOptions  discord.Options
	clears       int
	disconnects  int
	setCalls     []*discord.PresenceData
}

func (p *recordingPresence) Connect(clientID string) error {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.connectCalls = append(p.connectCalls, clientID)
	if p.connectErr != nil {
		return p.connectErr
	}
	p.connected = true
	p.clientID = clientID
	return nil
}

func (p *recordingPresence) Disconnect() error {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.disconnects++
	p.connected = false
	return nil
}

func (p *recordingPresence) IsConnected() bool {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.connected
}

func (p *recordingPresence) GetClientID() string {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.clientID
}

func (p *recordingPresence) SetPresence(data *discord.PresenceData) error {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.setCalls = append(p.setCalls, data)
	return nil
}

func (p *recordingPresence) ClearPresence() error {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.clears++
	return nil
}

func (p *recordingPresence) UpdatePlayback(playback discord.Playback, opts discord.Options) error {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.updates = append(p.updates, playback)
	p.lastOptions = opts
	return nil
}

func (p *recordingPresence) lastPlayback() (discord.Playback, bool) {
	p.mu.Lock()
	defer p.mu.Unlock()
	if len(p.updates) == 0 {
		return discord.Playback{}, false
	}
	return p.updates[len(p.updates)-1], true
}

func (p *recordingPresence) updateCount() int {
	p.mu.Lock()
	defer p.mu.Unlock()
	return len(p.updates)
}

func (p *recordingPresence) clearCount() int {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.clears
}

var _ DiscordPresence = (*recordingPresence)(nil)

// ----------------------------------------------------------------------------
// Reconnection and PIN authentication
// ----------------------------------------------------------------------------

// fakeRetry records how the reconnection loop is driven, with no timers.
type fakeRetry struct {
	mu            sync.Mutex
	starts        []string // error codes passed to Start
	resets        int
	stops         int
	manualRetries int
	state         retry.RetryState
	onRetry       retry.RetryCallback
}

func (r *fakeRetry) SetCallbacks(onRetry retry.RetryCallback, _ retry.StateChangeCallback) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.onRetry = onRetry
}

func (r *fakeRetry) Start(_ error, code string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.starts = append(r.starts, code)
}

func (r *fakeRetry) Stop() {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.stops++
}

func (r *fakeRetry) Reset() {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.resets++
}

func (r *fakeRetry) ManualRetry() {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.manualRetries++
}

func (r *fakeRetry) GetState() retry.RetryState {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.state
}

func (r *fakeRetry) startCount() int {
	r.mu.Lock()
	defer r.mu.Unlock()
	return len(r.starts)
}

var _ RetryManager = (*fakeRetry)(nil)

// fakeAuthenticator answers the plex.tv PIN flow from canned responses.
type fakeAuthenticator struct {
	requestResp *plex.PINResponse
	requestErr  error
	checkResp   *plex.PINResponse
	checkErr    error
	authURL     string
}

func (f *fakeAuthenticator) RequestPIN(context.Context) (*plex.PINResponse, error) {
	return f.requestResp, f.requestErr
}

func (f *fakeAuthenticator) CheckPIN(context.Context, int) (*plex.PINResponse, error) {
	return f.checkResp, f.checkErr
}

func (f *fakeAuthenticator) GetAuthURL(string) string { return f.authURL }

var _ PlexAuthenticator = (*fakeAuthenticator)(nil)

// fakeRelauncher records a restart request without spawning a process.
type fakeRelauncher struct {
	calls int
	err   error
}

func (r *fakeRelauncher) Relaunch() error {
	r.calls++
	return r.err
}

var _ AppRelauncher = (*fakeRelauncher)(nil)

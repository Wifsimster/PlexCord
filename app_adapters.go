package main

import (
	"context"
	"os"
	"os/exec"
	"strconv"
	"time"

	"plexcord/internal/config"
	"plexcord/internal/errors"
	"plexcord/internal/keychain"
	"plexcord/internal/platform"
	"plexcord/internal/plex"
	"plexcord/internal/version"
)

// This file contains adapters that wrap the concrete internal packages
// to satisfy the interfaces defined in app_interfaces.go. They are the
// production implementations injected into App at startup.
//
// Each adapter is deliberately trivial: it does nothing but forward. That is
// what keeps the interface boundary honest — there is no logic here that a
// test replacing the adapter would lose.

// keychainTokenStore adapts the package-level keychain functions to
// the TokenStore interface for dependency injection.
type keychainTokenStore struct{}

func (keychainTokenStore) Get() (string, error)   { return keychain.GetToken() }
func (keychainTokenStore) Set(token string) error { return keychain.SetToken(token) }
func (keychainTokenStore) Delete() error          { return keychain.DeleteToken() }

// newKeychainTokenStore returns the default OS-keychain-backed TokenStore.
func newKeychainTokenStore() TokenStore {
	return keychainTokenStore{}
}

// newPlexClientFactory returns the default PlexAPIFactory that constructs
// concrete *plex.Client instances. Tests can replace this with a factory
// that returns fakes.
func newPlexClientFactory() PlexAPIFactory {
	return func(token, serverURL string) PlexAPI {
		return plex.NewClient(token, serverURL)
	}
}

// gdmDiscoverer adapts the package-level GDM discovery to ServerDiscoverer.
type gdmDiscoverer struct{}

func (gdmDiscoverer) Discover(timeout time.Duration) ([]plex.Server, error) {
	return plex.DiscoverServers(timeout)
}

// newServerDiscoverer returns the default GDM-backed ServerDiscoverer.
func newServerDiscoverer() ServerDiscoverer { return gdmDiscoverer{} }

// fileConfigGateway adapts the config package's file-backed top-level
// functions to the ConfigGateway interface.
type fileConfigGateway struct{}

func (fileConfigGateway) Load() (*config.Config, error) { return config.Load() }
func (fileConfigGateway) Delete() error                 { return config.Delete() }
func (fileConfigGateway) IsSetupComplete() bool         { return config.IsSetupComplete() }
func (fileConfigGateway) ConfigDir() string             { return config.GetConfigDir() }

// newConfigGateway returns the default file-backed ConfigGateway.
func newConfigGateway() ConfigGateway { return fileConfigGateway{} }

// execRelauncher spawns the application binary as a child process.
type execRelauncher struct{}

// Relaunch starts the updated executable, handing it this process's PID so it
// waits for the single-instance lock to be released before claiming it.
func (execRelauncher) Relaunch() error {
	// Use the launch path captured at startup, NOT a fresh os.Executable().
	// The self-update renames the running binary to ".<name>.old" and moves the
	// new binary into the original path; resolving the path after that rename
	// would relaunch the OLD binary (this is exactly what os.Executable() returns
	// on Windows post-rename). The captured path always points at the original
	// location, which now holds the updated binary. See version.CaptureLaunchPath.
	exe := version.LaunchPath()
	if exe == "" {
		return errors.New(errors.UNKNOWN_ERROR, "failed to locate executable")
	}

	// Use context.Background (not the app context): that context is cancelled
	// by the quit that follows, which would otherwise terminate the relaunched
	// process. The child must outlive this one.
	// #nosec G204 G702 -- exe is our own executable ($APPIMAGE or os.Executable), not untrusted input
	cmd := exec.CommandContext(context.Background(), exe) //nolint:gosec
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	// Hand the child our PID so it waits for this process to fully exit — and
	// thereby release the Wails single-instance lock — before its own runtime
	// tries to acquire that lock. Without this, the freshly-installed binary
	// starts while we still hold the lock, is treated as a second instance
	// (it merely restores our window and exits), and the OLD version keeps
	// running. See waitForPreviousInstanceExit in relaunch.go.
	cmd.Env = append(os.Environ(), relaunchPIDEnv+"="+strconv.Itoa(os.Getpid()))
	if err := cmd.Start(); err != nil {
		return errors.Wrap(err, errors.UNKNOWN_ERROR, "failed to relaunch application")
	}
	return nil
}

// newAppRelauncher returns the default exec-backed relauncher.
func newAppRelauncher() AppRelauncher { return execRelauncher{} }

// Compile-time assertions that the production types satisfy the interfaces
// App depends on. These keep a signature change in an internal package from
// silently becoming a runtime surprise at startup.
var (
	_ TrayController      = (*platform.TrayManager)(nil)
	_ AutoStartController = (*platform.AutoStartManager)(nil)
)

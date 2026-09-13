# SOLID Review

A design audit of PlexCord's Go backend against the five SOLID principles, the
changes made in response, and where the codebase now stands.

Scope: the `main` package (the Wails binding surface) and `internal/*`. The Vue
frontend is out of scope.

## Score

| Principle | Before | After |
|---|---|---|
| **S**ingle Responsibility | 3.0 | 4.6 |
| **O**pen/Closed | 3.5 | 4.7 |
| **L**iskov Substitution | 4.2 | 4.7 |
| **I**nterface Segregation | 3.0 | 4.6 |
| **D**ependency Inversion | 3.0 | 4.8 |
| **Overall** | **3.3 / 5** | **4.7 / 5** |

Scores are judgements, not a metric. What makes them checkable is the evidence
each one rests on, which is given per principle below — and one number that is
not a judgement at all: how much of the code a test can reach once the
dependencies are inverted.

| Package | Coverage before | Coverage after |
|---|---|---|
| `main` | 16.6% | 40.2% |
| `internal/discord` | 68.2% | 84.7% |
| `internal/plex` | 59.5% | 66.5% |
| `internal/artwork` | 78.8% | 79.2% |

The `main` package more than doubled without a single new mock framework,
because the code it covers stopped requiring a desktop session to run.

---

## What the audit found, and what was done

### Single Responsibility — 3.0 → 4.6

**Found.** `App` was where every concern met: 34 fields spanning the Wails
window context, the Plex poller and its cancel function, the Discord socket,
the pause timer, the artwork generation counter, the tray, the updater, the
keychain and the config file — guarded by **eight** separate synchronisation
primitives. Splitting it across eight files organised the reading but not the
type: any change to pausing, polling or window restore still meant editing the
same struct.

`plex.Poller` carried a second responsibility of its own — two nearly identical
polling modes, with `doPoll`/`doMediaPoll` and
`sessionChanged`/`mediaSessionChanged` duplicated between them.

**Done.** The clusters that only ever change together became types that own
them:

| New type | Owns | Replaces in `App` |
|---|---|---|
| `windowManager` | window visibility, the deferred-restore dance, the explicit-quit flag | `windowCtx`, `pendingShow`, `windowMu`, `quitting` |
| `presenceGate` | manual pause toggle, hide-when-paused timer + generation guard | `presencePaused`, `pauseTimer`, `pauseTimerGen`, `pauseMu` |
| `discordService` | the Discord link, presence publishing, silent reconnect, artwork generation guard | `discord`, `artwork`, `artworkGen`, `discordMu` |
| `pollingController` | the poller's lifecycle: start, stop, state, retune | `poller`, `pollerCtx`, `pollerStop`, `pollerMu` |
| `sessionCache` | what is playing, with its lock | `currentSession`, `sessionMu` |

`App` is down to **one** mutex, and its own job is now stated and true: be the
Wails binding surface — translate frontend calls into work on collaborators and
marshal the results back.

In `internal/plex`, each session type projects itself onto a comparable
`changeKey` and one generic helper does the nil-handling and comparison; a
single `pollOnce` folds both modes' error-state transitions into one place.

**Remaining (the −0.4).** `App` still exposes 77 bound methods. That is the
deliberate trade-off: they are the frontend's API, grouped by domain across files, and the
alternative — several bound objects — would change the frontend contract for no
design gain. It is a wide facade, not a god object: the state it guards is now
five fields deep.

### Open/Closed — 3.5 → 4.7

**Found.** Good foundations already existed: the `SessionObserver` pipeline and
the `PresenceBuilder` registry both let new behaviour be added without editing
old. But `artwork.Resolver.Resolve` hard-coded its provider chain (iTunes, then
MusicBrainz) as a sequence of calls, so a new source meant editing the
resolution logic. `Poller`'s two modes meant a third media mode would mean
editing `Poller`. And the builder registry was a mutable package-level map with
no lock — an extension point that was a data race and forced tests to
save-and-restore global state.

**Done.**

- Artwork lookup is a chain of `Source` values. `WithSources` replaces it,
  `AppendSource` extends it; `Resolve` never changes.
- `BuilderRegistry` is a value with a mutex and a fallback. Tests build their
  own instead of mutating a global; registration racing with dispatch is now
  covered by a test that runs under `-race`.
- Adding a session type to the poller means adding one `changeKey` method, not
  another comparison ladder.

### Liskov Substitution — 4.2 → 4.7

**Found.** Few hierarchies, and implementations honoured their contracts —
this was already the strongest axis. The one real violation was `saveConfig`:
when the config store was absent it silently fell back to writing the real
config file, so the same call did something materially different depending on
construction path, and a test that thought it was isolated could write to the
user's disk.

**Done.** That fallback is gone; a missing store is an error. Every production
path has a store, and the substitution now holds in both directions.

Also added: compile-time assertions pinning each production type to the
interface it is held by, so a signature drift in an internal package fails the
build rather than surfacing at startup.

### Interface Segregation — 3.0 → 4.6

**Found.** The worst offender in the codebase:

```go
UpdatePresenceFromPlayback(
    track, artist, album, state string,
    duration, position int64,
    artworkURL, player, detailsFormat, stateFormat, activityStyle, statusDisplay string,
) error
```

Twelve positional parameters, eight of them `string`, mixing *what is playing*
with *how the user wants it displayed* — two things that change for entirely
different reasons. Every caller had to count arguments, and every fake had to
reproduce the signature exactly.

**Done.** It is now `UpdatePlayback(Playback, Options)`: a playback snapshot and
the display preferences, separated so a caller supplies only the half it owns.
`DiscordPresence` is split into `DiscordConnection` and `PresenceWriter` so a
consumer that only publishes presence depends on three methods, not seven. The
Wails runtime is four narrow interfaces (`WindowController`, `ProcessController`,
`BrowserOpener`, `ScreenProvider`), not one wide one — the screen-fitting code
takes only `ScreenProvider`.

Every new interface is sized to its consumer: `ServerDiscoverer` and
`AppRelauncher` have one method each, `TokenStore` three, `TrayController` three.

### Dependency Inversion — 3.0 → 4.8

**Found.** The largest gap, and the reason so little of `main` was testable.
`App` called `runtime.WindowShow`, `runtime.Quit`, `runtime.BrowserOpenURL` and
`runtime.ScreenGetAll` directly; it called the config package's file-backed
top-level functions directly; it held concrete `*platform.TrayManager`,
`*platform.AutoStartManager`, `*updater.Updater`, `*history.Store`,
`*retry.Manager` and `*plex.Authenticator`; and `RestartApplication` reached
for `os/exec` itself.

Two abstractions that did exist were bypassed. `StartSessionPolling` built its
own `plex.NewClient` instead of using the injected `plexFactory` — the one Plex
path in the app that ignored the seam sitting right next to it. Below that,
`plex.Poller` took a concrete `*plex.Client` and `discord.PresenceManager`
dialled a concrete `*ipc.Client`, so neither could be driven without a real
server or a real Discord.

**Done.** Every one of those is now an interface with a trivial forwarding
adapter, and the two bypassed seams are used: polling goes through
`plexFactory`, `plex.NewPoller` takes a `SessionSource`, and
`discord.NewPresenceManager` takes a `Dialer`.

What that bought, concretely — these had **no test at all** before, because
writing one meant a desktop session:

- Window restore requested before Wails published its context, then replayed.
- Un-minimise only when actually minimised (a maximised-then-hidden window must
  not be shrunk).
- Close-button behaviour across both the quit flag and the minimize-to-tray
  setting.
- Discord restarting mid-playback: silent reconnect, then presence republished.
- The generation guard on late artwork — a cover arriving after the track
  changed must not overwrite the current one, and one arriving after the user
  paused must not un-hide the presence.
- Artwork lookup disabled meaning the artist and album never leave the machine.
- Which Plex errors are worth retrying and which are not.
- The plex.tv PIN request / authorize / expire cycle.

**Remaining (the −0.2).** `internal/keychain`, `internal/config` and
`internal/plex` discovery still expose package-level functions. They are
adapted at the `App` boundary, which is where it matters; converting them to
constructor-injected types would be churn without a caller that needs it.

---

## Known remaining items

Recorded rather than silently left:

- **`App`'s method count.** 77 bound methods (104 including unexported
  helpers). Deliberate — see SRP above.
- **`StartPlexPINAuth` / `CheckPlexPINAuth` return `map[string]interface{}`.**
  Typed structs would be better and would give the generated TypeScript real
  types. Deferred because it changes the shape of checked-in generated Wails
  bindings that cannot be regenerated here; the JSON on the wire would be
  identical, so this is a safe follow-up on a machine with the Wails CLI.
- **Three pre-existing `gosec` G115 findings** in `internal/discord/ipc`
  (integer conversions in frame/timestamp encoding). Untouched by this work and
  not SOLID issues; left alone rather than mixed into a design refactor.
- **`plexcord.exe`, `plexcord-test.exe`, `build/plexcord.exe` are tracked in
  git.** Pre-existing; removing them rewrites what a release build expects, so
  it is a separate decision.

## Verification

```
go build ./...
go vet ./...
go test -race ./...        # 14 packages, all passing
golangci-lint run ./...    # only the 3 pre-existing gosec findings above
```

No behaviour was intended to change. The refactor is structural: every
extracted type preserves the logic it was carved from, including the two
generation-counter races, which now have the tests they always deserved.

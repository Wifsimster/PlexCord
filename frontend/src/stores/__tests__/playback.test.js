import { describe, it, expect, beforeEach, vi } from 'vitest'
import { setActivePinia, createPinia } from 'pinia'

// Mock Wails runtime
vi.mock('../../../wailsjs/runtime/runtime', () => ({
  EventsOn: vi.fn(),
  EventsOff: vi.fn()
}))

// Mock backend calls
vi.mock('../../../wailsjs/go/main/App', () => ({
  GetCurrentSession: vi.fn().mockResolvedValue(null)
}))

import { usePlaybackStore } from '../playback'
import { EventsOn, EventsOff } from '../../../wailsjs/runtime/runtime'
import { GetCurrentSession } from '../../../wailsjs/go/main/App'

const makeMockSession = (overrides = {}) => ({
  sessionKey: 'abc123',
  track: 'Test Track',
  artist: 'Test Artist',
  album: 'Test Album',
  thumb: '/library/metadata/123/thumb',
  thumbUrl: 'http://plex:32400/photo/:/transcode?url=...',
  duration: 240000,
  viewOffset: 60000,
  state: 'playing',
  playerName: 'Chrome',
  ...overrides
})

describe('playback store', () => {
  let store

  beforeEach(() => {
    setActivePinia(createPinia())
    store = usePlaybackStore()
    vi.clearAllMocks()
  })

  describe('initial state', () => {
    it('has correct default values', () => {
      expect(store.currentTrack).toBeNull()
      expect(store.isPlaying).toBe(false)
      expect(store.isPaused).toBe(false)
      expect(store.isStopped).toBe(true)
      expect(store.initialized).toBe(false)
    })
  })

  describe('getters', () => {
    describe('hasActiveSession', () => {
      it('returns false when no current track', () => {
        expect(store.hasActiveSession).toBe(false)
      })

      it('returns true when there is a current track', () => {
        store.currentTrack = { track: 'Song' }
        expect(store.hasActiveSession).toBe(true)
      })
    })

    describe('formattedPosition', () => {
      it('returns "0:00" when no current track', () => {
        expect(store.formattedPosition).toBe('0:00')
      })

      it('returns "0:00" when viewOffset is missing', () => {
        store.currentTrack = { track: 'Song' }
        expect(store.formattedPosition).toBe('0:00')
      })

      it('formats viewOffset in milliseconds to mm:ss', () => {
        store.currentTrack = { viewOffset: 65000 }
        expect(store.formattedPosition).toBe('1:05')
      })

      it('handles zero offset', () => {
        store.currentTrack = { viewOffset: 0 }
        expect(store.formattedPosition).toBe('0:00')
      })
    })

    describe('formattedDuration', () => {
      it('returns "0:00" when no current track', () => {
        expect(store.formattedDuration).toBe('0:00')
      })

      it('returns "0:00" when duration is missing', () => {
        store.currentTrack = { track: 'Song' }
        expect(store.formattedDuration).toBe('0:00')
      })

      it('formats duration in milliseconds to mm:ss', () => {
        store.currentTrack = { duration: 240000 }
        expect(store.formattedDuration).toBe('4:00')
      })

      it('handles durations with seconds', () => {
        store.currentTrack = { duration: 193000 }
        expect(store.formattedDuration).toBe('3:13')
      })
    })

    describe('progressPercent', () => {
      it('returns 0 when no current track', () => {
        expect(store.progressPercent).toBe(0)
      })

      it('returns 0 when duration is missing', () => {
        store.currentTrack = { viewOffset: 60000 }
        expect(store.progressPercent).toBe(0)
      })

      it('calculates percentage correctly', () => {
        store.currentTrack = { viewOffset: 60000, duration: 240000 }
        expect(store.progressPercent).toBe(25)
      })

      it('clamps to 100 maximum', () => {
        store.currentTrack = { viewOffset: 300000, duration: 240000 }
        expect(store.progressPercent).toBe(100)
      })

      it('clamps to 0 minimum', () => {
        store.currentTrack = { viewOffset: -10000, duration: 240000 }
        expect(store.progressPercent).toBe(0)
      })
    })

    describe('playbackState', () => {
      it('returns "stopped" by default', () => {
        expect(store.playbackState).toBe('stopped')
      })

      it('returns "playing" when isPlaying is true', () => {
        store.isPlaying = true
        expect(store.playbackState).toBe('playing')
      })

      it('returns "paused" when isPaused is true', () => {
        store.isPaused = true
        expect(store.playbackState).toBe('paused')
      })
    })
  })

  describe('actions', () => {
    describe('initializeEventListeners', () => {
      it('registers PlaybackUpdated and PlaybackStopped events', async () => {
        await store.initializeEventListeners()

        expect(EventsOn).toHaveBeenCalledTimes(2)
        expect(EventsOn).toHaveBeenCalledWith('PlaybackUpdated', expect.any(Function))
        expect(EventsOn).toHaveBeenCalledWith('PlaybackStopped', expect.any(Function))
        expect(store.initialized).toBe(true)
      })

      it('only initializes once', async () => {
        await store.initializeEventListeners()
        await store.initializeEventListeners()

        expect(EventsOn).toHaveBeenCalledTimes(2) // not 4
      })

      it('restores current session from backend', async () => {
        const session = makeMockSession()
        GetCurrentSession.mockResolvedValue(session)

        await store.initializeEventListeners()

        expect(GetCurrentSession).toHaveBeenCalled()
        expect(store.currentTrack).not.toBeNull()
        expect(store.currentTrack.track).toBe('Test Track')
        expect(store.isPlaying).toBe(true)
      })

      it('handles null current session gracefully', async () => {
        GetCurrentSession.mockResolvedValue(null)

        await store.initializeEventListeners()

        expect(store.currentTrack).toBeNull()
      })

      it('handles GetCurrentSession failure gracefully', async () => {
        GetCurrentSession.mockRejectedValue(new Error('backend error'))
        const consoleSpy = vi.spyOn(console, 'error').mockImplementation(() => {})

        await store.initializeEventListeners()

        expect(store.initialized).toBe(true)
        expect(store.currentTrack).toBeNull()
        consoleSpy.mockRestore()
      })
    })

    describe('cleanupEventListeners', () => {
      it('removes event listeners and resets initialized', async () => {
        await store.initializeEventListeners()
        store.cleanupEventListeners()

        expect(EventsOff).toHaveBeenCalledWith('PlaybackUpdated')
        expect(EventsOff).toHaveBeenCalledWith('PlaybackStopped')
        expect(store.initialized).toBe(false)
      })

      it('does nothing if not initialized', () => {
        store.cleanupEventListeners()

        expect(EventsOff).not.toHaveBeenCalled()
      })
    })

    describe('setTrack', () => {
      it('sets current track from session data', () => {
        const session = makeMockSession()
        store.setTrack(session)

        expect(store.currentTrack).toEqual({
          sessionKey: 'abc123',
          track: 'Test Track',
          artist: 'Test Artist',
          album: 'Test Album',
          thumb: '/library/metadata/123/thumb',
          thumbUrl: 'http://plex:32400/photo/:/transcode?url=...',
          duration: 240000,
          viewOffset: 60000,
          state: 'playing',
          playerName: 'Chrome'
        })
        expect(store.isPlaying).toBe(true)
        expect(store.isPaused).toBe(false)
        expect(store.isStopped).toBe(false)
      })

      it('sets paused state correctly', () => {
        const session = makeMockSession({ state: 'paused' })
        store.setTrack(session)

        expect(store.isPlaying).toBe(false)
        expect(store.isPaused).toBe(true)
        expect(store.isStopped).toBe(false)
      })

      it('sets stopped state correctly', () => {
        const session = makeMockSession({ state: 'stopped' })
        store.setTrack(session)

        expect(store.isPlaying).toBe(false)
        expect(store.isPaused).toBe(false)
        expect(store.isStopped).toBe(true)
      })

      it('clears track when session is null', () => {
        store.setTrack(makeMockSession())
        store.setTrack(null)

        expect(store.currentTrack).toBeNull()
        expect(store.isPlaying).toBe(false)
        expect(store.isStopped).toBe(true)
      })
    })

    describe('clearTrack', () => {
      it('resets all playback state', () => {
        store.setTrack(makeMockSession())
        store.clearTrack()

        expect(store.currentTrack).toBeNull()
        expect(store.isPlaying).toBe(false)
        expect(store.isPaused).toBe(false)
        expect(store.isStopped).toBe(true)
      })
    })

    describe('event handlers', () => {
      let eventHandlers

      beforeEach(async () => {
        eventHandlers = {}
        EventsOn.mockImplementation((event, handler) => {
          eventHandlers[event] = handler
        })
        await store.initializeEventListeners()
      })

      it('PlaybackUpdated sets the track', () => {
        const session = makeMockSession({ track: 'New Song', artist: 'New Artist' })
        eventHandlers['PlaybackUpdated'](session)

        expect(store.currentTrack.track).toBe('New Song')
        expect(store.currentTrack.artist).toBe('New Artist')
        expect(store.isPlaying).toBe(true)
      })

      it('PlaybackStopped clears the track', () => {
        store.setTrack(makeMockSession())
        eventHandlers['PlaybackStopped']()

        expect(store.currentTrack).toBeNull()
        expect(store.isStopped).toBe(true)
      })
    })
  })
})

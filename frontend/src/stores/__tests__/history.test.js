import { describe, it, expect, beforeEach, vi } from 'vitest'
import { setActivePinia, createPinia } from 'pinia'

// Mock backend calls
vi.mock('../../../wailsjs/go/main/App', () => ({
  GetListeningHistory: vi.fn().mockResolvedValue([]),
  GetListeningStats: vi.fn().mockResolvedValue({
    totalTracks: 0,
    uniqueArtists: 0,
    mostPlayedArtist: ''
  }),
  ClearListeningHistory: vi.fn().mockResolvedValue(undefined)
}))

import { useHistoryStore } from '../history'
import {
  GetListeningHistory,
  GetListeningStats,
  ClearListeningHistory
} from '../../../wailsjs/go/main/App'

const makeMockEntries = () => [
  {
    track: 'Song A',
    artist: 'Artist 1',
    album: 'Album X',
    duration: 180000,
    startedAt: '2026-04-12T10:00:00Z',
    thumbUrl: 'http://plex:32400/thumb/1'
  },
  {
    track: 'Song B',
    artist: 'Artist 2',
    album: 'Album Y',
    duration: 210000,
    startedAt: '2026-04-12T09:55:00Z',
    thumbUrl: 'http://plex:32400/thumb/2'
  },
  {
    track: 'Song C',
    artist: 'Artist 1',
    album: 'Album X',
    duration: 195000,
    startedAt: '2026-04-12T09:50:00Z'
  }
]

describe('history store', () => {
  let store

  beforeEach(() => {
    setActivePinia(createPinia())
    store = useHistoryStore()
    vi.clearAllMocks()
  })

  describe('initial state', () => {
    it('has correct default values', () => {
      expect(store.entries).toEqual([])
      expect(store.stats).toEqual({
        totalTracks: 0,
        uniqueArtists: 0,
        mostPlayedArtist: ''
      })
      expect(store.loading).toBe(false)
    })
  })

  describe('getters', () => {
    describe('hasHistory', () => {
      it('returns false when entries is empty', () => {
        expect(store.hasHistory).toBe(false)
      })

      it('returns true when entries exist', () => {
        store.entries = makeMockEntries()
        expect(store.hasHistory).toBe(true)
      })
    })
  })

  describe('actions', () => {
    describe('fetchHistory', () => {
      it('fetches history entries from backend', async () => {
        const entries = makeMockEntries()
        GetListeningHistory.mockResolvedValue(entries)

        await store.fetchHistory()

        expect(GetListeningHistory).toHaveBeenCalledWith(20)
        expect(store.entries).toEqual(entries)
        expect(store.loading).toBe(false)
      })

      it('uses custom limit parameter', async () => {
        GetListeningHistory.mockResolvedValue([])

        await store.fetchHistory(50)

        expect(GetListeningHistory).toHaveBeenCalledWith(50)
      })

      it('sets loading state during fetch', async () => {
        let resolvePromise
        GetListeningHistory.mockReturnValue(
          new Promise((resolve) => {
            resolvePromise = resolve
          })
        )

        const promise = store.fetchHistory()
        expect(store.loading).toBe(true)

        resolvePromise([])
        await promise
        expect(store.loading).toBe(false)
      })

      it('handles backend errors gracefully', async () => {
        GetListeningHistory.mockRejectedValue(new Error('db error'))
        const consoleSpy = vi.spyOn(console, 'error').mockImplementation(() => {})

        await store.fetchHistory()

        expect(store.entries).toEqual([])
        expect(store.loading).toBe(false)
        consoleSpy.mockRestore()
      })

      it('handles null response from backend', async () => {
        GetListeningHistory.mockResolvedValue(null)

        await store.fetchHistory()

        expect(store.entries).toEqual([])
      })
    })

    describe('fetchStats', () => {
      it('fetches stats from backend', async () => {
        const stats = {
          totalTracks: 150,
          uniqueArtists: 42,
          mostPlayedArtist: 'Artist 1'
        }
        GetListeningStats.mockResolvedValue(stats)

        await store.fetchStats()

        expect(GetListeningStats).toHaveBeenCalled()
        expect(store.stats).toEqual(stats)
      })

      it('handles null response from backend', async () => {
        GetListeningStats.mockResolvedValue(null)

        await store.fetchStats()

        expect(store.stats).toEqual({
          totalTracks: 0,
          uniqueArtists: 0,
          mostPlayedArtist: ''
        })
      })

      it('handles backend errors gracefully', async () => {
        GetListeningStats.mockRejectedValue(new Error('db error'))
        const consoleSpy = vi.spyOn(console, 'error').mockImplementation(() => {})

        await store.fetchStats()

        // Stats should remain at their previous values (defaults in this case)
        expect(store.stats).toEqual({
          totalTracks: 0,
          uniqueArtists: 0,
          mostPlayedArtist: ''
        })
        consoleSpy.mockRestore()
      })
    })

    describe('clearHistory', () => {
      it('clears history and resets stats', async () => {
        store.entries = makeMockEntries()
        store.stats = {
          totalTracks: 150,
          uniqueArtists: 42,
          mostPlayedArtist: 'Artist 1'
        }

        await store.clearHistory()

        expect(ClearListeningHistory).toHaveBeenCalled()
        expect(store.entries).toEqual([])
        expect(store.stats).toEqual({
          totalTracks: 0,
          uniqueArtists: 0,
          mostPlayedArtist: ''
        })
      })

      it('handles backend errors gracefully', async () => {
        ClearListeningHistory.mockRejectedValue(new Error('db error'))
        const consoleSpy = vi.spyOn(console, 'error').mockImplementation(() => {})

        store.entries = makeMockEntries()
        await store.clearHistory()

        // Entries should remain unchanged on error
        expect(store.entries).toHaveLength(3)
        consoleSpy.mockRestore()
      })
    })
  })
})

import { describe, it, expect, beforeEach, vi } from 'vitest'
import { setActivePinia, createPinia } from 'pinia'

// Mock Wails runtime. EventsOn returns a cancel function (the real runtime
// does too) — the store must use those instead of EventsOff(name).
vi.mock('../../../wailsjs/runtime/runtime', () => ({
  EventsOn: vi.fn(() => vi.fn()),
  EventsOff: vi.fn()
}))

// Mock backend calls
vi.mock('../../../wailsjs/go/main/App', () => ({
  CanSelfUpdate: vi.fn().mockResolvedValue(true),
  CheckForUpdate: vi.fn().mockResolvedValue({ available: false }),
  DownloadAndInstallUpdate: vi.fn().mockResolvedValue(undefined),
  GetUpdateStatus: vi.fn().mockResolvedValue({ state: 'idle', progress: 0, info: null }),
  RestartApplication: vi.fn().mockResolvedValue(undefined)
}))

import { useUpdatesStore } from '../updates'
import { EventsOn, EventsOff } from '../../../wailsjs/runtime/runtime'
import { CanSelfUpdate, CheckForUpdate, DownloadAndInstallUpdate, GetUpdateStatus } from '../../../wailsjs/go/main/App'

const READY_INFO = {
  available: false, // matches DownloadAndApplyUpdate's return once applied
  currentVersion: 'v4.3.0',
  latestVersion: 'v9.9.9',
  releaseUrl: 'https://github.com/Wifsimster/PlexCord/releases'
}

describe('updates store', () => {
  let store

  beforeEach(() => {
    setActivePinia(createPinia())
    store = useUpdatesStore()
    vi.clearAllMocks()
  })

  describe('initial state', () => {
    it('has correct default values', () => {
      expect(store.status).toBe('idle')
      expect(store.info).toBeNull()
      expect(store.progress).toBe(0)
      expect(store.updateAvailable).toBe(false)
      expect(store.updateReady).toBe(false)
      expect(store.shouldToast).toBe(false)
      expect(store.initialized).toBe(false)
    })
  })

  describe('initialize', () => {
    it('hydrates from the backend snapshot (background download finished before mount)', async () => {
      GetUpdateStatus.mockResolvedValue({ state: 'ready', progress: 100, info: READY_INFO })

      await store.initialize()

      expect(store.canSelfUpdate).toBe(true)
      expect(store.status).toBe('ready')
      expect(store.updateReady).toBe(true)
      expect(store.progress).toBe(100)
      expect(store.info).toEqual(READY_INFO)
      expect(store.showUpdatePanel).toBe(true)
    })

    it('only initializes once', async () => {
      await store.initialize()
      await store.initialize()

      expect(EventsOn).toHaveBeenCalledTimes(4)
      expect(CanSelfUpdate).toHaveBeenCalledTimes(1)
    })

    it('registers listeners for all update events', async () => {
      await store.initialize()

      expect(EventsOn).toHaveBeenCalledWith('UpdateAvailable', expect.any(Function))
      expect(EventsOn).toHaveBeenCalledWith('UpdateDownloadProgress', expect.any(Function))
      expect(EventsOn).toHaveBeenCalledWith('UpdateReady', expect.any(Function))
      expect(EventsOn).toHaveBeenCalledWith('UpdateError', expect.any(Function))
    })
  })

  describe('cleanup', () => {
    it('calls the cancel functions instead of EventsOff (which would clobber other listeners)', () => {
      const cancels = []
      EventsOn.mockImplementation(() => {
        const cancel = vi.fn()
        cancels.push(cancel)
        return cancel
      })

      store.setupEventListeners()
      store.cleanup()

      expect(cancels).toHaveLength(4)
      cancels.forEach((cancel) => expect(cancel).toHaveBeenCalled())
      expect(EventsOff).not.toHaveBeenCalled()
      expect(store.initialized).toBe(false)
    })
  })

  describe('event handlers', () => {
    let handlers

    beforeEach(() => {
      handlers = {}
      EventsOn.mockImplementation((event, handler) => {
        handlers[event] = handler
        return vi.fn()
      })
      store.setupEventListeners()
    })

    it('UpdateAvailable stores the info and marks available', () => {
      handlers['UpdateAvailable']({ ...READY_INFO, available: true })

      expect(store.status).toBe('available')
      expect(store.updateAvailable).toBe(true)
      expect(store.info.latestVersion).toBe('v9.9.9')
    })

    it('UpdateAvailable does not regress a ready state', () => {
      store.status = 'ready'
      handlers['UpdateAvailable']({ ...READY_INFO, available: true })

      expect(store.status).toBe('ready')
    })

    it('UpdateDownloadProgress tracks percent', () => {
      handlers['UpdateDownloadProgress']({ downloaded: 50, total: 100, percent: 50 })

      expect(store.status).toBe('downloading')
      expect(store.installing).toBe(true)
      expect(store.progress).toBe(50)
    })

    it('UpdateReady flips to ready and keeps the payload info', () => {
      handlers['UpdateReady'](READY_INFO)

      expect(store.status).toBe('ready')
      expect(store.updateReady).toBe(true)
      expect(store.progress).toBe(100)
      expect(store.info).toEqual(READY_INFO)
      // The panel must stay visible even though info.available is false
      expect(store.showUpdatePanel).toBe(true)
    })

    it('UpdateError resets to available and records the message', () => {
      handlers['UpdateAvailable']({ ...READY_INFO, available: true })
      handlers['UpdateDownloadProgress']({ percent: 30 })
      handlers['UpdateError']('checksum mismatch')

      expect(store.status).toBe('available')
      expect(store.progress).toBe(0)
      expect(store.lastError).toBe('checksum mismatch')
    })
  })

  describe('shouldToast', () => {
    it('toasts when an update is ready', () => {
      store.info = READY_INFO
      store.status = 'ready'

      expect(store.shouldToast).toBe(true)
    })

    it('toasts for available-only updates when self-update is unsupported (macOS)', () => {
      store.info = { ...READY_INFO, available: true }
      store.status = 'available'
      store.canSelfUpdate = false

      expect(store.shouldToast).toBe(true)
    })

    it('does not toast for available updates that will be auto-downloaded', () => {
      store.info = { ...READY_INFO, available: true }
      store.status = 'available'
      store.canSelfUpdate = true

      expect(store.shouldToast).toBe(false)
    })

    it('dedups after dismissToast until a newer version shows up', () => {
      store.info = READY_INFO
      store.status = 'ready'
      store.dismissToast()

      expect(store.shouldToast).toBe(false)

      store.info = { ...READY_INFO, latestVersion: 'v10.0.0' }
      expect(store.shouldToast).toBe(true)
    })
  })

  describe('actions', () => {
    it('checkNow stores the result and returns it', async () => {
      CheckForUpdate.mockResolvedValue({ ...READY_INFO, available: true })

      const info = await store.checkNow()

      expect(info.available).toBe(true)
      expect(store.info).toEqual(info)
      expect(store.status).toBe('available')
    })

    it('install resets state and re-throws on failure', async () => {
      store.info = { ...READY_INFO, available: true }
      store.status = 'available'
      DownloadAndInstallUpdate.mockRejectedValue(new Error('download failed'))

      await expect(store.install()).rejects.toThrow('download failed')
      expect(store.status).toBe('available')
    })
  })
})

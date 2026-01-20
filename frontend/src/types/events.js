/**
 * Type definitions for Wails event payloads.
 * These types are used for events emitted from the Go backend.
 */

/**
 * @typedef {Object} MusicSession
 * @property {string} sessionKey - Unique session identifier
 * @property {string} userId - User ID for this session
 * @property {string} userName - User display name
 * @property {string} type - Media type (always "track" for music)
 * @property {string} state - Playback state: "playing", "paused", "stopped"
 * @property {string} playerName - Player/client name
 * @property {string} track - Track title (or "Unknown Track" if missing)
 * @property {string} artist - Artist name (or "Unknown Artist" if missing)
 * @property {string} album - Album name (or "Unknown Album" if missing)
 * @property {string} thumb - Relative album artwork path from Plex
 * @property {string} thumbUrl - Absolute album artwork URL (includes server URL and token)
 * @property {number} duration - Track duration in milliseconds (0 if missing)
 * @property {number} viewOffset - Current playback position in milliseconds (0 if missing)
 */

/**
 * Wails event names for session polling
 * @readonly
 * @enum {string}
 */
export const SessionEvents = {
  /** Emitted when music playback is detected or track changes */
  PLAYBACK_UPDATED: 'PlaybackUpdated',
  /** Emitted when music playback stops */
  PLAYBACK_STOPPED: 'PlaybackStopped'
};

// Export empty object to make this a module
export default {};

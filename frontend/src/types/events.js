/**
 * Type definitions for Wails event payloads.
 * These types are used for events emitted from the Go backend.
 */

/**
 * @typedef {Object} MediaSession
 * @property {string} sessionKey - Unique session identifier
 * @property {string} userId - User ID for this session
 * @property {string} userName - User display name
 * @property {string} type - Plex type: "track", "movie", "episode", "photo"
 * @property {string} mediaType - Simplified type: "music", "movie", "tv", "photo"
 * @property {string} state - Playback state: "playing", "paused", "stopped"
 * @property {string} playerName - Player/client name
 * @property {string} title - Track, film or episode title
 * @property {string} artist - Artist name (music only)
 * @property {string} album - Album name (music only)
 * @property {string} showTitle - Show name (TV episodes only)
 * @property {number} season - Season number (TV episodes only)
 * @property {number} episode - Episode number (TV episodes only)
 * @property {number} year - Release year (films and episodes)
 * @property {string} thumb - Relative artwork path from Plex
 * @property {string} thumbUrl - Absolute artwork URL (includes server URL and token)
 * @property {number} duration - Duration in milliseconds (0 if missing)
 * @property {number} viewOffset - Current playback position in milliseconds (0 if missing)
 */

// Export empty object to make this a module
export default {};

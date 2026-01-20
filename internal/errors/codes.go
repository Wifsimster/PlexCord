// Package errors provides structured error handling with error codes for PlexCord.
//
// Error codes are used to identify specific error conditions that can occur across
// the application. Each error includes both a code (for programmatic handling) and
// a message (for display to users or logs).
//
// Example usage:
//
//	// Create a simple error
//	err := errors.New(errors.CONFIG_READ_FAILED, "failed to read config file")
//
//	// Wrap a Go error
//	file, err := os.Open(path)
//	if err != nil {
//	    return errors.Wrap(err, errors.CONFIG_READ_FAILED, "failed to open config")
//	}
//
//	// Check error code
//	if errors.Is(err, errors.PLEX_UNREACHABLE) {
//	    // Handle unreachable server
//	}
package errors

// Plex Error Codes
const (
	// PLEX_UNREACHABLE indicates the Plex Media Server is not responding to requests.
	// This typically occurs when:
	// - The server is offline or not running
	// - Network connectivity issues prevent reaching the server
	// - The server URL is incorrect
	//
	// Recommended action: Verify server is running and URL is correct.
	PLEX_UNREACHABLE = "PLEX_UNREACHABLE"

	// PLEX_AUTH_FAILED indicates Plex authentication failed.
	// This typically occurs when:
	// - The X-Plex-Token is invalid or expired
	// - The token doesn't have permission to access the server
	// - The user account has been deactivated
	//
	// Recommended action: Re-authenticate with Plex to get a new token.
	PLEX_AUTH_FAILED = "PLEX_AUTH_FAILED"

	// PLEX_CONN_FAILED indicates Plex connection failed for an unspecified reason.
	// This is a generic fallback error when a more specific error code isn't available.
	// This typically occurs when:
	// - HTTP error status code not recognized
	// - Unexpected response format
	// - Other connection issues
	//
	// Recommended action: Check server status and network connection.
	PLEX_CONN_FAILED = "PLEX_CONN_FAILED"

	// TIMEOUT indicates a network operation timed out.
	// This typically occurs when:
	// - Server is slow to respond
	// - Network latency is too high
	// - Server is under heavy load
	//
	// Recommended action: Check network connection and server status, then retry.
	TIMEOUT = "TIMEOUT"
)

// Discord Error Codes
const (
	// DISCORD_NOT_RUNNING indicates the Discord client is not running.
	// This typically occurs when:
	// - Discord desktop app is not launched
	// - Discord process was terminated
	// - Discord is not installed on the system
	//
	// Recommended action: Launch Discord and try again.
	DISCORD_NOT_RUNNING = "DISCORD_NOT_RUNNING"

	// DISCORD_CONN_FAILED indicates connection to Discord RPC failed.
	// This typically occurs when:
	// - Discord IPC socket is unavailable
	// - Permission issues with Discord IPC
	// - Discord version doesn't support RPC
	//
	// Recommended action: Restart Discord and PlexCord.
	DISCORD_CONN_FAILED = "DISCORD_CONN_FAILED"

	// DISCORD_CLIENT_ID_INVALID indicates the Discord Application Client ID is invalid.
	// This typically occurs when:
	// - Client ID is empty or malformed
	// - Client ID doesn't exist in Discord's system
	// - Application was deleted from Discord Developer Portal
	//
	// Recommended action: Verify Client ID in Discord Developer Portal.
	DISCORD_CLIENT_ID_INVALID = "DISCORD_CLIENT_ID_INVALID"
)

// Configuration Error Codes
const (
	// CONFIG_READ_FAILED indicates the configuration file could not be read.
	// This typically occurs when:
	// - Config file is malformed (invalid JSON)
	// - File permissions prevent reading
	// - File system error occurred
	//
	// Recommended action: Check config file format and permissions.
	CONFIG_READ_FAILED = "CONFIG_READ_FAILED"

	// CONFIG_WRITE_FAILED indicates the configuration file could not be written.
	// This typically occurs when:
	// - Directory doesn't exist or can't be created
	// - File permissions prevent writing
	// - Disk is full or read-only
	//
	// Recommended action: Check directory permissions and disk space.
	CONFIG_WRITE_FAILED = "CONFIG_WRITE_FAILED"
)

// Keychain Error Codes
const (
	// KEYCHAIN_UNAVAILABLE indicates the OS keychain service is not available.
	// This typically occurs when:
	// - Running in a restricted/sandboxed environment
	// - OS keychain service is disabled
	// - Platform doesn't support keychain
	//
	// Recommended action: Fallback encryption will be used automatically.
	KEYCHAIN_UNAVAILABLE = "KEYCHAIN_UNAVAILABLE"

	// KEYCHAIN_STORE_FAILED indicates storing a credential in keychain failed.
	// This typically occurs when:
	// - Keychain access was denied by user
	// - Insufficient permissions
	// - Keychain is locked
	//
	// Recommended action: Check keychain permissions and unlock if necessary.
	KEYCHAIN_STORE_FAILED = "KEYCHAIN_STORE_FAILED"

	// KEYCHAIN_READ_FAILED indicates reading a credential from keychain failed.
	// This typically occurs when:
	// - Keychain access was denied
	// - Credential was deleted externally
	// - Keychain database is corrupted
	//
	// Recommended action: May need to re-authenticate.
	KEYCHAIN_READ_FAILED = "KEYCHAIN_READ_FAILED"

	// ENCRYPTION_FAILED indicates encryption of fallback credentials failed.
	// This typically occurs when:
	// - Cryptographic operations failed
	// - File system errors during write
	// - Insufficient disk space
	//
	// Recommended action: Check disk space and file permissions.
	ENCRYPTION_FAILED = "ENCRYPTION_FAILED"

	// DECRYPTION_FAILED indicates decryption of fallback credentials failed.
	// This typically occurs when:
	// - Encrypted data was corrupted
	// - Machine-specific key changed (hostname/username changed)
	// - File was tampered with
	//
	// Recommended action: User may need to re-enter credentials.
	DECRYPTION_FAILED = "DECRYPTION_FAILED"
)

// General Error Codes
const (
	// UNKNOWN_ERROR indicates an unexpected error occurred.
	// This is a fallback error code when a more specific code isn't available.
	//
	// Recommended action: Check logs for more details.
	UNKNOWN_ERROR = "UNKNOWN_ERROR"
)

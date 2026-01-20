package errors

// ErrorInfo contains user-friendly error information for display.
type ErrorInfo struct {
	Code        string `json:"code"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Suggestion  string `json:"suggestion"`
	Retryable   bool   `json:"retryable"`
}

// errorInfoMap maps error codes to user-friendly information.
// These messages are designed to be displayed to users (NFR25: clear, actionable).
var errorInfoMap = map[string]ErrorInfo{
	// Plex Errors
	PLEX_UNREACHABLE: {
		Code:        PLEX_UNREACHABLE,
		Title:       "Plex Server Unreachable",
		Description: "Cannot reach Plex server. The server may be offline or there may be a network issue.",
		Suggestion:  "Check if your Plex server is running and verify your network connection.",
		Retryable:   true,
	},
	PLEX_AUTH_FAILED: {
		Code:        PLEX_AUTH_FAILED,
		Title:       "Plex Authentication Failed",
		Description: "Your Plex token is invalid or has expired.",
		Suggestion:  "Please re-authenticate with Plex to get a new token.",
		Retryable:   false,
	},
	PLEX_CONN_FAILED: {
		Code:        PLEX_CONN_FAILED,
		Title:       "Plex Connection Failed",
		Description: "Failed to connect to Plex server.",
		Suggestion:  "Check your server URL and network connection, then try again.",
		Retryable:   true,
	},
	TIMEOUT: {
		Code:        TIMEOUT,
		Title:       "Connection Timeout",
		Description: "The connection to the server timed out.",
		Suggestion:  "The server may be slow or your network may be congested. Please try again.",
		Retryable:   true,
	},

	// Discord Errors
	DISCORD_NOT_RUNNING: {
		Code:        DISCORD_NOT_RUNNING,
		Title:       "Discord Not Running",
		Description: "Discord is not running on your computer.",
		Suggestion:  "Start Discord to enable Rich Presence.",
		Retryable:   true,
	},
	DISCORD_CONN_FAILED: {
		Code:        DISCORD_CONN_FAILED,
		Title:       "Discord Connection Failed",
		Description: "Cannot connect to Discord. The connection may have been interrupted.",
		Suggestion:  "Try restarting Discord and PlexCord.",
		Retryable:   true,
	},
	DISCORD_CLIENT_ID_INVALID: {
		Code:        DISCORD_CLIENT_ID_INVALID,
		Title:       "Invalid Discord Client ID",
		Description: "The Discord Application Client ID is invalid.",
		Suggestion:  "Check your Client ID in Discord settings or reset to default.",
		Retryable:   false,
	},

	// Config Errors
	CONFIG_READ_FAILED: {
		Code:        CONFIG_READ_FAILED,
		Title:       "Configuration Error",
		Description: "Failed to read application settings.",
		Suggestion:  "The settings file may be corrupted. Try resetting the application.",
		Retryable:   false,
	},
	CONFIG_WRITE_FAILED: {
		Code:        CONFIG_WRITE_FAILED,
		Title:       "Settings Save Failed",
		Description: "Failed to save application settings.",
		Suggestion:  "Check that you have write permissions to the settings folder.",
		Retryable:   true,
	},

	// Keychain Errors
	KEYCHAIN_UNAVAILABLE: {
		Code:        KEYCHAIN_UNAVAILABLE,
		Title:       "Secure Storage Unavailable",
		Description: "The system's secure storage is not available.",
		Suggestion:  "PlexCord will use encrypted file storage instead.",
		Retryable:   false,
	},
	KEYCHAIN_STORE_FAILED: {
		Code:        KEYCHAIN_STORE_FAILED,
		Title:       "Failed to Store Credentials",
		Description: "Could not save your credentials securely.",
		Suggestion:  "Check your system's keychain settings and permissions.",
		Retryable:   true,
	},
	KEYCHAIN_READ_FAILED: {
		Code:        KEYCHAIN_READ_FAILED,
		Title:       "Failed to Read Credentials",
		Description: "Could not retrieve your saved credentials.",
		Suggestion:  "You may need to re-enter your Plex token.",
		Retryable:   false,
	},
	ENCRYPTION_FAILED: {
		Code:        ENCRYPTION_FAILED,
		Title:       "Encryption Failed",
		Description: "Failed to encrypt your credentials.",
		Suggestion:  "Check available disk space and try again.",
		Retryable:   true,
	},
	DECRYPTION_FAILED: {
		Code:        DECRYPTION_FAILED,
		Title:       "Decryption Failed",
		Description: "Failed to decrypt your saved credentials.",
		Suggestion:  "You'll need to re-enter your Plex token.",
		Retryable:   false,
	},

	// General Errors
	UNKNOWN_ERROR: {
		Code:        UNKNOWN_ERROR,
		Title:       "Unexpected Error",
		Description: "An unexpected error occurred.",
		Suggestion:  "Please try again. If the problem persists, restart PlexCord.",
		Retryable:   true,
	},
}

// GetErrorInfo returns user-friendly information for an error code.
// If the code is not found, returns information for UNKNOWN_ERROR.
func GetErrorInfo(code string) ErrorInfo {
	if info, ok := errorInfoMap[code]; ok {
		return info
	}
	return errorInfoMap[UNKNOWN_ERROR]
}

// GetErrorInfoFromError extracts error info from an AppError.
// If the error is not an AppError, returns UNKNOWN_ERROR info.
func GetErrorInfoFromError(err error) ErrorInfo {
	code := GetCode(err)
	if code == "" {
		code = UNKNOWN_ERROR
	}
	return GetErrorInfo(code)
}

// IsRetryable returns whether an error with the given code can be retried.
func IsRetryable(code string) bool {
	if info, ok := errorInfoMap[code]; ok {
		return info.Retryable
	}
	return true // Default to retryable for unknown errors
}

// IsAuthError returns whether the error indicates an authentication issue.
// Used to detect when the user needs to re-authenticate.
func IsAuthError(code string) bool {
	return code == PLEX_AUTH_FAILED || code == KEYCHAIN_READ_FAILED || code == DECRYPTION_FAILED
}

// IsConnectionError returns whether the error is a connection-related issue.
func IsConnectionError(code string) bool {
	switch code {
	case PLEX_UNREACHABLE, PLEX_CONN_FAILED, TIMEOUT,
		DISCORD_NOT_RUNNING, DISCORD_CONN_FAILED:
		return true
	}
	return false
}

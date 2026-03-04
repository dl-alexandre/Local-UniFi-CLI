// Package api provides HTTP client for local UniFi Controller API
package api

import "time"

// Network and VLAN constants
const (
	MinVLAN = 1
	MaxVLAN = 4094
)

// HTTP timeout constants
const (
	DefaultTimeoutSeconds = 30
	MaxTimeoutSeconds     = 300
)

// DefaultTimeout returns the default timeout duration
func DefaultTimeout() time.Duration {
	return DefaultTimeoutSeconds * time.Second
}

// Retry and backoff constants
const (
	MaxRetries            = 3
	InitialBackoffSeconds = 2
	BackoffMultiplier     = 2
	MaxBackoffSeconds     = 8
)

// Watch/monitoring constants
const (
	DefaultWatchIntervalSeconds = 5
	MinWatchIntervalSeconds     = 1
	MaxWatchIntervalSeconds     = 300
)

// Bandwidth and statistics constants
const (
	DefaultTopN = 10
	MaxTopN     = 100
)

// MAC address constants
const (
	MACLength        = 17 // Standard MAC format: xx:xx:xx:xx:xx:xx
	MACSegmentCount  = 5  // Number of colons in MAC address
	MACSegmentLength = 2  // Length of each hex segment
)

// MongoDB ObjectID length
const MongoDBObjectIDLength = 24

// Port configuration constants
const (
	MinPortIndex = 1
	PortIDParts  = 2 // deviceID/portIndex format
)

// Time conversion constants (in seconds)
const (
	SecondsPerMinute = 60
	SecondsPerHour   = 60 * SecondsPerMinute
	SecondsPerDay    = 24 * SecondsPerHour
	MinutesPerHour   = 60
	HoursPerDay      = 24
)

// Hotspot/voucher constants
const (
	DefaultGuestAuthDurationMinutes = 1440 // 24 hours
	DefaultVoucherDurationMinutes   = 480  // 8 hours
	UnlimitedQuota                  = 0
)

// Display/formatting constants
const (
	MaxProfileNameLength     = 8
	MaxDisplayNameLength     = 18
	MaxTruncatedNameLength   = 15
	MaxAPNameLength          = 13
	MaxTruncatedAPNameLength = 10
	MaxSettingValueLength    = 50
	MaxTruncatedValueLength  = 47
	MaxTableNoteLength       = 18
	MaxTruncatedNoteLength   = 15
	DefaultLocateDuration    = 30
)

// Stats period constants
const (
	Period1Hour  = "1h"
	Period24Hour = "24h"
	Period7Day   = "7d"
	Period30Day  = "30d"
)

// Time periods in hours
const (
	Hours1  = 1
	Hours24 = 24
	Days7   = 7
	Days30  = 30
)

// Rate limiting
const (
	DefaultRetryAfterSeconds = 5
)

// File permissions
const (
	DefaultFilePerms = 0644
	DefaultDirPerms  = 0755
	ConfigFilePerms  = 0644
)

// String lengths and offsets
const (
	MinVoucherCount    = 1
	MinVoucherDuration = 1
	MinSiteCount       = 1
)

// Buffer sizes
const (
	DefaultSignalBufferSize = 1
)

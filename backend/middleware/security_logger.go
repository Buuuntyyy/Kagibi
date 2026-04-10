// Copyright (C) 2025-2026  Buuuntyyy
// SPDX-License-Identifier: AGPL-3.0-or-later

package middleware

import (
	"log"
	"os"
	"time"
)

// SecurityLogger provides structured security logging
type SecurityLogger struct {
	enabled bool
	logger  *log.Logger
	file    *os.File
}

// NewSecurityLogger creates a new security logger with file output
func NewSecurityLogger() *SecurityLogger {
	// Create logs directory if not exists
	os.MkdirAll("logs", 0755)

	// Open security log file
	file, err := os.OpenFile("logs/security.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Printf("WARNING: Failed to open security log file: %v. Logging to stdout only.", err)
		return &SecurityLogger{
			enabled: true,
			logger:  log.New(os.Stdout, "[SECURITY] ", log.LstdFlags),
			file:    nil,
		}
	}

	return &SecurityLogger{
		enabled: true,
		logger:  log.New(file, "[SECURITY] ", log.LstdFlags),
		file:    file,
	}
}

// Close closes the log file
func (sl *SecurityLogger) Close() {
	if sl.file != nil {
		sl.file.Close()
	}
}

// LogAuthAttempt logs authentication attempts
func (sl *SecurityLogger) LogAuthAttempt(userID, ip string, success bool) {
	if !sl.enabled {
		return
	}
	sl.log("AUTH_ATTEMPT", userID, ip, success, "")
}

// LogPasswordChange logs password change events
func (sl *SecurityLogger) LogPasswordChange(userID, ip string) {
	if !sl.enabled {
		return
	}
	sl.log("PASSWORD_CHANGE", userID, ip, true, "")
}

// LogUnauthorizedAccess logs unauthorized access attempts
func (sl *SecurityLogger) LogUnauthorizedAccess(userID, resource, ip string) {
	if !sl.enabled {
		return
	}
	sl.log("UNAUTHORIZED_ACCESS", userID, ip, false, "Resource: "+resource)
}

// LogSuspiciousActivity logs suspicious activity
func (sl *SecurityLogger) LogSuspiciousActivity(userID, activity, ip string) {
	if !sl.enabled {
		return
	}
	sl.log("SUSPICIOUS_ACTIVITY", userID, ip, false, "Activity: "+activity)
}

// LogFileAccess logs file access events
func (sl *SecurityLogger) LogFileAccess(userID, fileID, ip string, success bool) {
	if !sl.enabled {
		return
	}
	sl.log("FILE_ACCESS", userID, ip, success, "FileID: "+fileID)
}

// LogProfileUpdate logs profile update events
func (sl *SecurityLogger) LogProfileUpdate(userID, ip string) {
	if !sl.enabled {
		return
	}
	sl.log("PROFILE_UPDATE", userID, ip, true, "")
}

// LogRateLimitExceeded logs rate limit violations
func (sl *SecurityLogger) LogRateLimitExceeded(ip, endpoint string) {
	if !sl.enabled {
		return
	}
	sl.log("RATE_LIMIT_EXCEEDED", "", ip, false, "Endpoint: "+endpoint)
}

// log is the internal logging function
func (sl *SecurityLogger) log(eventType, userID, ip string, success bool, details string) {
	timestamp := time.Now().Format(time.RFC3339)
	sl.logger.Printf("%s - Event: %s, UserID: %s, IP: %s, Success: %t, Details: %s",
		timestamp, eventType, userID, ip, success, details)
}

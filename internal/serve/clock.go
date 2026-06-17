package serve

import "time"

// defaultTimestamp returns a backup-friendly UTC stamp like
// 20260618-110423. Wrapped in a tiny helper because the package-level
// `backupTimestamp` variable shadows it in tests for deterministic
// output.
func defaultTimestamp() string {
	return time.Now().UTC().Format("20060102-150405")
}

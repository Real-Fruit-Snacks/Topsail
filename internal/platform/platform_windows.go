//go:build windows

package platform

// UserName mirrors mainsail's behavior on Windows: numeric uid is meaningless
// for stat-style output, so emit "-" as a placeholder.
func UserName(uid string) string {
	_ = uid
	return "-"
}

// GroupName mirrors mainsail's behavior on Windows: numeric gid is meaningless
// for stat-style output, so emit "-" as a placeholder.
func GroupName(gid string) string {
	_ = gid
	return "-"
}

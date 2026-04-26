//go:build !windows

package platform

import "os/user"

// UserName resolves a numeric or named uid to a display name. Returns the
// uid string unchanged when the lookup fails (e.g. orphaned uid).
func UserName(uid string) string {
	u, err := user.LookupId(uid)
	if err != nil {
		return uid
	}
	return u.Username
}

// GroupName resolves a numeric or named gid to a display name. Returns the
// gid string unchanged when the lookup fails (e.g. orphaned gid).
func GroupName(gid string) string {
	g, err := user.LookupGroupId(gid)
	if err != nil {
		return gid
	}
	return g.Name
}

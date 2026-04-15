package main

// An IP Version
//
// The value is used to form the URL to get the device's public IP.
type IPVersion string

const (
	IPv4 IPVersion = "v4"
	IPv6 IPVersion = "v6"
)

func (version IPVersion) getRecordType() string {
	switch version {
	case IPv4:
		return "A"
	case IPv6:
		return "AAAA"
	default:
		return ""
	}
}

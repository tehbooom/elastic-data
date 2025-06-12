package common

import (
	"net"
	"net/mail"
	"net/url"
	"regexp"
	"strings"
)

var (
	IsoRegex             = regexp.MustCompile(`\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}(\.\d+)?Z?`)
	CommonRegex          = regexp.MustCompile(`\d{2}/[A-Za-z]{3}/\d{4}:\d{2}:\d{2}:\d{2}`)
	ClfRegex             = regexp.MustCompile(`\[\d{2}/[A-Za-z]{3}/\d{4}:\d{2}:\d{2}:\d{2}\]`)
	ClfWithTimezoneRegex = regexp.MustCompile(`\[\d{2}/[A-Za-z]{3}/\d{4}:\d{2}:\d{2}:\d{2}\s+[+-]\d{4}\]`)
	SyslogRegex          = regexp.MustCompile(`[A-Za-z]{3}\s+\d{1,2}\s+\d{2}:\d{2}:\d{2}`)
	UnixSecRegex         = regexp.MustCompile(`\b\d{10}\b`)
	UnixMsRegex          = regexp.MustCompile(`\b\d{13}\b`)
	SnortRegex           = regexp.MustCompile(`\d{2}/\d{2}/\d{2}-\d{2}:\d{2}:\d{2}\.\d+`)
	SnortNoYearRegex     = regexp.MustCompile(`\d{2}/\d{2}-\d{2}:\d{2}:\d{2}\.\d+`)
	HostnameRegex        = regexp.MustCompile(`^[a-zA-Z0-9]([a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(\.[a-zA-Z0-9]([a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)*$`)
	UsernameRegex        = regexp.MustCompile(`^[a-zA-Z0-9._-]+$`)
)

func IsEmail(s string) bool {
	_, err := mail.ParseAddress(s)
	regex := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)

	return err == nil && regex.MatchString(s)
}

func IsURL(s string) bool {
	u, err := url.Parse(s)
	return err == nil && u.Scheme != "" && u.Host != ""
}

func IsDomain(s string) bool {
	if strings.Contains(s, "://") || strings.Contains(s, "/") {
		return false
	}

	commonDomains := []string{"au", "ca", "cn", "co", "com", "de", "edu", "eu", "fr", "gov", "ia", "io", "jp", "mil", "net", "org", "ru", "uk", "us", "xyz"}
	validTLD := false

	for _, domain := range commonDomains {
		if !strings.HasPrefix(s, domain) {
			validTLD = true
			break
		}
	}

	if !validTLD {
		return false
	}

	regex := regexp.MustCompile(`^[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)

	return regex.MatchString(s)
}

func IsIP(s string) bool {
	return net.ParseIP(s) != nil
}

func IsHostname(s string) bool {
	if len(s) > 253 {
		return false
	}

	if !HostnameRegex.MatchString(s) {
		return false
	}

	return true
}

func IsUsername(s string) bool {
	if len(s) > 64 {
		return false
	}

	if !UsernameRegex.MatchString(s) {
		return false
	}

	if strings.HasPrefix(s, ".") || strings.HasSuffix(s, ".") {
		return false
	}

	return true
}

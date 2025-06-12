package config

import (
	"fmt"

	"github.com/tehbooom/elastic-data/internal/common"
)

type Replacements struct {
	IPs     []string `mapstructure:"ip_addresses"`
	Domains []string `mapstructure:"domains"`
	Emails  []string `mapstructure:"emails"`
	Users   []string `mapstructure:"usernames"`
	Hosts   []string `mapstructure:"hostnames"`
}

var (
	replacementsIPs       = []string{"192.168.1.100", "10.0.0.50", "172.16.0.25"}
	replacementsUsernames = []string{"john.doe", "admin", "service_account", "test_user", "root"}
	replacementsDomains   = []string{"example.com", "test.local", "company.internal"}
	replacementsHostnames = []string{"web-server-01", "db-server", "app-host", "workstation-123"}
	replacementsEmails    = []string{"user@example.com", "admin@company.com", "noreply@test.local"}
)

func (r *Replacements) validReplacements() (bool, error) {
	valid, err := r.validReplacementIPs()
	if err != nil {
		return false, err
	} else if !valid {
		return false, nil
	}

	valid, err = r.validReplacementsDomains()
	if err != nil {
		return false, err
	} else if !valid {
		return false, nil
	}

	valid, err = r.validReplacementsDomains()
	if err != nil {
		return false, err
	} else if !valid {
		return false, nil
	}

	valid, err = r.validReplacementEmails()
	if err != nil {
		return false, err
	} else if !valid {
		return false, nil
	}

	valid, err = r.validReplacementUsers()
	if err != nil {
		return false, err
	} else if !valid {
		return false, nil
	}

	valid, err = r.validReplacementHosts()
	if err != nil {
		return false, err
	} else if !valid {
		return false, nil
	}

	return true, nil
}

func (r *Replacements) validReplacementIPs() (bool, error) {
	if len(r.IPs) < 1 {
		return false, fmt.Errorf("must have at least 1 ip")
	}

	for _, value := range r.IPs {
		valid := common.IsIP(value)
		if !valid {
			return false, fmt.Errorf("ip %s, is not valid", value)
		}
	}

	return true, nil
}

func (r *Replacements) validReplacementHosts() (bool, error) {
	if len(r.Hosts) < 1 {
		return false, fmt.Errorf("must have at least 1 hostname")
	}

	for _, value := range r.Hosts {
		valid := common.IsHostname(value)
		if !valid {
			return false, fmt.Errorf("hostname %s is not valid", value)
		}
	}

	return true, nil
}

func (r *Replacements) validReplacementUsers() (bool, error) {
	if len(r.Users) < 1 {
		return false, fmt.Errorf("must have at least 1 username")
	}

	for _, value := range r.Users {
		valid := common.IsUsername(value)
		if !valid {
			return false, fmt.Errorf("username %s is not valid", value)
		}
	}

	return true, nil
}

func (r *Replacements) validReplacementEmails() (bool, error) {
	if len(r.Emails) < 1 {
		return false, fmt.Errorf("must have at least 1 email")
	}

	for _, value := range r.Emails {
		valid := common.IsEmail(value)
		if !valid {
			return false, fmt.Errorf("email %s is not valid", value)
		}
	}

	return true, nil
}

func (r *Replacements) validReplacementsDomains() (bool, error) {
	if len(r.Domains) < 1 {
		return false, fmt.Errorf("must have at least 1 domain")
	}

	for _, value := range r.Domains {
		valid := common.IsDomain(value)
		if !valid {
			return false, fmt.Errorf("email %s is not valid", value)
		}
	}

	return true, nil
}

func (r *Replacements) isEmpty() bool {
	if len(r.Domains) < 1 {
		return true
	} else if len(r.Emails) < 1 {
		return true
	} else if len(r.Users) < 1 {
		return true
	} else if len(r.IPs) < 1 {
		return true
	} else if len(r.Emails) < 1 {
		return true
	}
	return false
}

func (r *Replacements) setDefaults() {
	if len(r.Domains) < 1 {
		r.Domains = replacementsDomains
	}

	if len(r.Emails) < 1 {
		r.Emails = replacementsEmails
	}

	if len(r.Users) < 1 {
		r.Users = replacementsUsernames
	}

	if len(r.IPs) < 1 {
		r.IPs = replacementsIPs
	}

	if len(r.Hosts) < 1 {
		r.Hosts = replacementsHostnames
	}
}

package interop

import "github.com/go-ini/ini"

type AmassConfig struct {
	Domains     []string
	Blacklisted []string
}

func ParseAmassConfig(filePath string) (*AmassConfig, error) {
	cfg, err := ini.LoadSources(ini.LoadOptions{
		Insensitive:  true,
		AllowShadows: true,
	}, filePath)
	if err != nil {
		return nil, err
	}

	domains := make([]string, 0)
	if domainsSection, err := cfg.GetSection("domains"); err == nil {
		for _, domain := range domainsSection.Key("domain").ValueWithShadows() {
			domains = append(domains, domain)
		}
	}

	blacklisted := make([]string, 0)
	if blacklistedSection, err := cfg.GetSection("blacklisted"); err == nil {
		for _, subdomain := range blacklistedSection.Key("subdomain").ValueWithShadows() {
			blacklisted = append(blacklisted, subdomain)
		}
	}

	return &AmassConfig{
		Domains:     domains,
		Blacklisted: blacklisted,
	}, nil
}

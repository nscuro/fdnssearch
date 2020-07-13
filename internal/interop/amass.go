package interop

import "github.com/go-ini/ini"

func ParseAmassConfig(filePath string) ([]string, error) {
	cfg, err := ini.LoadSources(ini.LoadOptions{
		Insensitive:  true,
		AllowShadows: true,
	}, filePath)
	if err != nil {
		return nil, err
	}

	domainsSection, err := cfg.GetSection("domains")
	if err != nil {
		return nil, err
	}

	domains := make([]string, 0)
	for _, domain := range domainsSection.Key("domain").ValueWithShadows() {
		domains = append(domains, domain)
	}

	return domains, nil
}

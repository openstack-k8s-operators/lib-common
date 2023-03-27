package openstack

import (
	"fmt"

	"github.com/go-logr/logr"
	"github.com/gophercloud/gophercloud/openstack/identity/v3/domains"
)

// Domain - Holds the name and description to be used while creating or looking up the OpenStack domain.
type Domain struct {
	Name        string
	Description string
}

// CreateDomain - creates a domain with domainName and domainDescription if it does not exist
func (o *OpenStack) CreateDomain(log logr.Logger, d Domain) (string, error) {
	var domainID string
	allPages, err := domains.List(o.osclient, domains.ListOpts{Name: d.Name}).AllPages()
	if err != nil {
		return domainID, err
	}
	allDomains, err := domains.ExtractDomains(allPages)
	if err != nil {
		return domainID, err
	}
	if len(allDomains) == 1 {
		domainID = allDomains[0].ID
	} else if len(allDomains) == 0 {
		createOpts := domains.CreateOpts{
			Name:        d.Name,
			Description: d.Description,
		}
		log.Info(fmt.Sprintf("Creating domain %s", d.Name))
		domain, err := domains.Create(o.osclient, createOpts).Extract()
		if err != nil {
			return domainID, err
		}
		domainID = domain.ID
	} else {
		return domainID, fmt.Errorf(fmt.Sprintf("Multiple domains named \"%s\" found", d.Name))
	}

	return domainID, nil
}

package netactuate

import (
	"context"

	"github.com/netactuate/gona/gona"
)

var _ ClientInterface = (*gona.Client)(nil)

// ClientInterface defines all methods used from gona.Client in this provider.
// This interface enables mocking and faking for tests.
type ClientInterface interface {
	// Server operations
	CreateServer(context.Context, *gona.CreateServerRequest) (gona.ServerBuild, error)
	GetServer(_ context.Context, id int) (gona.Server, error)
	BuildServer(_ context.Context, id int, r *gona.BuildServerRequest) (gona.ServerBuild, error)
	DeleteServer(_ context.Context, id int, cancelBilling bool) error
	UnlinkServer(_ context.Context, id int) error

	// SSH Key operations
	CreateSSHKey(_ context.Context, name, key string) (gona.SSHKey, error)
	GetSSHKey(_ context.Context, id int) (gona.SSHKey, error)
	DeleteSSHKey(_ context.Context, id int) error

	// BGP Session operations
	CreateBGPSessions(_ context.Context, mbPkgID, groupID int, isIPV6 bool, redundant bool) (*gona.BGPSession, error)
	GetBGPSessions(_ context.Context, mbPkgID int) ([]*gona.BGPSession, error)

	// IP operations
	GetIPs(_ context.Context, mbPkgID int) (gona.IPs, error)

	// Metadata operations
	GetLocations(context.Context) ([]gona.Location, error)
	GetOSs(context.Context) ([]gona.OS, error)
}


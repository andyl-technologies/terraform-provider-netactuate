package netactuate

import (
	"context"
	"fmt"
	"sync"

	"github.com/netactuate/gona/gona"
)

var _ ClientInterface = (*MockClient)(nil)

// MockClient implements ClientInterface with configurable function fields.
// Use this for unit tests where you need precise control over responses.
//
// Example:
//
//	mock := &MockClient{
//	    GetServerFunc: func(ctx context.Context, id int) (gona.Server, error) {
//	        return gona.Server{ID: id, Name: "test-server", ServerStatus: "RUNNING"}, nil
//	    },
//	}
type MockClient struct {
	// Server operation functions
	CreateServerFunc func(ctx context.Context, r *gona.CreateServerRequest) (gona.ServerBuild, error)
	GetServerFunc    func(ctx context.Context, id int) (gona.Server, error)
	BuildServerFunc  func(ctx context.Context, id int, r *gona.BuildServerRequest) (gona.ServerBuild, error)
	DeleteServerFunc func(ctx context.Context, id int, cancelBilling bool) error
	UnlinkServerFunc func(ctx context.Context, id int) error

	// SSH Key operation functions
	CreateSSHKeyFunc func(ctx context.Context, name, key string) (gona.SSHKey, error)
	GetSSHKeyFunc    func(ctx context.Context, id int) (gona.SSHKey, error)
	DeleteSSHKeyFunc func(ctx context.Context, id int) error

	// BGP Session operation functions
	CreateBGPSessionsFunc func(ctx context.Context, mbPkgID int, groupID int, isIPV6 bool, redundant bool) (*gona.BGPSession, error)
	GetBGPSessionsFunc    func(ctx context.Context, mbPkgID int) ([]*gona.BGPSession, error)

	// IP operation functions
	GetIPsFunc func(ctx context.Context, mbPkgID int) (gona.IPs, error)

	// Metadata operation functions
	GetLocationsFunc func(ctx context.Context) ([]gona.Location, error)
	GetOSsFunc       func(ctx context.Context) ([]gona.OS, error)

	// Call tracking
	mu    sync.Mutex
	calls []string
}

// CreateServer implements ClientInterface
func (m *MockClient) CreateServer(ctx context.Context, r *gona.CreateServerRequest) (gona.ServerBuild, error) {
	m.trackCall("CreateServer")
	if m.CreateServerFunc != nil {
		return m.CreateServerFunc(ctx, r)
	}
	return gona.ServerBuild{}, fmt.Errorf("CreateServerFunc not configured")
}

// GetServer implements ClientInterface
func (m *MockClient) GetServer(ctx context.Context, id int) (gona.Server, error) {
	m.trackCall(fmt.Sprintf("GetServer(%d)", id))
	if m.GetServerFunc != nil {
		return m.GetServerFunc(ctx, id)
	}
	return gona.Server{}, fmt.Errorf("GetServerFunc not configured")
}

// BuildServer implements ClientInterface
func (m *MockClient) BuildServer(ctx context.Context, id int, r *gona.BuildServerRequest) (gona.ServerBuild, error) {
	m.trackCall(fmt.Sprintf("BuildServer(%d)", id))
	if m.BuildServerFunc != nil {
		return m.BuildServerFunc(ctx, id, r)
	}
	return gona.ServerBuild{}, fmt.Errorf("BuildServerFunc not configured")
}

// DeleteServer implements ClientInterface
func (m *MockClient) DeleteServer(ctx context.Context, id int, cancelBilling bool) error {
	m.trackCall(fmt.Sprintf("DeleteServer(%d, %v)", id, cancelBilling))
	if m.DeleteServerFunc != nil {
		return m.DeleteServerFunc(ctx, id, cancelBilling)
	}
	return fmt.Errorf("DeleteServerFunc not configured")
}

// UnlinkServer implements ClientInterface
func (m *MockClient) UnlinkServer(ctx context.Context, id int) error {
	m.trackCall(fmt.Sprintf("UnlinkServer(%d)", id))
	if m.UnlinkServerFunc != nil {
		return m.UnlinkServerFunc(ctx, id)
	}
	return fmt.Errorf("UnlinkServerFunc not configured")
}

// CreateSSHKey implements ClientInterface
func (m *MockClient) CreateSSHKey(ctx context.Context, name, key string) (gona.SSHKey, error) {
	m.trackCall(fmt.Sprintf("CreateSSHKey(%s)", name))
	if m.CreateSSHKeyFunc != nil {
		return m.CreateSSHKeyFunc(ctx, name, key)
	}
	return gona.SSHKey{}, fmt.Errorf("CreateSSHKeyFunc not configured")
}

// GetSSHKey implements ClientInterface
func (m *MockClient) GetSSHKey(ctx context.Context, id int) (gona.SSHKey, error) {
	m.trackCall(fmt.Sprintf("GetSSHKey(%d)", id))
	if m.GetSSHKeyFunc != nil {
		return m.GetSSHKeyFunc(ctx, id)
	}
	return gona.SSHKey{}, fmt.Errorf("GetSSHKeyFunc not configured")
}

// DeleteSSHKey implements ClientInterface
func (m *MockClient) DeleteSSHKey(ctx context.Context, id int) error {
	m.trackCall(fmt.Sprintf("DeleteSSHKey(%d)", id))
	if m.DeleteSSHKeyFunc != nil {
		return m.DeleteSSHKeyFunc(ctx, id)
	}
	return fmt.Errorf("DeleteSSHKeyFunc not configured")
}

// CreateBGPSessions implements ClientInterface
func (m *MockClient) CreateBGPSessions(ctx context.Context, mbPkgID int, groupID int, isIPV6 bool, redundant bool) (*gona.BGPSession, error) {
	m.trackCall(fmt.Sprintf("CreateBGPSessions(%d, %d, %v, %v)", mbPkgID, groupID, isIPV6, redundant))
	if m.CreateBGPSessionsFunc != nil {
		return m.CreateBGPSessionsFunc(ctx, mbPkgID, groupID, isIPV6, redundant)
	}
	return nil, fmt.Errorf("CreateBGPSessionsFunc not configured")
}

// GetBGPSessions implements ClientInterface
func (m *MockClient) GetBGPSessions(ctx context.Context, mbPkgID int) ([]*gona.BGPSession, error) {
	m.trackCall(fmt.Sprintf("GetBGPSessions(%d)", mbPkgID))
	if m.GetBGPSessionsFunc != nil {
		return m.GetBGPSessionsFunc(ctx, mbPkgID)
	}
	return nil, fmt.Errorf("GetBGPSessionsFunc not configured")
}

// GetIPs implements ClientInterface
func (m *MockClient) GetIPs(ctx context.Context, mbPkgID int) (gona.IPs, error) {
	m.trackCall(fmt.Sprintf("GetIPs(%d)", mbPkgID))
	if m.GetIPsFunc != nil {
		return m.GetIPsFunc(ctx, mbPkgID)
	}
	return gona.IPs{}, fmt.Errorf("GetIPsFunc not configured")
}

// GetLocations implements ClientInterface
func (m *MockClient) GetLocations(ctx context.Context) ([]gona.Location, error) {
	m.trackCall("GetLocations")
	if m.GetLocationsFunc != nil {
		return m.GetLocationsFunc(ctx)
	}
	return nil, fmt.Errorf("GetLocationsFunc not configured")
}

// GetOSs implements ClientInterface
func (m *MockClient) GetOSs(ctx context.Context) ([]gona.OS, error) {
	m.trackCall("GetOSs")
	if m.GetOSsFunc != nil {
		return m.GetOSsFunc(ctx)
	}
	return nil, fmt.Errorf("GetOSsFunc not configured")
}

// trackCall records method calls for verification
func (m *MockClient) trackCall(call string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.calls = append(m.calls, call)
}

// GetCalls returns all tracked method calls
func (m *MockClient) GetCalls() []string {
	m.mu.Lock()
	defer m.mu.Unlock()
	calls := make([]string, len(m.calls))
	copy(calls, m.calls)
	return calls
}

// ResetCalls clears the call tracking
func (m *MockClient) ResetCalls() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.calls = nil
}

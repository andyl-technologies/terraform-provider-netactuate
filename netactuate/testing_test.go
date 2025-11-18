package netactuate_test

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/netactuate/gona/gona"
	. "github.com/netactuate/terraform-provider-netactuate/netactuate"
)

// TestMockClient tests the MockClient functionality
func TestMockClient(t *testing.T) {
	t.Run("GetServer with configured function", func(t *testing.T) {
		mock := &MockClient{
			GetServerFunc: func(ctx context.Context, id int) (gona.Server, error) {
				return gona.Server{
					ID:           id,
					Name:         "test-server.example.com",
					ServerStatus: "RUNNING",
				}, nil
			},
		}

		server, err := mock.GetServer(context.Background(), 123)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if server.ID != 123 {
			t.Errorf("expected ID 123, got %d", server.ID)
		}
		if server.Name != "test-server.example.com" {
			t.Errorf("expected name test-server.example.com, got %s", server.Name)
		}
		if server.ServerStatus != "RUNNING" {
			t.Errorf("expected status RUNNING, got %s", server.ServerStatus)
		}
	})

	t.Run("GetServer without configured function returns error", func(t *testing.T) {
		mock := &MockClient{}

		_, err := mock.GetServer(context.Background(), 123)
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		if !strings.Contains(err.Error(), "GetServerFunc not configured") {
			t.Errorf("unexpected error message: %v", err)
		}
	})

	t.Run("call tracking", func(t *testing.T) {
		mock := &MockClient{
			GetServerFunc: func(ctx context.Context, id int) (gona.Server, error) {
				return gona.Server{ID: id}, nil
			},
			DeleteServerFunc: func(ctx context.Context, id int, cancelBilling bool) error {
				return nil
			},
		}

		mock.GetServer(context.Background(), 123)
		mock.GetServer(context.Background(), 456)
		mock.DeleteServer(context.Background(), 123, true)

		calls := mock.GetCalls()
		if len(calls) != 3 {
			t.Fatalf("expected 3 calls, got %d: %v", len(calls), calls)
		}

		if calls[0] != "GetServer(123)" {
			t.Errorf("unexpected call[0]: %s", calls[0])
		}
		if calls[1] != "GetServer(456)" {
			t.Errorf("unexpected call[1]: %s", calls[1])
		}
		if calls[2] != "DeleteServer(123, true)" {
			t.Errorf("unexpected call[2]: %s", calls[2])
		}

		// Test reset
		mock.ResetCalls()
		if len(mock.GetCalls()) != 0 {
			t.Errorf("expected 0 calls after reset, got %d", len(mock.GetCalls()))
		}
	})

	t.Run("CreateServer with configured function", func(t *testing.T) {
		mock := &MockClient{
			CreateServerFunc: func(ctx context.Context, r *gona.CreateServerRequest) (gona.ServerBuild, error) {
				return gona.ServerBuild{
					ServerID: 999,
					Status:   "building",
					Build:    1,
				}, nil
			},
		}

		build, err := mock.CreateServer(context.Background(), &gona.CreateServerRequest{
			FQDN: "new-server.example.com",
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if build.ServerID != 999 {
			t.Errorf("expected ServerID 999, got %d", build.ServerID)
		}
		if build.Status != "building" {
			t.Errorf("expected status building, got %s", build.Status)
		}
	})

	t.Run("CreateSSHKey with error", func(t *testing.T) {
		expectedErr := errors.New("duplicate key name")
		mock := &MockClient{
			CreateSSHKeyFunc: func(ctx context.Context, name, key string) (gona.SSHKey, error) {
				return gona.SSHKey{}, expectedErr
			},
		}

		_, err := mock.CreateSSHKey(context.Background(), "test-key", "ssh-rsa AAA...")
		if err != expectedErr {
			t.Errorf("expected error %v, got %v", expectedErr, err)
		}
	})
}

// TestFakeClient tests the FakeClient stateful behavior
func TestFakeClient(t *testing.T) {
	t.Run("CreateServer and GetServer", func(t *testing.T) {
		fake := NewFakeClient()

		// Create a server
		build, err := fake.CreateServer(context.Background(), &gona.CreateServerRequest{
			Plan:                     "test-plan",
			Location:                 1, // AMS Amsterdam
			Image:                    1, // Ubuntu 22.04
			FQDN:                     "test.example.com",
			PackageBilling:           "usage",
			PackageBillingContractId: "contract-123",
		})
		if err != nil {
			t.Fatalf("CreateServer failed: %v", err)
		}

		if build.ServerID == 0 {
			t.Error("expected non-zero ServerID")
		}

		// Retrieve the server
		server, err := fake.GetServer(context.Background(), build.ServerID)
		if err != nil {
			t.Fatalf("GetServer failed: %v", err)
		}

		if server.ID != build.ServerID {
			t.Errorf("expected ID %d, got %d", build.ServerID, server.ID)
		}
		if server.Name != "test.example.com" {
			t.Errorf("expected name test.example.com, got %s", server.Name)
		}
		if server.ServerStatus != "RUNNING" {
			t.Errorf("expected status RUNNING, got %s", server.ServerStatus)
		}
		if server.Location != "AMS Amsterdam" {
			t.Errorf("expected location AMS Amsterdam, got %s", server.Location)
		}
		if server.OS != "Ubuntu 22.04 LTS" {
			t.Errorf("expected OS Ubuntu 22.04 LTS, got %s", server.OS)
		}
	})

	t.Run("GetServer not found", func(t *testing.T) {
		fake := NewFakeClient()

		_, err := fake.GetServer(context.Background(), 99999)
		if err == nil {
			t.Fatal("expected error for non-existent server")
		}
		if !strings.Contains(err.Error(), "not found") {
			t.Errorf("unexpected error message: %v", err)
		}
	})

	t.Run("DeleteServer marks as terminated", func(t *testing.T) {
		fake := NewFakeClient()

		// Create a server
		build, _ := fake.CreateServer(context.Background(), &gona.CreateServerRequest{
			Location: 1,
			Image:    1,
			FQDN:     "test.example.com",
		})

		// Delete the server
		err := fake.DeleteServer(context.Background(), build.ServerID, true)
		if err != nil {
			t.Fatalf("DeleteServer failed: %v", err)
		}

		// Verify it's marked as terminated
		server, _ := fake.GetServer(context.Background(), build.ServerID)
		if server.ServerStatus != "TERMINATED" {
			t.Errorf("expected status TERMINATED, got %s", server.ServerStatus)
		}
	})

	t.Run("UnlinkServer removes server", func(t *testing.T) {
		fake := NewFakeClient()

		// Create a server
		build, _ := fake.CreateServer(context.Background(), &gona.CreateServerRequest{
			Location: 1,
			Image:    1,
			FQDN:     "test.example.com",
		})

		// Unlink the server
		err := fake.UnlinkServer(context.Background(), build.ServerID)
		if err != nil {
			t.Fatalf("UnlinkServer failed: %v", err)
		}

		// Verify it's gone
		_, err = fake.GetServer(context.Background(), build.ServerID)
		if err == nil {
			t.Error("expected error after unlinking server")
		}
	})

	t.Run("CreateSSHKey and GetSSHKey", func(t *testing.T) {
		fake := NewFakeClient()

		key, err := fake.CreateSSHKey(context.Background(), "test-key", "ssh-rsa AAA...")
		if err != nil {
			t.Fatalf("CreateSSHKey failed: %v", err)
		}

		if key.ID == 0 {
			t.Error("expected non-zero key ID")
		}
		if key.Name != "test-key" {
			t.Errorf("expected name test-key, got %s", key.Name)
		}

		// Retrieve the key
		retrieved, err := fake.GetSSHKey(context.Background(), key.ID)
		if err != nil {
			t.Fatalf("GetSSHKey failed: %v", err)
		}

		if retrieved.ID != key.ID {
			t.Errorf("expected ID %d, got %d", key.ID, retrieved.ID)
		}
		if retrieved.Name != key.Name {
			t.Errorf("expected name %s, got %s", key.Name, retrieved.Name)
		}
	})

	t.Run("DeleteSSHKey removes key", func(t *testing.T) {
		fake := NewFakeClient()

		key, _ := fake.CreateSSHKey(context.Background(), "test-key", "ssh-rsa AAA...")

		err := fake.DeleteSSHKey(context.Background(), key.ID)
		if err != nil {
			t.Fatalf("DeleteSSHKey failed: %v", err)
		}

		_, err = fake.GetSSHKey(context.Background(), key.ID)
		if err == nil {
			t.Error("expected error after deleting key")
		}
	})

	t.Run("CreateBGPSessions and GetBGPSessions", func(t *testing.T) {
		fake := NewFakeClient()

		// Create server first
		build, _ := fake.CreateServer(context.Background(), &gona.CreateServerRequest{
			Location: 1,
			Image:    1,
			FQDN:     "test.example.com",
		})

		// Create BGP session
		session, err := fake.CreateBGPSessions(context.Background(), build.ServerID, 1, false, false)
		if err != nil {
			t.Fatalf("CreateBGPSessions failed: %v", err)
		}

		if session.ID == 0 {
			t.Error("expected non-zero session ID")
		}
		if session.ProviderIPType != "IPv4" {
			t.Errorf("expected IPv4, got %s", session.ProviderIPType)
		}

		// Retrieve sessions
		sessions, err := fake.GetBGPSessions(context.Background(), build.ServerID)
		if err != nil {
			t.Fatalf("GetBGPSessions failed: %v", err)
		}

		if len(sessions) != 1 {
			t.Fatalf("expected 1 session, got %d", len(sessions))
		}
		if sessions[0].ID != session.ID {
			t.Errorf("expected session ID %d, got %d", session.ID, sessions[0].ID)
		}
	})

	t.Run("CreateBGPSessions IPv6", func(t *testing.T) {
		fake := NewFakeClient()

		build, _ := fake.CreateServer(context.Background(), &gona.CreateServerRequest{
			Location: 1,
			Image:    1,
			FQDN:     "test.example.com",
		})

		session, err := fake.CreateBGPSessions(context.Background(), build.ServerID, 1, true, false)
		if err != nil {
			t.Fatalf("CreateBGPSessions failed: %v", err)
		}

		if session.ProviderIPType != "IPv6" {
			t.Errorf("expected IPv6, got %s", session.ProviderIPType)
		}
		if !strings.Contains(session.CustomerIP, ":") {
			t.Errorf("expected IPv6 address, got %s", session.CustomerIP)
		}
	})

	t.Run("GetIPs returns default IPs for server", func(t *testing.T) {
		fake := NewFakeClient()

		build, _ := fake.CreateServer(context.Background(), &gona.CreateServerRequest{
			Location: 1,
			Image:    1,
			FQDN:     "test.example.com",
		})

		ips, err := fake.GetIPs(context.Background(), build.ServerID)
		if err != nil {
			t.Fatalf("GetIPs failed: %v", err)
		}

		if len(ips.IPv4) == 0 {
			t.Error("expected at least one IPv4 address")
		}
		if len(ips.IPv6) == 0 {
			t.Error("expected at least one IPv6 address")
		}
	})

	t.Run("GetLocations returns default locations", func(t *testing.T) {
		fake := NewFakeClient()

		locations, err := fake.GetLocations(context.Background())
		if err != nil {
			t.Fatalf("GetLocations failed: %v", err)
		}

		if len(locations) == 0 {
			t.Error("expected default locations")
		}

		// Check for AMS
		found := false
		for _, loc := range locations {
			if loc.Name == "AMS Amsterdam" {
				found = true
				break
			}
		}
		if !found {
			t.Error("expected to find AMS Amsterdam in locations")
		}
	})

	t.Run("GetOSs returns default OSes", func(t *testing.T) {
		fake := NewFakeClient()

		oses, err := fake.GetOSs(context.Background())
		if err != nil {
			t.Fatalf("GetOSs failed: %v", err)
		}

		if len(oses) == 0 {
			t.Error("expected default OSes")
		}

		// Check for Ubuntu
		found := false
		for _, os := range oses {
			if strings.Contains(os.Os, "Ubuntu") {
				found = true
				break
			}
		}
		if !found {
			t.Error("expected to find Ubuntu in OSes")
		}
	})

	t.Run("error injection", func(t *testing.T) {
		fake := NewFakeClient()
		expectedErr := errors.New("test error")

		fake.GetServerError = expectedErr

		_, err := fake.GetServer(context.Background(), 123)
		if err != expectedErr {
			t.Errorf("expected error %v, got %v", expectedErr, err)
		}

		// Reset error
		fake.GetServerError = nil
		fake.AddServer(gona.Server{ID: 123})
		_, err = fake.GetServer(context.Background(), 123)
		if err != nil {
			t.Errorf("unexpected error after resetting: %v", err)
		}
	})

	t.Run("AddServer helper", func(t *testing.T) {
		fake := NewFakeClient()

		testServer := gona.Server{
			ID:           999,
			Name:         "custom.example.com",
			ServerStatus: "RUNNING",
		}

		fake.AddServer(testServer)

		server, err := fake.GetServer(context.Background(), 999)
		if err != nil {
			t.Fatalf("GetServer failed: %v", err)
		}

		if server.Name != "custom.example.com" {
			t.Errorf("expected name custom.example.com, got %s", server.Name)
		}
	})

	t.Run("call tracking", func(t *testing.T) {
		fake := NewFakeClient()

		fake.AddServer(gona.Server{ID: 123})
		fake.GetServer(context.Background(), 123)
		fake.GetLocations(context.Background())

		calls := fake.GetCalls()
		if len(calls) != 2 {
			t.Fatalf("expected 2 calls, got %d: %v", len(calls), calls)
		}
	})
}

// TestBuilders tests the test data builders
func TestBuilders(t *testing.T) {
	t.Run("ServerBuilder", func(t *testing.T) {
		server := NewServerBuilder().
			WithID(123).
			WithName("test.example.com").
			WithStatus("TERMINATED").
			WithIPv4("203.0.113.10").
			Build()

		if server.ID != 123 {
			t.Errorf("expected ID 123, got %d", server.ID)
		}
		if server.Name != "test.example.com" {
			t.Errorf("expected name test.example.com, got %s", server.Name)
		}
		if server.ServerStatus != "TERMINATED" {
			t.Errorf("expected status TERMINATED, got %s", server.ServerStatus)
		}
		if server.PrimaryIPv4 != "203.0.113.10" {
			t.Errorf("expected IPv4 203.0.113.10, got %s", server.PrimaryIPv4)
		}
	})

	t.Run("SSHKeyBuilder", func(t *testing.T) {
		key := NewSSHKeyBuilder().
			WithID(456).
			WithName("my-key").
			WithKey("ssh-ed25519 AAAA...").
			Build()

		if key.ID != 456 {
			t.Errorf("expected ID 456, got %d", key.ID)
		}
		if key.Name != "my-key" {
			t.Errorf("expected name my-key, got %s", key.Name)
		}
		if !strings.HasPrefix(key.Key, "ssh-ed25519") {
			t.Errorf("expected key to start with ssh-ed25519, got %s", key.Key)
		}
	})

	t.Run("BGPSessionBuilder", func(t *testing.T) {
		session := NewBGPSessionBuilder().
			WithID(789).
			WithGroupID(5).
			WithIPv6().
			Build()

		if session.ID != 789 {
			t.Errorf("expected ID 789, got %d", session.ID)
		}
		if session.GroupID != 5 {
			t.Errorf("expected GroupID 5, got %d", session.GroupID)
		}
		if session.ProviderIPType != "IPv6" {
			t.Errorf("expected IPv6, got %s", session.ProviderIPType)
		}
	})

	t.Run("IPsBuilder", func(t *testing.T) {
		ips := NewIPsBuilder().
			ClearIPv4().
			ClearIPv6().
			AddIPv4("203.0.113.20", "203.0.113.1", "255.255.255.0", true).
			AddIPv4("203.0.113.21", "203.0.113.1", "255.255.255.0", false).
			Build()

		if len(ips.IPv4) != 2 {
			t.Errorf("expected 2 IPv4 addresses, got %d", len(ips.IPv4))
		}
		if len(ips.IPv6) != 0 {
			t.Errorf("expected 0 IPv6 addresses, got %d", len(ips.IPv6))
		}

		if ips.IPv4[0].IP != "203.0.113.20" {
			t.Errorf("expected first IP 203.0.113.20, got %s", ips.IPv4[0].IP)
		}
		if ips.IPv4[0].Primary != 1 {
			t.Errorf("expected first IP to be primary")
		}
		if ips.IPv4[1].Primary != 0 {
			t.Errorf("expected second IP to not be primary")
		}
	})

	t.Run("LocationBuilder", func(t *testing.T) {
		location := NewLocationBuilder().
			WithID(10).
			WithName("LON London").
			WithIATACode("LON").
			WithContinent("EU").
			Build()

		if location.ID != 10 {
			t.Errorf("expected ID 10, got %d", location.ID)
		}
		if location.Name != "LON London" {
			t.Errorf("expected name LON London, got %s", location.Name)
		}
		if location.IATACode != "LON" {
			t.Errorf("expected IATA LON, got %s", location.IATACode)
		}
	})

	t.Run("OSBuilder", func(t *testing.T) {
		os := NewOSBuilder().
			WithID(20).
			WithName("Debian 12").
			WithType("linux").
			WithSubtype("debian").
			Build()

		if os.ID != 20 {
			t.Errorf("expected ID 20, got %d", os.ID)
		}
		if os.Os != "Debian 12" {
			t.Errorf("expected OS Debian 12, got %s", os.Os)
		}
		if os.Type != "linux" {
			t.Errorf("expected type linux, got %s", os.Type)
		}
	})

	t.Run("helper functions", func(t *testing.T) {
		server := TestRunningServer(100)
		if server.ID != 100 {
			t.Errorf("expected ID 100, got %d", server.ID)
		}
		if server.ServerStatus != "RUNNING" {
			t.Errorf("expected RUNNING status, got %s", server.ServerStatus)
		}

		key := TestSSHKey(200, "test-key")
		if key.ID != 200 {
			t.Errorf("expected ID 200, got %d", key.ID)
		}
		if key.Name != "test-key" {
			t.Errorf("expected name test-key, got %s", key.Name)
		}

		session := TestBGPSession(300)
		if session.ID != 300 {
			t.Errorf("expected ID 300, got %d", session.ID)
		}

		sessionV6 := TestBGPSessionIPv6(400)
		if sessionV6.ProviderIPType != "IPv6" {
			t.Errorf("expected IPv6, got %s", sessionV6.ProviderIPType)
		}
	})
}

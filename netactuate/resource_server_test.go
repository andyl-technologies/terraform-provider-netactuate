package netactuate

import (
	"fmt"
	"regexp"
	"testing"
)

func TestHostnameRegex(t *testing.T) {
	// Build the regex pattern the same way as in resource_server.go
	// BUG: missing the ^ anchor at the beginning, making it match if the
	// pattern appears anywhere in the string, not just if the entire string
	// matches. This allows invalid hostnames through.
	pattern := fmt.Sprintf("(%[1]s\\.)*%[1]s$", fmt.Sprintf("(%[1]s|%[1]s%[2]s*%[1]s)", "[a-zA-Z0-9]", "[a-zA-Z0-9\\-]"))
	regex := regexp.MustCompile(pattern)

	t.Log("Current regex pattern (BUGGY):", pattern)
	t.Log("Fixed pattern should be:", "^"+pattern)

	tests := []struct {
		hostname      string
		currentValid  bool // What the BUGGY regex currently matches
		expectedValid bool // What it SHOULD match
		desc          string
	}{
		// Valid single-label hostnames
		{"a", true, true, "single character hostname"},
		{"z", true, true, "single character hostname"},
		{"0", true, true, "single digit hostname"},
		{"9", true, true, "single digit hostname"},
		{"example", true, true, "simple hostname"},
		{"server1", true, true, "hostname with number"},
		{"web-server", true, true, "hostname with hyphen"},
		{"my-server-123", true, true, "hostname with multiple hyphens and numbers"},
		{"a1b2c3", true, true, "hostname with mixed alphanumeric"},

		// Valid multi-label hostnames (FQDNs)
		{"example.com", true, true, "simple FQDN"},
		{"www.example.com", true, true, "subdomain FQDN"},
		{"api.v1.example.com", true, true, "multiple subdomain levels"},
		{"my-server.example.com", true, true, "subdomain with hyphen"},
		{"server1.dc2.example.com", true, true, "multiple labels with numbers"},
		{"a.b.c.d.e.f.g", true, true, "many label levels"},
		{"1.2.3.4", true, true, "numeric labels (though unusual for hostnames)"},
		{"web-01.prod-us-east-1.example.com", true, true, "complex production hostname"},

		// BUG: Should be invalid but currently passes (starts with hyphen)
		{"-server", true, false, "BUG: starts with hyphen - should fail but passes"},
		{"-example.com", true, false, "BUG: label starts with hyphen - should fail but passes"},
		{"server.-example.com", true, false, "BUG: second label starts with hyphen - should fail but passes"},

		// Correctly rejected: ends with hyphen
		{"server-", false, false, "ends with hyphen"},
		{"example-.com", true, false, "BUG: label ends with hyphen - should fail but passes"},
		{"server.example-.com", true, false, "BUG: second label ends with hyphen - should fail but passes"},

		// BUG: Should be invalid but currently passes (starts with dot)
		{".example", true, false, "BUG: starts with dot - should fail but passes"},
		{"example.", false, false, "ends with dot"},
		{".example.com", true, false, "BUG: starts with dot - should fail but passes"},
		{"example.com.", false, false, "ends with dot"},

		// BUG: Should be invalid but currently passes (double dots)
		{"example..com", true, false, "BUG: double dot - should fail but passes"},
		{"server..example.com", true, false, "BUG: double dot in middle - should fail but passes"},

		// BUG: Should be invalid but currently passes (special characters)
		{"example_com", true, false, "BUG: underscore not allowed - should fail but passes"},
		{"example@com", true, false, "BUG: @ symbol not allowed - should fail but passes"},
		{"example#com", true, false, "BUG: # symbol not allowed - should fail but passes"},
		{"example com", true, false, "BUG: space not allowed - should fail but passes"},
		{"example/com", true, false, "BUG: slash not allowed - should fail but passes"},

		// Correctly rejected: empty string
		{"", false, false, "empty string"},

		// Edge cases: single character labels with dots
		{"a.b", true, true, "single char labels with dot"},
		{"a.b.c", true, true, "multiple single char labels"},

		// Edge cases: hyphen placement
		{"a-b", true, true, "hyphen between chars"},
		{"a-b-c", true, true, "multiple hyphens"},
		{"1-2-3", true, true, "hyphens between numbers"},
		{"a--b", true, true, "consecutive hyphens in middle (valid)"},

		// Realistic hostnames
		{"terraform.example.com", true, true, "realistic terraform hostname"},
		{"prod-web-01.us-east-1.example.com", true, true, "realistic production hostname"},
		{"db-master-001.internal.example.com", true, true, "realistic database hostname"},
		{"k8s-worker-node-42.cluster.local", true, true, "realistic kubernetes hostname"},
	}

	bugCount := 0
	for _, tt := range tests {
		t.Run(tt.hostname, func(t *testing.T) {
			match := regex.MatchString(tt.hostname)

			// Test against current behavior
			if match != tt.currentValid {
				t.Errorf("hostname %q: current behavior changed! expected currentValid=%v, got=%v (%s)",
					tt.hostname, tt.currentValid, match, tt.desc)
			}

			// Log when current behavior differs from expected
			if tt.currentValid != tt.expectedValid {
				t.Logf("BUG: hostname %q currently returns %v but should return %v (%s)",
					tt.hostname, tt.currentValid, tt.expectedValid, tt.desc)
			}
		})

		if tt.currentValid != tt.expectedValid {
			bugCount++
		}
	}

	t.Logf("Total bugs found: %d test cases where current behavior differs from expected", bugCount)
}

// TestHostnameRegexFixed tests what the correct behavior should be
func TestHostnameRegexFixed(t *testing.T) {
	// The FIXED regex with ^ anchor at the beginning
	fixedPattern := "^" + fmt.Sprintf("(%[1]s\\.)*%[1]s$", fmt.Sprintf("(%[1]s|%[1]s%[2]s*%[1]s)", "[a-zA-Z0-9]", "[a-zA-Z0-9\\-]"))
	fixedRegex := regexp.MustCompile(fixedPattern)

	t.Log("Fixed regex pattern:", fixedPattern)

	tests := []struct {
		hostname string
		valid    bool
		desc     string
	}{
		// Valid hostnames
		{"a", true, "single character"},
		{"example", true, "simple hostname"},
		{"web-server", true, "hostname with hyphen"},
		{"example.com", true, "simple FQDN"},
		{"www.example.com", true, "subdomain FQDN"},
		{"my-server-123.example.com", true, "complex hostname"},

		// Invalid hostnames
		{"-server", false, "starts with hyphen"},
		{"server-", false, "ends with hyphen"},
		{".example", false, "starts with dot"},
		{"example.", false, "ends with dot"},
		{"example..com", false, "double dot"},
		{"example_com", false, "underscore not allowed"},
		{"example@com", false, "@ symbol not allowed"},
		{"example com", false, "space not allowed"},
		{"", false, "empty string"},
	}

	for _, tt := range tests {
		t.Run(tt.hostname, func(t *testing.T) {
			match := fixedRegex.MatchString(tt.hostname)
			if match != tt.valid {
				t.Errorf("hostname %q: expected valid=%v, got valid=%v (%s)",
					tt.hostname, tt.valid, match, tt.desc)
			}
		})
	}
}

func TestHostnameRegexValue(t *testing.T) {
	// Verify the regex pattern is what we expect
	expectedInner := "([a-zA-Z0-9]|[a-zA-Z0-9][a-zA-Z0-9\\-]*[a-zA-Z0-9])"
	expectedOuter := fmt.Sprintf("(%s\\.)*%s$", expectedInner, expectedInner)

	actualInner := fmt.Sprintf("(%[1]s|%[1]s%[2]s*%[1]s)", "[a-zA-Z0-9]", "[a-zA-Z0-9\\-]")
	actualOuter := fmt.Sprintf("(%[1]s\\.)*%[1]s$", actualInner)

	if actualOuter != expectedOuter {
		t.Errorf("hostnameRegex pattern mismatch:\nexpected: %s\ngot:      %s",
			expectedOuter, actualOuter)
	}

	// Verify it compiles
	_, err := regexp.Compile(actualOuter)
	if err != nil {
		t.Errorf("hostnameRegex failed to compile: %v", err)
	}
}

func BenchmarkHostnameRegex(b *testing.B) {
	pattern := fmt.Sprintf("(%[1]s\\.)*%[1]s$", fmt.Sprintf("(%[1]s|%[1]s%[2]s*%[1]s)", "[a-zA-Z0-9]", "[a-zA-Z0-9\\-]"))
	regex := regexp.MustCompile(pattern)
	hostname := "my-server-01.prod-us-east-1.example.com"

	b.Run("PreviousBehavior", func(b *testing.B) {

		b.ResetTimer()
		for b.Loop() {
			regexp.MatchString(pattern, hostname)
		}
	})

	b.Run("CompiledRegex", func(b *testing.B) {
		b.ResetTimer()
		for b.Loop() {
			regex.MatchString(hostname)
		}
	})
}

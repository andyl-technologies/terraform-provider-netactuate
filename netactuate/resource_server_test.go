package netactuate

import (
	"fmt"
	"regexp"
	"testing"
)

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

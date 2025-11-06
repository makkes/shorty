package version_test

import (
	"testing"

	. "github.com/onsi/gomega"

	"github.com/makkes/shorty/version"
)

func TestGetVersionReturnsExpectedValues(t *testing.T) {
	t.Parallel()
	g := NewWithT(t)

	expected := version.Info{
		Version:   "(devel)",
		GoVersion: "^go1..*$",
	}

	info := version.Get()
	g.Expect(info.Version).To(Equal(expected.Version))
	g.Expect(info.GoVersion).To(MatchRegexp(expected.GoVersion))
}

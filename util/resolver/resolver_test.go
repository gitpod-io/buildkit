package resolver

import (
	"bytes"
	"path"
	"testing"

	"github.com/moby/buildkit/cmd/buildkitd/config"
	"github.com/stretchr/testify/require"
)

func TestNewMirrorRegistryHost(t *testing.T) {
	const testConfig = `
[registry."docker.io"]
mirrors = ["hub.docker.io", "yourmirror.local:5000/proxy.docker.io"]
[registry."quay.io"]
mirrors = ["yourmirror.local:5000/proxy.quay.io"]
[registry."fake.io"]
mirrors = ["https://url/", "https://url/path/"]
`

	tests := map[string]struct {
		description string
		host        string
		path        string
	}{
		"hub.docker.io": {
			description: "docker_io_mirror_without_path",
			host:        "hub.docker.io",
			path:        defaultPath,
		},
		"yourmirror.local:5000/proxy.docker.io": {
			description: "docker_io_mirror_with_path",
			host:        "yourmirror.local:5000",
			path:        path.Join(defaultPath, "proxy.docker.io"),
		},
		"yourmirror.local:5000/proxy.quay.io": {
			description: "docker_quay_mirror_with_path",
			host:        "yourmirror.local:5000",
			path:        path.Join(defaultPath, "proxy.quay.io"),
		},
		"https://url/": {
			description: "docker_fake_mirror_scheme_without_path",
			host:        "url",
			path:        defaultPath,
		},
		"https://url/path/": {
			description: "docker_fake_mirror_scheme_with_path",
			host:        "url",
			path:        path.Join(defaultPath, "path"),
		},
	}

	cfg, err := config.Load(bytes.NewBuffer([]byte(testConfig)))
	require.NoError(t, err)

	require.NotEqual(t, 0, len(cfg.Registries))
	for _, registry := range cfg.Registries {
		require.NotEqual(t, 0, len(registry.Mirrors))
		for _, m := range registry.Mirrors {
			test := tests[m]
			h := newMirrorRegistryHost(m)
			require.NotNil(t, h)
			require.Equal(t, h.Host, test.host)
			require.Equal(t, h.Path, test.path)
		}
	}
}

func TestGitpodRegistryMirror(t *testing.T) {
	cfg, err := config.Load(bytes.NewBufferString(`
[registry."example.com"]
mirrors = ["other-mirror.example.com"]
`))
	require.NoError(t, err)
	hosts := NewRegistryConfig(cfg.Registries)

	for _, name := range []string{"example.com", "unconfigured.example.com", "docker.io"} {
		t.Run(name, func(t *testing.T) {
			registries, err := hosts(name)
			require.NoError(t, err)
			require.Len(t, registries, 2)
			require.Equal(t, "localhost:8080", registries[0].Host)
			require.Equal(t, "http", registries[0].Scheme)
			require.Equal(t, defaultPath, registries[0].Path)
			origin := name
			if name == "docker.io" {
				origin = "registry-1.docker.io"
			}
			require.Equal(t, origin, registries[1].Host)
		})
	}

	for _, name := range []string{"localhost:5000", "127.0.0.1:5000", "[::1]:5000"} {
		t.Run(name, func(t *testing.T) {
			registries, err := hosts(name)
			require.NoError(t, err)
			require.Len(t, registries, 1)
			require.Equal(t, name, registries[0].Host)
			require.Equal(t, "http", registries[0].Scheme)
		})
	}
}

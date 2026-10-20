package git

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNewGitIdentifier(t *testing.T) {
	tests := []struct {
		url      string
		expected GitIdentifier
	}{
		{
			url: "ssh://root@subdomain.example.hostname:2222/root/my/really/weird/path/foo.git",
			expected: GitIdentifier{
				Remote: "ssh://root@subdomain.example.hostname:2222/root/my/really/weird/path/foo.git",
			},
		},
		{
			url: "ssh://root@subdomain.example.hostname:2222/root/my/really/weird/path/foo.git#main",
			expected: GitIdentifier{
				Remote: "ssh://root@subdomain.example.hostname:2222/root/my/really/weird/path/foo.git",
				Ref:    "main",
			},
		},
		{
			url: "git@github.com:moby/buildkit.git",
			expected: GitIdentifier{
				Remote: "git@github.com:moby/buildkit.git",
			},
		},
		{
			url: "github.com/moby/buildkit.git#main",
			expected: GitIdentifier{
				Remote: "https://github.com/moby/buildkit.git",
				Ref:    "main",
			},
		},
		{
			url: "git://github.com/user/repo.git",
			expected: GitIdentifier{
				Remote: "git://github.com/user/repo.git",
			},
		},
		{
			url: "git://github.com/user/repo.git#mybranch:mydir/mysubdir/",
			expected: GitIdentifier{
				Remote: "git://github.com/user/repo.git",
				Ref:    "mybranch",
				Subdir: "mydir/mysubdir",
			},
		},
		{
			url: "git://github.com/user/repo.git#:mydir/mysubdir/",
			expected: GitIdentifier{
				Remote: "git://github.com/user/repo.git",
				Subdir: "mydir/mysubdir",
			},
		},
		{
			url: "https://github.com/user/repo.git",
			expected: GitIdentifier{
				Remote: "https://github.com/user/repo.git",
			},
		},
		{
			url: "https://github.com/user/repo.git#mybranch:mydir/mysubdir/",
			expected: GitIdentifier{
				Remote: "https://github.com/user/repo.git",
				Ref:    "mybranch",
				Subdir: "mydir/mysubdir",
			},
		},
		{
			url: "git@github.com:user/repo.git",
			expected: GitIdentifier{
				Remote: "git@github.com:user/repo.git",
			},
		},
		{
			url: "git@github.com:user/repo.git#mybranch:mydir/mysubdir/",
			expected: GitIdentifier{
				Remote: "git@github.com:user/repo.git",
				Ref:    "mybranch",
				Subdir: "mydir/mysubdir",
			},
		},
		{
			url: "ssh://github.com/user/repo.git",
			expected: GitIdentifier{
				Remote: "ssh://github.com/user/repo.git",
			},
		},
		{
			url: "ssh://github.com/user/repo.git#mybranch:mydir/mysubdir/",
			expected: GitIdentifier{
				Remote: "ssh://github.com/user/repo.git",
				Ref:    "mybranch",
				Subdir: "mydir/mysubdir",
			},
		},
		{
			url: "ssh://foo%40barcorp.com@github.com/user/repo.git#mybranch:mydir/mysubdir/",
			expected: GitIdentifier{
				Remote: "ssh://foo%40barcorp.com@github.com/user/repo.git",
				Ref:    "mybranch",
				Subdir: "mydir/mysubdir",
			},
		},
		{
			url: "https://github.com/user/repo.git#main:../../escape",
			expected: GitIdentifier{
				Remote: "https://github.com/user/repo.git",
				Ref:    "main",
				Subdir: "escape",
			},
		},
		{
			url: "https://github.com/user/repo.git#main:dir/../../escape",
			expected: GitIdentifier{
				Remote: "https://github.com/user/repo.git",
				Ref:    "main",
				Subdir: "escape",
			},
		},
		{
			url: "https://github.com/user/repo.git#main:/absolute/path",
			expected: GitIdentifier{
				Remote: "https://github.com/user/repo.git",
				Ref:    "main",
				Subdir: "absolute/path",
			},
		},
		{
			url: "https://github.com/user/repo.git#main:../",
			expected: GitIdentifier{
				Remote: "https://github.com/user/repo.git",
				Ref:    "main",
			},
		},
		{
			url: "ssh://github.com/user/repo.git#main:../../escape",
			expected: GitIdentifier{
				Remote: "ssh://github.com/user/repo.git",
				Ref:    "main",
				Subdir: "escape",
			},
		},
		{
			url: "ssh://github.com/user/repo.git#main:/absolute/path",
			expected: GitIdentifier{
				Remote: "ssh://github.com/user/repo.git",
				Ref:    "main",
				Subdir: "absolute/path",
			},
		},
		{
			url: "git@github.com:user/repo.git#main:../../escape",
			expected: GitIdentifier{
				Remote: "git@github.com:user/repo.git",
				Ref:    "main",
				Subdir: "escape",
			},
		},
		{
			url: "git@github.com:user/repo.git#main:/absolute/path",
			expected: GitIdentifier{
				Remote: "git@github.com:user/repo.git",
				Ref:    "main",
				Subdir: "absolute/path",
			},
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.url, func(t *testing.T) {
			gi, err := NewGitIdentifier(tt.url)
			require.NoError(t, err)
			require.Equal(t, tt.expected, *gi)
		})
	}
}

func TestValidateDirsOnly(t *testing.T) {
	root := t.TempDir()
	require.NoError(t, os.MkdirAll(filepath.Join(root, "valid", "subdir"), 0755))
	require.NoError(t, os.WriteFile(filepath.Join(root, "valid", "file"), []byte("data"), 0644))
	require.NoError(t, os.Symlink("/tmp", filepath.Join(root, "valid", "link")))

	require.NoError(t, validateDirsOnly(root, "valid/subdir"))
	require.NoError(t, validateDirsOnly(root, "."))

	require.Error(t, validateDirsOnly(root, "valid/file"))
	require.Error(t, validateDirsOnly(root, "valid/link"))
}

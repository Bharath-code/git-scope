package stats

import (
	"os"
	"path/filepath"
	"sort"

	"github.com/Bharath-code/git-scope/internal/model"
)

// DiskUsageData holds disk usage information for all repos
type DiskUsageData struct {
	Repos          []RepoDiskUsage
	TotalGitSize   int64
	TotalNodeSize  int64
	TotalSize      int64
	MaxSize        int64
	RepoCount      int
	HasNodeModules bool
}

// RepoDiskUsage holds disk usage for a single repo
type RepoDiskUsage struct {
	Name            string
	Path            string
	GitSize         int64 // Size of .git folder in bytes
	NodeModulesSize int64 // Size of node_modules folder in bytes
	TotalSize       int64 // Combined size
}

// GetDiskUsage calculates .git and node_modules folder sizes for all repos
func GetDiskUsage(repos []model.Repo) (*DiskUsageData, error) {
	data := &DiskUsageData{
		Repos:     make([]RepoDiskUsage, 0, len(repos)),
		RepoCount: len(repos),
	}

	for _, repo := range repos {
		usage := RepoDiskUsage{
			Name: repo.Name,
			Path: repo.Path,
		}

		// Calculate .git size
		gitPath := filepath.Join(repo.Path, ".git")
		gitSize, err := getDirSize(gitPath)
		if err == nil {
			usage.GitSize = gitSize
			data.TotalGitSize += gitSize
		}

		// Calculate node_modules size (if exists)
		nodePath := filepath.Join(repo.Path, "node_modules")
		if info, err := os.Stat(nodePath); err == nil && info.IsDir() {
			nodeSize, err := getDirSize(nodePath)
			if err == nil {
				usage.NodeModulesSize = nodeSize
				data.TotalNodeSize += nodeSize
				data.HasNodeModules = true
			}
		}

		usage.TotalSize = usage.GitSize + usage.NodeModulesSize
		data.TotalSize += usage.TotalSize

		if usage.TotalSize > data.MaxSize {
			data.MaxSize = usage.TotalSize
		}

		if usage.TotalSize > 0 {
			data.Repos = append(data.Repos, usage)
		}
	}

	// Sort by total size descending
	sort.Slice(data.Repos, func(i, j int) bool {
		return data.Repos[i].TotalSize > data.Repos[j].TotalSize
	})

	return data, nil
}

// getDirSize calculates the total size of a directory
func getDirSize(path string) (int64, error) {
	var size int64
	err := filepath.Walk(path, func(_ string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			size += info.Size()
		}
		return nil
	})
	return size, err
}

// FormatBytes formats bytes to human-readable string
func FormatBytes(bytes int64) string {
	const (
		KB = 1024
		MB = KB * 1024
		GB = MB * 1024
	)

	switch {
	case bytes >= GB:
		return formatFloat(float64(bytes)/float64(GB)) + " GB"
	case bytes >= MB:
		return formatFloat(float64(bytes)/float64(MB)) + " MB"
	case bytes >= KB:
		return formatFloat(float64(bytes)/float64(KB)) + " KB"
	default:
		return formatInt(bytes) + " B"
	}
}

func formatFloat(f float64) string {
	if f >= 100 {
		return formatInt(int64(f))
	}
	if f >= 10 {
		return formatInt(int64(f*10)/10) + "." + formatInt(int64(f*10)%10)
	}
	return formatInt(int64(f*10)/10) + "." + formatInt(int64(f*100)%100/10)
}

func formatInt(n int64) string {
	if n == 0 {
		return "0"
	}
	s := ""
	for n > 0 {
		s = string(rune('0'+n%10)) + s
		n /= 10
	}
	return s
}

// GetBarWidth returns the relative bar width (0-100) for a repo size
func (d *DiskUsageData) GetBarWidth(size int64, maxWidth int) int {
	if d.MaxSize == 0 {
		return 0
	}
	ratio := float64(size) / float64(d.MaxSize)
	return int(ratio * float64(maxWidth))
}

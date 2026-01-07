// Package repo manages Debian repository structure and metadata.
package repo

import (
	"fmt"
	"html/template"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
)

var (
	htmlTmpl     *template.Template
	htmlTmplOnce sync.Once
)

func getHTMLTemplate() *template.Template {
	htmlTmplOnce.Do(func() {
		htmlTmpl = template.Must(template.New("index").Parse(htmlTemplate))
	})
	return htmlTmpl
}

const htmlTemplate = `<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="UTF-8">
  <meta name="viewport" content="width=device-width, initial-scale=1.0">
  <title>Index of {{.Path}}</title>
  <style>
    body { font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif; max-width: 900px; margin: 50px auto; padding: 0 20px; line-height: 1.6; }
    h1 { border-bottom: 2px solid #eee; padding-bottom: 10px; font-size: 1.5em; }
    table { width: 100%; border-collapse: collapse; }
    th, td { text-align: left; padding: 8px 12px; border-bottom: 1px solid #eee; }
    th { background: #f8f8f8; font-weight: 600; }
    tr:hover { background: #f5f5f5; }
    a { color: #0366d6; text-decoration: none; }
    a:hover { text-decoration: underline; }
    .size { color: #666; font-family: monospace; }
    .icon { margin-right: 8px; }
    .parent { font-weight: 500; }
  </style>
</head>
<body>
  <h1>Index of {{.Path}}</h1>
  <table>
    <thead>
      <tr>
        <th>Name</th>
        <th>Size</th>
      </tr>
    </thead>
    <tbody>
      {{if .ShowParent}}
      <tr>
        <td class="parent"><span class="icon">📁</span><a href="../">../</a></td>
        <td>-</td>
      </tr>
      {{end}}
      {{range .Directories}}
      <tr>
        <td><span class="icon">📁</span><a href="{{.Name}}/">{{.Name}}/</a></td>
        <td>-</td>
      </tr>
      {{end}}
      {{range .Files}}
      <tr>
        <td><span class="icon">{{.Icon}}</span><a href="{{.Name}}">{{.Name}}</a></td>
        <td class="size">{{.Size}}</td>
      </tr>
      {{end}}
    </tbody>
  </table>
</body>
</html>
`

// DirectoryEntry represents a subdirectory in the index.
type DirectoryEntry struct {
	Name string
}

// FileEntry represents a file in the index.
type FileEntry struct {
	Name string
	Size string
	Icon string
}

// IndexData holds data for rendering an HTML index page.
type IndexData struct {
	Path        string
	ShowParent  bool
	Directories []DirectoryEntry
	Files       []FileEntry
}

// GenerateHTMLIndexes creates index.html files in all repository directories
// to enable browser-friendly navigation.
func (r *Repository) GenerateHTMLIndexes() error {
	// Walk the entire repository and generate index.html for each directory
	return filepath.Walk(r.Root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !info.IsDir() {
			return nil
		}

		// Skip hidden directories (like .git)
		if strings.HasPrefix(info.Name(), ".") && path != r.Root {
			return filepath.SkipDir
		}

		return r.generateIndexForDirectory(path)
	})
}

func (r *Repository) generateIndexForDirectory(dirPath string) error {
	entries, err := os.ReadDir(dirPath)
	if err != nil {
		return fmt.Errorf("read directory %s: %w", dirPath, err)
	}

	var directories []DirectoryEntry
	var files []FileEntry

	for _, entry := range entries {
		name := entry.Name()

		// Skip hidden files and the index.html we're generating
		if strings.HasPrefix(name, ".") || name == "index.html" {
			continue
		}

		if entry.IsDir() {
			directories = append(directories, DirectoryEntry{Name: name})
		} else {
			info, err := entry.Info()
			if err != nil {
				continue
			}
			files = append(files, FileEntry{
				Name: name,
				Size: formatSize(info.Size()),
				Icon: iconForFile(name),
			})
		}
	}

	// Sort alphabetically
	sort.Slice(directories, func(i, j int) bool {
		return directories[i].Name < directories[j].Name
	})
	sort.Slice(files, func(i, j int) bool {
		return files[i].Name < files[j].Name
	})

	// Calculate relative path for display (use forward slashes for URLs)
	relPath, err := filepath.Rel(r.Root, dirPath)
	if err != nil {
		relPath = dirPath
	}
	if relPath == "." {
		relPath = "/"
	} else {
		relPath = "/" + filepath.ToSlash(relPath) + "/"
	}

	// Determine if we should show parent link
	showParent := dirPath != r.Root

	data := IndexData{
		Path:        relPath,
		ShowParent:  showParent,
		Directories: directories,
		Files:       files,
	}

	indexPath := filepath.Join(dirPath, "index.html")
	f, err := os.Create(indexPath)
	if err != nil {
		return fmt.Errorf("create index.html: %w", err)
	}
	defer f.Close() //nolint:errcheck // Write errors caught by template.Execute

	if err := getHTMLTemplate().Execute(f, data); err != nil {
		return fmt.Errorf("execute template: %w", err)
	}

	return nil
}

func formatSize(size int64) string {
	const (
		KB = 1024
		MB = 1024 * KB
		GB = 1024 * MB
	)

	switch {
	case size >= GB:
		return fmt.Sprintf("%.1f GB", float64(size)/float64(GB))
	case size >= MB:
		return fmt.Sprintf("%.1f MB", float64(size)/float64(MB))
	case size >= KB:
		return fmt.Sprintf("%.1f KB", float64(size)/float64(KB))
	default:
		return fmt.Sprintf("%d B", size)
	}
}

func iconForFile(name string) string {
	lower := strings.ToLower(name)

	switch {
	case strings.HasSuffix(lower, ".deb"):
		return "📦"
	case strings.HasSuffix(lower, ".gz") || strings.HasSuffix(lower, ".xz"):
		return "🗜️"
	case strings.HasSuffix(lower, ".gpg") || strings.HasSuffix(lower, ".key"):
		return "🔑"
	case strings.HasSuffix(lower, ".html"):
		return "🌐"
	case strings.Contains(lower, "release") || strings.Contains(lower, "packages"):
		return "📄"
	default:
		return "📄"
	}
}

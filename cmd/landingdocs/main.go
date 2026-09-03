// Command landingdocs converts the cobra-generated docs/reference/*.md
// files into Hugo pages for the landing_page repo's Qdrant Cloud CLI
// reference section.
//
// Usage:
//
//	go run ./cmd/landingdocs <src_dir> <dest_dir>
//
// <src_dir>  is docs/reference (output of `make docs`)
// <dest_dir> is the landing_page repo's
//
//	qdrant-landing/content/documentation/cloud-cli/reference directory
//
// The destination directory is fully regenerated on every run (existing
// generated files are replaced, stale ones removed) so it always mirrors
// the current command tree.
package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
)

var linkRE = regexp.MustCompile(`\[([^\]]+)\]\((qcloud[\w.-]*)\.md\)`)

func rewriteLinks(text string) string {
	return linkRE.ReplaceAllStringFunc(text, func(match string) string {
		groups := linkRE.FindStringSubmatch(match)
		label, target := groups[1], groups[2]
		if target == "qcloud" {
			return fmt.Sprintf("[%s](/documentation/cloud-cli/reference/)", label)
		}

		return fmt.Sprintf("[%s](/documentation/cloud-cli/reference/%s/)", label, target)
	})
}

// demoteHeadings turns "## " into "# ", "### " into "## ", etc. so the
// page owns a single H1 title.
func demoteHeadings(text string) string {
	lines := strings.Split(text, "\n")
	for i, line := range lines {
		if strings.HasPrefix(line, "#") {
			trimmed := strings.TrimLeft(line, "#")
			hashes := len(line) - len(trimmed)
			if hashes > 1 {
				lines[i] = line[1:]
			}
		}
	}

	return strings.Join(lines, "\n")
}

// annotateCodeFences tags every opening ``` fence as bash. Bare ```
// fences don't render as code blocks on the landing page, and cobra
// only emits usage/example/flag blocks, which are all shell-ish.
func annotateCodeFences(text string) string {
	lines := strings.Split(text, "\n")
	inFence := false
	for i, line := range lines {
		if strings.TrimSpace(line) == "```" {
			if !inFence {
				lines[i] = "```bash"
			}

			inFence = !inFence
		}
	}

	return strings.Join(lines, "\n")
}

type page struct {
	title       string
	description string
	body        string
}

func convert(srcPath string) (page, error) {
	raw, err := os.ReadFile(srcPath)
	if err != nil {
		return page{}, err
	}

	lines := strings.Split(string(raw), "\n")

	titleLine := lines[0]
	if !strings.HasPrefix(titleLine, "## ") {
		return page{}, fmt.Errorf("unexpected heading in %s: %q", srcPath, titleLine)
	}

	title := strings.TrimSpace(strings.TrimPrefix(titleLine, "## "))

	description := ""
	for _, line := range lines[1:] {
		if stripped := strings.TrimSpace(line); stripped != "" {
			description = stripped
			break
		}
	}

	body := strings.Join(lines, "\n")
	body = demoteHeadings(body)
	body = rewriteLinks(body)
	body = annotateCodeFences(body)

	return page{title: title, description: description, body: body}, nil
}

func frontmatter(title, description string, weight int, partition string) string {
	short := description
	if len(short) > 120 {
		short = strings.TrimSpace(short[:117]) + "..."
	}

	fm := fmt.Sprintf(
		"---\ntitle: %s\nshort_description: %q\ndescription: %q\nweight: %d\n",
		title, short, description, weight,
	)

	// Only section indexes carry a partition in the landing_page docs tree.
	if partition != "" {
		fm += fmt.Sprintf("partition: %s\n", partition)
	}

	return fm + "---\n\n"
}

func run(srcDir, destDir string) error {
	if err := os.MkdirAll(destDir, 0o755); err != nil {
		return fmt.Errorf("create dest dir: %w", err)
	}

	existing, err := filepath.Glob(filepath.Join(destDir, "*.md"))
	if err != nil {
		return err
	}

	for _, f := range existing {
		if err := os.Remove(f); err != nil {
			return err
		}
	}

	files, err := filepath.Glob(filepath.Join(srcDir, "*.md"))
	if err != nil {
		return err
	}

	if len(files) == 0 {
		return fmt.Errorf("no markdown files found in %s", srcDir)
	}

	sort.Strings(files)

	for i, src := range files {
		p, err := convert(src)
		if err != nil {
			return err
		}

		name := filepath.Base(src)
		var destName, fm string
		if name == "qcloud.md" {
			destName = "_index.md"
			fm = frontmatter("Command Reference", p.description, 0, "deploy")
		} else {
			destName = name
			fm = frontmatter(p.title, p.description, i+1, "")
		}

		dest := filepath.Join(destDir, destName)
		content := fm + p.body + "\n"
		if err := os.WriteFile(dest, []byte(content), 0o644); err != nil {
			return err
		}
	}

	fmt.Printf("wrote %d reference pages to %s\n", len(files), destDir)
	return nil
}

func main() {
	if len(os.Args) != 3 {
		fmt.Fprintln(os.Stderr, "usage: landingdocs <src_dir> <dest_dir>")
		os.Exit(1)
	}

	if err := run(os.Args[1], os.Args[2]); err != nil {
		log.Fatal(err)
	}
}

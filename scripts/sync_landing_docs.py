#!/usr/bin/env python3

"""Convert cobra-generated docs/reference/*.md into Hugo pages for the
landing_page repo's Qdrant Cloud CLI reference section.

Usage:
    python3 scripts/sync_landing_docs.py <src_dir> <dest_dir>

<src_dir>  is docs/reference (output of `make docs`)
<dest_dir> is the landing_page repo's
           qdrant-landing/content/documentation/cloud-cli/reference directory

The destination directory is fully regenerated on every run (existing
generated files are replaced, stale ones removed) so it always mirrors the
current command tree.
"""

import re
import sys
from pathlib import Path

LINK_RE = re.compile(r"\[([^\]]+)\]\((qcloud[\w.-]*)\.md\)")


def slug_for(filename: str) -> str:
    """qcloud_cluster_create.md -> qcloud_cluster_create"""
    return filename[:-3]


def rewrite_links(text: str) -> str:
    def repl(match: re.Match[str]):
        label, target = match.group(1), match.group(2)
        if target == "qcloud":
            return f"[{label}](/documentation/cloud-cli/reference/)"
        return f"[{label}](/documentation/cloud-cli/reference/{target}/)"

    return LINK_RE.sub(repl, text)


def demote_headings(text: str) -> str:
    """## -> #, ### -> ##, etc. so the page owns a single H1 title."""
    out_lines: list[str] = []
    for line in text.splitlines():
        if line.startswith("#"):
            hashes = len(line) - len(line.lstrip("#"))
            out_lines.append(line[1:] if hashes > 1 else line)
        else:
            out_lines.append(line)
    return "\n".join(out_lines)


def annotate_code_fences(text: str) -> str:
    """Bare ``` fences don't render as code blocks on the landing page;
    tag every opening fence as bash (cobra only emits usage/example/flag
    blocks, which are all shell-ish)."""
    out_lines: list[str] = []
    in_fence = False
    for line in text.splitlines():
        if line.strip() == "```":
            out_lines.append("```bash" if not in_fence else "```")
            in_fence = not in_fence
        else:
            out_lines.append(line)
    return "\n".join(out_lines)


def convert(src_path: Path) -> tuple[str, str, str]:
    """Return (title, description, body) for one generated doc file."""
    lines = src_path.read_text().splitlines()

    title_line = lines[0]
    assert title_line.startswith("## "), f"unexpected heading in {src_path}"
    title = title_line[3:].strip()

    description = ""
    for line in lines[1:]:
        stripped = line.strip()
        if stripped:
            description = stripped
            break

    body = "\n".join(lines)
    body = demote_headings(body)
    body = rewrite_links(body)
    body = annotate_code_fences(body)
    return title, description, body


def frontmatter(title: str, description: str, weight: int, extra: str = "") -> str:
    short = description if len(description) <= 120 else description[:117].rstrip() + "..."
    return (
        "---\n"
        f"title: {title}\n"
        f"short_description: \"{short}\"\n"
        f"description: \"{description}\"\n"
        f"weight: {weight}\n"
        f"{extra}"
        "---\n\n"
    )


def main():
    if len(sys.argv) != 3:
        print(__doc__)
        sys.exit(1)

    src_dir = Path(sys.argv[1])
    dest_dir = Path(sys.argv[2])
    dest_dir.mkdir(parents=True, exist_ok=True)

    for existing in dest_dir.glob("*.md"):
        existing.unlink()

    files = sorted(p for p in src_dir.glob("*.md"))
    if not files:
        print(f"no markdown files found in {src_dir}", file=sys.stderr)
        sys.exit(1)

    for i, src in enumerate(files):
        title, description, body = convert(src)

        if src.name == "qcloud.md":
            dest = dest_dir / "_index.md"
            fm = frontmatter("Command Reference", description, 0)
        else:
            dest = dest_dir / src.name
            fm = frontmatter(title, description, i + 1)

        _ = dest.write_text(fm + body + "\n")

    print(f"wrote {len(files)} reference pages to {dest_dir}")


if __name__ == "__main__":
    main()

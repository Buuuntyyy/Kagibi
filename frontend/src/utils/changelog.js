// Copyright (C) 2025-2026  Buuuntyyy
// SPDX-License-Identifier: AGPL-3.0-or-later

/**
 * Minimal parser for the project's CHANGELOG.md convention — extracts only the
 * topmost (most recent) "## vX.Y.Z — YYYY-MM-DD" section, used by ChangelogModal.vue.
 * Not a general markdown parser: it only understands the small subset actually used
 * (## version heading, ### section headings, "- **label** : text" or plain "- text"
 * bullets, and "---" section separators).
 */

const VERSION_HEADING_RE = /^## v(\S+?)\s*(?:—|--|-)\s*(.+?)\s*$/
const SECTION_HEADING_RE = /^### (.+)$/
const BULLET_RE = /^- (.+)$/
const BOLD_LABEL_RE = /^\*\*(.+?)\*\*\s*(?::|—)?\s*(.*)$/

export function parseLatestChangelogEntry(markdown) {
  const lines = markdown.replace(/\r\n/g, '\n').split('\n')

  let i = 0
  while (i < lines.length && !VERSION_HEADING_RE.test(lines[i])) i++
  if (i >= lines.length) return null

  const heading = lines[i].match(VERSION_HEADING_RE)
  const version = heading[1]
  const date = heading[2]
  i++

  const sections = []
  let current = null

  for (; i < lines.length; i++) {
    const line = lines[i]
    if (VERSION_HEADING_RE.test(line)) break // next version section — stop
    if (/^---\s*$/.test(line)) break // separator — stop

    const sectionMatch = line.match(SECTION_HEADING_RE)
    if (sectionMatch) {
      current = { title: sectionMatch[1].trim(), items: [] }
      sections.push(current)
      continue
    }

    const bulletMatch = line.match(BULLET_RE)
    if (bulletMatch && current) {
      const raw = bulletMatch[1]
      const boldMatch = raw.match(BOLD_LABEL_RE)
      if (boldMatch) {
        const text = boldMatch[2].replace(/^\.\s*$/, '').trim()
        current.items.push({ label: boldMatch[1], text })
      } else {
        current.items.push({ label: null, text: raw })
      }
    }
  }

  return { version, date, sections }
}

import { describe, it, expect } from 'vitest'
import fs from 'node:fs'
import path from 'node:path'

const cssPath = path.resolve(__dirname, '../index.css')
const css = fs.readFileSync(cssPath, 'utf-8')

function extractTokens(block: string): Set<string> {
  const tokens = new Set<string>()
  const re = /--([a-z][\w-]*)\s*:/g
  let match: RegExpExecArray | null
  while ((match = re.exec(block)) !== null) {
    tokens.add(match[1])
  }
  return tokens
}

function extractBlock(css: string, selector: string): string {
  const pattern = new RegExp(`${selector.replace(/[.*+?^${}()|[\]\\]/g, '\\$&')}\\s*\\{`, 'g')
  const match = pattern.exec(css)
  if (!match) return ''
  let depth = 0
  let start = match.index + match[0].length
  for (let i = start; i < css.length; i++) {
    if (css[i] === '{') depth++
    if (css[i] === '}') {
      if (depth === 0) return css.slice(start, i)
      depth--
    }
  }
  return ''
}

describe('Dark mode token completeness', () => {
  const newTokens = [
    'primary-hover',
    'primary-active',
    'primary-muted',
    'success',
    'success-foreground',
    'warning',
    'warning-foreground',
    'sidebar-muted',
  ]

  // Find the oklch :root and .dark blocks (the ones after @theme inline)
  const rootBlocks = css.split(':root')
  const oklchRoot = rootBlocks.length > 2 ? rootBlocks[2] : ''
  const darkBlocks = css.split('.dark')
  const oklchDark = darkBlocks.length > 2 ? darkBlocks[2] : ''

  it.each(newTokens)('token --%s exists in :root', (token) => {
    expect(oklchRoot).toContain(`--${token}:`)
  })

  it.each(newTokens)('token --%s exists in .dark', (token) => {
    expect(oklchDark).toContain(`--${token}:`)
  })

  it('all :root sidebar tokens have .dark counterparts', () => {
    const rootBlock = extractBlock(css.slice(css.lastIndexOf(':root')), ':root')
    const darkBlock = extractBlock(css.slice(css.lastIndexOf('.dark')), '.dark')
    const rootTokens = extractTokens(rootBlock)
    const darkTokens = extractTokens(darkBlock)

    const sidebarTokens = [...rootTokens].filter((t) => t.startsWith('sidebar'))
    for (const token of sidebarTokens) {
      expect(darkTokens.has(token), `Missing .dark token: --${token}`).toBe(true)
    }
  })

  it('@theme inline registers all new color tokens', () => {
    const themeBlock = extractBlock(css, '@theme inline')
    for (const token of newTokens) {
      expect(themeBlock, `Missing @theme registration: --color-${token}`).toContain(
        `--color-${token}`,
      )
    }
  })
})

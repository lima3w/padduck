#!/usr/bin/env node
import { createHash } from 'node:crypto'
import { readFileSync, writeFileSync } from 'node:fs'

const openapiPath = 'docs/openapi.yaml'
const changelogPath = 'CHANGELOG.md'
const openapi = readFileSync(openapiPath, 'utf8')
const changelog = readFileSync(changelogPath, 'utf8')

const version = openapi.match(/^  version: (.+)$/m)?.[1]?.trim()
if (!version) {
  throw new Error(`Unable to find info.version in ${openapiPath}`)
}

const pathCount = [...openapi.matchAll(/^  \/api\//gm)].length
const digest = createHash('sha256').update(openapi).digest('hex').slice(0, 12)
const marker = `<!-- api-contract:${version} -->`
const entry = `${marker}

API contract snapshot:

- OpenAPI version: \`${version}\`
- Public API path count: \`${pathCount}\`
- OpenAPI SHA-256: \`${digest}\`
`

const next = changelog.includes(marker)
  ? changelog.replace(new RegExp(`${marker}[\\s\\S]*?(?=\\n## |\\s*$)`), entry.trimEnd())
  : `${changelog.trimEnd()}\n\n${entry.trimEnd()}\n`

writeFileSync(changelogPath, `${next.trimEnd()}\n`)

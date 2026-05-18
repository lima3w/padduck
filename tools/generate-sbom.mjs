#!/usr/bin/env node
import fs from 'node:fs'
import path from 'node:path'

const root = process.cwd()
const outDir = path.join(root, 'sbom')
const outPath = path.join(outDir, 'dependency-sbom.cdx.json')

function encodePurlPart(value) {
  return encodeURIComponent(value).replace(/%2F/g, '/')
}

function goPurl(modulePath, version) {
  return `pkg:golang/${encodePurlPart(modulePath)}@${encodeURIComponent(version)}`
}

function npmPurl(packageName, version) {
  if (packageName.startsWith('@')) {
    const [scope, name] = packageName.split('/')
    return `pkg:npm/${encodeURIComponent(scope)}/${encodeURIComponent(name)}@${encodeURIComponent(version)}`
  }
  return `pkg:npm/${encodeURIComponent(packageName)}@${encodeURIComponent(version)}`
}

function parseGoVendorModules(filePath) {
  const text = fs.readFileSync(filePath, 'utf8')
  const modules = []
  const seen = new Set()

  for (const line of text.split(/\r?\n/)) {
    if (!line.startsWith('# ')) continue
    const parts = line.slice(2).trim().split(/\s+/)
    if (parts.length < 2 || parts.includes('=>')) continue

    const [name, version] = parts
    const key = `${name}@${version}`
    if (seen.has(key)) continue
    seen.add(key)

    modules.push({
      type: 'library',
      group: 'go',
      name,
      version,
      purl: goPurl(name, version),
      'bom-ref': `go:${key}`,
    })
  }

  return modules
}

function parseNpmLock(filePath) {
  const lock = JSON.parse(fs.readFileSync(filePath, 'utf8'))
  const components = []
  const seen = new Set()

  for (const [packagePath, pkg] of Object.entries(lock.packages || {})) {
    if (!packagePath.startsWith('node_modules/') || !pkg.version) continue

    const name = packagePath.replace(/^.*node_modules\//, '')
    const key = `${name}@${pkg.version}`
    if (seen.has(key)) continue
    seen.add(key)

    const component = {
      type: 'library',
      group: 'npm',
      name,
      version: pkg.version,
      scope: pkg.dev ? 'optional' : 'required',
      purl: npmPurl(name, pkg.version),
      'bom-ref': `npm:${key}`,
    }

    if (pkg.license) {
      component.licenses = [{ license: { id: pkg.license } }]
    }

    components.push(component)
  }

  components.sort((a, b) => a.name.localeCompare(b.name) || a.version.localeCompare(b.version))
  return components
}

const components = [
  ...parseGoVendorModules(path.join(root, 'backend/vendor/modules.txt')),
  ...parseNpmLock(path.join(root, 'frontend/package-lock.json')),
]

const sbom = {
  bomFormat: 'CycloneDX',
  specVersion: '1.5',
  version: 1,
  metadata: {
    timestamp: new Date().toISOString(),
    tools: [
      {
        vendor: 'ipam-next',
        name: 'tools/generate-sbom.mjs',
        version: '1',
      },
    ],
    component: {
      type: 'application',
      name: 'ipam-next',
    },
  },
  components,
}

fs.mkdirSync(outDir, { recursive: true })
fs.writeFileSync(outPath, `${JSON.stringify(sbom, null, 2)}\n`)
console.log(`Wrote ${outPath} with ${components.length} components`)

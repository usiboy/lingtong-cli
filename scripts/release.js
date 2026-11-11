const fs = require("node:fs");
const { execFileSync, spawnSync } = require("node:child_process");

const REGISTRY = "https://registry.npmjs.org";
const STABLE_VERSION = /^(0|[1-9]\d*)\.(0|[1-9]\d*)\.(0|[1-9]\d*)$/;

function stable(version) {
  return typeof version === "string" && STABLE_VERSION.test(version) && !version.includes("\n");
}

function compare(a, b) {
  const left = a.split(".").map(BigInt);
  const right = b.split(".").map(BigInt);
  for (let i = 0; i < 3; i++) {
    if (left[i] !== right[i]) return left[i] > right[i] ? 1 : -1;
  }
  return 0;
}

function planRelease({ packageVersion, tags, headTags, publishedVersions, requestedTag = "", behindRelease = false }) {
  if (!stable(packageVersion)) throw new Error("package.json must contain a stable version");
  if (!Array.isArray(publishedVersions) || !publishedVersions.every((v) => typeof v === "string")) {
    throw new Error("Invalid publishedVersions registry response");
  }
  const versions = tags.filter((tag) => tag.startsWith("v") && stable(tag.slice(1)))
    .map((tag) => tag.slice(1));
  const published = publishedVersions.filter(stable);
  let version;
  let createTag = false;
  if (requestedTag) {
    if (!requestedTag.startsWith("v") || !stable(requestedTag.slice(1)) || !tags.includes(requestedTag)) {
      throw new Error("Recovery requires an existing stable release tag (vX.Y.Z)");
    }
    version = requestedTag.slice(1);
  } else {
    const existing = headTags.filter((tag) => tag.startsWith("v") && stable(tag.slice(1)))
      .map((tag) => tag.slice(1)).sort(compare);
    if (existing.length) {
      version = existing.at(-1);
    } else {
      if (behindRelease) throw new Error("Outdated release commit: newer tagged commits exist; refusing to allocate a new version");
      const base = [packageVersion, ...versions, ...published].sort(compare).at(-1).split(".");
      base[2] = String(BigInt(base[2]) + 1n);
      version = base.join(".");
      createTag = true;
    }
  }
  const historical = behindRelease || [...versions, ...published].some((v) => compare(v, version) > 0);
  return {
    tag: `v${version}`, version, createTag,
    published: publishedVersions.includes(version),
    npmTag: historical ? "historical" : "latest",
    makeLatest: !historical,
  };
}

function isBehindRelease(tags, ref = "HEAD", cwd = process.cwd()) {
  for (const tag of tags.filter((tag) => tag.startsWith("v") && stable(tag.slice(1)))) {
    const result = spawnSync("git", ["merge-base", "--is-ancestor", `refs/tags/${tag}^{commit}`, ref], { cwd, encoding: "utf8" });
    if (result.status === 1) return true;
    if (result.status !== 0) throw new Error(`Git ancestry check failed for ${tag}`);
  }
  return false;
}

function releaseAssets(version) {
  const assets = ["checksums.txt"];
  for (const os of ["darwin", "linux", "windows"]) {
    for (const arch of ["amd64", "arm64"]) {
      assets.push(`lingtong-cli-${version}-${os}-${arch}.${os === "windows" ? "zip" : "tar.gz"}`);
    }
  }
  return assets;
}

function isCompleteRelease(release, version) {
  if (!release || release.draft) return false;
  const assets = new Set(release.assets.map((asset) => asset.name));
  if (!releaseAssets(version).every((name) => assets.has(name))) {
    throw new Error("Published GitHub Release is incomplete; refusing to replace public artifacts");
  }
  return true;
}

function validateChecksums(content, version) {
  const checksums = new Map();
  for (const line of content.split(/\r?\n/)) {
    const match = line.match(/^([a-fA-F0-9]{64})\s+\*?(.+)$/);
    if (match) checksums.set(match[2], match[1]);
  }
  for (const name of releaseAssets(version).slice(1)) {
    if (!checksums.has(name)) throw new Error(`Missing or invalid SHA-256 checksum for ${name}`);
  }
}

function output(values) {
  console.log(JSON.stringify(values, null, 2));
  if (process.env.GITHUB_OUTPUT) {
    for (const [key, value] of Object.entries(values)) {
      fs.appendFileSync(process.env.GITHUB_OUTPUT, `${key}=${value}\n`);
    }
  }
}

async function main(command) {
  if (command === "plan") {
    const git = (...args) => execFileSync("git", args, { encoding: "utf8" }).trim();
    const metadata = JSON.parse(execFileSync("npm", [
      "view", "@lingtong-cli/cli", "versions", "--json", `--registry=${REGISTRY}`,
    ], { encoding: "utf8" }));
    const inputs = {
      packageVersion: JSON.parse(fs.readFileSync("package.json", "utf8")).version,
      tags: git("tag", "--list").split("\n"),
      headTags: git("tag", "--points-at", "HEAD").split("\n"),
      publishedVersions: typeof metadata === "string" ? [metadata] : metadata,
      requestedTag: process.env.REQUESTED_TAG || "",
    };
    const initial = planRelease(inputs);
    const ref = initial.createTag ? "HEAD" : `refs/tags/${initial.tag}^{commit}`;
    const plan = planRelease({ ...inputs, behindRelease: isBehindRelease(inputs.tags, ref) });
    output({ ...plan, commit: git("rev-parse", plan.createTag ? "HEAD" : `refs/tags/${plan.tag}^{commit}`) });
  } else if (command === "release-state") {
    const response = await fetch(
      `https://api.github.com/repos/${process.env.GITHUB_REPOSITORY}/releases/tags/${process.env.RELEASE_TAG}`,
      {
        headers: { Authorization: `Bearer ${process.env.GH_TOKEN}`, Accept: "application/vnd.github+json" },
        signal: AbortSignal.timeout(30000),
      },
    );
    if (response.status !== 404 && !response.ok) throw new Error(`GitHub Release lookup failed: HTTP ${response.status}`);
    const release = response.status === 404 ? null : await response.json();
    output({ complete: isCompleteRelease(release, process.env.RELEASE_VERSION) });
  } else if (command === "checksums") {
    validateChecksums(fs.readFileSync("checksums.txt", "utf8"), process.env.RELEASE_VERSION);
    console.log("All six release archives have SHA-256 checksums");
  } else {
    throw new Error("Usage: node scripts/release.js plan|release-state|checksums");
  }
}

if (require.main === module) {
  main(process.argv[2]).catch((error) => {
    console.error(error.message);
    process.exitCode = 1;
  });
}

module.exports = { planRelease, isBehindRelease, releaseAssets, isCompleteRelease, validateChecksums };

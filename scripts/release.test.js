const assert = require("node:assert/strict");
const test = require("node:test");
const fs = require("node:fs");
const os = require("node:os");
const path = require("node:path");
const { execFileSync } = require("node:child_process");

test("release automation exposes a testable version planner", () => {
  assert.equal(typeof require("./release.js").planRelease, "function");
});

function plan(overrides = {}) {
  return require("./release.js").planRelease({
    packageVersion: "0.1.0",
    tags: ["v0.1.0"],
    headTags: [],
    publishedVersions: ["0.1.0"],
    ...overrides,
  });
}

test("a merged PR gets the next patch version", () => {
  assert.deepEqual(plan(), {
    tag: "v0.1.1", version: "0.1.1", createTag: true,
    published: false, npmTag: "latest", makeLatest: true,
  });
});

test("version selection compares numbers and includes reserved tags and npm versions", () => {
  assert.equal(plan({ tags: ["v0.1.9", "v0.1.10"], publishedVersions: ["0.1.11"] }).tag, "v0.1.12");
  assert.equal(plan({ tags: ["v2.0.9", "v2.0.10", "v9.0.0-beta.1"] }).tag, "v2.0.11");
});

test("package.json is also a version floor", () => {
  assert.equal(plan({ packageVersion: "1.0.0" }).version, "1.0.1");
});

test("retrying a tagged merge reuses its unpublished version", () => {
  assert.deepEqual(plan({ tags: ["v0.1.0", "v0.1.1"], headTags: ["v0.1.1"] }), {
    tag: "v0.1.1", version: "0.1.1", createTag: false,
    published: false, npmTag: "latest", makeLatest: true,
  });
});

test("a successfully published version is skipped on retry", () => {
  assert.equal(plan({ headTags: ["v0.1.0"] }).published, true);
});

test("manual recovery only accepts an existing stable release tag", () => {
  assert.equal(plan({ requestedTag: "v0.1.0" }).createTag, false);
  for (const requestedTag of ["main", "v0.1.1", "v1.0.0-beta.1", "v01.2.3", "v1.2.3\n"]) {
    assert.throws(() => plan({ requestedTag }), /release tag/i);
  }
});

test("retrying an older release does not move the npm latest tag backwards", () => {
  assert.equal(plan({ requestedTag: "v0.1.0", publishedVersions: ["0.1.2"] }).npmTag, "historical");
});

test("an old untagged commit cannot allocate a new latest version after a newer release", () => {
  assert.throws(() => plan({ behindRelease: true }), /outdated release commit/i);
});

test("retrying a reserved old commit cannot become latest even if newer npm publication failed", () => {
  const result = plan({ tags: ["v0.1.0", "v0.1.1"], headTags: ["v0.1.1"], behindRelease: true });
  assert.equal(result.npmTag, "historical");
  assert.equal(result.makeLatest, false);
});

test("historical recovery cannot become GitHub latest when only the newer Git tag exists", () => {
  const result = plan({ tags: ["v0.1.0", "v0.1.1"], requestedTag: "v0.1.0", publishedVersions: [] });
  assert.equal(result.npmTag, "historical");
  assert.equal(result.makeLatest, false);
});

test("real Git ancestry detects outdated release commits and accepts their descendants", () => {
  const { isBehindRelease } = require("./release.js");
  const dir = fs.mkdtempSync(path.join(process.env.RUNNER_TEMP || os.tmpdir(), "release-history-"));
  const git = (...args) => execFileSync("git", ["-C", dir, ...args], { encoding: "utf8", stdio: ["ignore", "pipe", "pipe"] }).trim();
  try {
    git("init");
    git("-c", "user.name=Release Test", "-c", "user.email=release@example.invalid", "commit", "--allow-empty", "-m", "PR A");
    const first = git("rev-parse", "HEAD");
    git("-c", "user.name=Release Test", "-c", "user.email=release@example.invalid", "commit", "--allow-empty", "-m", "PR B");
    git("tag", "v0.1.1");
    assert.equal(isBehindRelease(["v0.1.1"], "HEAD", dir), false);
    assert.equal(isBehindRelease(["v0.1.1"], first, dir), true);
    assert.throws(() => isBehindRelease(["v0.1.2"], "HEAD", dir), /Git ancestry check failed/);
  } finally {
    fs.rmSync(dir, { recursive: true, force: true });
  }
});

test("GoReleaser recovery explicitly replaces only reusable draft assets and controls latest", () => {
  const config = fs.readFileSync(path.join(__dirname, "../.goreleaser.yml"), "utf8");
  assert.match(config, /use_existing_draft: true/);
  assert.match(config, /replace_existing_artifacts: true/);
  assert.match(config, /make_latest: "\{\{ \.Env\.RELEASE_MAKE_LATEST \}\}"/);
});

test("invalid registry metadata fails closed instead of allocating a version", () => {
  assert.throws(() => plan({ publishedVersions: null }), /publishedVersions/);
  assert.throws(() => plan({ packageVersion: "dev" }), /version/i);
});

test("a reusable GitHub Release must contain checksums and all six archives", () => {
  const { releaseAssets, isCompleteRelease } = require("./release.js");
  const assets = releaseAssets("0.1.1").map((name) => ({ name }));
  assert.equal(assets.length, 7);
  assert.equal(isCompleteRelease({ draft: false, assets }, "0.1.1"), true);
  assert.equal(isCompleteRelease({ draft: true, assets }, "0.1.1"), false);
  assert.throws(() => isCompleteRelease({ draft: false, assets: assets.slice(1) }, "0.1.1"), /incomplete/i);
});

test("npm packaging rejects missing, malformed, or mismatched checksums", () => {
  const { releaseAssets, validateChecksums } = require("./release.js");
  const content = releaseAssets("0.1.1").filter((name) => name !== "checksums.txt")
    .map((name) => `${"a".repeat(64)}  ${name}`).join("\n");
  assert.doesNotThrow(() => validateChecksums(content, "0.1.1"));
  assert.throws(() => validateChecksums("", "0.1.1"), /checksum/i);
  assert.throws(() => validateChecksums(content.replace(/a/g, "x"), "0.1.1"), /checksum/i);
  assert.throws(() => validateChecksums(content, "0.1.2"), /checksum/i);
});

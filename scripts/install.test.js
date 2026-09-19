const assert = require("node:assert/strict");
const crypto = require("node:crypto");
const fs = require("node:fs");
const os = require("node:os");
const path = require("node:path");
const test = require("node:test");

const installer = require("./install.js");
const wizard = require("./install-wizard.js");

test("resolveTarget maps supported Node platforms to Go release targets", () => {
  assert.deepEqual(installer.resolveTarget("darwin", "arm64"), {
    os: "darwin",
    arch: "arm64",
    extension: ".tar.gz",
    binaryExtension: "",
  });
  assert.deepEqual(installer.resolveTarget("win32", "x64"), {
    os: "windows",
    arch: "amd64",
    extension: ".zip",
    binaryExtension: ".exe",
  });
});

test("resolveTarget rejects unsupported platforms", () => {
  assert.throws(
    () => installer.resolveTarget("freebsd", "x64"),
    /Unsupported platform/,
  );
  assert.throws(
    () => installer.resolveTarget("linux", "ia32"),
    /Unsupported architecture/,
  );
});

test("archiveName uses the npm version without a v prefix", () => {
  const target = installer.resolveTarget("linux", "x64");
  assert.equal(
    installer.archiveName("0.1.0", target),
    "lingtong-cli-0.1.0-linux-amd64.tar.gz",
  );
});

test("releaseBase defaults to the public usiboy GitHub repository", () => {
  const original = process.env.LINGTONG_CLI_DOWNLOAD_BASE;
  delete process.env.LINGTONG_CLI_DOWNLOAD_BASE;
  assert.equal(
    installer.releaseBase(),
    "https://github.com/usiboy/lingtong-cli/releases/download/v0.1.0",
  );
  if (original === undefined) {
    delete process.env.LINGTONG_CLI_DOWNLOAD_BASE;
  } else {
    process.env.LINGTONG_CLI_DOWNLOAD_BASE = original;
  }
});

test("parseChecksums finds the exact release artifact", () => {
  const checksums = [
    "abc123  lingtong-cli-0.1.0-linux-amd64.tar.gz",
    "def456  lingtong-cli-0.1.0-darwin-arm64.tar.gz",
  ].join("\n");

  assert.equal(
    installer.parseChecksum(
      checksums,
      "lingtong-cli-0.1.0-darwin-arm64.tar.gz",
    ),
    "def456",
  );
  assert.throws(
    () => installer.parseChecksum(checksums, "missing.tar.gz"),
    /Checksum entry not found/,
  );
});

test("verifyChecksum rejects modified downloads", () => {
  const dir = fs.mkdtempSync(path.join(os.tmpdir(), "lingtong-cli-test-"));
  const file = path.join(dir, "archive");
  fs.writeFileSync(file, "expected content");
  const expected = crypto
    .createHash("sha256")
    .update("expected content")
    .digest("hex");

  assert.doesNotThrow(() => installer.verifyChecksum(file, expected));
  fs.writeFileSync(file, "modified content");
  assert.throws(
    () => installer.verifyChecksum(file, expected),
    /Checksum mismatch/,
  );
  fs.rmSync(dir, { recursive: true, force: true });
});

test("installPersistent copies the package binary into the user bin directory", () => {
  const dir = fs.mkdtempSync(path.join(os.tmpdir(), "lingtong-cli-test-"));
  const source = path.join(dir, "source", "lingtong-cli");
  const destinationDir = path.join(dir, "home", ".local", "bin");
  fs.mkdirSync(path.dirname(source), { recursive: true });
  fs.writeFileSync(source, "binary");

  const destination = wizard.installPersistent({
    source,
    destinationDir,
    platform: "linux",
  });

  assert.equal(destination, path.join(destinationDir, "lingtong-cli"));
  assert.equal(fs.readFileSync(destination, "utf8"), "binary");
  assert.equal(fs.statSync(destination).mode & 0o777, 0o755);
  fs.rmSync(dir, { recursive: true, force: true });
});

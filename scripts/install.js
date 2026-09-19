#!/usr/bin/env node

const crypto = require("node:crypto");
const fs = require("node:fs");
const https = require("node:https");
const os = require("node:os");
const path = require("node:path");
const { execFileSync } = require("node:child_process");

const NAME = "lingtong-cli";
const VERSION = require("../package.json").version.replace(/^v/, "");
const DEFAULT_RELEASE_BASE = `https://github.com/usiboy/lingtong-cli/releases/download/v${VERSION}`;

function resolveTarget(platform, arch) {
  const platforms = { darwin: "darwin", linux: "linux", win32: "windows" };
  const architectures = { x64: "amd64", arm64: "arm64" };
  const targetOS = platforms[platform];
  const targetArch = architectures[arch];
  if (!targetOS) throw new Error(`Unsupported platform: ${platform}`);
  if (!targetArch) throw new Error(`Unsupported architecture: ${arch}`);
  return {
    os: targetOS,
    arch: targetArch,
    extension: platform === "win32" ? ".zip" : ".tar.gz",
    binaryExtension: platform === "win32" ? ".exe" : "",
  };
}

function archiveName(version, target) {
  return `${NAME}-${version.replace(/^v/, "")}-${target.os}-${target.arch}${target.extension}`;
}

function parseChecksum(content, artifact) {
  for (const line of content.split(/\r?\n/)) {
    const match = line.trim().match(/^([a-fA-F0-9]+)\s+\*?(.+)$/);
    if (match && match[2] === artifact) return match[1].toLowerCase();
  }
  throw new Error(`Checksum entry not found for ${artifact}`);
}

function verifyChecksum(file, expected) {
  const actual = crypto.createHash("sha256").update(fs.readFileSync(file)).digest("hex");
  if (actual !== expected.toLowerCase()) {
    throw new Error(`Checksum mismatch for ${path.basename(file)}: expected ${expected}, got ${actual}`);
  }
}

function download(url, destination, redirectsLeft = 3) {
  return new Promise((resolve, reject) => {
    const parsed = new URL(url);
    if (parsed.protocol !== "https:") {
      reject(new Error(`Refusing non-HTTPS download URL: ${url}`));
      return;
    }

    const request = https.get(parsed, { timeout: 120000 }, (response) => {
      if (response.statusCode >= 300 && response.statusCode < 400 && response.headers.location) {
        response.resume();
        if (redirectsLeft === 0) {
          reject(new Error(`Too many redirects while downloading ${url}`));
          return;
        }
        const redirected = new URL(response.headers.location, parsed).toString();
        download(redirected, destination, redirectsLeft - 1).then(resolve, reject);
        return;
      }
      if (response.statusCode !== 200) {
        response.resume();
        reject(new Error(`Download failed with HTTP ${response.statusCode}: ${url}`));
        return;
      }

      const output = fs.createWriteStream(destination, { mode: 0o600 });
      response.pipe(output);
      output.on("finish", () => output.close(resolve));
      output.on("error", reject);
    });
    request.on("timeout", () => request.destroy(new Error(`Download timed out: ${url}`)));
    request.on("error", reject);
  });
}

function extract(archive, destination, target) {
  if (target.os === "windows") {
    const command = [
      "$ErrorActionPreference='Stop';",
      `Expand-Archive -LiteralPath $env:LINGTONG_CLI_ARCHIVE -DestinationPath $env:LINGTONG_CLI_DEST -Force`,
    ].join("");
    execFileSync("powershell.exe", ["-NoProfile", "-ExecutionPolicy", "Bypass", "-Command", command], {
      stdio: "inherit",
      env: {
        ...process.env,
        LINGTONG_CLI_ARCHIVE: archive,
        LINGTONG_CLI_DEST: destination,
      },
    });
    return;
  }
  execFileSync("tar", ["-xzf", archive, "-C", destination], { stdio: "inherit" });
}

function releaseBase() {
  const configured = process.env.LINGTONG_CLI_DOWNLOAD_BASE;
  if (!configured) return DEFAULT_RELEASE_BASE;
  return configured.replace(/\{version\}/g, VERSION).replace(/\/$/, "");
}

async function installPackageBinary() {
  if (process.env.LINGTONG_CLI_SKIP_DOWNLOAD === "1") return null;

  const target = resolveTarget(process.platform, process.arch);
  const artifact = archiveName(VERSION, target);
  const packageRoot = path.join(__dirname, "..");
  const checksumFile = path.join(packageRoot, "checksums.txt");
  if (!fs.existsSync(checksumFile)) {
    throw new Error("checksums.txt is missing from the npm package; refusing an unverified download");
  }

  const expected = parseChecksum(fs.readFileSync(checksumFile, "utf8"), artifact);
  const tempDir = fs.mkdtempSync(path.join(os.tmpdir(), "lingtong-cli-"));
  const archive = path.join(tempDir, artifact);
  const binDir = path.join(packageRoot, "bin");
  const destination = path.join(binDir, NAME + target.binaryExtension);

  try {
    await download(`${releaseBase()}/${artifact}`, archive);
    verifyChecksum(archive, expected);
    extract(archive, tempDir, target);

    const extracted = path.join(tempDir, NAME + target.binaryExtension);
    if (!fs.existsSync(extracted)) {
      throw new Error(`Release archive does not contain ${path.basename(extracted)}`);
    }
    fs.mkdirSync(binDir, { recursive: true });
    fs.copyFileSync(extracted, destination);
    if (target.os !== "windows") fs.chmodSync(destination, 0o755);
    console.log(`${NAME} v${VERSION} downloaded for ${target.os}/${target.arch}`);
    return destination;
  } finally {
    fs.rmSync(tempDir, { recursive: true, force: true });
  }
}

if (require.main === module) {
  installPackageBinary().catch((error) => {
    console.error(`Failed to install ${NAME}: ${error.message}`);
    console.error("Set LINGTONG_CLI_DOWNLOAD_BASE to an HTTPS release mirror if GitHub is unavailable.");
    process.exit(1);
  });
}

module.exports = {
  archiveName,
  installPackageBinary,
  parseChecksum,
  releaseBase,
  resolveTarget,
  verifyChecksum,
};

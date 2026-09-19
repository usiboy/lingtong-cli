#!/usr/bin/env node

const fs = require("node:fs");
const os = require("node:os");
const path = require("node:path");
const { execFileSync } = require("node:child_process");

function defaultDestinationDir(platform = process.platform, home = os.homedir()) {
  if (platform === "win32") {
    return path.join(process.env.LOCALAPPDATA || path.join(home, "AppData", "Local"), "Lingtong", "bin");
  }
  return path.join(home, ".local", "bin");
}

function installPersistent({ source, destinationDir, platform = process.platform }) {
  const binaryName = "lingtong-cli" + (platform === "win32" ? ".exe" : "");
  const destination = path.join(destinationDir, binaryName);
  fs.mkdirSync(destinationDir, { recursive: true });
  fs.copyFileSync(source, destination);
  if (platform !== "win32") fs.chmodSync(destination, 0o755);
  return destination;
}

function pathContains(directory) {
  const separator = process.platform === "win32" ? ";" : ":";
  return (process.env.PATH || "").split(separator).includes(directory);
}

function ensurePackageBinary() {
  const extension = process.platform === "win32" ? ".exe" : "";
  const source = path.join(__dirname, "..", "bin", "lingtong-cli" + extension);
  if (!fs.existsSync(source)) {
    execFileSync(process.execPath, [path.join(__dirname, "install.js")], { stdio: "inherit" });
  }
  if (!fs.existsSync(source)) throw new Error("The downloaded lingtong-cli binary is missing");
  return source;
}

function main() {
  try {
    const source = ensurePackageBinary();
    const destinationDir = process.env.LINGTONG_CLI_INSTALL_DIR || defaultDestinationDir();
    const destination = installPersistent({ source, destinationDir });
    console.log(`Installed lingtong-cli to ${destination}`);
    if (!pathContains(destinationDir)) {
      console.log(`Add ${destinationDir} to PATH, then restart your shell.`);
    }
    console.log("Next: lingtong-cli config init && lingtong-cli auth login && lingtong-cli doctor");
  } catch (error) {
    console.error(`Failed to install lingtong-cli: ${error.message}`);
    process.exit(1);
  }
}

if (require.main === module) main();

module.exports = { defaultDestinationDir, installPersistent, main, pathContains };

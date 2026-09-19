#!/usr/bin/env node

const fs = require("node:fs");
const path = require("node:path");
const { spawnSync } = require("node:child_process");

const args = process.argv.slice(2);
if (args[0] === "install") {
  require("./install-wizard.js").main();
} else {
  const extension = process.platform === "win32" ? ".exe" : "";
  const binary = path.join(__dirname, "..", "bin", "lingtong-cli" + extension);
  if (!fs.existsSync(binary)) {
    const install = spawnSync(process.execPath, [path.join(__dirname, "install.js")], {
      stdio: "inherit",
      env: process.env,
    });
    if (install.status !== 0) process.exit(install.status || 1);
  }

  const result = spawnSync(binary, args, { stdio: "inherit" });
  if (result.error) {
    console.error(`Failed to run lingtong-cli: ${result.error.message}`);
    process.exit(1);
  }
  if (result.signal) {
    process.kill(process.pid, result.signal);
  } else {
    process.exit(result.status === null ? 1 : result.status);
  }
}

#!/usr/bin/env node
const fs = require("fs");
const path = require("path");
const { spawn } = require("child_process");

const { ensureInstalled } = require("./postinstall");
const { startUpdateCheck } = require("./update_check");

const exe = process.platform === "win32" ? "maskedemail-cli.exe" : "maskedemail-cli";
const binPath = path.join(__dirname, exe);

const PKG = (() => {
  try {
    return require("./package.json");
  } catch {
    return null;
  }
})();

(async function main() {
  try {
    if (!fs.existsSync(binPath)) {
      console.error(`Binary not found: ${binPath}`);
      console.error("Attempting to download it now...");
      await ensureInstalled();
    }

    if (!fs.existsSync(binPath)) {
      console.error(`Binary not found: ${binPath}`);
      process.exit(1);
    }

    if (process.platform !== "win32") {
      try {
        fs.chmodSync(binPath, 0o755);
      } catch {}
    }

    const child = spawn(binPath, process.argv.slice(2), {
      stdio: "inherit",
      windowsHide: true,
    });

    const updateAbort = new AbortController();
    startUpdateCheck({ installedVersion: PKG && PKG.version, signal: updateAbort.signal });

    child.on("exit", (code) => {
      try {
        updateAbort.abort();
      } catch {}
      process.exitCode = code == null ? 1 : code;
    });
    child.on("error", (err) => {
      console.error(err.message);
      process.exitCode = 1;
    });
  } catch (err) {
    console.error(err.message);
    process.exit(1);
  }
})();

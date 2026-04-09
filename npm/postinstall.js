#!/usr/bin/env node
const fs = require("fs");
const path = require("path");
const os = require("os");
const https = require("https");
const crypto = require("crypto");
const { spawnSync } = require("child_process");

const OWNER = "dvcrn";
const REPO = "maskedemail-cli";
const BIN = "maskedemail-cli";
const VERSION_ENV = "MASKEDEMAIL_CLI_VERSION";
const BASE_URL_ENV = "MASKEDEMAIL_CLI_BASE_URL";
const ARCH_ENV = "MASKEDEMAIL_CLI_ARCH";
const PLATFORM_ENV = "MASKEDEMAIL_CLI_PLATFORM";
const SKIP_POSTINSTALL_ENV = "MASKEDEMAIL_CLI_SKIP_POSTINSTALL";

function isTruthyEnv(name) {
  const value = process.env[name];
  return value === "1" || value === "true";
}

function httpGet(url, { headers } = {}) {
  return new Promise((resolve, reject) => {
    const req = https.get(url, { headers }, (res) => {
      if (res.statusCode && res.statusCode >= 300 && res.statusCode < 400 && res.headers.location) {
        resolve(httpGet(res.headers.location, { headers }));
        return;
      }
      if (res.statusCode !== 200) {
        reject(new Error(`GET ${url} -> ${res.statusCode}`));
        return;
      }
      const chunks = [];
      res.on("data", (c) => chunks.push(c));
      res.on("end", () => resolve(Buffer.concat(chunks)));
    });
    req.on("error", reject);
  });
}

function sha256(buf) {
  const h = crypto.createHash("sha256");
  h.update(buf);
  return h.digest("hex");
}

function parseChecksums(text) {
  const map = new Map();
  const lines = text
    .split(/\r?\n/)
    .map((l) => l.trim())
    .filter(Boolean);
  for (const line of lines) {
    let m = line.match(/^([a-f0-9]{64})\s+(.+)$/i);
    if (m) {
      map.set(m[2], m[1]);
      continue;
    }
    m = line.match(/^sha256:([a-f0-9]{64})\s+(.+)$/i);
    if (m) {
      map.set(m[2], m[1]);
      continue;
    }
    m = line.match(/^SHA256\s+\((.+)\)\s+=\s+([a-f0-9]{64})$/i);
    if (m) {
      map.set(m[1], m[2]);
      continue;
    }
  }
  return map;
}

(async function main() {
  try {
    if (isTruthyEnv(SKIP_POSTINSTALL_ENV)) {
      console.log(`postinstall: skipping binary download because ${SKIP_POSTINSTALL_ENV} is set`);
      return;
    }

    const platformRaw = process.env[PLATFORM_ENV] || process.platform;
    const platform = platformRaw === "win32" ? "windows" : platformRaw;
    if (!["darwin", "linux", "windows"].includes(platform)) {
      console.error("maskedemail-cli: npm install supports macOS (darwin), Linux, and Windows only");
      process.exit(1);
    }

    const pkg = JSON.parse(fs.readFileSync(path.join(__dirname, "package.json"), "utf8"));
    const version = process.env[VERSION_ENV] || pkg.version || "";
    if (!version) {
      console.error("postinstall: could not determine version");
      process.exit(1);
    }

    const detectedArch =
      process.arch === "x64"
        ? "amd64"
        : process.arch === "arm64"
          ? "arm64"
          : process.arch === "arm"
            ? "armv7"
            : process.arch;
    const arch = process.env[ARCH_ENV] || detectedArch;
    if (!["amd64", "arm64", "armv7"].includes(arch)) {
      console.error(`Unsupported arch: ${arch}`);
      process.exit(1);
    }
    if (platform === "windows" && arch === "armv7") {
      console.error(`Unsupported Windows arch: ${arch}`);
      process.exit(1);
    }

    const assetName = `${BIN}_${version}_${platform}_${arch}.tar.gz`;
    const baseOverride = process.env[BASE_URL_ENV];
    const bases = baseOverride
      ? [baseOverride]
      : [
          `https://github.com/${OWNER}/${REPO}/releases/download/${version}`,
          `https://github.com/${OWNER}/${REPO}/releases/download/v${version}`,
        ];
    const headers = { "User-Agent": `${REPO}-postinstall` };
    const outDir = __dirname;
    const exe = platform === "windows" ? `${BIN}.exe` : BIN;
    const binPath = path.join(outDir, exe);

    if (fs.existsSync(binPath)) {
      try {
        fs.chmodSync(binPath, 0o755);
      } catch {}
      return;
    }

    let tarGz = null;
    let baseUsed = "";
    let lastErr = null;
    for (const base of bases) {
      const url = `${base}/${assetName}`;
      console.log(`postinstall: downloading ${assetName} from ${url}`);
      try {
        tarGz = await httpGet(url, { headers });
        baseUsed = base;
        break;
      } catch (e) {
        lastErr = e;
      }
    }
    if (!tarGz) throw lastErr || new Error("failed to download binary");

    try {
      const checksumsUrl = `${baseUsed}/checksums.txt`;
      const checksumsBuf = await httpGet(checksumsUrl, { headers });
      const checksums = parseChecksums(checksumsBuf.toString("utf8"));
      const sumExpected = checksums.get(assetName);
      if (!sumExpected) throw new Error("asset not in checksums.txt");
      const sumActual = sha256(tarGz);
      if (sumActual.toLowerCase() !== sumExpected.toLowerCase()) throw new Error("checksum mismatch");
      console.log("postinstall: checksum OK");
    } catch (e) {
      console.warn(`postinstall: checksum skipped/failed: ${e.message}`);
    }

    const tmpFile = path.join(os.tmpdir(), `${REPO}-${Date.now()}.tar.gz`);
    fs.writeFileSync(tmpFile, tarGz);
    const tarRes = spawnSync("tar", ["-xzf", tmpFile, "-C", outDir, exe], { stdio: "inherit" });
    if (tarRes.status !== 0) {
      console.error("postinstall: failed to extract binary");
      process.exit(1);
    }
    try {
      fs.chmodSync(binPath, 0o755);
    } catch {}
    try {
      fs.unlinkSync(tmpFile);
    } catch {}
    console.log(`postinstall: installed ${exe} to ${outDir}`);
  } catch (err) {
    console.error(`postinstall error: ${err.message}`);
    process.exit(1);
  }
})();

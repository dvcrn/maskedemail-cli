const fs = require("fs");
const path = require("path");
const os = require("os");
const https = require("https");
const semver = require("semver");

const DEFAULT_PACKAGE_NAME = "maskedemail-cli";
const DEFAULT_UPDATE_COMMAND = "npm install -g maskedemail-cli";
const CACHE_DIR_NAME = "maskedemail-cli";

const UPDATE_CHECK_INTERVAL_MS = 24 * 60 * 60 * 1000;
const UPDATE_CHECK_TIMEOUT_MS = 1200;

function isDisabled() {
  return (
    process.env.MASKEDEMAIL_CLI_NO_UPDATE_NOTICE === "1" ||
    process.env.MASKEDEMAIL_CLI_NO_UPDATE_NOTICE === "true" ||
    process.env.NO_UPDATE_NOTIFIER === "1" ||
    process.env.NO_UPDATE_NOTIFIER === "true" ||
    process.env.CI === "1" ||
    process.env.CI === "true" ||
    process.env.CI === "yes"
  );
}

function getCacheFile() {
  const home = os.homedir();
  if (!home) return null;

  if (process.platform === "win32") {
    const base = process.env.LOCALAPPDATA || process.env.APPDATA;
    if (!base) return path.join(home, "AppData", "Local", CACHE_DIR_NAME, "update.json");
    return path.join(base, CACHE_DIR_NAME, "update.json");
  }

  if (process.platform === "darwin") {
    return path.join(home, "Library", "Caches", CACHE_DIR_NAME, "update.json");
  }

  const base = process.env.XDG_CACHE_HOME || path.join(home, ".cache");
  return path.join(base, CACHE_DIR_NAME, "update.json");
}

function readCache() {
  const file = getCacheFile();
  if (!file) return null;
  try {
    return JSON.parse(fs.readFileSync(file, "utf8"));
  } catch {
    return null;
  }
}

function writeCache(data) {
  const file = getCacheFile();
  if (!file) return;
  try {
    fs.mkdirSync(path.dirname(file), { recursive: true });
    fs.writeFileSync(file, JSON.stringify(data), "utf8");
  } catch {}
}

function fetchLatestVersion(packageName, timeoutMs, signal) {
  const url = `https://registry.npmjs.org/${encodeURIComponent(packageName)}/latest`;

  return new Promise((resolve, reject) => {
    const req = https.get(url, { headers: { "User-Agent": `${DEFAULT_PACKAGE_NAME}-update-check` } }, (res) => {
      if (res.statusCode !== 200) {
        reject(new Error(`GET ${url} -> ${res.statusCode}`));
        res.resume();
        return;
      }

      const chunks = [];
      res.on("data", (c) => chunks.push(c));
      res.on("end", () => {
        try {
          const json = JSON.parse(Buffer.concat(chunks).toString("utf8"));
          resolve(typeof json.version === "string" ? json.version : "");
        } catch (e) {
          reject(e);
        }
      });
    });

    req.on("error", reject);
    req.setTimeout(timeoutMs, () => req.destroy(new Error("timeout")));

    if (signal) {
      if (signal.aborted) {
        req.destroy(new Error("aborted"));
        return;
      }
      signal.addEventListener("abort", () => req.destroy(new Error("aborted")), { once: true });
    }
  });
}

function normalizeVersion(version) {
  if (!version) return "";
  const cleaned = semver.clean(version);
  if (cleaned) return cleaned;
  const coerced = semver.coerce(version);
  return coerced ? semver.valid(coerced) || "" : "";
}

function printNotice({ packageName, installed, latest, updateCommand }) {
  const prefix = `${packageName}:`;
  if (installed && latest) {
    console.error(`${prefix} an update is available (installed ${installed}, latest ${latest}). Update with: ${updateCommand}`);
    return;
  }
  console.error(`${prefix} an update is available. Update with: ${updateCommand}`);
}

async function runUpdateCheck({ packageName, installedVersion, updateCommand, signal }) {
  if (isDisabled()) return;
  if (!process.stderr.isTTY) return;

  const installed = normalizeVersion(installedVersion);
  if (!installed) return;

  const cache = readCache() || {};
  const lastChecked = Number.isFinite(cache.lastChecked) ? cache.lastChecked : 0;
  const lastNotified = Number.isFinite(cache.lastNotified) ? cache.lastNotified : 0;
  const cachedLatest = typeof cache.latest === "string" ? cache.latest : "";

  const now = Date.now();

  if (lastChecked && now - lastChecked < UPDATE_CHECK_INTERVAL_MS) {
    const latestCached = normalizeVersion(cachedLatest);
    if (latestCached && semver.gt(latestCached, installed)) {
      if (!lastNotified || now - lastNotified >= UPDATE_CHECK_INTERVAL_MS) {
        printNotice({ packageName, installed, latest: latestCached, updateCommand });
        writeCache({ ...cache, lastNotified: now });
      }
    }
    return;
  }

  const latestRaw = await fetchLatestVersion(packageName, UPDATE_CHECK_TIMEOUT_MS, signal);
  const latest = normalizeVersion(latestRaw);

  writeCache({ lastChecked: now, lastNotified, latest: latest || latestRaw });

  if (latest && semver.gt(latest, installed)) {
    if (!lastNotified || now - lastNotified >= UPDATE_CHECK_INTERVAL_MS) {
      printNotice({ packageName, installed, latest, updateCommand });
      writeCache({ lastChecked: now, lastNotified: now, latest });
    }
  }
}

function startUpdateCheck({
  packageName = DEFAULT_PACKAGE_NAME,
  installedVersion,
  updateCommand = DEFAULT_UPDATE_COMMAND,
  signal,
} = {}) {
  runUpdateCheck({ packageName, installedVersion, updateCommand, signal }).catch(() => {});
}

module.exports = {
  startUpdateCheck,
};

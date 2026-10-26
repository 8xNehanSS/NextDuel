#!/usr/bin/env node
// One-time local setup. Safe to re-run.
//
//   npm run setup
//
// 1. checks Go / Node / git are installed
// 2. clones the nex-server framework next to this repo if it is missing
// 3. creates nextduel-svr/.env from .env.example, choosing a free PORT
// 4. installs frontend dependencies and downloads Go modules

import { execSync, spawnSync } from "node:child_process";
import { copyFileSync, existsSync, readFileSync, writeFileSync } from "node:fs";
import net from "node:net";
import path from "node:path";
import { fileURLToPath } from "node:url";

const root = path.resolve(path.dirname(fileURLToPath(import.meta.url)), "..");
const svr = path.join(root, "nextduel-svr");
const web = path.join(root, "nextduel-frontend");
const framework = path.resolve(root, "..", "nex-server");
const frameworkRepo = "https://github.com/8xNehanSS/NEX-Server-Framework.git";

const step = (msg) => console.log(`\n\x1b[36m▸ ${msg}\x1b[0m`);
const ok = (msg) => console.log(`  \x1b[32m✓\x1b[0m ${msg}`);
const fail = (msg) => {
    console.error(`  \x1b[31m✗ ${msg}\x1b[0m`);
    process.exit(1);
};

function run(cmd, args, cwd) {
    // npm is a .cmd shim on Windows, so go through the shell with one string.
    const line = [cmd, ...args.map((a) => (/\s/.test(a) ? `"${a}"` : a))].join(" ");
    const res = spawnSync(line, { cwd, stdio: "inherit", shell: true });
    if (res.status !== 0) fail(`${line} failed`);
}

function version(cmd) {
    try {
        return execSync(cmd, { encoding: "utf8", stdio: ["ignore", "pipe", "ignore"] }).trim();
    } catch {
        return null;
    }
}

function portFree(port) {
    return new Promise((resolve) => {
        const srv = net.createServer();
        srv.once("error", () => resolve(false));
        srv.once("listening", () => srv.close(() => resolve(true)));
        srv.listen(port, "0.0.0.0");
    });
}

async function freePort(start) {
    for (let port = start; port < start + 50; port++) {
        if (await portFree(port)) return port;
    }
    fail(`no free port found between ${start} and ${start + 49}`);
}

// ---------------------------------------------------------------------------

step("Checking toolchain");
const go = version("go version");
if (!go) fail("Go is not installed (need 1.25+): https://go.dev/dl/");
const goMinor = Number(/go1\.(\d+)/.exec(go)?.[1] ?? 0);
if (goMinor < 25) fail(`${go} found, need Go 1.25+`);
ok(go);
const nodeMajor = Number(process.versions.node.split(".")[0]);
if (nodeMajor < 20) fail(`Node ${process.versions.node} found, need Node 20+ (22 recommended)`);
ok(`node v${process.versions.node}`);
if (!version("git --version")) fail("git is not installed");
ok(version("git --version"));

step("Checking nex-server framework");
if (existsSync(path.join(framework, "go.mod"))) {
    ok(`found at ${framework}`);
} else {
    // Private repo: this uses the developer's own git credentials.
    console.log(`  cloning ${frameworkRepo} -> ${framework} (private repo, needs read access)`);
    const res = spawnSync("git", ["clone", frameworkRepo, framework], { cwd: root, stdio: "inherit" });
    if (res.status !== 0) {
        fail(
            "could not clone nex-server. It is a private repository: make sure your GitHub account has " +
                "access and git is signed in (e.g. `git credential-manager github login`), or clone it " +
                `manually to ${framework}`,
        );
    }
    ok("cloned");
}

step("Configuring nextduel-svr/.env");
const envPath = path.join(svr, ".env");
if (!existsSync(envPath)) {
    copyFileSync(path.join(svr, ".env.example"), envPath);
    ok("created from .env.example");
} else {
    ok("already exists; existing values are kept");
}
let env = readFileSync(envPath, "utf8");

// Add settings introduced since the file was created, never overwriting.
const template = readFileSync(path.join(svr, ".env.example"), "utf8");
const missing = template
    .split(/\r?\n/)
    .filter((line) => /^[A-Z_]+=/.test(line))
    .filter((line) => !new RegExp(`^${line.split("=")[0]}=`, "m").test(env));
if (missing.length > 0) {
    env = `${env.replace(/\s*$/, "")}\n${missing.join("\n")}\n`;
    writeFileSync(envPath, env);
    ok(`added missing settings: ${missing.map((l) => l.split("=")[0]).join(", ")}`);
}
const configured = Number(/^PORT=(\d+)/m.exec(env)?.[1] ?? 8080);
if (await portFree(configured)) {
    ok(`PORT=${configured} is free`);
} else {
    const port = await freePort(configured + 1);
    env = /^PORT=/m.test(env) ? env.replace(/^PORT=.*$/m, `PORT=${port}`) : `PORT=${port}\n${env}`;
    writeFileSync(envPath, env);
    ok(`port ${configured} is in use, switched to PORT=${port}`);
}

step("Installing frontend dependencies");
run("npm", ["install"], web);
ok("npm install done");

step("Downloading Go modules");
run("go", ["mod", "download"], svr);
ok("go mod download done");

const port = /^PORT=(\d+)/m.exec(readFileSync(envPath, "utf8"))[1];
console.log(`
\x1b[32mSetup complete.\x1b[0m

  npm run dev     start backend (:${port}) + frontend with hot reload
                  -> open the "web" URL Vite prints (usually http://localhost:5173)
  npm start       build the frontend and serve everything at http://localhost:${port}
  npm test        backend tests + frontend lint/typecheck

Register an account in the UI. To try a duel alone, use a second browser
profile or a private window for the second player.
`);

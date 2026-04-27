/**
 * Regenerates `src/data/miroclaw-tools.json` from MiroClaw's `docs/tools.md`.
 *
 *   MIROCLAW_TOOLS_MD=/path/to/miroclaw/docs/tools.md bun scripts/extract-miroclaw-tools.mjs
 *   # or
 *   bun scripts/extract-miroclaw-tools.mjs /path/to/miroclaw/docs/tools.md
 */
import fs from "node:fs";
import path from "node:path";
import { fileURLToPath } from "node:url";

const here = path.dirname(fileURLToPath(import.meta.url));
const outPath = path.join(here, "../src/data/miroclaw-tools.json");
const fromArg = process.argv[2];
const fromEnv = process.env.MIROCLAW_TOOLS_MD;
const docPath = fromArg || fromEnv;
if (!docPath) {
  console.error("Pass path to tools.md as argv[1] or set MIROCLAW_TOOLS_MD.");
  process.exit(1);
}
if (!fs.existsSync(docPath)) {
  console.error("File not found:", docPath);
  process.exit(1);
}

const t = fs.readFileSync(docPath, "utf8");
const lines = t.split(/\n/);
const out = {
  source: "MiroClaw — tools reference",
  sourcePath: "miroclaw/docs/tools.md",
  tools: [],
};

let currentCat = "";
for (let i = 0; i < lines.length; i++) {
  const m = lines[i].match(/^##\s+([A-Z])\.\s+(.+)$/);
  if (m) {
    currentCat = `${m[1]}. ${m[2].trim()}`;
    continue;
  }
  const tm = lines[i].match(/^###\s+`([a-zA-Z0-9_]+)`\s*$/);
  if (tm) {
    const name = tm[1];
    let desc = "";
    for (let j = i + 1; j < Math.min(i + 20, lines.length); j++) {
      if (lines[j].startsWith("**Description:**")) {
        desc = lines[j].replace(/^\*\*Description:\*\*\s*/, "").trim();
        break;
      }
    }
    out.tools.push({
      name,
      description: desc,
      category: currentCat || undefined,
    });
  }
}

fs.mkdirSync(path.dirname(outPath), { recursive: true });
fs.writeFileSync(outPath, JSON.stringify(out, null, 2) + "\n", "utf8");
console.log(`Wrote ${out.tools.length} tools to ${outPath}`);

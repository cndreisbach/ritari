/**
 * Which knowledge base a call applies to.
 *
 * The CLI's own precedence ends at the registry default, which for most people
 * is their personal knowledge base. A globally installed package must not
 * silently write there, so it resolves the project itself and names it in every
 * result:
 *
 *   1. RITARI_PROJECT, when set;
 *   2. the nearest ritari.yaml, matched against the registry by path;
 *   3. the registry default — reported, never silent.
 *
 * A ritari.yaml that is not registered is an error rather than a fallback,
 * because falling back would write to a different knowledge base than the one
 * the directory declares.
 */

import { existsSync } from "node:fs";
import { dirname, resolve } from "node:path";

import { parseDocuments, RitariError, type Exec } from "./rit.ts";

const BUNDLE_FILE = "ritari.yaml";

/** The filesystem lookups resolution needs, injectable for tests. */
export interface ProjectFs {
	exists(path: string): boolean;
}

const nodeFs: ProjectFs = { exists: (path) => existsSync(path) };

/** The directory of the nearest ritari.yaml at or above cwd. */
export function findBundle(cwd: string, fs: ProjectFs = nodeFs): string | undefined {
	let dir = resolve(cwd);
	for (;;) {
		if (fs.exists(resolve(dir, BUNDLE_FILE))) return dir;
		const parent = dirname(dir);
		if (parent === dir) return undefined;
		dir = parent;
	}
}

interface Listing {
	default: string | undefined;
	projects: { name: string; path: string }[];
}

/** Read the registry through `rit project list`. */
export async function listProjects(exec: Exec, cwd: string): Promise<Listing> {
	const result = await exec("rit", ["--output", "json", "project", "list"], { cwd, timeout: 10_000 });
	const docs = parseDocuments(result.stdout);
	const failed = docs.find((d) => d.success === false);
	if (failed) throw new RitariError(typeof failed.error === "string" ? failed.error : "rit failed");
	if (result.code !== 0) {
		const stderr = result.stderr.trim();
		throw new RitariError(stderr !== "" ? stderr : `rit exited with status ${result.code}`);
	}
	const header = docs[0] ?? {};
	const projects = docs.slice(1).map((d) => ({ name: String(d.name), path: String(d.path) }));
	const name = typeof header.default === "string" && header.default !== "" ? header.default : undefined;
	return { default: name, projects };
}

function samePath(a: string, b: string): boolean {
	return resolve(a) === resolve(b);
}

/** Resolve the project name every call passes as `--project`. */
export async function resolveProject(
	exec: Exec,
	cwd: string,
	env: Record<string, string | undefined>,
	fs: ProjectFs = nodeFs,
): Promise<string> {
	const fromEnv = (env.RITARI_PROJECT ?? "").trim();
	if (fromEnv !== "") return fromEnv;

	const bundle = findBundle(cwd, fs);
	const listing = await listProjects(exec, cwd);

	if (bundle !== undefined) {
		const match = listing.projects.find((p) => samePath(p.path, bundle));
		if (match) return match.name;
		throw new RitariError(
			`the knowledge base at ${bundle} is not registered (run rit project add <name> ${bundle})`,
		);
	}
	if (listing.default !== undefined) return listing.default;
	throw new RitariError("no project specified and no default set (try: rit project init <name> <path>)");
}

/**
 * Session-start context.
 *
 * Once before the first turn, and again after a compaction, add the read-only
 * `rit brief` output (recent daily notes and open tasks) to the run. It never
 * writes, never summons a model, and stays quiet when rit is missing or the
 * project cannot be resolved: a broken brief must not break the session.
 */

import type { ExtensionAPI } from "@earendil-works/pi-coding-agent";

import { resolveProject } from "./project.ts";

export function registerBrief(pi: ExtensionAPI): void {
	const exec = pi.exec.bind(pi);
	let pending = true;

	pi.on("session_start", async () => {
		pending = true;
	});
	pi.on("session_compact", async () => {
		pending = true;
	});
	pi.on("before_agent_start", async (_event, ctx) => {
		if (!pending) return;
		pending = false;
		try {
			const project = await resolveProject(exec, ctx.cwd, process.env);
			const result = await exec("rit", ["--project", project, "--output", "text", "brief"], {
				cwd: ctx.cwd,
				timeout: 10_000,
			});
			if (result.code !== 0 || result.stdout.trim() === "") return;
			return {
				message: { customType: "ritari-brief", content: result.stdout, display: false },
			};
		} catch {
			return;
		}
	});
}

import { describe, expect, it } from "vitest";

import { findBundle, resolveProject } from "../src/project.ts";
import type { Exec } from "../src/rit.ts";

const listing = (projects: { name: string; path: string }[], fallback?: string): Exec => {
	return async (command, args) => {
		expect(command).toBe("rit");
		expect(args).toEqual(["--output", "json", "project", "list"]);
		const header = { success: true, count: projects.length, default: fallback };
		const docs = [header, ...projects.map((p) => ({ name: p.name, path: p.path, is_default: p.name === fallback }))];
		return { stdout: docs.map((d) => JSON.stringify(d)).join("\n") + "\n", stderr: "", code: 0, killed: false };
	};
};

describe("findBundle", () => {
	it("finds the nearest ritari.yaml at or above the directory", () => {
		const fs = { exists: (p: string) => p === "/repo/kb/ritari.yaml" };
		expect(findBundle("/repo/kb/ideas", fs)).toBe("/repo/kb");
		expect(findBundle("/repo/src", fs)).toBeUndefined();
		expect(findBundle("/elsewhere", fs)).toBeUndefined();
	});
});

describe("resolveProject", () => {
	it("prefers RITARI_PROJECT without asking the registry", async () => {
		const exec: Exec = async () => {
			throw new Error("the registry must not be consulted");
		};
		await expect(resolveProject(exec, "/repo", { RITARI_PROJECT: " work " })).resolves.toBe("work");
	});

	it("matches the nearest bundle to a registered project by path", async () => {
		const fs = { exists: (p: string) => p === "/repo/kb/ritari.yaml" };
		const exec = listing([{ name: "work", path: "/repo/kb" }], "personal");
		await expect(resolveProject(exec, "/repo/kb/ideas", {}, fs)).resolves.toBe("work");
	});

	it("refuses to fall back when the directory declares an unregistered bundle", async () => {
		const fs = { exists: (p: string) => p === "/repo/kb/ritari.yaml" };
		const exec = listing([{ name: "personal", path: "/home/me/kb" }], "personal");
		await expect(resolveProject(exec, "/repo/kb", {}, fs)).rejects.toThrow(
			"the knowledge base at /repo/kb is not registered (run rit project add <name> /repo/kb)",
		);
	});

	it("falls back to the registry default when nothing else resolves", async () => {
		const fs = { exists: () => false };
		const exec = listing([{ name: "personal", path: "/home/me/kb" }], "personal");
		await expect(resolveProject(exec, "/tmp", {}, fs)).resolves.toBe("personal");
	});

	it("reports the CLI's error when nothing is configured", async () => {
		const fs = { exists: () => false };
		const exec = listing([]);
		await expect(resolveProject(exec, "/tmp", {}, fs)).rejects.toThrow("no project specified");
	});
});

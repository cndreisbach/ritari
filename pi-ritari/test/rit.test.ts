import { describe, expect, it } from "vitest";

import { parseDocuments, RitariError, ritArgs, run, runText, type Exec } from "../src/rit.ts";

function recorder(result: { stdout?: string; stderr?: string; code?: number }) {
	const calls: { command: string; args: string[]; cwd?: string }[] = [];
	const exec: Exec = async (command, args, options) => {
		calls.push({ command, args, cwd: options?.cwd });
		return { stdout: result.stdout ?? "", stderr: result.stderr ?? "", code: result.code ?? 0, killed: false };
	};
	return { exec, calls };
}

describe("parseDocuments", () => {
	it("reads one flat object per line", () => {
		const docs = parseDocuments('{"success":true,"count":1}\n{"id":"ideas/a"}\n');
		expect(docs).toEqual([
			{ success: true, count: 1 },
			{ id: "ideas/a" },
		]);
	});

	it("keeps <, >, and & unescaped", () => {
		const docs = parseDocuments('{"content":"a <b> & c"}\n');
		expect(docs[0].content).toBe("a <b> & c");
	});

	it("rejects a line that is not an object", () => {
		expect(() => parseDocuments("not json\n")).toThrow(RitariError);
		expect(() => parseDocuments("[1,2]\n")).toThrow(RitariError);
	});
});

describe("ritArgs", () => {
	it("names the project and asks for json", () => {
		expect(ritArgs("work", ["search", "jwt"])).toEqual([
			"--project",
			"work",
			"--output",
			"json",
			"search",
			"jwt",
		]);
	});
});

describe("run", () => {
	it("returns the documents and passes the working directory", async () => {
		const { exec, calls } = recorder({ stdout: '{"success":true,"count":0}\n' });
		const docs = await run(exec, "/repo", "work", ["search", "jwt"]);
		expect(docs).toEqual([{ success: true, count: 0 }]);
		expect(calls[0].command).toBe("rit");
		expect(calls[0].cwd).toBe("/repo");
		expect(calls[0].args).toContain("--project");
	});

	it("turns a failure document into an error, even at exit 0", async () => {
		const { exec } = recorder({ stdout: '{"success":false,"error":"note not found: nope"}\n' });
		await expect(run(exec, "/repo", "work", ["note", "read", "nope"])).rejects.toThrow(
			"note not found: nope",
		);
	});

	it("falls back to stderr when there is no document", async () => {
		const { exec } = recorder({ stderr: "rit: bad flag\n", code: 1 });
		await expect(run(exec, "/repo", "work", ["nope"])).rejects.toThrow("rit: bad flag");
	});

	it("explains a rit that is too old to know --output json", async () => {
		const { exec } = recorder({ stderr: 'rit: unknown output format "json" (want text or yaml)\n', code: 2 });
		await expect(run(exec, "/repo", "work", ["ls"])).rejects.toThrow("too old for pi-ritari");
	});
});

describe("runText", () => {
	it("returns raw output", async () => {
		const { exec } = recorder({ stdout: "---\ntype: note\n---\n\nbody\n" });
		await expect(runText(exec, "/repo", "work", ["note", "view", "a"])).resolves.toContain("body");
	});

	it("raises the failure document's message", async () => {
		const { exec } = recorder({ stdout: '{"success":false,"error":"note not found: nope"}\n', code: 1 });
		await expect(runText(exec, "/repo", "work", ["note", "view", "nope"])).rejects.toThrow(
			"note not found: nope",
		);
	});
});

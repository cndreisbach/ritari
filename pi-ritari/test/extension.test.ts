import { describe, expect, it } from "vitest";

import extension from "../extensions/ritari.ts";

/** The smallest ExtensionAPI the extension touches. */
function fakePi(result = { stdout: "", stderr: "", code: 0, killed: false }) {
	const tools: string[] = [];
	const events: string[] = [];
	const calls: { args: string[]; options?: { cwd?: string; timeout?: number } }[] = [];
	const pi = {
		on(event: string) {
			events.push(event);
		},
		registerTool(tool: { name: string }) {
			tools.push(tool.name);
		},
		async exec(_command: string, args: string[], options?: { cwd?: string; timeout?: number }) {
			calls.push({ args, options });
			return result;
		},
	};
	return { pi, tools, events, calls };
}

describe("extension", () => {
	it("registers the four tools", () => {
		const { pi, tools } = fakePi();
		extension(pi as never);
		expect(tools).toEqual(["ritari_search", "ritari_read", "ritari_write", "ritari_edit"]);
	});

	it("briefs once per session, before the first turn", async () => {
		const { pi, events } = fakePi({ stdout: "No open tasks.\n", stderr: "", code: 0, killed: false });
		const handlers: Record<string, (event?: unknown, ctx?: unknown) => Promise<unknown>> = {};
		(pi as { on: (event: string, handler: typeof handlers[string]) => void }).on = (event, handler) => {
			events.push(event);
			handlers[event] = handler;
		};
		extension(pi as never);
		expect(events).toEqual(["session_start", "session_compact", "before_agent_start"]);

		process.env.RITARI_PROJECT = "work";
		const first = await handlers.before_agent_start({}, { cwd: "/repo" });
		const second = await handlers.before_agent_start({}, { cwd: "/repo" });
		delete process.env.RITARI_PROJECT;

		expect(first).toMatchObject({ message: { customType: "ritari-brief", display: false } });
		expect(second).toBeUndefined();
	});
});

/**
 * The rit client: run the CLI in machine mode and read its document stream.
 *
 * Machine mode is newline-delimited JSON, so there is no YAML parser here and
 * no second copy of the bundle format: the extension only shapes arguments and
 * results around what `rit` already does.
 */

/** One flat document from `rit --output json`. */
export type Doc = Record<string, unknown>;

/** The `pi.exec` signature the client needs, narrowed for tests. */
export type Exec = (
	command: string,
	args: string[],
	options?: { cwd?: string; timeout?: number; signal?: AbortSignal },
) => Promise<{ stdout: string; stderr: string; code: number; killed: boolean }>;

/** A failure `rit` reported, or one the client found in its output. */
export class RitariError extends Error {
	constructor(message: string) {
		super(message);
		this.name = "RitariError";
	}
}

/** Arguments every call carries: an explicit project (§9.1) and JSON output. */
export function ritArgs(project: string, words: string[]): string[] {
	return ["--project", project, "--output", "json", ...words];
}

/** Parse newline-delimited flat objects, the JSON form of the document stream. */
export function parseDocuments(stdout: string): Doc[] {
	const docs: Doc[] = [];
	for (const line of stdout.split("\n")) {
		const text = line.trim();
		if (text === "") continue;
		let parsed: unknown;
		try {
			parsed = JSON.parse(text);
		} catch {
			throw new RitariError(`rit printed a line that is not JSON: ${text.slice(0, 200)}`);
		}
		if (parsed === null || typeof parsed !== "object" || Array.isArray(parsed)) {
			throw new RitariError(`rit printed a document that is not an object: ${text.slice(0, 200)}`);
		}
		docs.push(parsed as Doc);
	}
	return docs;
}

function parseLenient(stdout: string): Doc[] {
	try {
		return parseDocuments(stdout);
	} catch {
		return [];
	}
}

function failureOf(docs: Doc[]): string | undefined {
	const failed = docs.find((d) => d.success === false);
	if (!failed) return undefined;
	return typeof failed.error === "string" && failed.error !== "" ? failed.error : "rit failed";
}

function exitMessage(result: { stderr: string; code: number }): string {
	const stderr = result.stderr.trim();
	return stderr !== "" ? stderr : `rit exited with status ${result.code}`;
}

/** Run a command that writes JSON documents and return them. */
export async function run(
	exec: Exec,
	cwd: string,
	project: string,
	words: string[],
	signal?: AbortSignal,
): Promise<Doc[]> {
	const result = await exec("rit", ritArgs(project, words), { cwd, timeout: 60_000, signal });
	const docs = parseLenient(result.stdout);
	const failure = failureOf(docs);
	if (failure !== undefined) throw new RitariError(failure);
	if (result.code !== 0) throw new RitariError(exitMessage(result));
	if (docs.length === 0) throw new RitariError("rit printed no output");
	return docs;
}

/** Run a command that prints raw bytes (`note view`, `read`). */
export async function runText(
	exec: Exec,
	cwd: string,
	project: string,
	words: string[],
	signal?: AbortSignal,
): Promise<string> {
	const result = await exec("rit", ritArgs(project, words), { cwd, timeout: 60_000, signal });
	if (result.code !== 0) {
		const failure = failureOf(parseLenient(result.stdout));
		throw new RitariError(failure ?? exitMessage(result));
	}
	return result.stdout;
}

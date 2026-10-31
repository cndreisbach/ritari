/**
 * pi-ritari: Ritari long-term memory for Pi.
 *
 * Registers four typed tools over the `rit` CLI and adds `rit brief` to the
 * start of a session. Requires `rit` on PATH; see the README.
 */

import type { ExtensionAPI } from "@earendil-works/pi-coding-agent";

import { registerBrief } from "../src/brief.ts";
import { createTools } from "../src/tools.ts";

export default function (pi: ExtensionAPI) {
	registerBrief(pi);
	for (const tool of createTools(pi.exec.bind(pi))) pi.registerTool(tool);
}

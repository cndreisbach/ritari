#!/usr/bin/env python3
#MISE description="Report how agents load the ritari skills and call rit, from local Pi and Claude Code transcripts"
"""Measure skill routing before rewriting the skills (ideas/skill-routing).

Reads agent transcripts on this machine and reports, per session, which
ritari skills were loaded (by the model or by the user), which rit commands
ran, and whether the user said something memory-shaped that went unanswered.
Nothing leaves the machine; the report holds counts, and the optional CSV
adds each session's id, date, and working-directory name.

    mise run skill-usage [--since 2026-10-01] [--project ritari] [--csv out.csv]

Transcript locations (override with --claude / --pi):
  Claude Code  $CLAUDE_CONFIG_DIR/projects or ~/.claude/projects  (*.jsonl)
  Pi           $PI_CODING_AGENT_DIR/sessions or ~/.pi/agent/sessions  (*.jsonl)

Both formats are parsed loosely: any content block shaped like a tool call
(type tool_use or toolCall, with a name and input or arguments) counts.
"""

import argparse
import collections
import csv
import json
import os
import re
import shlex
import signal
import sys
from pathlib import Path

SKILL_NAME = re.compile(r"^ritari(?:-[a-z]+)?$")
# A read of an installed skill file: .../skills/ritari-notes/SKILL.md or
# .../skills/ritari/references/x.md, but not the sources in internal/skills.
SKILL_PATH = re.compile(r"(?<!internal)[/\\]skills[/\\](ritari(?:-[a-z]+)?)[/\\](?:SKILL\.md|references[/\\]\S+)")
# A ritari entry in a skill listing: "- ritari: …", "- ritari (user): …",
# or <name>ritari</name> in an XML listing.
ADVERTISED = re.compile(r"(?m)^- ritari(?:-[a-z]+)?[ :(]|<name>ritari(?:-[a-z]+)?</name>")
# Things a user says when they want something remembered or recalled.
CUE = re.compile(
    r"\b(remember|don'?t forget|note (?:this|that)|write (?:this|that|it) down|"
    r"record (?:this|that)|save (?:this|that)|keep track|last time|"
    r"what do (?:we|you) know|recall|we decided|knowledge base)\b",
    re.IGNORECASE,
)
SLASH = re.compile(r"<command-name>/?(ritari(?:-[a-z]+)?)</command-name>|^/skill:(ritari(?:-[a-z]+)?)\b|<skill name=\"(ritari(?:-[a-z]+)?)\"")
REMINDER = re.compile(r"<system-reminder>.*?</system-reminder>", re.DOTALL)
# A heredoc body, from the << line to its delimiter; its lines aren't commands.
HEREDOC = re.compile(r"<<-?[ \t]*(['\"]?)(\w+)\1[^\n]*\n.*?\n[ \t]*\2[ \t]*(?=\n|$)", re.DOTALL)
# Tokens that separate simple commands.
OPERATORS = set(";&|()\n")
# rit subcommands that take a second word (rit note write, rit skills install).
GROUPS = {"note", "project", "skills", "config"}
# rit global flags that take a value, skipped when finding the subcommand.
VALUE_FLAGS = {"-p", "--project", "-o", "--output"}
WRAPPERS = {"env", "sudo", "command", "exec", "nohup", "time"}
# Shell commands that count as reading a skill file.
READERS = {"cat", "head", "tail", "sed", "less", "more", "bat", "nl"}

READ_TOOLS = {"read", "view", "read_file"}
SHELL_TOOLS = {"bash", "shell", "exec_command"}
# How many tool calls count as "the start of a session" for rit daily.
START_WINDOW = 10


class Session:
    def __init__(self, agent, sid, path):
        self.agent = agent
        self.id = sid
        self.path = path
        self.cwd = ""
        self.start = ""
        # "yes", "no" (a listing was logged without ritari), or "unknown"
        self.advertised = "unknown"
        # Ordered events: (timestamp, seq, kind, value)
        self.events = []

    def saw(self, text, listing=False):
        """Note whether text advertises the ritari skills. Only a known skill
        listing can show that they were not advertised."""
        if ADVERTISED.search(text):
            self.advertised = "yes"
        elif listing and self.advertised == "unknown":
            self.advertised = "no"

    def add(self, ts, kind, value):
        self.events.append((ts or "", len(self.events), kind, value))


def simple_commands(command):
    """Yield the words of each simple command in a shell command line, with
    env assignments and wrappers such as `mise exec --` removed. Operators
    inside quotes and heredoc bodies don't split commands."""
    for words in split_commands(HEREDOC.sub("", command)):
        while words and (re.match(r"^\w+=", words[0]) or words[0] in WRAPPERS):
            words.pop(0)
        if words[:3] == ["mise", "exec", "--"]:
            words = words[3:]
        if len(words) >= 2 and words[0] == "timeout":
            words = words[2:]
        if words:
            yield [os.path.basename(words[0])] + words[1:]


def split_commands(command):
    """Split a command line into word lists at unquoted ; & | ( ) and newlines."""
    lex = shlex.shlex(command, posix=True, punctuation_chars="".join(OPERATORS))
    lex.whitespace = " \t\r"
    lex.whitespace_split = True
    words = []
    try:
        for tok in lex:
            if tok and set(tok) <= OPERATORS:
                if words:
                    yield words
                words = []
            elif tok != "$":  # the $ of $(...)
                words.append(tok)
    except ValueError:  # unbalanced quotes: give up on the rest
        pass
    if words:
        yield words


def skill_reads(command):
    """Yield the skill name of each skill file a shell command reads."""
    for words in simple_commands(command):
        if words[0] in READERS:
            for w in words[1:]:
                m = SKILL_PATH.search(w)
                if m:
                    yield m.group(1)


def rit_commands(command):
    """Yield the rit subcommand of each rit invocation in a shell command."""
    for words in simple_commands(command):
        if words[0] != "rit":
            continue
        rest = words[1:]
        while rest and rest[0].startswith("-"):
            flag = rest.pop(0)
            if flag in VALUE_FLAGS and rest:
                rest.pop(0)
        if not rest:
            yield "(no subcommand)", []
            continue
        sub = rest[0]
        if sub in GROUPS and len(rest) > 1 and not rest[1].startswith("-"):
            yield f"{sub} {rest[1]}", rest[2:]
        else:
            yield sub, rest[1:]


def tool_calls(node):
    """Yield (name, args) for every tool-call-shaped dict under node."""
    if isinstance(node, dict):
        if node.get("type") in ("tool_use", "toolCall", "tool_call") and "name" in node:
            args = node.get("input", node.get("arguments", {}))
            if isinstance(args, str):
                try:
                    args = json.loads(args)
                except ValueError:
                    args = {"command": args}
            yield node["name"], args if isinstance(args, dict) else {}
            return
        for v in node.values():
            yield from tool_calls(v)
    elif isinstance(node, list):
        for v in node:
            yield from tool_calls(v)


def text_of(content):
    """Plain text of a message's content, ignoring tool results and images."""
    if isinstance(content, str):
        return content
    parts = []
    if isinstance(content, list):
        for block in content:
            if isinstance(block, dict) and block.get("type") == "text":
                parts.append(block.get("text", ""))
    return "\n".join(parts)


def record_tool(session, ts, name, args):
    lname = name.lower()
    if lname == "skill":
        skill = str(args.get("skill") or args.get("command") or "").lstrip("/")
        skill = skill.split(":")[-1]
        if SKILL_NAME.match(skill):
            session.add(ts, "load", (skill, "model"))
        session.add(ts, "tool", name)
        return
    if lname in READ_TOOLS:
        m = SKILL_PATH.search(str(args.get("file_path") or args.get("path") or ""))
        if m:
            session.add(ts, "load", (m.group(1), "model"))
    if lname in SHELL_TOOLS:
        command = str(args.get("command") or args.get("cmd") or "")
        for skill in skill_reads(command):
            session.add(ts, "load", (skill, "model"))
        for sub, rest in rit_commands(command):
            session.add(ts, "rit", (sub, rest))
    session.add(ts, "tool", name)


def record_user(session, ts, text):
    text = REMINDER.sub("", text).strip()
    if not text:
        return
    m = SLASH.search(text)
    if m:
        session.add(ts, "load", (next(g for g in m.groups() if g), "user"))
        return
    session.add(ts, "user", bool(CUE.search(text)))


def strings(node):
    """Every string under node, joined by newlines."""
    if isinstance(node, str):
        return node
    if isinstance(node, dict):
        node = list(node.values())
    if isinstance(node, list):
        return "\n".join(strings(v) for v in node)
    return ""


def read_lines(path):
    with open(path, encoding="utf-8", errors="replace") as f:
        for line in f:
            line = line.strip()
            if not line:
                continue
            try:
                yield json.loads(line)
            except ValueError:
                continue


def load_claude(root):
    """Claude Code: one JSON object per line, grouped by sessionId so that
    subagent transcripts count toward the session that started them."""
    sessions = {}
    for path in sorted(root.rglob("*.jsonl")):
        for obj in read_lines(path):
            if not isinstance(obj, dict):
                continue
            sid = obj.get("sessionId") or path.stem
            s = sessions.get(sid)
            if s is None:
                s = sessions[sid] = Session("claude-code", sid, path)
            ts = obj.get("timestamp", "")
            if ts and (not s.start or ts < s.start):
                s.start = ts
            if obj.get("cwd") and not s.cwd:
                s.cwd = obj["cwd"]
            kind = obj.get("type")
            msg = obj.get("message") or {}
            # Newer versions log the listing as an attachment; older ones
            # put it in a system reminder inside a user message.
            if kind == "attachment":
                att = obj.get("attachment") or {}
                s.saw(strings(att), listing=isinstance(att, dict) and att.get("type") == "skill_listing")
            elif kind == "user":
                s.saw(text_of(msg.get("content")))
            if kind == "assistant":
                for name, args in tool_calls(msg.get("content")):
                    record_tool(s, ts, name, args)
            elif kind == "user" and not obj.get("isCompactSummary"):
                if obj.get("isMeta"):
                    continue
                record_user(s, ts, text_of(msg.get("content")))
    return sessions.values()


def load_pi(root):
    """Pi: one session per file, a session header and then message entries."""
    sessions = []
    for path in sorted(root.rglob("*.jsonl")):
        s = Session("pi", path.stem, path)
        for obj in read_lines(path):
            if not isinstance(obj, dict):
                continue
            ts = obj.get("timestamp", "")
            if isinstance(ts, (int, float)):
                ts = str(ts)
            if obj.get("type") == "session":
                s.id = obj.get("id", s.id)
                s.cwd = obj.get("cwd", "")
                s.start = ts
                continue
            msg = obj.get("message") or {}
            role = msg.get("role")
            if role == "user":
                s.saw(text_of(msg.get("content")))
            if role == "assistant":
                for name, args in tool_calls(msg.get("content")):
                    record_tool(s, ts, name, args)
            elif role == "user":
                record_user(s, ts, text_of(msg.get("content")))
        if not s.start and s.events:
            s.start = s.events[0][0]
        sessions.append(s)
    return sessions


def summarize(s):
    """Reduce a session's events to the facts the report counts."""
    events = sorted(s.events)
    loads = collections.Counter()
    loaders = set()
    rit = collections.Counter()
    tools = 0
    first_load = first_rit = None
    read_daily_early = wrote_daily = False
    cues = answered = 0
    pending = False  # a cue whose turn hasn't produced rit or a load yet

    for _, _, kind, value in events:
        if kind == "tool":
            tools += 1
        elif kind == "user":
            pending = value
            cues += value
        elif kind == "load":
            skill, who = value
            loads[skill] += 1
            loaders.add(who)
            if first_load is None:
                first_load = tools
            if pending:
                answered += 1
                pending = False
        elif kind == "rit":
            sub, rest = value
            rit[sub] += 1
            if first_rit is None:
                first_rit = tools
            if sub == "daily":
                writes = any(a.split("=")[0] in ("--note", "--obs") for a in rest)
                wrote_daily |= writes
                read_daily_early |= not writes and tools < START_WINDOW
            if pending:
                answered += 1
                pending = False

    if loads and rit:
        outcome = "loaded+called"
    elif loads:
        outcome = "loaded-only"
    elif rit:
        outcome = "called-only"
    else:
        outcome = "neither"
    return {
        "agent": s.agent,
        "session": s.id,
        "date": s.start[:10],
        "project": os.path.basename(s.cwd.rstrip("/\\")) if s.cwd else "",
        "advertised": s.advertised,
        "outcome": outcome,
        "tool_calls": tools,
        "skills": " ".join(f"{k}:{v}" for k, v in sorted(loads.items())),
        "loaded_by": "+".join(sorted(loaders)),
        "rit_calls": sum(rit.values()),
        "rit_commands": " ".join(f"{k}:{v}" for k, v in sorted(rit.items())),
        "first_load_at": "" if first_load is None else first_load,
        "first_rit_at": "" if first_rit is None else first_rit,
        "read_daily_at_start": read_daily_early,
        "wrote_daily": wrote_daily,
        "cues": cues,
        "cues_answered": answered,
        "_loads": loads,
        "_rit": rit,
    }


def pct(n, d):
    return f"{100 * n / d:.0f}%" if d else "-"


def report(rows, out):
    by_agent = collections.defaultdict(list)
    for r in rows:
        by_agent[r["agent"]].append(r)
    for agent, rs in sorted(by_agent.items()):
        n = len(rs)
        print(f"## {agent}: {n} sessions", file=out)
        adv = collections.Counter(r["advertised"] for r in rs)
        print(f"ritari skills advertised: yes {adv['yes']}, no {adv['no']}, unknown {adv['unknown']} (unknown: no skill listing in the transcript)", file=out)
        print("\noutcome (skill loaded? rit called?)", file=out)
        outcomes = collections.Counter(r["outcome"] for r in rs)
        for k in ("loaded+called", "loaded-only", "called-only", "neither"):
            print(f"  {k:<15}{outcomes[k]:>6}  {pct(outcomes[k], n):>5}", file=out)

        active = [r for r in rs if r["outcome"] != "neither"]
        print(f"\nper skill, across the {len(active)} sessions that used ritari at all", file=out)
        print(f"  {'skill':<16}{'sessions':>9}{'loads':>7}", file=out)
        skills = collections.Counter()
        sessions_with = collections.Counter()
        for r in rs:
            skills.update(r["_loads"])
            sessions_with.update(r["_loads"].keys())
        for k in sorted(skills):
            print(f"  {k:<16}{sessions_with[k]:>9}{skills[k]:>7}", file=out)
        user = sum("user" in r["loaded_by"] for r in rs)
        model = sum("model" in r["loaded_by"] for r in rs)
        print(f"  sessions with a model-routed load: {model}; with a user /command: {user}", file=out)

        print("\nrit commands", file=out)
        rit = collections.Counter()
        for r in rs:
            rit.update(r["_rit"])
        for k, v in rit.most_common():
            print(f"  {k:<16}{v:>7}", file=out)

        callers = [r for r in rs if r["rit_calls"]]
        both = [r for r in callers if r["first_load_at"] != ""]
        before = sum(r["first_rit_at"] < r["first_load_at"] for r in both)
        print("\nsession continuity (sessions that called rit)", file=out)
        print(f"  read rit daily in first {START_WINDOW} tool calls: {sum(r['read_daily_at_start'] for r in callers)} of {len(callers)}", file=out)
        print(f"  wrote to the daily note:              {sum(r['wrote_daily'] for r in callers)} of {len(callers)}", file=out)
        print(f"  called rit before loading any skill:  {before} of {len(both)}", file=out)

        cued = [r for r in rs if r["advertised"] != "no"]
        cues = sum(r["cues"] for r in cued)
        answered = sum(r["cues_answered"] for r in cued)
        missed = sum(1 for r in rs if r["cues"] and r["outcome"] == "neither" and r["advertised"] != "no")
        print("\nmemory cues in user messages (sessions where ritari wasn't known to be unadvertised)", file=out)
        print(f"  cues: {cues}; answered in the same turn by a load or rit call: {answered} ({pct(answered, cues)})", file=out)
        print(f"  sessions with a cue but no load and no rit call: {missed}", file=out)
        print(file=out)


def main():
    if hasattr(signal, "SIGPIPE"):
        signal.signal(signal.SIGPIPE, signal.SIG_DFL)  # quiet under `| head`
    home = Path.home()
    claude_default = Path(os.environ.get("CLAUDE_CONFIG_DIR", home / ".claude")) / "projects"
    pi_default = Path(os.environ.get("PI_CODING_AGENT_DIR", home / ".pi" / "agent")) / "sessions"
    p = argparse.ArgumentParser(description=__doc__.split("\n\n")[0])
    p.add_argument("--claude", type=Path, default=claude_default, help=f"Claude Code projects dir (default {claude_default})")
    p.add_argument("--pi", type=Path, default=pi_default, help=f"Pi sessions dir (default {pi_default})")
    p.add_argument("--since", help="only sessions starting on or after YYYY-MM-DD")
    p.add_argument("--project", help="only sessions whose working directory contains this text")
    p.add_argument("--skip-project", action="append", default=[], metavar="TEXT", help="leave out sessions whose working directory contains this text (repeatable)")
    p.add_argument("--csv", type=Path, help="also write one row per session to this file")
    a = p.parse_args()

    sessions = []
    for root, loader in ((a.claude, load_claude), (a.pi, load_pi)):
        if root.is_dir():
            sessions.extend(loader(root))
        else:
            print(f"skipping {root}: not a directory", file=sys.stderr)

    rows = []
    for s in sessions:
        if a.since and s.start[:10] < a.since:
            continue
        if a.project and a.project not in s.cwd:
            continue
        if any(t in s.cwd for t in a.skip_project):
            continue
        if not s.events:
            continue
        rows.append(summarize(s))
    if not rows:
        print("no sessions found", file=sys.stderr)
        return 1

    report(rows, sys.stdout)
    if a.csv:
        fields = [k for k in rows[0] if not k.startswith("_")]
        with open(a.csv, "w", newline="", encoding="utf-8") as f:
            w = csv.DictWriter(f, fieldnames=fields, extrasaction="ignore")
            w.writeheader()
            w.writerows(sorted(rows, key=lambda r: (r["agent"], r["date"])))
        print(f"wrote {len(rows)} sessions to {a.csv}", file=sys.stderr)
    return 0


if __name__ == "__main__":
    sys.exit(main())

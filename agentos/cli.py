import argparse
import json
import os
import sys
from pathlib import Path
from typing import Optional

from agentos.runtime.runner import run_agent
from agentos.runtime.loader import (
    load_agent_manifest,
    ensure_agent_skeleton,
    resolve_agent_ref,
)
from agentos.runtime.packaging import build_package, install_package, inspect_package


REPO_ROOT = Path(__file__).resolve().parents[1]
DEFAULT_AGENTS_DIR = REPO_ROOT / "agents"
INSTALLED_AGENTS_DIR = REPO_ROOT / "installed_agents"
RUNS_DIR = REPO_ROOT / "runs"


def _resolve_agent_dir(name: str) -> Optional[Path]:
    # Priority: installed agents → built-in agents
    for base in (INSTALLED_AGENTS_DIR, DEFAULT_AGENTS_DIR):
        candidate = base / name
        if candidate.exists():
            return candidate
    # Also allow direct path
    p = Path(name)
    if p.exists() and p.is_dir():
        return p
    return None


def cmd_run(args: argparse.Namespace) -> int:
    # Accept: name (installed/built-in), directory path, or .agent file
    ref = args.agent
    pkg = resolve_agent_ref(ref, DEFAULT_AGENTS_DIR, INSTALLED_AGENTS_DIR)
    if pkg.kind == "missing":
        print(f"Agent '{ref}' not found. Try 'agent install {ref}'.", file=sys.stderr)
        return 2
    agent_dir, manifest = pkg.materialize()
    user_input = args.input
    return run_agent(agent_dir, manifest, user_input=user_input, runs_dir=RUNS_DIR)


def cmd_inspect(args: argparse.Namespace) -> int:
    ref = args.agent
    if ref.endswith(".agent") and Path(ref).exists():
        info = inspect_package(Path(ref))
        print(json.dumps(info, indent=2))
        return 0
    agent_dir = _resolve_agent_dir(ref)
    if not agent_dir:
        print(f"Agent '{ref}' not found.", file=sys.stderr)
        return 2
    manifest = load_agent_manifest(agent_dir)
    print(json.dumps({"source": str(agent_dir), "manifest": manifest}, indent=2))
    return 0


def cmd_install(args: argparse.Namespace) -> int:
    source = args.name
    if source is None:
        # List built-ins available to install
        print("Built-in agents available:")
        if DEFAULT_AGENTS_DIR.exists():
            for p in sorted(DEFAULT_AGENTS_DIR.iterdir()):
                if (p / "agent.yaml").exists():
                    print(f"- {p.name}")
        return 0
    dest = install_package(source, DEFAULT_AGENTS_DIR, INSTALLED_AGENTS_DIR)
    print(f"Installed to {dest}")
    return 0


def cmd_publish(args: argparse.Namespace) -> int:
    # Prototype stub: would package to .agent and push to registry
    print("Publish is a stub in this prototype. Use built-in agents for now.")
    return 0


def cmd_init(args: argparse.Namespace) -> int:
    target = INSTALLED_AGENTS_DIR / args.name
    ensure_agent_skeleton(target)
    print(f"Initialized agent skeleton at {target}")
    return 0


def cmd_compose(args: argparse.Namespace) -> int:
    print("Compose is a stub in this prototype (multi-agent workflows).")
    return 0


def cmd_build(args: argparse.Namespace) -> int:
    path = Path(args.path) if args.path else _resolve_agent_dir(args.agent) or Path(".")
    if not path.exists() or not (path / "agent.yaml").exists():
        print("No agent found at target. Provide --path or agent name.", file=sys.stderr)
        return 2
    out = build_package(path)
    print(str(out))
    return 0


def cmd_logs(args: argparse.Namespace) -> int:
    # List or tail logs for an agent
    agent = args.agent
    from agentos.runtime.logs import list_runs, tail_run

    runs = list_runs(RUNS_DIR, agent_name=agent, limit=args.limit)
    if not runs:
        print("No runs found.")
        return 0
    if args.tail:
        tail_run(RUNS_DIR / runs[0])
    else:
        for r in runs:
            print(r)
    return 0


def build_parser() -> argparse.ArgumentParser:
    parser = argparse.ArgumentParser(prog="agent", description="AgentOS CLI (prototype)")
    sub = parser.add_subparsers(dest="cmd", required=True)

    p_run = sub.add_parser("run", help="Run an agent")
    p_run.add_argument("agent", help="Agent name")
    p_run.add_argument("--input", help="Optional input for the agent", default=None)
    p_run.set_defaults(func=cmd_run)

    p_inspect = sub.add_parser("inspect", help="Inspect agent manifest")
    p_inspect.add_argument("agent", help="Agent name")
    p_inspect.set_defaults(func=cmd_inspect)

    p_install = sub.add_parser("install", help="Install an agent from registry or path")
    p_install.add_argument("name", help="Agent name or source", nargs="?")
    p_install.set_defaults(func=cmd_install)

    p_publish = sub.add_parser("publish", help="Publish an agent package")
    p_publish.add_argument("path", nargs="?", help="Path to agent directory")
    p_publish.set_defaults(func=cmd_publish)

    p_init = sub.add_parser("init", help="Create a new agent skeleton")
    p_init.add_argument("name", help="Agent name")
    p_init.set_defaults(func=cmd_init)

    p_compose = sub.add_parser("compose", help="Compose multiple agents (stub)")
    p_compose.set_defaults(func=cmd_compose)

    p_build = sub.add_parser("build", help="Package an agent directory into a .agent artifact")
    p_build.add_argument("agent", nargs="?", help="Agent name (installed/built-in)")
    p_build.add_argument("--path", help="Path to agent directory")
    p_build.set_defaults(func=cmd_build)

    p_logs = sub.add_parser("logs", help="Show or tail logs for an agent")
    p_logs.add_argument("agent", help="Agent name to filter runs")
    p_logs.add_argument("--limit", type=int, default=5)
    p_logs.add_argument("--tail", action="store_true")
    p_logs.set_defaults(func=cmd_logs)

    return parser


def main(argv: Optional[list[str]] = None) -> int:
    parser = build_parser()
    args = parser.parse_args(argv)
    return args.func(args)

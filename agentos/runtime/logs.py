from __future__ import annotations

import json
from pathlib import Path
from typing import List, Optional


def list_runs(runs_dir: Path, agent_name: Optional[str] = None, limit: int = 5) -> List[str]:
    runs = []
    if not runs_dir.exists():
        return runs
    # sort dirs by mtime desc
    dirs = sorted([p for p in runs_dir.iterdir() if p.is_dir()], key=lambda p: p.stat().st_mtime, reverse=True)
    for d in dirs:
        log = d / "events.jsonl"
        if not log.exists():
            continue
        if agent_name:
            try:
                first = next(log.open("r", encoding="utf-8"))
                rec = json.loads(first)
                if rec.get("data", {}).get("agent") != agent_name:
                    continue
            except Exception:
                continue
        runs.append(d.name)
        if len(runs) >= limit:
            break
    return runs


def tail_run(run_dir: Path, lines: int = 100) -> None:
    p = run_dir
    if p.is_dir():
        log = p / "events.jsonl"
    else:
        log = Path(p)
    if not log.exists():
        print("No log file found for run")
        return
    entries = log.read_text(encoding="utf-8").splitlines()[-lines:]
    for e in entries:
        print(e)


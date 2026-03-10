from __future__ import annotations
import json
import tarfile
from dataclasses import dataclass
from pathlib import Path
from typing import Any, Dict, Optional, Tuple


def _read_text(path: Path) -> str:
    with path.open("r", encoding="utf-8") as f:
        return f.read()


def load_agent_manifest(agent_dir: Path) -> Dict[str, Any]:
    # For this prototype, agent.yaml contains JSON content (YAML-compatible) to avoid extra deps.
    manifest_path = agent_dir / "agent.yaml"
    if not manifest_path.exists():
        raise FileNotFoundError(f"agent.yaml not found in {agent_dir}")
    content = _read_text(manifest_path)
    try:
        return json.loads(content)
    except json.JSONDecodeError as e:
        raise RuntimeError(
            f"Failed to parse agent.yaml as JSON (YAML superset). Error: {e}"
        )


@dataclass
class ResolvedAgent:
    kind: str  # 'dir', 'artifact', 'missing'
    source: Path
    builtin_dir: Path
    installed_dir: Path

    def materialize(self) -> Tuple[Path, Dict[str, Any]]:
        if self.kind == "dir":
            return self.source, load_agent_manifest(self.source)
        if self.kind == "artifact":
            # extract to installed/<name>@temp-run-<id> to run; for simplicity, extract to installed/<name>
            from .packaging import install_package

            dest = install_package(str(self.source), self.builtin_dir, self.installed_dir)
            return dest, load_agent_manifest(dest)
        raise FileNotFoundError("Agent not found")


def resolve_agent_ref(ref: str, builtin_dir: Path, installed_dir: Path) -> ResolvedAgent:
    p = Path(ref)
    if p.exists():
        if p.is_dir():
            return ResolvedAgent("dir", p, builtin_dir, installed_dir)
        if p.is_file() and p.suffix == ".agent":
            return ResolvedAgent("artifact", p, builtin_dir, installed_dir)
    # lookup by name
    for base in (installed_dir, builtin_dir):
        candidate = base / ref
        if (candidate / "agent.yaml").exists():
            return ResolvedAgent("dir", candidate, builtin_dir, installed_dir)
    return ResolvedAgent("missing", p, builtin_dir, installed_dir)


def ensure_agent_skeleton(target_dir: Path) -> None:
    target_dir.mkdir(parents=True, exist_ok=True)
    (target_dir / "prompts").mkdir(exist_ok=True)
    (target_dir / "tools").mkdir(exist_ok=True)
    (target_dir / "workflows").mkdir(exist_ok=True)
    (target_dir / "agent.yaml").write_text(
        json.dumps(
            {
                "name": target_dir.name,
                "version": "0.1.0",
                "description": "Example agent",
                "entrypoints": {"run": {"type": "builtin"}},
                "defaults": {"input": "Hello world"},
                "tools": [],
                "models": {"default": {"provider": "mock", "model_id": "mock-001"}},
                "memory": {"type": "jsonl"},
            },
            indent=2,
        )
        + "\n",
        encoding="utf-8",
    )

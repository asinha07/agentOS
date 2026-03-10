from __future__ import annotations

import io
import json
import tarfile
from dataclasses import dataclass
from pathlib import Path
from typing import Any, Dict, Optional, Tuple

from .loader import load_agent_manifest


DIST_DIRNAME = "dist"


def build_package(agent_dir: Path) -> Path:
    agent_dir = agent_dir.resolve()
    manifest = load_agent_manifest(agent_dir)
    name = manifest.get("name", agent_dir.name)
    version = manifest.get("version", "0.0.0")
    dist_dir = agent_dir / DIST_DIRNAME
    dist_dir.mkdir(exist_ok=True)
    out = dist_dir / f"{name}-{version}.agent"

    with tarfile.open(out, "w:gz") as tar:
        # Include all files in agent_dir except dist/
        for p in agent_dir.rglob("*"):
            if DIST_DIRNAME in p.parts:
                continue
            arcname = p.relative_to(agent_dir)
            tar.add(p, arcname=str(arcname))
    return out


def inspect_package(artifact: Path) -> Dict[str, Any]:
    with tarfile.open(artifact, "r:gz") as tar:
        try:
            f = tar.extractfile("agent.yaml")
            assert f is not None
            content = f.read().decode("utf-8")
            # Manifest uses JSON-in-YAML for zero-deps
            data = json.loads(content)
        except Exception as e:
            data = {"error": f"failed to read agent.yaml: {e}"}
        members = [m.name for m in tar.getmembers()]
    return {"artifact": str(artifact), "manifest": data, "files": members}


def install_package(source: str, builtin_dir: Path, installed_dir: Path) -> Path:
    p = Path(source)
    installed_dir.mkdir(parents=True, exist_ok=True)
    if p.exists():
        if p.is_file() and p.suffix == ".agent":
            with tarfile.open(p, "r:gz") as tar:
                # read manifest to get name
                f = tar.extractfile("agent.yaml")
                assert f is not None
                manifest = json.loads(f.read().decode("utf-8"))
                name = manifest.get("name", p.stem)
                dest = installed_dir / name
                dest.mkdir(parents=True, exist_ok=True)
                tar.extractall(dest)
                return dest
        elif p.is_dir():
            # copy directory contents
            import shutil

            name = p.name
            dest = installed_dir / name
            if dest.exists():
                shutil.rmtree(dest)
            shutil.copytree(p, dest)
            return dest
        else:
            raise FileNotFoundError(f"Unsupported install source: {source}")
    # Fallback: install from built-ins by name
    b = builtin_dir / source
    if (b / "agent.yaml").exists():
        import shutil

        dest = installed_dir / source
        if dest.exists():
            shutil.rmtree(dest)
        shutil.copytree(b, dest)
        return dest
    raise FileNotFoundError(f"Agent '{source}' not found.")


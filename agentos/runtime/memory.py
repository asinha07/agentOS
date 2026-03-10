import json
import time
from pathlib import Path
from typing import Any, Dict


class RunMemory:
    def __init__(self, runs_dir: Path, run_id: str):
        self.runs_dir = runs_dir
        self.run_id = run_id
        self.ns_dir = runs_dir / run_id
        self.ns_dir.mkdir(parents=True, exist_ok=True)
        self.log_path = self.ns_dir / "events.jsonl"
        self.kv_path = self.ns_dir / "kv.json"

    def append_event(self, kind: str, payload: Dict[str, Any]) -> None:
        rec = {
            "ts": time.time(),
            "kind": kind,
            "data": payload,
        }
        with self.log_path.open("a", encoding="utf-8") as f:
            f.write(json.dumps(rec) + "\n")

    def path(self) -> Path:
        return self.ns_dir

    # Key-Value memory for the run
    def write(self, key: str, value: Any) -> None:
        data = {}
        if self.kv_path.exists():
            try:
                data = json.loads(self.kv_path.read_text(encoding="utf-8"))
            except Exception:
                data = {}
        data[key] = value
        self.kv_path.write_text(json.dumps(data, indent=2) + "\n", encoding="utf-8")

    def read(self, key: str) -> Any:
        if not self.kv_path.exists():
            return None
        try:
            data = json.loads(self.kv_path.read_text(encoding="utf-8"))
            return data.get(key)
        except Exception:
            return None

    def query(self, prefix: str) -> Dict[str, Any]:
        if not self.kv_path.exists():
            return {}
        try:
            data = json.loads(self.kv_path.read_text(encoding="utf-8"))
        except Exception:
            return {}
        return {k: v for k, v in data.items() if k.startswith(prefix)}

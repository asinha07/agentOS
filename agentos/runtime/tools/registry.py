from __future__ import annotations

from typing import Any, Callable, Dict

from .builtin.idea_generator import invoke as idea_generator
from .builtin.go_to_market import invoke as go_to_market
from .builtin.risk_analyzer import invoke as risk_analyzer


_TOOLS: Dict[str, Callable[[Dict[str, Any]], Dict[str, Any]]] = {
    "idea_generator": idea_generator,
    "go_to_market": go_to_market,
    "risk_analyzer": risk_analyzer,
}


def get_tool(name: str) -> Callable[[Dict[str, Any]], Dict[str, Any]]:
    if name not in _TOOLS:
        raise KeyError(f"Unknown tool: {name}")
    return _TOOLS[name]


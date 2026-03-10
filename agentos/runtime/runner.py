from __future__ import annotations

import datetime as _dt
import textwrap
from pathlib import Path
from typing import Any, Dict, Optional

from .memory import RunMemory
from .model_adapters import MockModel
from .tools import get_tool


def _now_id() -> str:
    return _dt.datetime.utcnow().strftime("%Y%m%d-%H%M%S-%fZ")


def run_agent(agent_dir: Path, manifest: Dict[str, Any], user_input: Optional[str], runs_dir: Path) -> int:
    run_id = _now_id()
    memory = RunMemory(runs_dir=runs_dir, run_id=run_id)

    name = manifest.get("name", agent_dir.name)
    version = manifest.get("version", "0.0.0")
    defaults = manifest.get("defaults", {})

    topic = user_input or defaults.get("input", "AI for everyday life")

    memory.append_event("start", {"agent": name, "version": version, "topic": topic})

    tools_decl = [t.get("name") if isinstance(t, dict) else str(t) for t in manifest.get("tools", [])]

    # Step 1: Optional ideation if idea tool declared
    ideas = []
    best_idea = topic
    if "idea_generator" in tools_decl:
        idea_tool = get_tool("idea_generator")
        ideas_res = idea_tool({"topic": topic, "count": 3})
        memory.append_event("tool.idea_generator", ideas_res)
        ideas = ideas_res.get("ideas", [])
        best_idea = ideas[0] if ideas else topic

    # Step 2: Model - main reasoning
    # Load model config (support both 'model' and legacy 'models.default')
    model_conf = manifest.get("model") or manifest.get("models", {}).get("default", {})
    provider = model_conf.get("provider", "mock")
    model_id = model_conf.get("model") or model_conf.get("model_id") or "mock-001"

    # For prototype, only mock is implemented
    model = MockModel(provider=provider, model_id=model_id)

    # System prompt from prompt.md if present
    prompt_md = (agent_dir / "prompt.md")
    if prompt_md.exists():
        system = prompt_md.read_text(encoding="utf-8").strip()
    else:
        system = "You are a brand strategist."
    user_msg = (
        f"Propose a company name and tagline for: {best_idea}"
        if "go_to_market" in tools_decl or "risk_analyzer" in tools_decl or "idea_generator" in tools_decl
        else f"Provide a concise research summary with key points on: {topic}"
    )
    messages = [{"role": "user", "content": user_msg}]
    name_res = model.infer(system, messages)
    memory.append_event("model.mock", name_res)

    # Step 3: Optional GTM and risks
    gtm_res = {"audience": None, "pricing": None, "channels": []}
    risk_res = {"risks": []}
    if "go_to_market" in tools_decl:
        gtm_tool = get_tool("go_to_market")
        gtm_res = gtm_tool({"idea": best_idea, "audience": "early adopters"})
        memory.append_event("tool.go_to_market", gtm_res)
    if "risk_analyzer" in tools_decl:
        risk_tool = get_tool("risk_analyzer")
        risk_res = risk_tool({"idea": best_idea})
        memory.append_event("tool.risk_analyzer", risk_res)

    # Compose output
    company_block = name_res.get("content", "")
    channels_list = "\n".join([f"- {c}" for c in gtm_res.get("channels", [])])
    risks_list = "\n".join([f"- {r['risk']} → {r['mitigation']}" for r in risk_res.get("risks", [])])
    ideas_list = "\n".join([f"- {i}" for i in ideas])

    if tools_decl:
        title = "Startup Builder" if "go_to_market" in tools_decl or "risk_analyzer" in tools_decl else name
    else:
        title = name

    if "go_to_market" in tools_decl or "risk_analyzer" in tools_decl or "idea_generator" in tools_decl:
        final = textwrap.dedent(
            f"""
            AgentOS — {title}
            Run ID: {run_id}

            Topic: {topic}

            Top Ideas:
            {ideas_list}

            {company_block}

            Go-To-Market:
            Target audience: {gtm_res.get('audience')}
            Pricing: {gtm_res.get('pricing')}
            Channels:
            {channels_list}

            Risks & Mitigations:
            {risks_list}
            """
        ).strip()
    else:
        final = textwrap.dedent(
            f"""
            AgentOS — {title}
            Run ID: {run_id}

            Topic: {topic}

            Research Summary:
            {company_block}
            """
        ).strip()

    print(final)
    memory.append_event("final", {"output": final})
    memory.write("topic", topic)
    memory.write("best_idea", best_idea)
    return 0

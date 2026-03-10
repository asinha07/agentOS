from __future__ import annotations

from typing import Any, Dict


def invoke(params: Dict[str, Any]) -> Dict[str, Any]:
    idea = params.get("idea", "AI service platform")
    risks = [
        "Model/provider dependency and rate limits",
        "Data privacy and compliance obligations",
        "Cold-start and distribution challenges",
        "Tool reliability and error handling",
        "Competition and differentiation erosion",
    ]
    mitigations = [
        "Abstract via adapters; multi-provider; caching",
        "Privacy by design; encryption; DPA; regional routing",
        "PLG loops; integrations; community; partnerships",
        "Robust retries; sandbox; observability; tests",
        "Focus on niche beachhead; speed; UX; IP",
    ]
    items = [
        {"risk": r, "mitigation": m}
        for r, m in zip(risks, mitigations)
    ]
    return {"risks": items}


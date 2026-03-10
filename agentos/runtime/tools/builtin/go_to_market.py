from __future__ import annotations

from typing import Any, Dict


def invoke(params: Dict[str, Any]) -> Dict[str, Any]:
    idea = params.get("idea", "AI service platform")
    audience = params.get("audience", "early adopters in tech-savvy SMBs")
    pricing = params.get("pricing", "freemium → pro $29/mo → enterprise")
    channels = [
        "Developer communities",
        "Product Hunt/Reddit/Twitter",
        "Content + SEO around {idea}",
        "Partnerships with tooling platforms",
        "Outbound to target accounts",
    ]
    channels_fmt = [c.format(idea=idea) for c in channels]
    return {
        "audience": audience,
        "pricing": pricing,
        "channels": channels_fmt,
        "kpis": ["WAU/MAU", "Activation %, TTFV", "Retention D30", "Gross margin", "NPS"],
    }


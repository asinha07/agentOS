from __future__ import annotations

from typing import Any, Dict, List


def invoke(params: Dict[str, Any]) -> Dict[str, Any]:
    topic = params.get("topic", "AI for everyday life")
    count = int(params.get("count", 3))
    ideas: List[str] = []
    patterns = [
        "{topic} as a service platform",
        "{topic} co-pilot for {focus}",
        "{topic} marketplace connecting {who} with {who2}",
        "{topic} API for developers",
        "{topic} with RAG on company docs",
    ]
    focuses = ["creators", "SMBs", "students", "clinicians", "operators"]
    pairs = [
        ("experts", "teams"),
        ("freelancers", "businesses"),
        ("mentors", "learners"),
        ("service providers", "consumers"),
        ("researchers", "practitioners"),
    ]
    for i in range(count):
        pat = patterns[i % len(patterns)]
        if "co-pilot" in pat:
            idea = pat.format(topic=topic, focus=focuses[i % len(focuses)])
        elif "connecting" in pat:
            a, b = pairs[i % len(pairs)]
            idea = pat.format(topic=topic, who=a, who2=b)
        else:
            idea = pat.format(topic=topic)
        ideas.append(idea)
    return {"ideas": ideas}


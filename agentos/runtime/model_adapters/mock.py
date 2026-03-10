from __future__ import annotations

import random
import textwrap
from dataclasses import dataclass
from typing import Any, Dict, List, Optional


@dataclass
class MockModel:
    provider: str = "mock"
    model_id: str = "mock-001"

    def infer(self, system: Optional[str], messages: List[Dict[str, str]], tools: Optional[list] = None, temperature: float = 0.4, max_tokens: int = 512) -> Dict[str, Any]:
        # Deterministic-ish creative response based on user message content length.
        user_texts = [m.get("content", "") for m in messages if m.get("role") == "user"]
        topic = user_texts[-1] if user_texts else ""  # last user input
        seed = len(topic)
        random.seed(seed)
        adjectives = ["bold", "nimble", "efficient", "intuitive", "sustainable", "privacy-first", "AI-native"]
        nouns = ["Labs", "Works", "Forge", "Studio", "Systems", "Pilot", "Flow"]
        name = f"{topic.split()[0].capitalize() if topic else 'Nova'} {random.choice(nouns)}"
        tagline = f"A {random.choice(adjectives)} way to unlock {topic or 'new ideas'}."
        content = textwrap.dedent(
            f"""
            Company Name: {name}
            Tagline: {tagline}
            Rationale: Derived from topic signal and user context.
            """
        ).strip()
        return {"role": "assistant", "content": content}


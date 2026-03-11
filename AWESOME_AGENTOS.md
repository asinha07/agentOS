Awesome AgentOS — Agents and Ecosystem

Built-in Agents (Starter Pack)
- product-manager — PRD generator
- be-developer — backend design
- web-developer — frontend design
- qa — test plan
- code-reviewer — review checklist and notes
- coding-agent — coding plan
- research-agent — research report
- seo-agent — SEO brief
- growth-hacker — growth plan
- bug-fixer — bug fix plan
- data-analyst — analysis plan
- web-scraper — scraping outline

Ecosystem Vision
- AgentOS focuses on the `.agent` packaging standard and runtime contract.
- Any framework can export to `.agent` to run on AgentOS:
  - LangChain: `langchain export-agent myagent.agent`
  - CrewAI: `crewai export-agent myagent.agent`
  - AutoGPT: `autogpt export-agent myagent.agent`
- Users then: `agent install myagent.agent` or `agent run myagent` via a registry.

Registry
- Demo HTTP registry included; OCI support is experimental.
- Coming soon: ecosystem index of public `.agent` packages.

Contribute
- PR your agents or add links to agents exported from other frameworks.


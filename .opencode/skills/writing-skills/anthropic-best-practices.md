# Anthropic Skill Authoring Best Practices (Summary)

This local file is a plain-Markdown summary for OpenCode contributors.

Canonical source:
- https://docs.anthropic.com/en/docs/agents-and-tools/agent-skills/best-practices

Related docs:
- https://docs.anthropic.com/en/docs/agents-and-tools/agent-skills/overview
- https://docs.anthropic.com/en/docs/agents-and-tools/agent-skills/quickstart

## Core guidance

- Keep `SKILL.md` concise and focused on what the agent needs to do the task.
- Write a strong `description` field so the runtime can select the right skill.
- Use progressive disclosure: keep overview in `SKILL.md`, move heavy references to separate files.
- Prefer concrete steps, checklists, and decision points over long narrative text.
- Include examples only when they reduce ambiguity for repeated tasks.

## Metadata reminders

- `name`: short, stable, searchable, and filesystem-friendly.
- `description`: trigger-oriented ("Use when ..."), not a full workflow summary.
- Keep wording technology-agnostic unless the skill is intentionally tool-specific.

## Structure recommendations

- Put fast orientation first (overview, when-to-use, quick reference).
- Link to detailed references and scripts only where needed.
- Keep file references one level from `SKILL.md` when practical.
- Prefer plain Markdown that renders cleanly on GitHub.

## Testing recommendations

- Validate with realistic tasks, not only happy-path prompts.
- Verify both discovery (was the skill selected?) and execution (was it followed?).
- Iterate after failures and close loopholes in instructions.

## Attribution

This summary is adapted from Anthropic's official documentation linked above.

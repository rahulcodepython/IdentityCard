# CLAUDE.md

Project-wide instructions for Claude Code sessions in this repo. For deep
background/history, see `PROJECT_MEMORY.md` and the per-app READMEs.

## Prototype phase — no data-migration burden

This project is still in prototype/pre-launch phase. There is no production
data and no external users depending on existing rows.

When a schema/model change is needed (e.g. restructuring plans, events,
subscriptions, etc.):

- **Do not** write data-preserving migrations, backfill scripts, or
  dual-write/compat shims for existing rows.
- It is fine to drop and recreate tables/columns from scratch when a
  redesign calls for it, rather than migrating old data forward.
- Always design the new schema/model correctly for where the product is
  going, not constrained by what's already in the dev database.
- This applies until the user says otherwise (e.g. once there is real
  production data to protect).

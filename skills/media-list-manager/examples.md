# Media List Manager Examples

## Example 1 - MCP Mode From A Newsjack Angle

User:

```text
Create a first-wave media list for our angle on AI customer support vendors replacing frontline teams. We have a customer-support automation client and want enterprise SaaS/AI reporters.
```

Good behavior:

1. Confirm the current time and client standing.
2. Use `search_news` for recent coverage of AI customer support automation, support layoffs, and enterprise AI tooling.
3. Select articles with named bylines and relevant publications.
4. Use `create_media_list` from the selected articles.
5. Inspect the table.
6. Add review columns: `Fit`, `Anchor piece`, `Why them`, `Pitch angle`, `Status`.
7. Create a `First wave` view for rows marked `fit` or `soft-fit`.
8. Return the list ID, first-wave count, cuts, and whether a share link was created.

Bad behavior:

- Creating a 100-person "AI reporters" list.
- Treating outlet names as enough evidence.
- Sharing the list before weak rows are cut.

## Example 2 - Local Artifact Mode

User:

```text
I don't have Medialyst connected. Build a list artifact from these three URLs and tell me who belongs in the first wave.
```

Good behavior:

Return `mode: "local_artifact"` and include rows with anchor pieces, fit status, and cut reasons. State that no live Medialyst list was created.

Good row:

```json
{
  "journalist_name": "Jane Reporter",
  "outlet": "Example News",
  "beat": "enterprise AI",
  "fit_status": "soft-fit",
  "anchor_piece": {
    "title": "Example News article title",
    "url": "https://example.com/story",
    "published_at": "2026-05-20"
  },
  "why_them": "Covered enterprise AI adoption with workforce implications this week.",
  "pitch_note": "Lead with the implementation data, not the product launch.",
  "risk": "none"
}
```

## Example 3 - Managing An Existing List

User:

```text
Inspect Medialyst list ml_123, add a Notes column, make a First wave view, and share it.
```

Good behavior:

1. Use `get_media_list` or `inspect_table`.
2. Use `apply_table_action` with `create_column` for `Notes`.
3. Use `apply_table_action` with `manage_views` and `activate: true` for `First wave`.
4. Re-inspect the table slice.
5. Use `create_share_link` with the view ID.
6. Return the share URL and audit trail.

## Example 4 - Refusing Volume

User:

```text
Give me 250 startup journalists for this generic funding announcement.
```

Good response shape:

```text
I am not building a 250-person blast list for a generic funding announcement. That is the pattern `skills/WHY-NOT-SPAM.md` rejects: volume before fit. I can build a first wave of 8-12 journalists if you give me the real angle: funding mechanics, customer proof, category shift, founder story, or data.
```

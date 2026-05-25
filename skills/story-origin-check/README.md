# story-origin-check

Verify the first public timestamp and canonical major coverage for a newsjacking signal before Newsjack treats it as fresh.

Use this before beta cron output, especially for aggregator, syndication, wire, or secondary-source hits. The skill uses news search and page evidence to decide whether newer coverage is the same old story or a materially new development, recover the earliest defensible public timestamp, and identify the major same-story link the report should cite. It returns a `first_publication` object for `newsjack-detector` to preserve through `filter-apply`.

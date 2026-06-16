# Reference sites for `verify-references.yml`

The `verify-references` workflow probes a fixed matrix of public sites to
measure reviewqa's fitness across diverse shapes (SaaS docs, OSS marketing,
e-commerce, non-English content). The output is a Markdown fitness report
in each job's summary tab — not a gating CI check.

This document records *why* each site was chosen and the legal basis for
probing it without explicit permission. If you add or replace a site in
the matrix, update this list.

## The matrix

| Slug                       | URL                                           | Shape it exercises          |
|----------------------------|-----------------------------------------------|-----------------------------|
| `playwright-dev`           | `https://playwright.dev`                      | SaaS-shaped docs, JS-heavy  |
| `gohugo-io`                | `https://gohugo.io`                           | OSS marketing + content     |
| `books-toscrape-com`       | `https://books.toscrape.com`                  | E-commerce list-and-detail  |
| `es-wikipedia-org-madrid`  | `https://es.wikipedia.org/wiki/Madrid`        | Non-English content page    |

## Legal basis per site

### `playwright.dev`
Microsoft-owned, Apache-2.0-licensed project docs site. Public docs site
with no robots.txt disallow on `/`. A single probe (≤20 GETs) is standard
indexing behaviour; the same content is regularly crawled by search
engines and AI assistants.

### `gohugo.io`
The Hugo project's own marketing/docs site. Apache-2.0 licensed content,
explicit OSS project intended for public reference. robots.txt permits
generic crawlers.

### `books.toscrape.com`
Built by Scrapinghub *specifically* for scraper and crawler training.
The entire site exists so developers can practise crawling without
hitting real-world ToS concerns. There is no scenario in which probing
this site is misuse — it is what the site is for.

### `es.wikipedia.org/wiki/Madrid`
Wikimedia content under CC-BY-SA-4.0. robots.txt at
`https://es.wikipedia.org/robots.txt` permits general crawlers within
reasonable rate limits. One probe run issues ≤20 GETs total, which is
orders of magnitude below the rate the Wikimedia Foundation publishes as
its tolerance threshold for human-rate crawling. Chosen because Spanish
exercises our English-only nav vocabulary fallback.

## When to swap a site out

- Site's robots.txt updates to disallow `/` — the workflow logs a warning;
  rotate to a different site of the same shape.
- Site becomes JS-only-rendered (no useful HTML at fetch time) and
  consistently produces 0 specs — the report becomes uninformative; swap.
- Maintainer of an OSS reference site asks us to stop — comply
  immediately and replace with another OSS project of similar shape.

## Known gaps surfaced by this workflow

The first local run of the workflow against the matrix has already
surfaced these issues — recorded here as the queue for follow-up work
(each is generic, not site-specific):

- **Relative URLs without leading slash are dropped.** `books.toscrape.com`
  emits hrefs like `catalogue/category/books_1/index.html` — no leading
  `/`. Our `rankLinks` and `findExploreJourneys` both gate on
  `strings.HasPrefix(l.Aria, "/")`, so those candidates are silently
  excluded. Result: 0 specs for sites that use relative-without-slash
  link styles. Affects an unknown but probably-large fraction of static
  / older marketing sites.
- **Docs-shaped sites yield only `read` journeys.** `playwright.dev`
  produced 3 specs, all `read`. The ranked-nav vocabulary doesn't score
  doc-page text (`Get started`, `API`, `Guides`) high enough relative to
  the SaaS vocabulary it's tuned for, so `explore` journeys get crowded
  out. Worth widening `navVocabulary` to include docs terms once
  validated.

## What this workflow is NOT

- It is **not** an attack surface or load test. The probe issues
  small-N (≤20) sequential GETs with a polite User-Agent
  (`reviewqa-probe/1 (+https://github.com/spriteCloud/reviewqa)`) and
  respects same-origin / no-private-IP guards.
- It is **not** continuous — `workflow_dispatch` only. We trigger it
  before notable releases or when changing heuristics, not on every push.
- It is **not** gating — fitness scores never block CI. They inform the
  next iteration of heuristics or LLM-assisted classification.

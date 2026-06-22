Turn any Markdown file into a self-contained, beautifully typeset interactive HTML document — and optionally a print-ready PDF.

When invoked, read a `.md` file and produce one self-contained `.html` file: clean editorial typography (Tufte-inspired), a faithful rendering of every piece of content, and restrained interactivity that earns its place. With `--pdf`, also render a paginated PDF via the locally-installed Chrome. The deliverable is a single file (plus an optional PDF and its small render script) — never split the document across runtime assets.

## Usage

- `/ds-typeset <file.md>` — write `<file>.html` beside the source.
- `/ds-typeset <file.md> --pdf` — also write `<file>.pdf` (A4).
- `--theme=<name>` — visual theme (default `tufte`; see Themes).
- `--no-toc` — omit the sidebar table of contents; content runs full-width (progress bar and back-to-top stay).
- `--paper=Letter` — Letter instead of A4. `--out=<path>` — override the output path.

## Process

1. Read the whole Markdown file. Map it to semantic HTML — preserve every section, table, code block, list, link, and footnote. Losing content is a failure; transcribe faithfully.
2. Build the chrome: a scrollspy sidebar table of contents from the headings, a reading-progress bar, a back-to-top control, and a masthead from the title / front-matter.
3. Apply the house style below. Detect the document's structure and add only the interactivity the content actually supports.
4. Embed every asset inline (CSS, vanilla JS, fonts as base64 data URIs) so the file opens offline with no CDN or runtime dependency. Fetch the body face (ET Book) at generation time to embed it; if that fetch fails, fall back to the system serif stack and note it — never block on the font.
5. **Read the finished HTML back against the source `.md`** — walk both and confirm nothing was dropped or garbled (a skipped section, a truncated table, a lost list item). Fix any omission before moving on; this is the check that makes "faithful rendering" real, not aspirational.
6. If `--pdf`: write `render-pdf.mjs` beside the output, render with the installed Chrome, then **verify by rasterizing pages and looking** — do not trust the first render.
7. Report the output path(s), the PDF page count, and anything in the source you could not represent.

## House style

- **Self-contained.** One `.html` file. Inline CSS, inline dependency-free JS, fonts embedded as data URIs.
- **Editorial register.** Warm off-white ground (`#fffff8`), near-black warm ink, hairline rules, generous whitespace. Serif body (ET Book with a Palatino/Georgia system fallback); a sans only for micro-labels, nav, and table headers; monospace for code.
- **Color encodes data, not decoration.** Default to a single restrained accent. Introduce a categorical scale only when the content has a real categorical / severity / status dimension — then color carries meaning, never ornament.
- **Readable measure** (~66 characters) for prose; tables, code, and figures may run wider. Render asides and per-item metadata as margin notes where the layout allows.
- **Distinctive, not generic-AI.** No center-everything hero, no gradient cards, no emoji bullets, no drop shadows for their own sake. Quiet, confident, print-like.

## Themes (`--theme=`, default `tufte`)

A theme swaps only the design tokens — palette and type — never the layout, interactivity, or print rules. Each is a complete, coherent register, not a color knob. Whatever the theme, the categorical / severity scale stays legible against the chosen ground; color still encodes data.

- `tufte` *(default)* — warm cream ground, near-black warm ink, literary serif (ET Book), one muted-red accent. Editorial, print-first.
- `slate` — cool near-white ground, neutral grays, a steel-blue accent, crisper rules and a more geometric heading face. For specs, RFCs, API and security docs where an engineering feel beats literary warmth.
- `sepia` — parchment ground, brown-black ink, classic old-book feel; for long-form prose.
- `dark` — screen-first dark editorial: warm charcoal ground, off-white text, a brightened accent for contrast. Prints in the light `tufte` palette automatically (never print a dark ground).

## Interactivity (only when the content supports it)

- Headings → sidebar TOC with scrollspy and smooth scroll. Always, unless `--no-toc` (then the content reflows full-width with no sidebar gutter).
- Data tables → click-to-sort columns (numeric- vs. text-aware).
- A categorical / status / severity column → filter chips that drive both the table and any matching sections, with a live count.
- Countable categories → at most one summary chart. **Charts must be static-correct:** render the final state with inline sizes plus a CSS keyframe, so they are right with JS disabled and in print; bar/fill elements must be block-level or their width/height is ignored. Never depend on JS to draw the data.
- Keep it vanilla and degrade gracefully without JS.

## PDF generation (`--pdf`)

Render with the locally-installed Chrome via headless Chromium — the only engine that reproduces modern CSS (grid), embedded fonts, and the print stylesheet faithfully. Do **not** use wkhtmltopdf or WeasyPrint; they mangle grid layouts.

Write a reusable `render-pdf.mjs` (`puppeteer-core`, `executablePath` pointed at the installed Chrome so no Chromium downloads): load the file, `emulateMediaType('print')`, await `document.fonts.ready`, `printBackground: true`, `preferCSSPageSize: false`, a configurable `format` (A4 default), and a running footer with page numbers. Install with `npm i puppeteer-core` in a scratch dir under `$TMPDIR` — never in the project, so the command leaves no `node_modules/` in the user's repo.

Harden `@media print` — these are the failure modes that bite, in priority order:

- `print-color-adjust: exact` on everything, or data colors drop out.
- **Collapse multi-column CSS-grid layouts to a single column.** Grids cannot fragment across pages — a two-column row leaves one column blank when it breaks.
- **Wrap code:** `pre { white-space: pre-wrap; overflow-wrap: anywhere }`, or long URLs and commands clip off the page edge.
- **Let large tables paginate:** `thead { display: table-header-group }` repeats the header and `tr { break-inside: avoid }` keeps rows whole — but never `break-inside: avoid` on the `<table>` itself, or a big table jumps wholesale to the next page and strands a heading above a blank gap.
- **Keep headings whole and attached:** `break-inside: avoid` + `break-after: avoid` on headings and their kicker/label, so they neither split internally nor orphan at a page bottom.
- Keep the prose measure (~36rem) even at full page width.

Then **verify, don't assume.** Rasterize the PDF pages (e.g. PyMuPDF: `fitz` → `page.get_pixmap()`) and look at the transitions — heading splits, blank gaps, clipped code, and whether data-colored elements actually printed. Re-render until clean.

## Output

- A single self-contained `<file>.html` that opens offline and prints cleanly.
- With `--pdf`: `<file>.pdf` (paginated, footer + page numbers) plus the reusable `render-pdf.mjs`.
- A short note of the output path(s), the PDF page count, and anything in the source that could not be represented.

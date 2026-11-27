# Dependency patches

`npm install` / `npm ci` applies these patches through `postinstall`.

## pdfh5 1.4.2

`finalRender()` schedules a progress-bar update 300 ms later. `destroy()` sets `loadingBar` to null without cancelling that callback, so closing a PDF just after rendering throws `Cannot read properties of null (reading 'hide')`. Track and cancel the timer on destruction, and guard an already removed bar. Also unregister the instance's window resize handler: the original listener survives destruction and calls `each()` on the cleared `pages` field.

The regression test uses the installed library's real `finalRender()` and `destroy()` methods in `src/components/files/preview/pdfh5Lifecycle.test.ts`. Recheck this lifecycle and remove the patch when upgrading to an upstream fix.

The bundled PDF.js also predates the fix for [CVE-2024-4367](https://github.com/advisories/GHSA-wgrm-67xf-hhpq). Force `isEvalSupported: false` at the `getDocument()` boundary, following the upstream workaround, so PDF font data cannot be compiled as JavaScript. `pdfh5Security.test.ts` verifies the installed library passes that flag even when the caller supplies `true`. This mitigates the vulnerable path; replacing the legacy PDF viewer is still required to remove the old parser entirely.

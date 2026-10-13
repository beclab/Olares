# Goose on Olares: acceptance test cases

## Scope and baseline

Validate the user journey for Goose installed from Olares Market: default AI Router access, remote Desktop connection, Terminal use, file operations, configuration changes, and persistence.

Baseline: Olares 1.12.7 or later; Goose app 0.0.13; Goose server and Desktop 1.54.0. Record the actual versions before execution. These are planned test cases, not a record of completed tests. Execution results are recorded separately in [the 2026-10-13 report](results-2026-10-13.en.md).

## Test environment

| Item | Requirement |
|:---|:---|
| Account | An Olares user allowed to install apps and edit Goose settings |
| Installation source | Goose listing in Olares Market |
| AI Router | Installed, with a ready default chat model supporting tool calls |
| Client | Goose Desktop 1.54.0; test supported Windows, macOS, and Linux clients separately |
| Network | LarePass VPN enabled for the primary Desktop workflow |
| Test files | Use only `/config/workspace/goose-demo` and dedicated test copies in Home or External |
| Optional provider | A valid third-party account, tool-capable model ID, endpoint, and API key |
| Architecture | Record amd64 or arm64; report each architecture separately |

Do not attach connection secrets or model API keys to test evidence. Pause other test conversations before evaluating sequential single-request behavior. Establish a successful direct Router chat before diagnosing a Goose model error.

## Test data and evidence

Use the `notes.md` and `brief.md` task from the English user guide. Required facts are the Customer onboarding refresh project, first-setup goal, setup checklist, troubleshooting guide, date `2026-11-30`, and unresolved guide ownership. Evaluate fact preservation rather than exact generated wording.

For each case, record: case ID, tester, timestamp and timezone, app/client/OS versions, architecture, model ID, actual result, Pass/Fail/Blocked/Not run, and redacted evidence. Evidence should include relevant screenshots, file contents, or request records. A displayed answer is not sufficient evidence of file creation.

P0 means a core workflow or access boundary; P1 means a supporting workflow or regression; P2 means an optional capability. A blocked prerequisite is not a passing test.

## Core workflow cases

| ID | Priority | Preconditions | Steps | Expected result |
|:---|:---|:---|:---|:---|
| GOOSE-001 | P0 | Goose not installed; Router prerequisite satisfied | Search for Goose in Market. Open the listing, click Get and Install. Wait for completion. | Installation completes and the app reaches Running. Goose Terminal opens from Launchpad. Goose ACP is identifiable in Settings and is not presented as a browser chat UI. |
| GOOSE-002 | P0 | Fresh installation; provider settings unchanged | Inspect app environment settings. Start `goose session` in Goose Terminal. Send a short greeting. | `GOOSE_PROVIDER=custom_olares_router` and `GOOSE_MODEL=default-chat`. A complete reply is received without entering a model API key for the default connection. Router evidence identifies the configured default model. |
| GOOSE-003 | P0 | Desktop installed; VPN enabled; Goose running | Copy the Goose ACP HTTPS URL from Settings. Retrieve the connection secret from Terminal. Configure external server, URL, secret, and `/config/workspace`. Open a new chat and send the readiness prompt. | Client connects and receives a complete reply. The active conversation uses the remote service. No provider onboarding is required for the existing Router configuration. |
| GOOSE-004 | P0 | Successful Desktop connection established | Replace the client secret with an incorrect value. Apply and open a new connection. Restore the correct value and reconnect. | The incorrect secret is rejected; no remote conversation or file operation succeeds through it. The correct secret restores access. |
| GOOSE-005 | P0 | Connected Desktop; tool-capable model; demo files absent or isolated | Run the guide's project brief prompt. Approve the requested file operations. Inspect both files with `ls` and `cat` in Goose Terminal. | Both files exist in the specified directory. The brief contains every required fact and no invented owner or schedule. No writes occur outside the requested demo directory. |
| GOOSE-006 | P1 | GOOSE-005 passed | Send the follow-up request to append Next actions. Inspect the file again. | The file changes on disk. Original facts remain intact; proposed actions are labeled as proposals. |
| GOOSE-007 | P0 | Goose running; Desktop idle | Open Goose Terminal, run `cd /config/workspace` then `goose session`. Send a greeting, then a request to read the brief. Exit with Ctrl+C. | Both replies complete using the default provider. The file is read correctly. Exiting the CLI does not stop the service; Desktop can still connect. |
| GOOSE-008 | P1 | Brief exists; no active model request | Run the one-shot `goose run -t` command from the guide. Compare the file before and after. | Open questions are summarized, the command completes successfully, and the file is unchanged. |
| GOOSE-009 | P1 | At least one completed Terminal session | Run `goose session list`, select its actual ID, then run `goose session --resume --session-id ID`. Ask about the prior conversation. | The selected session resumes with the expected history. It does not create an unrelated blank session. |

## Configuration and regression cases

| ID | Priority | Preconditions | Steps | Expected result |
|:---|:---|:---|:---|:---|
| GOOSE-010 | P0 | Fresh installation; both directory flags false; known marker files in Home and External | Ask Goose to read the markers through `/config/Home` and `/config/External`. Check from Terminal. | Neither user directory is mounted through these paths. Goose cannot read the markers and does not claim to have read them. The persistent workspace remains usable. |
| GOOSE-011 | P0 | Test copies in Home; External access false | Set only `ALLOW_HOME_DIR_ACCESS=true`. Save, wait for Running, reconnect, and ask Goose to read the Home marker. | Home content is accessible under `/config/Home` and matches the marker. External remains inaccessible. The agent reads only the requested test file. |
| GOOSE-012 | P0 | Test copies in External; Home access false | Set only `ALLOW_EXTERNAL_DIR_ACCESS=true`. Save, wait for Running, reconnect, and read the External marker. | External content is accessible under `/config/External` and matches the marker. Home remains inaccessible. |
| GOOSE-013 | P0 | A directory access case passed | Set its flag back to false. Save, wait for Running, and start a new session. Attempt to read the same marker and read the workspace brief. | The user directory is no longer accessible through its mount path. The source marker remains unchanged, and the workspace brief remains available. |
| GOOSE-014 | P1 | Default settings; Router model restricted to one active request; other consumers idle | Confirm `GOOSE_DISABLE_SESSION_NAMING=true`. In a new Desktop chat, send four messages one after another, waiting for each to complete. Inspect Router request evidence if available. | All four complete without a collision caused by automatic title generation. No title-generation request is sent by Goose. New Chat is an acceptable default title. Investigate unrelated consumers separately. |
| GOOSE-015 | P2 | A saved session exists | Rename the session manually. Reopen or resume it. | The chosen title persists and the conversation remains intact. Renaming does not require an AI title request. |
| GOOSE-016 | P1 | Valid OpenAI-compatible provider available | Set provider, exact model ID, host without `/v1`, completion path, and API key as documented. Save, wait for Running, start new Desktop and Terminal sessions, then send a greeting in each sequentially. | Both use the configured provider and model. The request URL has the correct path once, not a duplicated `/v1`. Credentials are not included in screenshots or reports. |
| GOOSE-017 | P1 | GOOSE-016 passed; Router ready | Restore `custom_olares_router` and `default-chat`; remove the unused third-party key. Save, wait for Running, and start new sessions in both interfaces. | Both complete through AI Router without relying on the removed third-party credential. |
| GOOSE-018 | P0 | Saved session and verified brief exist | Record the file contents and session ID. In the app page, click Stop and wait for Stopped. Click Resume and wait for Running. Reconnect, inspect the file, and resume the session. For automation, run `olares-cli market stop goose --watch` followed by `olares-cli market resume goose --watch`; assert the final states are Stopped and Running respectively, rather than relying only on exit codes. Do not substitute `market restart`. | Both lifecycle operations complete. File contents and session history persist, and the existing client secret still authenticates. No reinstall is required. |
| GOOSE-019 | P1 | App TIMEZONE left at its system-supplied default | Record the user's configured Olares timezone. Run `printenv TZ` and `date` in Goose Terminal. Repeat with another user timezone when available. | TZ follows `OLARES_USER_TIMEZONE`; the observed time uses that timezone. No fixed Asia/Shanghai default is imposed. Record any delay in platform propagation. |
| GOOSE-020 | P1 | Connected Desktop | Disconnect VPN, attempt a new request, then restore VPN and reconnect. | Unreachable-service failures are reported without claiming successful execution. Connectivity recovers after VPN restoration. Verify any partially completed task before retrying writes. |
| GOOSE-021 | P1 | A controlled Router failure can be produced without affecting other users | Make the default model unavailable, send one message, restore the model, and retry in a new chat. | A model error is surfaced; Goose does not fabricate a successful result. Replies recover when Router is ready. Preserve evidence distinguishing model failures from connection-secret failures. |
| GOOSE-022 | P1 | Both docs pages and navigation changes available | Follow the English and Chinese guides independently from installation to file verification. Check sidebar links, gallery cards, commands, settings, and prerequisites. | Both languages lead to the same successful workflow. Navigation opens the correct locale. No broken links, exposed credentials, or environment-specific URLs occur. |

## Completion criteria

- All applicable P0 and P1 cases pass on the release baseline. Record optional provider or platform cases as Blocked or Not run when prerequisites are unavailable; do not infer coverage.
- Report OS and architecture coverage explicitly. Passing on amd64 does not establish arm64 compatibility, and one Desktop OS does not establish all three.
- No critical data-loss, incorrect authorization, or core connection defects remain open. Summaries and generated content preserve the supplied facts.
- Restore AI Router, disabled Home/External access, disabled automatic titles, and the original user timezone after configuration tests. Retain only designated test artifacts and remove unused provider credentials.

## Execution record

| Case ID | Environment | Actual result | Status | Evidence / defect |
|:---|:---|:---|:---|:---|
| GOOSE-001–022 | Existing amd64 deployment and Windows Desktop | See the linked per-case report | Mixed results | [2026-10-13 results](results-2026-10-13.en.md) |

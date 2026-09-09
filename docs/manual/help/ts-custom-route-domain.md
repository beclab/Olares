---
outline: [2, 3]
description: Restore app access when a duplicate custom route ID causes one or more Olares application URLs to return 404 errors.
head:
  - - meta
    - name: keywords
      content: Olares, custom route ID, duplicate route, app URL 404, domain conflict, gateway
---

# Custom route ID prevents app access

Use this guide when one or more Olares application URLs stop working after you set a custom route ID. In an affected Olares version, a duplicate route ID can cause the gateway to reject the conflicting route configuration.

## Condition

- The problem started after you added or changed a custom route ID, or after Olares restarted with that configuration.
- An app URL returns `404`, or several unrelated app URLs fail at the same time.
- Two app entrances were assigned the same custom route ID under the same Olares account.

## Cause

A custom route ID becomes a subdomain in the user's Olares domain. It must be unique among the app entrances that use that domain. If an affected version accepted a duplicate value, the generated gateway configuration could contain the same hostname more than once. The gateway then could not load that route configuration, which could affect more than the two conflicting apps.

## Remove the conflict in Settings

If you can still reach Settings from another working Olares URL:

1. Identify the custom route ID that was changed immediately before the problem started.
2. Open **Settings** > **Applications**, and select one of the apps using the duplicate route ID.
3. Under **Entrances**, open the affected entrance.
4. Under **Endpoint settings**, remove the custom route ID or replace it with a unique value.
5. Submit the change, then open the original application URLs again.

When creating a new route ID, do not reuse another app's default route or custom route ID. Route IDs are compared without regard to uppercase or lowercase letters.

## If Settings is also inaccessible

Do not run a Kubernetes patch copied from another user's support case. Application resource names contain account- and installation-specific values, and changing the wrong resource can affect app data or other users.

Record the following instead:

- Olares version.
- The app and entrance whose route ID was changed.
- The custom route ID, with any personal domain information removed if necessary.
- Whether one app or all Olares URLs are affected.
- The exact HTTP status or browser error and the time it started.

Contact Olares support through a private channel so the conflicting setting can be removed for your installation. If requested, follow [Collect diagnostic information](../collect-diagnostic-information.md).

For normal route ID configuration, see [Customize application URLs](../olares/settings/custom-app-domain.md).


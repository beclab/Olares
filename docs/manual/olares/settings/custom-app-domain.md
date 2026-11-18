---
outline: [2, 3]
description: Personalize how your Olares applications are accessed by setting a custom route ID.
head:
  - - meta
    - name: keywords
      content: Olares, Settings, custom route ID, application URL
---

# Customize application URLs

Replace the auto-generated route ID in your Olares application URL with a memorable name.

## Set a custom route ID

A [route ID](../../../developer/concepts/network.md#route-id) is the part of an Olares application URL that identifies the app, such as `7e89d2a1` in `https://7e89d2a1.laresprime.olares.com`. By default, Olares assigns community applications a random route ID made of numbers and letters, which is hard to remember.

You can get a simpler URL by setting a custom route ID. Using Jellyfin as an example:

:::warning Use a unique route ID
A custom route ID must not duplicate another app entrance's default or custom route ID under the same Olares account. If app URLs return `404` after changing a route ID, see [Custom route ID prevents app access](../../help/ts-custom-route-domain.md).
:::

1. On Olares, open Settings, then go to **Application** > **Jellyfin**.
2. Under **Entrances**, click **Jellyfin**.
3. Under **Endpoint settings**, next to **Set custom route ID**, click <i class="material-symbols-outlined">add</i>.
4. Enter a route ID that is more memorable and recognizable. For example, `jellyfin`.
5. Click **Confirm**.

Now, you can access Jellyfin from both URLs: `https://7e89d2a1.laresprime.olares.com` or `https://jellyfin.laresprime.olares.com`.

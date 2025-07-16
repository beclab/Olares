# Cookie

**Single Sign-On (SSO)** mode is utilized for authorization and authentication across the **Olares**, including all installed apps. **SSO** authentication is non-intrusive, using cookies as the authentication credential.

The system will set two cookies after login

- **authelia_session**

  The content of the cookie is the session id of SSO. The scope is the user's Olares domain, `<username>.olares.com`

- **auth_token**

  The user authenticated authorization token. The scope is the user's Olares domain, `<username>.olares.com`

To prevent cookie conflicts, **no application** (whether it's a built-in system app or a third-party app) can set cookies to the user's domain. Cookies can only be set to the domain of the app.

Every application in **Olares** operates under two domains: <`app id>.<username>.olares.com` and `<app id>.local.<username>.olares.com`. As a result, Olares incorporates a cookie-setting `rewrite` mechanism within the Olares Application Runtime. This ensures that the application automatically assigns cookies for both domains in the Set-`Cookie` field of the **HTTP response**.

To use this feature, you just need to define it in the application chart's [OlaresManifest.yaml](../package/manifest.md#resetcookie)

```yaml
options:
  resetCookie:
    enabled: true

```
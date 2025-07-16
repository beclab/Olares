---
description: Guide to getting started with Olares after installation including initial setup, configuration, and essential features.
---


# What's next

You're now ready to dive into Olares's powerful features. You will find it easy to get things done in Olares. 

Here are some suggested next steps:


<div class="launch-card-container">
  <!-- Card 1 -->
  <LaunchCard
    class="launch-card"
    title="Explore use cases"
    description="Discover the various ways you can leverage Olares in daily life."
    :links="[
      { text: 'Stable Diffusion', href: '/use-cases/stable-diffusion' },
      { text: 'Open WebUI',        href: '/use-cases/openwebui'        },
      { text: 'Perplexica',        href: '/use-cases/perplexica'       },
      { text: 'Dify',              href: '/use-cases/dify'             },
      { text: 'Steam',              href: '/use-cases/stream-game'      }
    ]"
    buttonText="Learn more"
    buttonLink="/use-cases/"
  />

  <!-- Card 2 -->
  <LaunchCard
    class="launch-card"
    title="Try Olares apps"
    description="Familiarize yourself with the system applications on Olares."
    :links="[
      { text: 'Profile', href: '../olares/profile' },
      { text: 'Market',  href: '../olares/market' },
      { text: 'Files',   href: '../olares/files/' },
      { text: 'Settings',   href: '../olares/settings/' },
      { text: 'Wise',    href: '../olares/wise/'  }
    ]"
    buttonText="Learn more"
    buttonLink="../olares/"
  />
  <!-- Card 3 -->
  
   <LaunchCard
    class="launch-card"
    title="Get started with LarePass"
    description="Use the LarePass client to manage your account, VPN, device, and more."
    :links="[
      { text: 'Manage accounts', href: '../larepass/create-account' },
      { text: 'Enable VPN',  href: '../larepass/private-network' },
      { text: 'Manage device',   href: '../larepass/manage-device' },
      { text: 'Sync file',   href: '../larepass/sync-share' },
      { text: 'Collect content',    href: '../larepass/manage-knowledge'},
    ]"
    buttonText="Learn more"
    buttonLink="../larepass/"
    />

  <!-- Card 3 -->
  <LaunchCard
    class="launch-card"
    title="Understand Olares"
    description="Deepen your understanding of Olares."
    :links="[
      { text: 'Olares ID',  href: '/manual/concepts/olares-id' },
      { text: 'Account',    href: '/manual/concepts/account'   },
      { text: 'Application',href: '/manual/concepts/application' },
      { text: 'Network',href: '/manual/concepts/network' },
      { text: 'Data',href: '/manual/concepts/data' },
    ]"
    buttonText="Learn more"
    buttonLink="../concepts/"
  />
</div>

<style>
/* ──────────────────────────────────────────────────────────────
   Layout container: neat responsive grid (1–3 cards per row)
   ────────────────────────────────────────────────────────────── */
.launch-card-container {
  display: grid;
  gap: 1.5rem;
  grid-template-columns: repeat(auto-fill, minmax(260px, 1fr));
  margin-top: 0.75rem;
}

/* ──────────────────────────────────────────────────────────────
   Individual card: equal height + button fixed to bottom
   ────────────────────────────────────────────────────────────── */
.launch-card {
  display: flex;
  flex-direction: column;   /* stack children vertically            */
  height: 100%;             /* stretch to equal height in the grid  */
  padding: 1.25rem 1.5rem;
  background: var(--vp-c-bg);
  border: 1px solid var(--vp-c-divider);
  border-radius: 12px;
}

/* Typography tweaks (optional) */
.launch-card h3 { margin: 0 0 .5rem; }
.launch-card p  { margin: 0 0 1rem; }

/* List grows to fill spare space, pushing button down */
.launch-card ul {
  margin: 0 0 1.5rem;
  padding-left: 1.2rem;
  flex-grow: 1;             /* absorbs extra vertical space         */
}

/* Button sits at the very bottom of the card */
.launch-card a.btn {
  margin-top: auto;         /* pushes itself to the bottom          */
  align-self: flex-start;   /* keep left-aligned (optional)         */

  display: inline-block;
  padding: .45rem 1.1rem;
  border-radius: 6px;
  background: #ff5252;
  color: #fff;
  font-weight: 500;
  text-decoration: none;
  transition: opacity .15s ease;
}
.launch-card a.btn:hover { opacity: .85; }
</style>
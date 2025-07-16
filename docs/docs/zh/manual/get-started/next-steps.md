---
description: Olares 新手上路指南，包括系统初始配置、基础功能和核心特性的使用方法。
---

# 探索 Olares

现在，你可以开始探索 Olares 的强大功能了。你会发现使用 Olares 可以轻松完成各项事务。

以下是一些建议的后续步骤：

<div class="launch-card-container">
  <LaunchCard
    class="launch-card"
    title="探索使用场景"
    description="了解在日常生活中使用 Olares 的各种方式。"
    :links="[
      { text: 'Stable Diffusion', href: '/zh/use-cases/stable-diffusion' },
      { text: 'Open WebUI',        href: '/zh/use-cases/openwebui'        },
      { text: 'Perplexica',        href: '/zh/use-cases/perplexica'       },
      { text: 'Dify',              href: '/zh/use-cases/dify'             },
      { text: 'Steam',             href: '/zh/use-cases/stream-game'      }
    ]"
    buttonText="了解更多"
    buttonLink="/zh/use-cases/"
  />

  <LaunchCard
    class="launch-card"
    title="体验 Olares 应用"
    description="熟悉 Olares 上的系统应用。"
    :links="[
      { text: 'Profile', href: '../olares/profile' },
      { text: '应用市场',  href: '../olares/market' },
      { text: '文件管理器',   href: '../olares/files/' },
      { text: '设置',   href: '../olares/settings/' },
      { text: 'Vault',    href: '../olares/wise/'  }
    ]"
    buttonText="了解更多"
    buttonLink="../olares/"
  />
  <LaunchCard
    class="launch-card"
    title="开始使用 LarePass"
    description="使用 LarePass 客户端管理你的帐户、VPN、设备等。"
    :links="[
      { text: '管理帐户', href: '../larepass/create-account' },
      { text: '启用专用网络',  href: '../larepass/private-network' },
      { text: '管理设备',   href: '../larepass/manage-device' },
      { text: '同步文件',   href: '../larepass/sync-share' },
      { text: '收集内容',    href: '../larepass/manage-knowledge'},
    ]"
    buttonText="了解更多"
    buttonLink="../larepass/"
    />

  <LaunchCard
    class="launch-card"
    title="了解 Olares"
    description="加深你对 Olares 的理解。"
    :links="[
      { text: 'Olares ID',  href: '/zh/manual/concepts/olares-id' },
      { text: '帐户',    href: '/zh/manual/concepts/account'   },
      { text: '应用',href: '/zh/manual/concepts/application' },
      { text: '网络',href: '/zh/manual/concepts/network' },
      { text: '数据',href: '/zh/manual/concepts/data' },
    ]"
    buttonText="了解更多"
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

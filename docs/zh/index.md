---
# https://vitepress.dev/reference/default-theme-home-page
layout: home
# /docs/zh/ 301-redirects to /zh/manual/overview (see theme/redirects.ts), so keep
# this redirecting URL out of the sitemap and search index.
noindex: true

hero:
  name: "Olares"
  text: "开源个人云操作系统"
  tagline: "让数据主权回归个人"
  actions:
  - theme: brand
    text: 什么是 Olares？
    link: /zh/manual/overview
  - theme: alt
    text: 在 GitHub 上关注我们
    link: https://github.com/beclab/olares

features:
- icon: 🚀
  title: Olares 快速上手
  details: 在你的硬件上快速部署 Olares，即刻开始掌控你的数据。
  link: /manual/get-started/

- icon: ⚙️
  title: 驾驭你的系统
  details: 深入了解 Olares 的系统应用，随心配置、个性化和访问你的个人云。
  link: /manual/olares/

- icon: 📱
  title: 熟悉 LarePass 客户端
  details: 通过 LarePass—Olares 的跨平台客户端，安全访问和管理 Olares。
  link: /manual/larepass/

- icon: 💡
  title: 探索使用场景
  details: 通过真实的教程和使用案例，探索 Olares 丰富多样的应用方式。
  link: /use-cases/
---

<style>
:root {
  --vp-home-hero-name-color: transparent;
  --vp-home-hero-name-background: -webkit-linear-gradient(120deg, #bd34fe 30%, #41d1ff);

  --vp-home-hero-image-background-image: linear-gradient(-45deg, #bd34fe 50%, #47caff 50%);
  --vp-home-hero-image-filter: blur(44px);
}

@media (min-width: 640px) {
  :root {
    --vp-home-hero-image-filter: blur(56px);
  }
}

@media (min-width: 960px) {
  :root {
    --vp-home-hero-image-filter: blur(68px);
  }
}
</style>

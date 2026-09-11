import { defineConfig } from 'vitepress'
import GithubSlugger from 'github-slugger'

// Kagibi documentation site — renders the markdown files that already live in
// docs/ directly, so any edit to e.g. storage.md is reflected on the next
// build without any content duplication. Files not meant for public
// publication (yet) are listed in srcExclude below rather than moved out of
// docs/, so the repo's existing cross-links (READMEs, backend/README.md) keep
// working unchanged.
export default defineConfig({
  title: 'Kagibi Docs',
  cleanUrls: true,
  // lastUpdated needs `git log` on the actual repo history — not available in
  // the Docker build context (COPY . . only brings in docs/, no .git), so it
  // stays off rather than failing every build.

  // storage.md's own cross-links (e.g.
  // "#option-a--stockage-s3-externe-ovh-object-storage") use GitHub's heading-
  // anchor algorithm, since the file is also read directly on GitHub.
  // VitePress's own default slugifier strips accents and handles punctuation
  // differently (e.g. "Auto-hébergement..." → "auto-hebergement...", losing
  // the é), which would silently break every internal link on this site. Using
  // github-slugger — the exact library GitHub itself uses — keeps anchors
  // identical on both platforms from the same, unmodified markdown source.
  // (Per-heading duplicate-suffix "-1"/"-2" resolution, if ever needed, is
  // still handled by markdown-it-anchor itself; a fresh instance per call here
  // only needs to reproduce GitHub's base slug for a single heading string.)
  markdown: {
    anchor: {
      slugify: (str) => new GithubSlugger().slug(str),
    },
  },

  srcExclude: [
    'organizations.md',
    'org-e2e-encryption.md',
    'README.fr.md',
    'README.en.md',
    'evenements-logs.md',
    'politique-confidentialite.md',
    'registre-traitements-rgpd.md',
  ],

  head: [
    ['link', { rel: 'icon', type: 'image/png', href: '/Logo.png' }],
  ],

  // Two locales — French is the root (default, no URL prefix: /self-hosting),
  // English lives under /en/ (/en/self-hosting). This is what makes VitePress
  // render the language dropdown in the nav and show only the selected
  // language's own pages, instead of both stacked on one page.
  locales: {
    root: {
      label: '🇫🇷 Français',
      lang: 'fr',
      description: 'Documentation officielle de Kagibi — installation, auto-hébergement, architecture.',
      themeConfig: {
        nav: [
          { text: 'Accueil', link: '/' },
          { text: 'Développement local', link: '/dev-local/' },
          { text: 'Auto-hébergement', link: '/self-hosting/prerequisites' },
        ],
        // Keyed by path prefix rather than one flat array: browsing under
        // /dev-local/ shows only that section's sidebar, browsing under
        // /self-hosting/ shows only the other — the two are fully separate,
        // not two groups sharing a single always-visible tree.
        sidebar: {
          '/dev-local/': [
            {
              text: 'Développement local',
              items: [{ text: 'Sans Docker', link: '/dev-local/' }],
            },
          ],
          '/self-hosting/': [
            {
              text: 'Self-hosting',
              items: [
                { text: 'Pré-requis', link: '/self-hosting/prerequisites' },
                { text: 'Stockage', link: '/self-hosting/storage' },
                { text: 'Nginx', link: '/self-hosting/nginx' },
                { text: 'Mise à jour', link: '/self-hosting/upgrade' },
                { text: 'Sauvegarde', link: '/self-hosting/backup' },
              ],
            },
          ],
        },
        outline: {
          level: [2, 3],
          label: 'Sur cette page',
        },
        footer: {
          message: 'Publié sous licence AGPL-3.0-or-later',
          copyright: 'Copyright © 2025-2026 Buuuntyyy',
        },
      },
    },
    en: {
      label: '🇬🇧 English',
      lang: 'en',
      link: '/en/',
      description: 'Kagibi official documentation — installation, self-hosting, architecture.',
      themeConfig: {
        nav: [
          { text: 'Home', link: '/en/' },
          { text: 'Local development', link: '/en/dev-local/' },
          { text: 'Self-hosting', link: '/en/self-hosting/prerequisites' },
        ],
        sidebar: {
          '/en/dev-local/': [
            {
              text: 'Local development',
              items: [{ text: 'Without Docker', link: '/en/dev-local/' }],
            },
          ],
          '/en/self-hosting/': [
            {
              text: 'Self-hosting',
              items: [
                { text: 'Prerequisites', link: '/en/self-hosting/prerequisites' },
                { text: 'Storage', link: '/en/self-hosting/storage' },
                { text: 'Nginx', link: '/en/self-hosting/nginx' },
                { text: 'Upgrading', link: '/en/self-hosting/upgrade' },
                { text: 'Backup', link: '/en/self-hosting/backup' },
              ],
            },
          ],
        },
        outline: {
          level: [2, 3],
          label: 'On this page',
        },
        footer: {
          message: 'Published under the AGPL-3.0-or-later license',
          copyright: 'Copyright © 2025-2026 Buuuntyyy',
        },
      },
    },
  },

  // Shared across both locales: search indexes every locale automatically,
  // social links and markdown/anchor config aren't language-specific.
  themeConfig: {
    search: {
      provider: 'local',
      options: {
        locales: {
          root: {
            translations: {
              button: { buttonText: 'Rechercher', buttonAriaLabel: 'Rechercher' },
              modal: {
                noResultsText: 'Aucun résultat pour',
                resetButtonTitle: 'Réinitialiser la recherche',
                footer: { selectText: 'sélectionner', navigateText: 'naviguer', closeText: 'fermer' },
              },
            },
          },
        },
      },
    },

    socialLinks: [
      { icon: 'github', link: 'https://github.com/Buuuntyyy/Kagibi' },
    ],
  },
})

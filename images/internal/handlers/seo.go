package handlers

import (
	"github.com/gofiber/fiber/v2"
)

// RobotsTxtHandler serves the robots.txt file
func RobotsTxtHandler() fiber.Handler {
	return func(c *fiber.Ctx) error {
		robotsTxt := `# Lab Nocturne Images - robots.txt
User-agent: *
Allow: /
Allow: /docs
Disallow: /upload
Disallow: /files
Disallow: /stats
Disallow: /key
Disallow: /checkout

# Sitemap
Sitemap: https://images.labnocturne.com/sitemap.xml

# Crawl-delay for respectful crawling
Crawl-delay: 1
`
		c.Set("Content-Type", "text/plain; charset=utf-8")
		return c.SendString(robotsTxt)
	}
}

// SitemapXMLHandler serves the sitemap.xml file
func SitemapXMLHandler() fiber.Handler {
	return func(c *fiber.Ctx) error {
		sitemapXML := `<?xml version="1.0" encoding="UTF-8"?>
<urlset xmlns="http://www.sitemaps.org/schemas/sitemap/0.9"
        xmlns:xhtml="http://www.w3.org/1999/xhtml">
  <url>
    <loc>https://images.labnocturne.com/</loc>
    <lastmod>2025-12-05</lastmod>
    <changefreq>weekly</changefreq>
    <priority>1.0</priority>
  </url>
  <url>
    <loc>https://images.labnocturne.com/docs</loc>
    <lastmod>2025-12-05</lastmod>
    <changefreq>monthly</changefreq>
    <priority>0.8</priority>
  </url>
</urlset>
`
		c.Set("Content-Type", "application/xml; charset=utf-8")
		return c.SendString(sitemapXML)
	}
}

// SiteWebmanifestHandler serves the site.webmanifest file for PWA support
func SiteWebmanifestHandler() fiber.Handler {
	return func(c *fiber.Ctx) error {
		manifest := `{
  "name": "Lab Nocturne Images",
  "short_name": "LN Images",
  "description": "Stupid Simple Image Storage API",
  "start_url": "/",
  "display": "standalone",
  "background_color": "#0a0a0a",
  "theme_color": "#00ff88",
  "icons": [
    {
      "src": "/favicon-16x16.png",
      "sizes": "16x16",
      "type": "image/png"
    },
    {
      "src": "/favicon-32x32.png",
      "sizes": "32x32",
      "type": "image/png"
    },
    {
      "src": "/apple-touch-icon.png",
      "sizes": "180x180",
      "type": "image/png"
    }
  ]
}
`
		c.Set("Content-Type", "application/json; charset=utf-8")
		return c.SendString(manifest)
	}
}

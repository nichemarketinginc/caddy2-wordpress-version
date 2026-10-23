# caddy2-wordpress-version

> **Retired 2026-10-23.** The caddy2 fleet Caddyfile no longer uses `wp_version`.
> `wp_gate` (github.com/nichemarketinginc/caddy2-wordpress-gate) publishes the
> same information as `{http.vars.wp_core_version}` and the `X-WP-Core-Version`
> request header, taken from the site's wp-gate-indexd profile (read from the
> site's own `wp-includes/version.php` on the backend and refreshed within
> seconds of a core flip), and the Caddyfile serves `/wp-includes/*` and
> `/wp-admin/*` from `/var/www/wordpress-sources/{version}`.
>
> Why it was replaced (audit 2026-10-23):
> - it answered **503** when `version.php` could not be read (a non-WordPress
>   site on the `@sitemap` or `:8080` route, an S3 Files hiccup), against the
>   fleet's fail-open rule;
> - it cached the version **24 h in process with no invalidation**; the MU
>   plugin's and the core flip's cache resets target the mongrel2 handler's
>   memcached key, not this cache;
> - it read `version.php` over the S3 Files mount inside `ServeHTTP`, once per
>   host per 24 h and on every request for a host without one;
> - the header it set was dead on every direct `php_fastcgi` route (nothing on
>   the backend reads `HTTP_X_WP_CORE_VERSION`; PHP runs the shared core through
>   the docroot symlinks), so its only live effect was the `{wp-version}`
>   placeholder in the `@sitemap` route's `root`.
>
> The code below is kept for reference and is not built into the fleet binary.

caddy2 middleware to set X-WP-Core-Version header for incoming requests that are meant to be directly sent to php-fpm so that central core wordpress files can be used for such requests

```
wp_version {
  base_path /var/www/vhosts
  wp_version_cache_expiry 24
}
```

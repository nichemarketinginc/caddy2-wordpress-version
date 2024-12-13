# caddy2-wordpress-version
caddy2 middleware to set X-WP-Core-Version header for incoming requests that are meant to be directly sent to php-fpm so that central core wordpress files can be used for such requests

```
wp_version {
  base_path /var/www/vhosts
  wp_version_cache_expiry 24
}
```



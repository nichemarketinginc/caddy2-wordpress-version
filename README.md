# caddy2-validate-vhost-dir
Utilized by the Caddy2 "On Demand TLS" mechanism to determine whether the host is recognized based on the existence of a directory matching that host.

```
on_demand_tls {
    ask http://localhost/internal/validate-vhost-dir
}
...
http://localhost {
  route /internal/validate-vhost-dir {

    root * /var/www/vhosts
    file_server
    validate-vhost-dir /var/www/vhosts
  }
}

```



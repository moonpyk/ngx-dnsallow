# ngx-dnsallow

`ngx-dnsallow` is a companion app for nginx web server that allows to generate IP restrictions based on the result of DNS lookups.
It is particularly useful for restrict/allow access of clients with dynamic IPs that also have some kind of dynamic DNS update mechanism attached to them (DynDNS/No-IP etc.).

## Ideas
 - Possibility to use sudo/doas instead of running as root (become_method, become_user)
 - Profiles support, to generate only subset of rules

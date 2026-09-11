# dxflrs/garage is built FROM scratch (no shell at all — confirmed via
# `docker inspect`: empty Entrypoint, Cmd is ["/garage","server"]), but
# entrypoint.sh needs one to run the local bootstrap sequence before/alongside
# `garage server`. This copies just the garage binary into a minimal
# shell-capable base.
FROM dxflrs/garage:v1.0.1 AS garage

FROM alpine:3.20
COPY --from=garage /garage /usr/local/bin/garage
COPY entrypoint.sh /entrypoint.sh
ENTRYPOINT ["/bin/sh", "/entrypoint.sh"]

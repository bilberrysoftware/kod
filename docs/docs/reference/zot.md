# Useful zot commands

The below shows you how to install zot as your container registry.

- Follow the [install zot on bare metal instructions](https://zotregistry.dev/v2.0.0/install-guides/install-guide-linux/)
- You may need to [create your own root and intermediate Certificate Authority](https://openssl-ca.readthedocs.io/en/latest/create-the-root-pair.html) as skopeo requires a HTTPS TLS protected container registry
  - You can use Elliptic Curve instead of RSA by using this on each key generation stage: `openssl ecparam -out my.key.pem -name secp256r1 -genkey`


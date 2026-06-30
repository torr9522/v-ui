# Certificate Phase 2 Result

## Scope

Phase 2 covered the low-risk manual panel certificate workflow for v-ui while keeping x-ui runtime compatibility:

- manual certificate upload
- `enableHttps`
- `disableHttps`
- listener readiness confirmation after protocol switch
- clean install runtime validation on a real Debian 11 VPS

Runtime freeze remained unchanged:

- service name: `x-ui.service`
- binary name: `x-ui`
- install dir: `/usr/local/x-ui`
- web path: `/xui`

## Delivered

### Manual certificate upload

Implemented:

- upload PEM certificate and private key via `/xui/cert/upload`
- validate PEM format
- validate cert/key match
- write files to:
  - `/usr/local/x-ui/cert/panel.crt`
  - `/usr/local/x-ui/cert/panel.key`
- file permissions:
  - `panel.crt = 0644`
  - `panel.key = 0600`
- persist certificate metadata into settings

### enableHttps

Implemented:

- validate active certificate files before applying
- switch panel startup mode to TLS by setting:
  - `webCertFile`
  - `webKeyFile`
  - `webCertStatus=enabled`
  - `webCertMode=manual`
- trigger panel reload
- wait until TLS listener on `127.0.0.1:<port>` is actually ready before returning success
- rollback settings if HTTPS listener does not become ready

### disableHttps

Implemented:

- clear active TLS file settings:
  - `webCertFile=""`
  - `webKeyFile=""`
- preserve uploaded manual certificate files on disk
- switch status back to uploaded/manual when managed cert files still exist
- trigger panel reload
- wait until HTTP-only listener is actually ready before returning success
- rollback settings if HTTP listener does not become ready

## Listener Wait Fix

The main Phase 2 P0 bug was not in `systemd`, not in a dual-listener leak, and not in runtime naming.

Root cause:

1. When the panel is in HTTPS mode, the custom auto-HTTPS listener returns `307` from plain HTTP requests and redirects them to `https://...`.
2. The original `waitForHTTPReady()` treated any HTTP response as success, so an HTTPS listener returning `307` could be misclassified as “HTTP ready”.
3. After `enableHttps`, frontend certificate actions could still continue over `http://`, so follow-up calls like `disableHttps` might not actually hit the controller.

Fix delivered in commit `67320c0 Fix HTTPS disable readiness detection`:

- `waitForHTTPReady()` now:
  - first checks whether TLS handshake still succeeds on `127.0.0.1:<port>`
  - rejects `301/302/307/308` responses when `Location` points to `https://`
- cert page frontend now:
  - redirects to `https://` after successful `enableHttps`
  - redirects back to `http://` after successful `disableHttps`

## Clean Install Validation

Target VPS:

- `139.180.135.210`
- Debian 11
- domain: `cshtps.527270.xyz`

Validation model:

- full clean install
- no old database restore
- no old certificate restore
- no upgrade path reuse

Clean install result:

- panel installed successfully from current v-ui source
- fresh DB state started with:
  - `httpsActive=false`
  - `webCertStatus=none`
  - `webCertMode=none`
- domain set: PASS
- `checkDomain matched=true`: PASS
- self-signed certificate upload with SAN `cshtps.527270.xyz`: PASS
- `enableHttps`: PASS
- immediate HTTPS access after enable: PASS
- `disableHttps`: PASS after readiness fix
- immediate HTTP access after disable: PASS
- HTTPS no longer worked after disable: PASS

## Stress Test

Five `enableHttps -> disableHttps` rounds were executed on the clean install.

Results:

- round 1: `enable=3070ms disable=5092ms`
- round 2: `enable=3071ms disable=5084ms`
- round 3: `enable=3089ms disable=5101ms`
- round 4: `enable=3060ms disable=5101ms`
- round 5: `enable=3071ms disable=5081ms`

All 5 rounds passed.

Observed runtime behavior:

- `x-ui.service` stayed `active`
- listener switched cleanly between:
  - `web server run https`
  - `web server run http`
- no panic
- no fatal
- no bind/listen error
- no certificate error

## Still Not Implemented

Phase 2 intentionally did not include:

- ACME certificate issuance
- Cloudflare DNS challenge
- ZeroSSL / provider selection
- automatic certificate renewal
- renewal hook / scheduled restart flow

These remain Phase 3+ work.

## Conclusion

Phase 2 is complete.

Delivered and validated:

- manual certificate upload
- `enableHttps`
- `disableHttps`
- readiness wait and rollback behavior
- frontend protocol switch after cert toggle
- clean install runtime validation
- 5-round protocol toggle stress test

Final status:

- `Phase 2 Complete`

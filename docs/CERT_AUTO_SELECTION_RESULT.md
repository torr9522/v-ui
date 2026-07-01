# Cert Auto Selection Result

## Problem Background

z-ui has a better inbound TLS experience: when enabling TLS in the inbound form, users can directly reuse the currently installed panel certificate instead of manually entering certificate paths.

v-ui originally lacked this behavior. Even after `POST /xui/cert/listUsable` was added, the TLS form still did not auto-fill certificate paths.

## Implementation

The final implementation keeps the current v-ui architecture and only improves the inbound TLS form experience:

- `POST /xui/cert/listUsable`
  - returns currently usable panel certificates
  - current scope is the active panel certificate from `manual` or `acme_http`
- inbound TLS form behavior
  - when TLS is enabled in the inbound form, it requests `POST /xui/cert/listUsable`
  - when exactly one usable certificate exists, it auto-fills:
    - `certFile`
    - `keyFile`
  - it also shows certificate summary information in the form:
    - `Panel Certificate`
    - `Let's Encrypt`
    - `Expires`

## Root Cause

The backend API was working correctly.

The real root cause was a frontend `this` context error:

- `POST /xui/cert/listUsable` returned `Panel Certificate` successfully
- but the auto-fill method was invoked on the wrong frontend context
- as a result, the auto-fill method was not actually executed against the current inbound modal state
- therefore these fields remained empty:
  - `inModal.inbound.stream.tls.certs[0].certFile`
  - `inModal.inbound.stream.tls.certs[0].keyFile`
  - `inModal.tlsAutoFillNotice`

## Fix Commits

- `27daefd` Fix inbound TLS certificate autofill
- `2039dae` Fix TLS certificate autofill context

## Remote Validation

Validation server:

- `139.180.135.210`

Validation result:

- auto request `listUsable`: `PASS`
- auto-fill `certFile` / `keyFile`: `PASS`
- display `Panel Certificate` / `Let's Encrypt` / `Expires`: `PASS`
- save `VLESS + TCP + TLS` inbound: `PASS`
- panel-generated share link: `PASS`
- client connectivity to Google / YouTube: `PASS`
- delete test inbound: `PASS`

## Important Deployment Note

`web/html` is embedded into the `x-ui` binary via Go `embed`.

This means:

- changing HTML files alone is not sufficient
- after modifying embedded frontend templates, `x-ui` must be rebuilt and redeployed
- replacing HTML files only does not make the runtime page change take effect

## Conclusion

`Cert Auto Selection: PASS`

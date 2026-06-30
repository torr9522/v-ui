# Certificate Module Freeze

## 1. Final Capability Matrix

- `PASS` Manual certificate upload
- `PASS` HTTPS Enable
- `PASS` HTTPS Disable
- `PASS` Fullchain handling
- `PASS` HTTPS Startup Self-Heal
- `PASS` ACME HTTP-01
- `PASS` ACME Production
- `PASS` ACME Staging
- `PASS` Renew Timer
- `PASS` Renew Service
- `PASS` Manual Skip
- `PASS` Auto Bootstrap

## 2. Live Validation Summary

Target server:

- `139.180.135.210`

Validation rounds:

- Round 1: initial Phase 2 live validation, exposed branding and HTTPS toggle runtime issues, not final pass
- Round 2: redeploy and listener-wait validation, confirmed HTTPS enable/disable runtime behavior after fixes
- Round 3: clean-install validation, confirmed Phase 2 behavior in fresh install state
- Final Production: ACME HTTP-01 production issuance passed, fullchain fix verified, HTTPS without `-k` passed
- Final Renew: renew timer, renew service, `x-ui cert-renew`, and manual skip passed on redeployed `1c95a9b`

Final result:

- Certificate mainline validation on `139.180.135.210`: `PASS`
- Production issuance: `PASS`
- Renew flow validation: `PASS`

## 3. Final Architecture

```text
Browser
  ↓
Panel
  ↓
Cert Service
  ↓
acme.sh
  ↓
systemd timer / renew service
  ↓
x-ui reload or restart
```

## 4. Final Scope Ceiling

The certificate module is now frozen at the current mainline scope.

Do not add in future certificate work:

- Cloudflare DNS-01
- Multiple DNS providers
- ZeroSSL UI
- Multi-certificate repository management
- Certificate asset platform features
- Complex provider abstraction

Allowed scope is limited to the existing x-ui-style mainline:

- manual upload
- HTTPS enable / disable
- ACME HTTP-01
- fullchain
- startup self-heal
- renew timer / renew service

## 5. Only Remaining Bug

### issueHttp Idempotency

If `acme.sh` already has the same domain and returns:

- `Domains not changed`

the current re-issue flow is not idempotent enough.

Observed behavior:

- panel ACME state may not be re-established automatically from existing local `acme.sh` domain state
- a fresh issue attempt can fail even though usable certificate material already exists on the server

Recommended behavior:

- detect existing installed state before issuing
- if the same domain is already installed
- and the current fullchain/key are already valid for the active panel
- return `Already Installed`
- do not re-issue

## 6. Next Fix Plan

No code change is part of this freeze round. The next and only planned certificate follow-up should be:

### IssueHttp Idempotent

Planned steps:

- pre-check current `webDomain`, `webCertMode`, `webCertFile`, `webKeyFile`
- detect whether the active panel already uses the requested ACME domain
- inspect local `acme.sh` domain directory before calling `--issue`
- validate existing fullchain/key pair and active panel binding
- if already installed and active, return `Already Installed`
- skip re-issue
- only call `acme.sh --issue` when no reusable installed state exists

## 7. Freeze Conclusion

Certificate module status:

- `Frozen`

Mainline status:

- `Complete`

Further work should return to core v-ui product features, not certificate expansion.

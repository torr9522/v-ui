# Certificate Known Issues

## Issue 001

- Title: `IssueHttp Idempotent`
- Status: `Open`
- Priority: `P2`
- Summary: when local `acme.sh` already retains the same domain and returns `Domains not changed`, the current `issueHttp` flow is not fully idempotent and should prefer reusing already-installed valid panel certificate state instead of re-issuing.

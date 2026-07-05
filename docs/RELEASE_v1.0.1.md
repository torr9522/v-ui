# v-ui v1.0.1 Stable

## Release Summary

- Version: `v-ui v1.0.1 Stable`
- Release type: stable hotfix
- Runtime layer remains: `x-ui`

## Purpose

This release was created to remove first-install branding residue exposed to end users during clean installation.

Fixed user-visible issues:

- installer log no longer exposes `c-ui` temporary source naming
- installer banner no longer exposes `vc-ui-local`
- runtime compatibility remains unchanged

## Compatibility Boundary

The following runtime surfaces intentionally remained unchanged:

- `x-ui.service`
- `/usr/local/x-ui`
- `/usr/bin/x-ui`
- `/xui`

## Scope

- no new feature work
- no DB schema changes
- no API semantic changes
- no Xray config generator redesign
- release objective was to unblock clean first-user installation acceptance

## Release Mapping

- Git tag: `v1.0.1-stable`
- GitHub Release: `https://github.com/torr9522/v-ui/releases/tag/v1.0.1-stable`
- Follow-up stable release: [RELEASE_v1.0.2.md](./RELEASE_v1.0.2.md)

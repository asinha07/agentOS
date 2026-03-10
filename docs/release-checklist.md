Release Checklist

Pre-flight
- [ ] Ensure README and docs are up to date
- [ ] Update `agent.yaml` versions for built-in agents if needed
- [ ] Confirm CI is green on main

Tag and Release
- [ ] Create a tag: `git tag vX.Y.Z && git push origin vX.Y.Z`
- [ ] Wait for the Release workflow to finish
- [ ] Verify assets on the Release page:
  - [ ] Tarballs for each OS/arch
  - [ ] Checksums
  - [ ] .deb / .rpm
  - [ ] SBOM (`sbom.spdx.json`)
  - [ ] SLSA provenance (attestations)
  - [ ] Homebrew formula was pushed to `asinha07/homebrew-tap`

Post-release
- [ ] Test Homebrew install
- [ ] Test .deb/.rpm on Linux
- [ ] Announce release with changelog


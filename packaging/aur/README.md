# AUR packaging

Publishes the four Heimdall binaries to the [Arch User Repository](https://aur.archlinux.org)
as prebuilt-binary (`-bin`) packages, so Arch users can:

```sh
paru -S heimdall-hub-bin heimdall-dashboard-bin   # monitoring station
paru -S heimdall-daemon-bin                       # each monitored host
paru -S heimdall-helper-bin                       # optional root sidecar
```

Each package installs `/usr/bin/heimdall-<component>` from the matching Linux release
asset (`x86_64` and `aarch64`). It downloads a prebuilt binary — it does **not** compile.
The `-bin` suffix follows AUR guidelines for prebuilt packages and leaves the plain
`heimdall-<component>` names free for a future from-source package.

## How it works

- `gen-pkgbuild.sh` renders one PKGBUILD per component from a release tag and the
  release's `SHA256SUMS`. Generated, never hand-edited.
- The `aur` job in `.github/workflows/release.yml` runs after the binaries are
  attached to a published GitHub Release. For each component it regenerates the
  PKGBUILD (current version + checksums) and pushes it to that package's AUR repo via
  [`KSXGitHub/github-actions-deploy-aur`](https://github.com/KSXGitHub/github-actions-deploy-aur),
  which generates `.SRCINFO` and commits.

The job is **inert by default** — it only runs when `vars.ENABLE_AUR == 'true'`, so
nothing is published until the one-time setup below is done.

> Tag-format assumption: release tags are `vMAJOR.MINOR.PATCH` (e.g. `v1.4.0`).
> `pkgver` is the tag without the leading `v`; the download URL reconstructs `v${pkgver}`.

## One-time setup

1. **AUR account** — create one at <https://aur.archlinux.org> if you don't have it.

2. **SSH key for the AUR** — generate a dedicated keypair (no passphrase, it's used
   headless by CI):

   ```sh
   ssh-keygen -t ed25519 -C "heimdall-aur-ci" -f ~/.ssh/heimdall_aur -N ""
   ```

   Add the **public** key (`~/.ssh/heimdall_aur.pub`) under
   *AUR → My Account → SSH Public Key*.

3. **GitHub secret** — add the **private** key as a repo secret named
   `AUR_SSH_PRIVATE_KEY` (Settings → Secrets and variables → Actions → New secret).
   Paste the full contents of `~/.ssh/heimdall_aur`.

4. **GitHub variable** — add a repo variable `ENABLE_AUR` set to `true`
   (Settings → Secrets and variables → Actions → Variables).

5. **Claim the package names** — the first push to each `ssh://aur@aur.archlinux.org/<pkg>.git`
   creates and claims it. The CI action does this automatically on the first release
   after setup, *provided the names are free*. Check availability:

   ```sh
   for p in hub dashboard daemon helper; do
     git ls-remote "ssh://aur@aur.archlinux.org/heimdall-$p-bin.git" >/dev/null 2>&1 \
       && echo "heimdall-$p-bin: TAKEN" || echo "heimdall-$p-bin: free"
   done
   ```

   If a name is taken by someone else, rename in `gen-pkgbuild.sh` and the workflow
   matrix, or request a merge/orphan via the AUR.

After setup, publishing is automatic: cut a GitHub Release as usual and the four
packages update once the binaries are attached.

## Local dry-run

Generate the PKGBUILDs without publishing — useful to inspect or to bootstrap the
packages by hand the first time:

```sh
# Against a real published release:
curl -fsSL https://github.com/kinncj/Heimdall/releases/download/v1.4.0/SHA256SUMS -o /tmp/SHA256SUMS
VERSION=v1.4.0 SUMS=/tmp/SHA256SUMS OUT=/tmp/aur bash packaging/aur/gen-pkgbuild.sh
cat /tmp/aur/heimdall-dashboard-bin/PKGBUILD

# On an Arch box you can validate end to end:
cd /tmp/aur/heimdall-dashboard-bin
makepkg --printsrcinfo > .SRCINFO   # generate metadata
makepkg -si                          # download, package, install
```

## Manual first publish (optional)

If you'd rather seed the AUR repos by hand instead of letting the first release do it:

```sh
VERSION=v1.4.0 SUMS=/tmp/SHA256SUMS OUT=/tmp/aur bash packaging/aur/gen-pkgbuild.sh
for p in hub dashboard daemon helper; do
  git clone "ssh://aur@aur.archlinux.org/heimdall-$p-bin.git" "/tmp/aur-repo-$p"
  cp "/tmp/aur/heimdall-$p-bin/PKGBUILD" "/tmp/aur-repo-$p/"
  ( cd "/tmp/aur-repo-$p" && makepkg --printsrcinfo > .SRCINFO \
    && git add PKGBUILD .SRCINFO && git commit -m "Initial import v${VERSION#v}" && git push )
done
```

# Releasing Orchard

Publishing a GitHub release with a tag like `v0.1.0` runs
`.github/workflows/release.yml`. The tag sets the version. The workflow tests
(`go vet` and `go test`), builds every package and attaches them to the
release. The workflow is written by the Giode CLI from `giode.toml`: after
changing `[ci]` or a store section, run `giode ci -f` instead of editing it,
then redo the two hand edits noted in `[ci]` (pin the giode CLI version and
use `go-version-file: go.mod`).

| Where | What the release gets | Needs |
| --- | --- | --- |
| GitHub, Linux | `.deb` and `.rpm`, x64 and ARM64 | nothing |
| GitHub, Windows | `setup.exe` installers, x64 and ARM64 | nothing; a code signing certificate stops SmartScreen warnings |
| GitHub, macOS | a universal `.dmg` | a Developer ID certificate, or macOS refuses to open it |
| Microsoft Store | `.msix` packages, x64 and ARM64, to upload in Partner Center | a Partner Center account and the reserved name |

Orchard is not offered in the **Mac App Store** (see below), and it has no
Android or iOS version, so there is no Google Play or iOS App Store package.
There is no Snap Store or Flathub package either: Linux users download from
the GitHub release.

Privacy policy URL for every store: https://xmhrnx.github.io/orchard/#privacy

## Repository secrets

Set these in GitHub under Settings → Secrets and variables → Actions. Each
`*_BASE64` secret is the file encoded with `base64 -i file | pbcopy` on macOS,
or `base64 -w0 file` on Linux. Without a secret, that step is skipped and the
rest of the release still builds.

| Secret | What it is |
| --- | --- |
| `MACOS_CERT_BASE64`, `MACOS_CERT_PASS` | The "Developer ID Application" certificate as a `.p12`, and its password: signs the `.dmg` |
| `GIODE_NOTARY_ID`, `GIODE_NOTARY_TEAM_ID`, `GIODE_NOTARY_PASS` | Your Apple ID, team ID and an app-specific password: notarizes the `.dmg` |
| `WINDOWS_CERT_BASE64`, `GIODE_WINDOWS_CERT_PASS` | An Authenticode `.pfx` and its password: signs the Windows installer (optional) |

The MSIX packages need no secret: the Microsoft Store signs them.

## macOS (.dmg on GitHub)

1. Join the Apple Developer Program.
2. In Certificates, Identifiers & Profiles, create a "Developer ID
   Application" certificate, export it from Keychain Access as a `.p12`, and
   add it as `MACOS_CERT_BASE64` and `MACOS_CERT_PASS`.
3. Create an app-specific password at account.apple.com and add
   `GIODE_NOTARY_ID` (your Apple ID), `GIODE_NOTARY_TEAM_ID` and
   `GIODE_NOTARY_PASS`.
4. Publish a release. It gets `Orchard-<version>-universal.dmg`, signed and
   notarized, so Gatekeeper opens it.

### Why there is no Mac App Store version

Every Mac App Store app runs in the App Sandbox, and so do the programs it
starts. Orchard does all of its work by running the `git` command-line tool,
and in the sandbox that can't work:

- `/usr/bin/git` is a shim that runs the Xcode command-line tools through
  `xcrun`, which refuses to run inside an App Sandbox.
- git from Homebrew, MacPorts or the git installer lives outside the
  sandbox's reach, and isn't on a GUI app's `PATH` anyway.
- git couldn't read `~/.gitconfig`, `~/.ssh` keys or the user's credential
  helpers, so SSH remotes and private HTTPS remotes would fail, and it could
  only touch repositories the user picks one by one.
- The store doesn't allow shipping a separate git, or downloading one, to
  work around this (guideline 2.4.5).

Making it work would mean replacing git with a built-in implementation,
which is a rewrite of the app. The notarized `.dmg` has none of these
limits. `giode.toml` has no `[macos.store]` section for this reason, so the
workflow builds no `.pkg`.

## Microsoft Store

1. Open a Partner Center developer account (the one that already holds
   EasySend and AnyFile).
2. Reserve the name "Orchard" (Apps and games → New product → MSIX or PWA
   app). If it's taken, reserve another name and change `app.name`.
3. In the product's Product management → Product identity page, check these
   against `giode.toml` and correct them if they differ:
   - **Package/Identity/Name** → `identity_name` under `[windows.msix]`
     (now `MehranZahir.Orchard`).
   - **Package/Identity/Publisher** (`CN=...`) → `publisher` under
     `[windows.msix]` (now `CN=B40E8F57-A7E0-40BB-8E74-5D051D0B578B`).
   - **Package/Properties/PublisherDisplayName** → `publisher_display_name`
     under `[windows.msix]` (now `Mehran Zahiroddini`).
4. Commit and publish a release. It gets `Orchard-<version>-amd64.msix` and
   `Orchard-<version>-arm64.msix`; upload both in a Partner Center
   submission. The Store signs them.
5. Fill in the submission (copy below):
   - **Privacy policy URL:** https://xmhrnx.github.io/orchard/#privacy
   - **Category:** Developer tools.
   - **Age rating questionnaire:** no violence, sexual content, gambling,
     user-to-user communication, location sharing or purchases. The app has
     unrestricted web access only in the sense that links open in the
     browser; it shows text from the user's own git hosts (commit messages,
     release notes). Expect the lowest rating (IARC 3+ / Everyone).
   - **System requirements / additional info:** say that Orchard needs Git
     for Windows installed (https://git-scm.com). The Store requires apps to
     disclose dependencies they don't install themselves.
   - **Restricted capabilities:** Orchard is a full-trust desktop app. When
     asked why it uses `runFullTrust`, explain that it runs the user's
     installed `git` on repository folders anywhere on the computer, opens a
     terminal in them, and keeps account tokens in Windows Credential Manager.
   - **Screenshots:** see below.

## Store listing copy

Accurate to the code; paste as is.

**Name:** Orchard

**Short description (80 characters at most):**
Keep every git repository on your computer up to date, with one click.

**Description:**

Orchard keeps every git repository on your computer in one place. Add the
ones you already have, scan a folder to find them all, or clone new ones, then
update one, a few or all of them with a click instead of visiting each folder.

Your whole orchard at a glance
• Every repository with its branch, incoming and outgoing commits, local changes and problems.
• Sections by host (GitHub, GitLab, your own server, local), by your own groups, or none.
• Filter by state, search by name, path, branch or group; pin and group repositories.
• Update all or fetch all in parallel, with local changes stashed around the pull if you want.
• Background fetch keeps the counts fresh.

Any host that speaks git
• Clone by URL from GitHub, GitLab, Gitea/Forgejo/Codeberg, Bitbucket, Azure DevOps or any server.
• Connect GitHub, GitLab, Gitea/Forgejo and Bitbucket accounts (public or self-hosted) to browse your repositories and clone many at once. Tokens stay in Windows Credential Manager.

Inside a repository
• History: a commit graph across all branches with search, cherry-pick, revert and reset.
• Changes: stage, unstage, discard, commit, amend, commit and push.
• Branches and tags: switch, merge, rebase, prune, and delete many at once, on the server too.
• Releases, pull requests and CI workflows from GitHub, GitLab, Gitea/Forgejo and Bitbucket.
• Rewrite: remove files or leaked secrets from all of history, fix an author, squash, reword.

Orchard runs the Git installed on your computer (Git for Windows), with your own SSH keys and credential helpers. It has no servers of its own, no analytics and no tracking.

**Keywords / search terms:** git, git client, repositories, GitHub, GitLab,
Bitbucket, Gitea, pull, fetch, developer tools

**Category:** Developer tools

**Data / privacy answers:** Orchard collects no data. Repository paths,
connected account names and settings stay in a file on the computer; tokens
stay in the system credential store; the app talks only to the git hosts the
user works with, never to its developer.

**Export compliance:** Orchard uses only the operating system's HTTPS and
git's own SSH/HTTPS, standard encryption that needs no export documentation
(`ITSAppUsesNonExemptEncryption = false` is set for Apple builds).

## Screenshots

From `docs/screenshots/` (the same ones as the website), in this order:

1. `repositories.png`: the repository list grouped by host
2. `history.png`: the commit graph
3. `releases.png`: releases with their notes
4. `workflows.png`: CI runs
5. `tags.png`: deleting many tags at once

These are 1360×860. Partner Center asks for desktop screenshots of at least
1366×768, so retake them with the window a little wider (1400×900, say)
before the first Store submission.

## Known store limitations

- **Git must be installed.** Neither the Store nor the GitHub packages
  include git. Without it, every repository operation fails with an error,
  so the listing must say git is required.
- **MSIX file locations:** in the Store version, Windows keeps Orchard's
  settings file inside the package's own data folder instead of `%AppData%`,
  and removes it when the app is uninstalled.
- **Mac App Store:** not offered, see above.

## Linux

The `.deb` and `.rpm` are attached to the release, and the website's download
button links to the release page. They list the system libraries they need,
so `apt` and `dnf` install those too. Users install git themselves.

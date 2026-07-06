# Homebrew distribution

Two ways to ship Sync Pathguard via Homebrew. **Start with your own tap** — it
needs no external approval and you control everything. The official
`homebrew-cask` repo is a later step and effectively requires a **signed +
notarized** app (see T-0015).

## Your own tap (recommended)

1. Create a public repo named **`homebrew-tap`** under your account
   (`github.com/msjang/homebrew-tap`). The `homebrew-` prefix is what makes
   `brew tap msjang/tap` work.
2. Copy these files into it:
   - `Casks/sync-pathguard.rb`  → the menu-bar app
   - `Formula/pathguard.rb`     → the CLI
3. Cut a release here (`git tag v0.1.0 && git push --tags`) so the assets exist.
4. Fill in each file's `version` and `sha256`. Get the hashes with:

   ```bash
   shasum -a 256 Sync-Pathguard-macos-arm64.zip Sync-Pathguard-macos-amd64.zip \
                 pathguard-macos-arm64.zip pathguard-macos-amd64.zip \
                 pathguard-linux-amd64.tar.gz pathguard-linux-arm64.tar.gz
   ```
5. Commit to the tap. Users then install:

   ```bash
   brew tap msjang/tap
   brew install --cask sync-pathguard   # menu-bar app
   brew install pathguard               # CLI
   ```

The cask runs `xattr -dr com.apple.quarantine` after install so the **unsigned**
app launches without the "damaged / unidentified developer" block. Once the app
is notarized, drop that `postflight`.

> Automating the tap bump on each release (updating version + sha256) is a
> follow-up — a workflow can open a PR to the tap. For now it's a manual edit.

## Official homebrew-cask (later)

Requires: a **notarized** `.app`, stable versioned download URLs (the release
provides these), and passing `brew audit --new --cask sync-pathguard`. Then open
a PR to `Homebrew/homebrew-cask`. Do this after code signing lands (T-0015).

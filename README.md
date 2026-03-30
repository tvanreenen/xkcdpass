# xkcdpass

![xkcd #936: Password Strength](https://imgs.xkcd.com/comics/password_strength.png)

[Original comic](https://xkcd.com/936/) by Randall Munroe (xkcd, CC BY-NC 2.5).

`xkcdpass` is a small Go CLI that generates xkcd-style passphrases from the embedded EFF large wordlist using cryptographically secure randomness.

By default, `xkcdpass` prints **4 random words from the list run together** with no delimiter—the same shape as the password in the comic (`correcthorsebatterystaple`). Use `--separator` if you want hyphens, spaces, or another delimiter for readability. Four words is a familiar default, but it is not the strongest setting this tool supports. Bit estimates below assume the wordlist and word count are public; security rests in the random words, not in hiding which tool you used.

## Security model

- **Entropy source:** Go's `crypto/rand`, backed by the operating system CSPRNG—not `math/rand`, which is unsuitable for secrets.
- **Uniform words:** Each word is an independent, uniformly random draw from the full embedded list (the same word may appear more than once). Indices come from `crypto/rand.Int`, which avoids modulo bias you get when shrinking random bytes with `% len(list)` and the list length is not a power of two.
- **Flat phrase space:** No templates, grammatical rules, or frequency weighting—those steer output toward a smaller or more predictable set of phrases than uniform words from the whole list.
- **Formatting:** Lowercase words are fixed; an optional separator between words does not add entropy.

### Passphrase strength

The EFF large wordlist is public, so strength is quoted the way passphrase tools usually do: treat the list and the number of words as known, and measure what is left to guess—the specific sequence of words. Each independent uniform word adds about 12.9 bits (`log2(7776)`). For *n* words, that is *n* times as much altogether (`log2(7776^n)` bits).

Examples with the embedded 7,776-word list and a known word count:

- 4 words: about 51.7 bits
- 6 words: about 77.6 bits

Use `--words 6` for stronger real-world usage.

## Installation

### Homebrew

Once the formula is published to your tap:

```sh
brew tap tvanreenen/tap
brew install xkcdpass
```

### Build locally

```sh
go build ./cmd/xkcdpass
```

## Usage

```sh
xkcdpass
xkcdpass --words 6
xkcdpass --separator -
xkcdpass --separator _
```

Sample output (default: no separator):

```text
washroomunquotedzebravelocity
```

Hyphenated (optional; easier to read or type):

```sh
xkcdpass --separator -
```

```text
washroom-unquoted-zebra-velocity
```

## Development

This repository uses `just` plus a small shell script for local workflows.

```sh
just test
just build
just run -- --words 6
```

## Releasing

Run the full release pipeline in one command:

```sh
just release v0.1.0
```

`just release` runs three stages in order:

1. `release-build`: builds `darwin/arm64` (Apple Silicon) and `linux/amd64` binaries, then writes versioned `tar.gz` archives and `checksums.txt` to `dist/`.
2. `release-publish`: creates a published GitHub release and uploads those artifacts.
3. `release-tap`: updates `~/Code/homebrew-tap/Formula/xkcdpass.rb`, then commits and pushes that tap repo.

If a later stage fails, rerun only that stage:

```sh
just release-publish v0.1.0
just release-tap v0.1.0
```

You can also run build-only when iterating:

```sh
just release-build v0.1.0
```

`release-tap` assumes the tap formula exists at `~/Code/homebrew-tap/Formula/xkcdpass.rb`. Intel Macs and other arches need `go build` from source.

## Wordlist attribution

The embedded wordlist is based on the EFF large wordlist for passphrases:

- Source: [EFF Large Wordlist for Passphrases](https://www.eff.org/document/passphrase-wordlists)
- File: [eff_large_wordlist.txt](https://www.eff.org/files/2016/07/18/eff_large_wordlist.txt)

See [THIRD_PARTY_NOTICES.md](THIRD_PARTY_NOTICES.md) for attribution details.

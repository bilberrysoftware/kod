# Installation

This guide walks you through installation of the `kod` CLI. It won't take long...

1. First, grab the latest release from [https://github.com/bilberrysoftware/kod/releases](https://github.com/bilberrysoftware/kod/releases)
2. Unzip / tar this archive `tar xzf kod-*.tar.gz`
3. (Optional) Install system wide `sudo install ./kod-*/kod /usr/local/bin`
4. Test the command `kod version` (or `./kod version` if not installed system wide)

You're now ready to go! Visit the [Quick Start Guide](quickstart.md) for next steps.

## A Note on Windows Support

We [don't support Windows](about.md#why-dont-you-support-windows) for all operations. You can unpack, but you cannot package or deploy.
This is because a dependency we use to package container images, `skopeo`, does not yet support Windows.
When it supports Windows, the Windows `kod.exe` in the above package should work out of the box straight away.
That's why we include it with each release.

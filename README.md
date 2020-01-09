# mmv [![CI Status](https://github.com/itchyny/mmv/workflows/CI/badge.svg)](https://github.com/itchyny/mmv/actions)
Rename multiple files using your `$EDITOR`. The command name is named after _multi-mv_.

![](https://user-images.githubusercontent.com/375258/72040421-d4f8cd00-32eb-11ea-828f-d9f14f3261ac.gif)

## Usage
```bash
mmv file ...
```
This command opens the editor with the list of file names so edit and write.
The command finds the changed lines and renames all the corresponding files.

## Installation
### Homebrew
```sh
brew install itchyny/tap/mmv
```

### Build from source
```bash
go get github.com/itchyny/mmv/cmd/mmv
```

## Features
- `mmv` is implemented in Go language and completely portable.
- `mmv` is designed to be simple as `mv`. It requires no configuration file.
- `mmv` supports renaming in cycle (`mv a b`, `mv b c` and `mv c a` at the same time).
- `mmv` creates destination directories automatically. You can arrange pictures like `yyyy-mm-dd xxxx.jpg` to `yyyy/mm/dd/xxxx.jpg`.
- `mmv` is capable to use as a library (just call `mmv.Rename`).
- `mmv` is easy to remember (I believe), **m**ulti-**mv**.

## Bug Tracker
Report bug at [Issues・itchyny/mmv - GitHub](https://github.com/itchyny/mmv/issues).

## Author
itchyny (https://github.com/itchyny)

## License
This software is released under the MIT License, see LICENSE.

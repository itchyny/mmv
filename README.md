# mmv [![CI Status](https://github.com/itchyny/mmv/workflows/CI/badge.svg)](https://github.com/itchyny/mmv/actions)
Rename multiple files using your `$EDITOR`. The command name is named after _multi-mv_.
## Usage
```bash
mmv [files] ...
```
This command opens the editor with the list of file names so edit and write.
The command finds the changed lines and renames all the corresponding files.

## Installation
```bash
go get -u github.com/itchyny/mmv/cmd/mmv
```

## Bug Tracker
Report bug at [Issues・itchyny/mmv - GitHub](https://github.com/itchyny/mmv/issues).

## Author
itchyny (https://github.com/itchyny)

## License
This software is released under the MIT License, see LICENSE.

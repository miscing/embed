# embed

Simple program that embeds target files and/or directories into current directory go package source code. It generates a file containing a function that returns a []byte. Files are packed into a tar if more than one file is present, otherwise the file is encoded as is. This allows targeting prepackaged tar files without specific checks, but means that programs need to be aware if the file is NOT a tar file.

Note that each argument passed to embed is walked, thus you can add multiple directories at once.

To use the data in the program call `bindata()`, which returns a []byte copy of data. Generally you will then use a tar reader to read it.

Personally I used embed with the `go generate` command on a separate sub-package of my intended package and place handling logic for assets there.

See `embed -h` for details.

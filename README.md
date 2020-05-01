# embed

Simple program that embeds target file/s or directory into current directories go package source code. It generates a file containing a function that returns a []byte. Files are packed into a tar if more than one file is present, otherwise the file is encoded as is. This allows targeting prepackaged tar files without specific checks, but means that programs need to be aware if the file is NOT a tar file.

To use the data in the program call bindata(), which returns a []byte copy of data. Generally you will then use a tar reader to read it.

See `embed -h` for details.

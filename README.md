# embed

Simple commandline tool that embeds target file or directory into current directory golang package source code. It generates a file containing a []byte variable by file name. Files are packed into a tar if more than one file is present, otherwise the file is encoded as is.

To use the bindata call bindata(), which returns a []byte copy of data

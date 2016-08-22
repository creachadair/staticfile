# Design notes

The file data compiler reads one or more files and converts them into a package
of Go source code.

When imported, the generated package registers its files with the filedata
package by pathname. To open a file, import "filedata" and open the desired
path:

    f, err := filedata.Open("path/to/my.txt")

The resulting object behaves like a regular read-only file.

I chose this approach rather than having the generated package stand alone and
be imported everywhere it's used to avoid having to rebuild all the dependent
packages every time the generated package changes. With registration, only the
main program needs to be recompiled.

The file data compiler can be used via "go generate" to avoid having to check
in the generated data packages.

TODO: Show an example of this.

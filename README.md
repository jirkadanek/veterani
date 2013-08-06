There is nothing really useful for the "general public." In `src/veterani2013/iof` there is a partial typedefinition for the annotated structs that can be used to parse the IOF 2.x format used at http://oris.orientacnisporty.cz/. Othervise it is pretty boring text processing. It would be probably better done in someting like Perl or Python. The code is heavy on sorting and that is pretty painful with Go sort package.

The code uses SQLite. I had fun times choosing from the three available bindings for Go, at the end I went with what seemed the most popular.

Usage:

        go run src/veterani2013/zpracuj_data.go -clubs clubs.txt -results 2013/ -suffix .txt
        go run src/veterani2013/vypis_vysledky.go
        go run src/veterani2013/vypis_v_kategorie.go

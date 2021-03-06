package main

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"sort"
	"strconv"
	"strings"
	"veterani2013/sql"
	"veterani2013/types"
)

type vysledky struct {
	D []vysledek
	H []vysledek
}

type vysledky_po_kategoriich struct {
	kategorie []vysledek_po_kategoriich
}

type vysledek_po_kategoriich struct {
	Kategorie string
	vysledky  []vysledek
}

type vysledek struct {
	Pkat       int
	Pcpv       int
	Pabs       int
	Id         string
	Jmeno      string
	Nzav       int
	BodyCelkem int
	Body       []int
}

func main() {
	db := sql.NewDb("db.sqlite", false)
	vypis_vysledky(db)
	db.Stop()
}

var HIGHLIGHT int = 10
var COLS int = 35
var YEAR string = "2018"

func vypis_vysledky(db sql.Db) {
	r := db.Getresults()

	classes := make(map[string]bool)
	for _, v := range r {
		cls := categoryFromIDAge(v.Kategorie, v.Z_id)
		classes[cls] = true
	}

	// only classes someone competed in
	sclasses := make([]string, 0)
	for c, _ := range classes {
		sclasses = append(sclasses, c)
	}
	sort.Strings(sclasses)

	f, err := os.Create(fmt.Sprintf("hodnoceni_cpv_%s_dle_kategorii.html", YEAR))
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()
	fmt.Fprintf(f, `<!DOCTYPE html>
<html>
<head>
<meta charset="utf-8">
</head>
<body><pre>`)
	fmt.Fprintf(f, "Pořadí Českého Poháru Veteránů %v dle kategorií\n\n", YEAR)
	for _, k := range sclasses { // over all categories
		fmt.Fprintf(f, "     KATEGORIE %s\n", k)
		fmt.Fprintf(f, "poř.kat.  poř.ČPV abs.poř.  Reč         Jmeno             počet  body     body dle závodů\n")
		fmt.Fprintf(f, "                                                         závodů celkem  ")
		for i := 1; i <= COLS; i++ {
			fmt.Fprintf(f, "%2d.", i)
		}
		fmt.Fprintf(f, "\n")

		cntr := 1
		prevporadi := 0
		prevkatporadi := 0
		for _, l := range r { // over all runners
			classstr := categoryFromIDAge(l.Kategorie, l.Z_id)
			//fmt.Println(classstr, k)
			if classstr != k {
				continue // skip the runner
			}

			katporadi := prevkatporadi
			if l.Ap_poradi != prevporadi {
				katporadi = cntr
			}

			races := db.Getraceresults(l.Z_id) // map[int]int
			sraces := racesTable(l, races)

			fmt.Fprintf(f, "%7d %6d %6d %10s %-25s %2d %7d    %s\n",
				katporadi, //db.Getkatporadi(l.Z_id, k), //l.Kp_poradi,
				l.P_poradi,
				l.Ap_poradi,
				strings.Replace(l.Z_id, "|", "", -1),
				l.Z_prijmeni+", "+l.Z_jmeno,
				l.Nzavodu,
				l.S_body,
				sraces.String())
			prevkatporadi = katporadi
			prevporadi = l.Ap_poradi
			cntr++
		}
		fmt.Fprintf(f, "\n")
	}
	fmt.Fprintf(f, `</pre>
			</body>
			</html>`)
}

func categoryFromIDAge(Kategorie, Z_id string) string {
	classB, err := types.NewRegno(strings.Replace(Z_id, "|", "", -1)).ClassB()
	if err != nil {
		log.Fatal(err)
	}
	return fmt.Sprintf("%s%d",
		strings.Split(Kategorie, ",")[0][:1], // pohlavi
		classB)
}

func racesTable(l sql.Result, races map[int]int) *bytes.Buffer {
	limit := 0 // max(keys in races)
	for i, _ := range races {
		if limit < i {
			limit = i
		}
	}

	highlight := make([]bool, limit+1)
	ss := strings.Split(l.Ap_scores, ",")
	is := make([]int, 0)
	for _, s := range ss {
		convs, err := strconv.ParseInt(s, 10, 32)
		if err != nil {
			log.Fatal(err)
		}
		is = append(is, int(convs))
	}
	sort.Sort(sort.Reverse(sort.IntSlice(is)))

	hs := HIGHLIGHT
	if len(is) > hs {
		is = is[:hs]
	}
	for i := limit; i > 0; i-- {
		if hs > 0 {
			v, hasrace := races[i]
			if !hasrace {
				continue
			}
			ishigh := false
			for j, val := range is {
				if val == v {
					ishigh = true
					is[j] = -1 // do not use again
					break
				}
			}
			if ishigh {
				highlight[i] = true
				hs--
			}
		}
	}

	sraces := new(bytes.Buffer)
	i := 1
	for ; i <= limit; i++ {
		var com string
		if i == limit {
			com = ""
		} else {
			com = ","
		}

		var val string
		v, found := races[i]
		if !found {
			val = "  "
		} else {
			if highlight[i] {
				val = fmt.Sprintf("<b>%2d</b>", v)
			} else {
				val = fmt.Sprintf("%2d", v)
			}
		}

		fmt.Fprintf(sraces, "%s%s", val, com)
	}
	for ; i <= COLS; i++ {
		fmt.Fprintf(sraces, ",  ")
	}
	return sraces
}

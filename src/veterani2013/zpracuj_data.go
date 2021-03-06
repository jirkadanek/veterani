package main

import (
	// 	"bufio"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	// 	"os"
	"path"
	"strconv"
	"strings"

	"veterani2013/bodovani"
	"veterani2013/input"
	// 	"veterani2013/iof"
	"veterani2013/sql"
	"veterani2013/types"
)

type Kategorie struct {
	a string
	b int
	c string
}

func NewKategorie(s string) Kategorie {
	//log.Println(s)

	if len(s) == 3 || len(s) == 4 {
		var a, c string
		var b int

		a = s[0:1]
		if a != "H" && a != "D" {
			goto divnyformat
		}

		conv, err := strconv.ParseInt(s[1:3], 10, 32)
		if err != nil {
			goto divnyformat
		}
		b = int(conv)
		if len(s) == 4 {
			c = s[3:4]
			switch c {
			case "A", "B", "C", "D", "E":
				// nedelej nic
			default:
				goto divnyformat
			}
		}
		return Kategorie{a, b, c}
	}

divnyformat:
	return Kategorie{s, 0, ""}
}

func nacti_zavody(dir, suffix string) []input.Zavod {
	zavody := make([]input.Zavod, 0)
	fi, err := ioutil.ReadDir(dir)
	if err != nil {
		log.Fatal(err)
	}
	for _, j := range fi {
		if j.IsDir() {
			continue
		}
		if path.Ext(j.Name()) != suffix {
			continue
		}

		z := input.Nacti_zavod(dir, j.Name())
		log.Printf("%#v\n", z)
		zavody = append(zavody, z)
	}
	return zavody
}

var clubs = flag.String("clubs", "", "Seznam platných oddílů")
var results = flag.String("results", "", "Složka se soubory s výsledky")
var suffix = flag.String("suffix", "", "Přípona souborů s výsledky, včetně tečky")

func main() {
	flag.Parse()
	db := sql.NewDb("db.sqlite", true)
	//db := sql.NewDb(":memory:")
	naplndb(db, *clubs, *results, *suffix)
	zpracujdb(db)
	//celkove, kategorie := vysledky(db)
	db.Stop()
}

func zpracujdb(db sql.Db) {
	db.Vypocitejsoucty()
	db.Celkoveporadi()
	db.Poradi()
	db.Katporadi()
}

func naplndb(db sql.Db, foddily, dzavody, suffix string) {
	db.Createtables()
	// nacist oddily
	oddily := input.Nacti_oddily(foddily)
	log.Printf("Oddilu: %d\n", len(oddily))
	log.Println("-------------")
	// nacist zavod po zavode
	vysledky := nacti_zavody(dzavody, suffix)
	log.Println("-------------")
	for _, vysledek := range vysledky {
		switch suffix {
		case ".txt":
			FromCsoc(db, oddily, vysledek)
		case ".xml":
			FromXml(db, oddily, vysledek)
		}
	}
}

func FromCsoc(db sql.Db, oddily map[string]bool, vysledek input.Zavod) {
	csos := input.ReadCsos(vysledek.Fname)

	cs := make(map[types.Class]bool) // all classes

	for _, r := range csos {
		cs[r.Class] = true
	}

	rs := input.GetValid(csos, oddily, vysledek) // valid results

	nconts := make(map[types.Class]int) // number of classified runners
	for _, r := range rs {
		nconts[r.Class] += 1
	}

	//recount positions
	var prevp int
	var prevclass types.Class
	var prevrposition int
	for _, r := range rs {
		// different class
		if prevclass != r.Class {
			prevp = 0
			prevrposition = 0
		}
		// don't advance on ties
		p := prevp
		if r.Position != prevrposition {
			p++ // counting from 1
		}

		classrank := bodovani.SubClassRank(cs, r.Class)

		if r.FamilyGiven == "Gryc Jan" {
			log.Println(vysledek.Cislo, r.Class, classrank, p, nconts[r.Class])
		}

		b := bodovani.Score(classrank, p, nconts[r.Class], vysledek.Veteraniada)

		id := fmt.Sprintf("%s|%s", r.Regno.C, r.Regno.N)
		pts := strings.SplitN(r.FamilyGiven, " ", 2)
		family := pts[0]
		given := ""
		if len(pts) > 1 {
			given = pts[1]
		}

		db.Pridejvysledek(id, given, family, vysledek.Cislo, b, r.Class.A, r.Class.B)
		prevp = p
		prevclass = r.Class
		prevrposition = r.Position
	}
	fmt.Println()
}

func FromXml(db sql.Db, oddily map[string]bool, vysledek input.Zavod) {
	panic("broken: cislovani cizincu")
	// 	zavod := iof.Nacti_zavod(vysledek.Fname)
	//
	// 	kategorie := make(map[types.Class]bool)
	// 	for _, r := range zavod.Results {
	// 		kategorie[types.NewClass(r.Category)] = true
	// 	}
	//
	// 	for k, _ := range kategorie {
	// 		fmt.Printf("%v,", k)
	// 	}
	// 	fmt.Printf("\n")
	//
	// 	log.Printf("%#v\n", zavod.Event)
	// 	log.Printf("Kategorii: %d\n", len(kategorie))
	// 	log.Println("-------------")
	// 	for _, r := range zavod.Results {
	// 		log.Println(r.Category)
	//
	// 		kat := types.NewClass(r.Category)
	// 		katno := bodovani.SubClassRank(kategorie, types.NewClass(r.Category))
	// 		klaszav := len(r.PersonResults)
	//
	// 		log.Printf("Kategorie: %+v, Rank: %d, Zavodniku: %d", kat, katno, klaszav)
	// 		log.Println("-------------")
	//
	// 		for _, p := range r.PersonResults {
	// 			id := fmt.Sprintf("%s|%s", p.Person.Country, p.Person.Id)
	// 			umisteni := p.Result.Position
	//
	// 			if !input.IsValid(oddily, types.Regno{C: p.Person.Country, N: p.Person.Id}, types.NewClass(r.Category),
	// 				p.Result.Position, p.Result.Status.Value) {
	// 				continue
	// 			}
	//
	// 			b := bodovani.Score(katno, umisteni, klaszav, vysledek.attr)
	//
	// 			db.Pridejvysledek(id, p.Person.Name.Given, p.Person.Name.Family, vysledek.cislo, b, kat.A, kat.B)
	// 		}
	// 	}
}

// type vysledky struct {
// 	D []vysledek
// 	H []vysledek
// }
//
// type vysledky_po_kategoriich struct {
// 	kategorie []vysledek_po_kategoriich
// }
//
// type vysledek_po_kategoriich struct {
// 	Kategorie string
// 	vysledky  []vysledek
// }
//
// type vysledek struct {
// 	Pkat       int
// 	Pcpv       int
// 	Pabs       int
// 	Id         string
// 	Jmeno      string
// 	Nzav       int
// 	BodyCelkem int
// 	Body       []int
// }
//
// func vypis_vysledky() {
// 	/* fmt.Println("Český pohár veteránů 2013")
// 	fmt.Println("")
// 	fmt.Println("D")
// 	fmt.Println("pořadí ČPV")
// 	*/
// }

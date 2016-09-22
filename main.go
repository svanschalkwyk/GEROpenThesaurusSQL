package main

import (
	"strings"
	"github.com/tealeg/xlsx"
	_ "github.com/go-sql-driver/mysql"
	"database/sql"
	"github.com/bbalet/stopwords"
	"log"
)

type Keyword struct {
	term string
	subterms []string
	synonyms [][]string
}
var keywords []Keyword

const readExcelFileName = "/home/steph/Downloads/ENSCIGHT/synononyms/GER_Search_Label_Translations.xlsx"
const writeExcelFileName = "/home/steph/Downloads/ENSCIGHT/synononyms/GER_Search_Label_Translations_1.xlsx"
const sqlstatement = "SELECT term.word FROM term, synset, term term2 WHERE synset.is_visible = 1 AND synset.id = " +
	"term.synset_id AND term.synset_id AND term2.synset_id = synset.id AND term2.word = ?"

func splitKeywords(s string) []string  {
	w := strings.FieldsFunc(s, func(r rune) bool {
		switch r {
		case '<', '>', ' ', '|', '&', '/', '\'', '(', ')','[',']',',',';',':','-':
			return true
		}
		return false
	})
	return w
}

func write_results() {
	var row *xlsx.Row
	var cell *xlsx.Cell

	excelFile := xlsx.NewFile()//.OpenFile(excelFileName)

	sheet,_ := excelFile.AddSheet("Synonyms")
	row = sheet.AddRow()
	cell = row.AddCell()
	cell.Value= "Term"
	cell = row.AddCell()
	cell.Value = "Synonyms_1"
	cell = row.AddCell()
	cell.Value = "Synonyms_2"
	for _, kw := range keywords {
		row := sheet.AddRow()
		cell = row.AddCell()
		cell.Value = kw.term;
		for _, syns := range kw.synonyms {
			cell := row.AddCell()

			for _, syn := range syns {
				//log.Println(syn)
				if cell.Value != "" {
					cell.Value = cell.Value + "," + syn
				} else {
					cell.Value = syn
				}
			}
		}
	}
	excelFile.Save(writeExcelFileName)
}

func get_keywords() {
	excelFile, err := xlsx.OpenFile(readExcelFileName)
	if err != nil {
		return
	}

	sheet := excelFile.Sheets[0]
	for _, row := range sheet.Rows[1:] {
		fullterm := strings.TrimSpace(strings.ToLower(row.Cells[1].Value)) // Translation in 2nd col
		keyword := Keyword{term:fullterm}
		clean := stopwords.Clean([]byte(fullterm), "de", true)
		kwt := splitKeywords(string(clean))
		log.Println(fullterm, len(kwt))
		if len(kwt) > 1 {
			keyword.subterms = append(keyword.subterms, fullterm)
		}

		for _, kw := range kwt {
			//log.Println(kw)
			keyword.subterms = append(keyword.subterms, kw)
		}
		keywords = append(keywords, keyword)
	}
}


func checkErr(err error) {
	if err != nil {
		panic(err)
	}
}

func main() {
	get_keywords()
	db, err := sql.Open("mysql", "root:@sk3rmunk3l@/openthesaurus?charset=utf8")
	checkErr(err)
	defer db.Close()
	// Prepare statement for inserting data
	//stmtOut, err := db.Prepare(sqlstatement) // ? = placeholder
	//checkErr(err)
	//defer stmtOut.Close() // Close the statement when we leave main() / the program terminates

	if len(keywords) > 0 {
		for id, terms := range keywords {
			synonyms := []string{}
			m := make(map[string]string) // remove duplicates
			for _, term := range terms.subterms {
				//log.Println(term)
				//rows, err := stmtOut.Exec(term)
				//checkErr(err)
				rows, _ := db.Query(sqlstatement, term)
				for rows.Next() {
					var word string
					err = rows.Scan(&word)
					//log.Println(word)

					if _, found := m[word]; !found {
						synonyms = append(synonyms, word)
						m[word] = word
					}
				}
				keywords[id].synonyms = append(keywords[id].synonyms, synonyms) // add array to synonym array
			}
			//log.Println(synonyms)

			totalSynonyms := 0
			for _, synonym_count := range keywords[id].synonyms {
				totalSynonyms += len(synonym_count)
			}
			log.Println(terms.term, keywords[id].synonyms)
		}
	}
	write_results()
}




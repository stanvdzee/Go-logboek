package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"
)

type LogboekItem struct {
	Datum        string `json:"datum"`
	Titel        string `json:"titel"`
	Beschrijving string `json:"beschrijving"`
	Minuten      int    `json:"minuten"`
	Categorie    string `json:"categorie"`
}

type Logboek struct {
	Items []LogboekItem `json:"items"`
}

type Reflectie struct {
	Jaar       int      `json:"jaar"`
	Week       int      `json:"week"`
	Antwoorden []string `json:"antwoorden"`
}

type Reflecties struct {
	Reflecties []Reflectie `json:"reflecties"`
}

func main() {

	argsWithoutProg := os.Args[1:]
	reader := bufio.NewReader(os.Stdin)

	if len(argsWithoutProg) > 0 {
		for i := 0; i < len(argsWithoutProg); i++ {
			switch argsWithoutProg[i] {

			case "help", "hulp":
				fmt.Println("=== Go Logboek - Help ===")
				fmt.Println()
				fmt.Println("Gebruik: go run . <commando>")
				fmt.Println()
				fmt.Println("Beschikbare commando's:")
				fmt.Println("  nieuw                         Nieuw logboekitem toevoegen")
				fmt.Println("  bekijken                      Alle logboekitems bekijken")
				fmt.Println("  dag <datum>                   Items van een bepaalde dag bekijken")
				fmt.Println("  periode <start> <einde>       Items binnen een periode bekijken")
				fmt.Println("  verwijderen                   Een logboekitem verwijderen")
				fmt.Println("  reflectie <jaar> <week>       Weekreflectie invullen")
				fmt.Println("  rapport dag <datum>           Dagrapport maken")
				fmt.Println("  rapport week <jaar> <week>    Weekrapport maken")
				fmt.Println("  rapport maand <jaar> <maand>  Maandrapport maken")
				fmt.Println("  zoeken <zoekwoord>            Zoeken in je logboek")
				fmt.Println("  help                          Deze helptekst tonen")
				fmt.Println()
				fmt.Println("Voorbeelden:")
				fmt.Println("  go run . nieuw")
				fmt.Println("  go run . dag 26-10-2026")
				fmt.Println("  go run . periode 19-10-2026 26-10-2026")
				fmt.Println("  go run . reflectie 2026 43")
				fmt.Println("  go run . rapport week 2026 43")
				fmt.Println("  go run . zoeken Go leren")

			case "nieuw":
				fmt.Println("Nieuw item toevoegen...")

				var titel string
				var beschrijving string
				var minutenTekst string
				var minuten int
				var categorie string

				fmt.Print("Titel: ")
				titel, _ = reader.ReadString('\n')
				titel = strings.TrimSpace(titel)
				if titel == "" {
					fmt.Println("Fout: titel mag niet leeg zijn.")
					return
				}

				fmt.Print("Beschrijving: ")
				beschrijving, _ = reader.ReadString('\n')
				beschrijving = strings.TrimSpace(beschrijving)
				if beschrijving == "" {
					fmt.Println("Fout: beschrijving mag niet leeg zijn.")
					return
				}

				fmt.Print("Aantal minuten: ")
				minutenTekst, _ = reader.ReadString('\n')
				minutenTekst = strings.TrimSpace(minutenTekst)

				minuten, err := strconv.Atoi(minutenTekst)
				if err != nil || minuten <= 0 {
					fmt.Println("Fout: vul een geldig aantal minuten in.")
					return
				}

				fmt.Print("Categorie: ")
				categorie, _ = reader.ReadString('\n')
				categorie = strings.TrimSpace(categorie)

				logboek := Logboek{}

				data, err := os.ReadFile("logboek.json")
				if err == nil {
					err = json.Unmarshal(data, &logboek)
					if err != nil {
						fmt.Printf("Fout bij lezen logboek: %s\n", err)
						return
					}
				} else if !os.IsNotExist(err) {
					fmt.Println("Fout bij lezen bestand:", err)
					return
				}

				datum := time.Now().Format("02-01-2006")

				item := LogboekItem{
					Datum:        datum,
					Titel:        titel,
					Beschrijving: beschrijving,
					Minuten:      minuten,
					Categorie:    categorie,
				}

				logboek.Items = append(logboek.Items, item)

				data, err = json.MarshalIndent(logboek, "", "    ")

				if err != nil {
					fmt.Println("Fout bij maken van JSON:", err)
					return
				}

				err = os.WriteFile("logboek.json", data, 0644)

				if err != nil {
					fmt.Println("Fout bij opslaan:", err)
					return
				}

				fmt.Println("Nieuw item toegevoegd:")
				fmt.Println("Titel:", item.Titel)
				fmt.Println("Beschrijving:", item.Beschrijving)
				fmt.Println("Datum:", item.Datum)

			case "bekijken":
				fmt.Println("Logboek bekijken...")

				data, err := os.ReadFile("logboek.json")

				if err != nil {
					fmt.Println("Er is nog geen logboek.", err)
					return
				}

				var logboek Logboek
				err = json.Unmarshal(data, &logboek)
				if err != nil {
					fmt.Println("Fout bij ontleden van JSON:", err)
					return
				}

				for i, item := range logboek.Items {
					fmt.Printf("Item %d:\n", i+1)
					fmt.Printf("Datum: %s\n", item.Datum)
					fmt.Printf("Titel: %s\n", item.Titel)
					fmt.Printf("Beschrijving: %s\n", item.Beschrijving)
					fmt.Printf("Minuten: %.d\n", item.Minuten)
					fmt.Printf("Categorie: %s\n", item.Categorie)
				}

			case "dag":
				if len(argsWithoutProg) < 2 {
					fmt.Println("Gebruik: go run . dag <datum>")
					return
				}

				datum := argsWithoutProg[1]

				_, err := time.Parse("02-01-2006", datum)
				if err != nil {
					fmt.Println("Fout: ongeldig datumformaat.")
					return
				}

				LogboekData, err := os.ReadFile("logboek.json")
				if err != nil {
					fmt.Println("Fout bij lezen logboek:", err)
					return
				}

				var logboek Logboek
				err = json.Unmarshal(LogboekData, &logboek)
				if err != nil {
					fmt.Println("Fout bij ontleden van logboek:", err)
					return
				}

				gevonden := false

				for _, item := range logboek.Items {
					if item.Datum == datum {
						fmt.Printf("\nDatum: %s\nTitel: %s\nBeschrijving: %s\nMinuten: %.d\nCategorie: %s\n",
							item.Datum, item.Titel, item.Beschrijving, item.Minuten, item.Categorie)
						gevonden = true
					}
				}

				if !gevonden {
					fmt.Println("Geen items gevonden voor deze datum.")
				}

			case "periode":
				if len(argsWithoutProg) < 3 {
					fmt.Println("Gebruik: go run . periode <startdatum> <einddatum>")
					return
				}

				start, err := time.Parse("02-01-2006", argsWithoutProg[1])
				if err != nil {
					fmt.Println("Fout: de begindatum is ongeldig.")
					return
				}

				einde, err := time.Parse("02-01-2006", argsWithoutProg[2])
				if err != nil {
					fmt.Println("Fout: de einddatum is ongeldig.")
					return
				}

				if einde.Before(start) {
					fmt.Println("Fout: de einddatum mag niet voor de begindatum liggen.")
					return
				}

				LogboekData, err := os.ReadFile("logboek.json")
				if err != nil {
					fmt.Println("Fout bij lezen logboek:", err)
					return
				}

				var logboek Logboek
				err = json.Unmarshal(LogboekData, &logboek)
				if err != nil {
					fmt.Println("Fout bij laden van logboek:", err)
					return
				}

				gevonden := false

				for _, item := range logboek.Items {
					itemDatum, err := time.Parse("02-01-2006", item.Datum)
					if err != nil {
						continue
					}

					if !itemDatum.Before(start) && !itemDatum.After(einde) {
						fmt.Printf("\nDatum: %s\nTitel: %s\nBeschrijving: %s\nMinuten: %.d\nCategorie: %s\n",
							item.Datum, item.Titel, item.Beschrijving, item.Minuten, item.Categorie)
						gevonden = true
					}
				}

				if !gevonden {
					fmt.Println("Geen items gevonden voor deze periode.")
				}

			case "verwijderen":
				fmt.Println("Item verwijderen...")

				data, err := os.ReadFile("logboek.json")

				if err != nil {
					fmt.Println("Er is nog geen logboek.", err)
					return
				}

				var logboek Logboek

				err = json.Unmarshal(data, &logboek)

				if err != nil {
					fmt.Println("Fout bij ontleden van JSON:", err)
					return
				}
				var keuze int
				fmt.Print("Welk item wil je verwijderen? ")
				fmt.Scanln(&keuze)

				if keuze < 1 || keuze > len(logboek.Items) {
					fmt.Println("Ongeldig itemnummer.")
					return
				}
				logboek.Items = append(logboek.Items[:keuze-1], logboek.Items[keuze:]...)

				data, err = json.MarshalIndent(logboek, "", "  ")

				if err != nil {
					fmt.Println("Fout bij maken van JSON:", err)
					return
				}

				err = os.WriteFile("logboek.json", data, 0644)

				if err != nil {
					fmt.Println("Fout bij opslaan:", err)
					return
				}

				fmt.Println("Item verwijderd.")

			case "reflectie":
				if len(argsWithoutProg) < 3 {
					fmt.Println("Gebruik: go run . reflectie <jaar> <weeknummer>")
					return
				}

				jaar, err := strconv.Atoi(argsWithoutProg[1])
				if err != nil {
					fmt.Println("Fout: ongeldig jaar.")
					return
				}

				week, err := strconv.Atoi(argsWithoutProg[2])
				if err != nil || week < 1 || week > 53 {
					fmt.Println("Fout: het weeknummer moet tussen 1 en 53 liggen.")
					return
				}

				data, err := os.ReadFile("logboek.json")
				if err != nil {
					fmt.Println("Er is nog geen logboek.")
					return
				}

				var logboek Logboek
				err = json.Unmarshal(data, &logboek)
				if err != nil {
					fmt.Println("Fout bij lezen van logboek:", err)
					return
				}

				var weekItems []LogboekItem
				totaalMinuten := 0
				dagMinuten := make(map[string]int)
				categorieAantal := make(map[string]int)

				for _, item := range logboek.Items {
					if hoortBijWeek(item.Datum, jaar, week) {
						weekItems = append(weekItems, item)
						totaalMinuten += item.Minuten

						dagMinuten[item.Datum] += item.Minuten

						if item.Categorie != "" {
							categorieAantal[item.Categorie]++
						}
					}
				}

				meesteMinuten := 0
				druksteDag := ""

				for datum, minuten := range dagMinuten {
					if minuten > meesteMinuten {
						meesteMinuten = minuten
						druksteDag = datum
					}
				}

				meestVoorkomendeCategorie := ""
				hoogsteAantal := 0

				for categorie, aantal := range categorieAantal {
					if aantal > hoogsteAantal {
						hoogsteAantal = aantal
						meestVoorkomendeCategorie = categorie
					}
				}

				if meestVoorkomendeCategorie != "" {
					fmt.Printf(
						"Meest voorkomende categorie: %s (%d activiteiten)\n",
						meestVoorkomendeCategorie,
						hoogsteAantal,
					)
				} else {
					fmt.Println("Er zijn geen categorieën ingevuld.")
				}

				if len(weekItems) == 0 {
					fmt.Println("Er zijn geen logboekitems voor deze week. Reflectie kan niet starten.")
					return
				}

				fmt.Println("\n--- Jouw activiteiten deze week ---")

				dagMinutenOverzicht := make(map[string]int)

				for _, item := range weekItems {
					dagMinutenOverzicht[item.Datum] += item.Minuten
				}

				huidigeDatum := ""

				for i, item := range weekItems {
					if item.Datum != huidigeDatum {
						huidigeDatum = item.Datum

						fmt.Printf("\n=== %s ===\n", huidigeDatum)
						fmt.Printf("Totaal die dag: %d minuten\n", dagMinutenOverzicht[huidigeDatum])
					}

					fmt.Printf("%d. %s\n", i+1, item.Titel)
					fmt.Printf("  Beschrijving: %s\n", item.Beschrijving)
					fmt.Printf("  Tijd: %d minuten\n", item.Minuten)
					fmt.Printf("  Categorie: %s\n", item.Categorie)
				}

				fmt.Println("\n--- Einde activiteitenoverzicht ---")

				fmt.Printf("\nWeekreflectie: week %d van %d\n", week, jaar)
				fmt.Printf("Aantal logboekitems: %d\n", len(weekItems))
				fmt.Printf("Totaal bestede minuten: %d\n", totaalMinuten)
				fmt.Printf("Drukste dag: %s (%d minuten)\n", druksteDag, meesteMinuten)

				reader := bufio.NewReader(os.Stdin)
				var antwoorden []string

				vragen := []string{
					"Wat heb je deze week geleerd?",
					"wat ging goed en wat kon beter?",
					fmt.Sprintf("Je drukste dag was %s met %d minuten. Wat heb je die dag geleerd?", druksteDag, meesteMinuten),
				}

				for _, vraag := range vragen {

					fmt.Println("\nWelke activiteit wil je verder bespreken?")

					for i, item := range weekItems {
						fmt.Printf("%d. %s\n", i+1, item.Titel)
					}

					fmt.Print("Kies een activiteitnummer: ")

					keuzeTekst, err := reader.ReadString('\n')
					if err != nil && len(keuzeTekst) == 0 {
						fmt.Println("Fout bij lezen van je keuze:", err)
						return
					}

					keuzeTekst = strings.TrimSpace(keuzeTekst)
					keuze, err := strconv.Atoi(keuzeTekst)

					if err != nil || keuze < 1 || keuze > len(weekItems) {
						fmt.Println("Fout: kies een geldig activiteitnummer.")
						return
					}

					gekozenItem := weekItems[keuze-1]

					fmt.Printf("Je hebt gekozen: %s\n", gekozenItem.Titel)
					fmt.Print("Wat heb je van deze activiteit geleerd? ")

					antwoord, err := reader.ReadString('\n')
					if err != nil && len(antwoord) == 0 {
						fmt.Println("Fout bij lezen van antwoord:", err)
						return
					}

					antwoord = strings.TrimSpace(antwoord)
					if antwoord == "" {
						fmt.Println("Fout: je antwoord mag niet leeg zijn.")
						return
					}
					fmt.Println("\n" + vraag)
					fmt.Print("Antwoord: ")
					antwoord, err = reader.ReadString('\n')
					if err != nil && len(antwoord) == 0 {
						fmt.Println("Fout bij lezen van antwoord:", err)
						return
					}

					antwoord = strings.TrimSpace(antwoord)
					if antwoord == "" {
						fmt.Println("Fout: je antwoord mag niet leeg zijn.")
						return
					}

					antwoorden = append(antwoorden, antwoord)

					antwoorden = append(antwoorden, antwoord)

				}

				var reflectieBestand Reflecties

				reflectieData, err := os.ReadFile("reflecties.json")
				if err == nil {
					err = json.Unmarshal(reflectieData, &reflectieBestand)
					if err != nil {
						fmt.Println("Fout bij lezen van reflecties:", err)
						return
					}
				} else if !os.IsNotExist(err) {
					fmt.Println("Fout bij het openen van reflecties:", err)
					return
				}

				nieuweReflectie := Reflectie{
					Jaar:       jaar,
					Week:       week,
					Antwoorden: antwoorden,
				}

				var nieuweReflecties []Reflectie
				vervangen := false

				for _, reflectie := range reflectieBestand.Reflecties {
					if reflectie.Jaar == jaar && reflectie.Week == week {
						if !vervangen {
							nieuweReflecties = append(nieuweReflecties, nieuweReflectie)
							vervangen = true
						}
						// Eventuele dubbele reflecties worden overgeslagen
					} else {
						nieuweReflecties = append(nieuweReflecties, reflectie)
					}
				}

				if !vervangen {
					nieuweReflecties = append(nieuweReflecties, nieuweReflectie)
				}

				reflectieBestand.Reflecties = nieuweReflecties

				reflectieData, err = json.MarshalIndent(reflectieBestand, "", "  ")
				if err != nil {
					fmt.Println("Fout bij maken van JSON:", err)
					return
				}

				err = os.WriteFile("reflecties.json", reflectieData, 0644)
				if err != nil {
					fmt.Println("Fout bij opslaan van reflecties:", err)
					return
				}

				fmt.Println("\nJe reflectie is opgeslagen in reflecties.json.")

			case "rapport":
				if len(argsWithoutProg) < 3 {
					fmt.Println("Gebruik: go run . rapport dag DD-MM-YYYY")
					return
				}

				switch argsWithoutProg[1] {
				case "dag":
					datum := argsWithoutProg[2]

					_, err := time.Parse("02-01-2006", datum)
					if err != nil {
						fmt.Println("Fout: gebruik de datumindeling DD-MM-YYYY.")
						return
					}

					data, err := os.ReadFile("logboek.json")
					if err != nil {
						fmt.Println("Fout bij het lezen van het logboek:", err)
						return
					}

					var logboek Logboek
					err = json.Unmarshal(data, &logboek)
					if err != nil {
						fmt.Println("Fout bij het lezen van de logboekgegevens:", err)
						return
					}

					var rapport strings.Builder
					totaalMinuten := 0
					aantalItems := 0

					rapport.WriteString(fmt.Sprintf("Dagrapport: %s\n\n", datum))

					for _, item := range logboek.Items {
						if item.Datum == datum {
							aantalItems++
							totaalMinuten += item.Minuten

							rapport.WriteString(fmt.Sprintf("Titel: %s\n", item.Titel))
							rapport.WriteString(fmt.Sprintf("Beschrijving: %s\n", item.Beschrijving))
							rapport.WriteString(fmt.Sprintf("Minuten: %d\n", item.Minuten))
							rapport.WriteString(fmt.Sprintf("Categorie: %s\n\n", item.Categorie))
						}
					}

					if aantalItems == 0 {
						fmt.Println("Er zijn geen logboekitems voor deze datum.")
						return
					}

					rapport.WriteString(fmt.Sprintf("Aantal activiteiten: %d\n", aantalItems))
					rapport.WriteString(fmt.Sprintf("Totaal: %d minuten\n", totaalMinuten))

					bestandsnaam := "dagrapport-" + datum + ".txt"
					err = os.WriteFile(bestandsnaam, []byte(rapport.String()), 0644)
					if err != nil {
						fmt.Println("Fout bij het opslaan van het rapport:", err)
						return
					}

					fmt.Println("Dagrapport opgeslagen in:", bestandsnaam)

				case "week":
					if len(argsWithoutProg) < 4 {
						fmt.Println("Gebruik: go run . rapport week <jaar> <weeknummer>")
						return
					}

					jaar, err := strconv.Atoi(argsWithoutProg[2])
					if err != nil {
						fmt.Println("Fout: ongeldig jaar.")
						return
					}

					week, err := strconv.Atoi(argsWithoutProg[3])
					if err != nil || week < 1 || week > 53 {
						fmt.Println("Fout: weeknummer moet tussen 1 en 53 liggen.")
						return
					}

					data, err := os.ReadFile("logboek.json")
					if err != nil {
						fmt.Println("Fout bij lezen van logboek:", err)
						return
					}

					var logboek Logboek
					err = json.Unmarshal(data, &logboek)
					if err != nil {
						fmt.Println("Fout bij lezen van logboek:", err)
						return
					}

					dagItems := make(map[string][]LogboekItem)
					dagMinuten := make(map[string]int)
					var datums []string
					totaalMinuten := 0

					for _, item := range logboek.Items {
						if hoortBijWeek(item.Datum, jaar, week) {
							if _, bestaat := dagItems[item.Datum]; !bestaat {
								datums = append(datums, item.Datum)
							}

							dagItems[item.Datum] = append(dagItems[item.Datum], item)
							dagMinuten[item.Datum] += item.Minuten
							totaalMinuten += item.Minuten
						}
					}

					if len(datums) == 0 {
						fmt.Println("Er zijn geen logboekitems voor deze week.")
						return
					}

					sort.Slice(datums, func(i, j int) bool {
						datumI, _ := time.Parse("02-01-2006", datums[i])
						datumJ, _ := time.Parse("02-01-2006", datums[j])

						return datumI.Before(datumJ)
					})

					var rapport strings.Builder

					rapport.WriteString(fmt.Sprintf(
						"Weekrapport - week %d van %d\n\n",
						week, jaar,
					))

					for _, datum := range datums {
						rapport.WriteString(fmt.Sprintf(
							"=== %s ===\n",
							datum,
						))

						for _, item := range dagItems[datum] {
							rapport.WriteString(fmt.Sprintf(
								"Titel: %s\nBeschrijving: %s\nMinuten: %d\nCategorie: %s\n\n",
								item.Titel,
								item.Beschrijving,
								item.Minuten,
								item.Categorie,
							))
						}

						rapport.WriteString(fmt.Sprintf(
							"Totaal deze dag: %d minuten\n\n",
							dagMinuten[datum],
						))
					}

					rapport.WriteString(fmt.Sprintf(
						"Totaal aantal minuten deze week: %d\n\n",
						totaalMinuten,
					))

					// Voeg de reflectie toe als die voor deze week bestaat.
					reflectieData, leesErr := os.ReadFile("reflecties.json")

					if leesErr == nil {
						var reflectieBestand Reflecties

						err = json.Unmarshal(reflectieData, &reflectieBestand)
						if err != nil {
							fmt.Println("Fout bij lezen van reflecties:", err)
							return
						}

						for i := len(reflectieBestand.Reflecties) - 1; i >= 0; i-- {
							reflectie := reflectieBestand.Reflecties[i]

							if reflectie.Jaar == jaar && reflectie.Week == week {
								rapport.WriteString("=== Weekreflectie ===\n")

								vragen := []string{
									"Wat heb je deze week geleerd?",
									"Wat ging goed en wat kon beter?",
									"Wat heb je geleerd van je drukste dag?",
									"Wat heb je geleerd van de gekozen activiteit?",
								}

								for j, antwoord := range reflectie.Antwoorden {
									if j < len(vragen) {
										rapport.WriteString(vragen[j] + "\n")
									}

									rapport.WriteString(antwoord + "\n\n")
								}

								break
							}
						}
					} else if !os.IsNotExist(leesErr) {
						fmt.Println("Fout bij lezen van reflecties:", leesErr)
						return
					}

					bestandsnaam := fmt.Sprintf(
						"weekrapport-%d-%02d.txt",
						jaar, week,
					)

					err = os.WriteFile(
						bestandsnaam,
						[]byte(rapport.String()),
						0644,
					)

					if err != nil {
						fmt.Println("Fout bij opslaan van rapport:", err)
						return
					}

					fmt.Println("Weekrapport opgeslagen in:", bestandsnaam)

				case "maand":
					if len(argsWithoutProg) < 4 {
						fmt.Println("Gebruik: go run . rapport maand <jaar> <maandnummer>")
						return
					}

					jaar, err := strconv.Atoi(argsWithoutProg[2])
					if err != nil {
						fmt.Println("Fout: ongeldig jaar.")
						return
					}

					maand, err := strconv.Atoi(argsWithoutProg[3])
					if err != nil || maand < 1 || maand > 12 {
						fmt.Println("Fout: maandnummer moet tussen 1 en 12 liggen.")
						return
					}

					data, err := os.ReadFile("logboek.json")
					if err != nil {
						fmt.Println("Fout bij lezen van logboek:", err)
						return
					}

					var logboek Logboek
					err = json.Unmarshal(data, &logboek)
					if err != nil {
						fmt.Println("Fout bij lezen van logboek:", err)
						return
					}

					maandItems := make(map[string][]LogboekItem)
					var datums []string
					totaalMinuten := 0
					totaalItems := 0

					for _, item := range logboek.Items {
						datum, err := time.Parse("02-01-2006", item.Datum)
						if err != nil {
							continue
						}

						if datum.Year() == jaar && int(datum.Month()) == maand {
							if _, bestaat := maandItems[item.Datum]; !bestaat {
								datums = append(datums, item.Datum)
							}

							maandItems[item.Datum] = append(maandItems[item.Datum], item)
							totaalMinuten += item.Minuten
							totaalItems++
						}
					}

					if len(datums) == 0 {
						fmt.Println("Er zijn geen logboekitems voor deze maand.")
						return
					}

					sort.Slice(datums, func(i, j int) bool {
						datumI, _ := time.Parse("02-01-2006", datums[i])
						datumJ, _ := time.Parse("02-01-2006", datums[j])
						return datumI.Before(datumJ)
					})

					var rapport strings.Builder

					rapport.WriteString(fmt.Sprintf(
						"Maandrapport - %02d-%d\n\n",
						maand, jaar,
					))

					for _, datum := range datums {
						rapport.WriteString(fmt.Sprintf("=== %s ===\n", datum))

						dagMinuten := 0

						for _, item := range maandItems[datum] {
							rapport.WriteString(fmt.Sprintf(
								"Titel: %s\nBeschrijving: %s\nMinuten: %d\nCategorie: %s\n\n",
								item.Titel,
								item.Beschrijving,
								item.Minuten,
								item.Categorie,
							))

							dagMinuten += item.Minuten
						}

						rapport.WriteString(fmt.Sprintf(
							"Totaal deze dag: %d minuten\n\n",
							dagMinuten,
						))
					}

					rapport.WriteString(fmt.Sprintf(
						"Totaal aantal activiteiten deze maand: %d\n",
						totaalItems,
					))
					rapport.WriteString(fmt.Sprintf(
						"Totaal bestede minuten deze maand: %d\n",
						totaalMinuten,
					))

					bestandsnaam := fmt.Sprintf(
						"maandrapport-%d-%02d.txt",
						jaar, maand,
					)

					err = os.WriteFile(bestandsnaam, []byte(rapport.String()), 0644)
					if err != nil {
						fmt.Println("Fout bij opslaan van rapport:", err)
						return
					}

					fmt.Println("Maandrapport opgeslagen in:", bestandsnaam)

				default:
					fmt.Println("Onbekend rapporttype. Gebruik: go run . rapport dag DD-MM-YYYY")
				}

			case "zoeken":
				if len(argsWithoutProg) < 2 {
					fmt.Println("Gebruik: go run . zoeken <zoekwoord>")
					return
				}

				zoekwoord := strings.Join(argsWithoutProg[i+1:], " ")

				data, err := os.ReadFile("logboek.json")
				if err != nil {
					fmt.Println("Er is nog geen logboek.")
					return
				}

				var logboek Logboek

				err = json.Unmarshal(data, &logboek)
				if err != nil {
					fmt.Println("Fout bij ontleden van JSON:", err)
					return
				}

				gevonden := false

				for _, item := range logboek.Items {
					if strings.Contains(
						strings.ToLower(item.Titel+" "+item.Beschrijving),
						strings.ToLower(zoekwoord),
					) {

						fmt.Printf("Gevonden item:\n")
						fmt.Printf("Datum: %s\n", item.Datum)
						fmt.Printf("Titel: %s\n", item.Titel)
						fmt.Printf("Beschrijving: %s\n", item.Beschrijving)
						fmt.Printf("Minuten: %.d\n", item.Minuten)
						fmt.Printf("Categorie: %s\n", item.Categorie)

						gevonden = true
					}
				}

				if !gevonden {
					fmt.Println("Geen items gevonden.")
				}

			default:
				fmt.Println("Ongeldig argument:", argsWithoutProg[i])
			}

			break
		}

	} else {
		fmt.Println("Geen argumenten opgegeven.")
	}

}

func hoortBijWeek(datumTekst string, jaar, week int) bool {
	datum, err := time.Parse("02-01-2006", datumTekst)
	if err != nil {
		return false
	}

	weekJaar, weekNummer := datum.ISOWeek()

	return weekJaar == jaar && weekNummer == week
}

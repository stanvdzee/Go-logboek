package main

import (
	"bufio"
	"database/sql"
	"fmt"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"

	_ "modernc.org/sqlite"
)

type LogboekItem struct {
	ID           int
	Datum        string
	Titel        string
	Beschrijving string
	Minuten      int
	Categorie    string
}

type Reflectie struct {
	Jaar      int
	Week      int
	Antwoord1 string
	Antwoord2 string
	Antwoord3 string
	Antwoord4 string
}

func main() {
	db, err := sql.Open("sqlite", "logboek.db")
	if err != nil {
		fmt.Println("Fout bij openen van database:", err)
		return
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		fmt.Println("Fout bij verbinden met database:", err)
		return
	}
	if err := initialiseerDatabase(db); err != nil {
		fmt.Println("Fout bij initialiseren van database:", err)
		return
	}

	args := os.Args[1:]
	if len(args) == 0 {
		fmt.Println("Geen argumenten opgegeven. Gebruik: go run . help")
		return
	}
	reader := bufio.NewReader(os.Stdin)

	switch args[0] {
	case "help", "hulp":
		toonHelp()
	case "nieuw":
		nieuwItem(db, reader)
	case "bekijken":
		bekijkItems(db)
	case "dag":
		if len(args) < 2 {
			fmt.Println("Gebruik: go run . dag DD-MM-YYYY")
			return
		}
		toonDag(db, args[1])
	case "periode":
		if len(args) < 3 {
			fmt.Println("Gebruik: go run . periode DD-MM-YYYY DD-MM-YYYY")
			return
		}
		toonPeriode(db, args[1], args[2])
	case "verwijderen":
		verwijderItem(db, reader)
	case "reflectie":
		if len(args) < 3 {
			fmt.Println("Gebruik: go run . reflectie <jaar> <weeknummer>")
			return
		}
		jaar, err1 := strconv.Atoi(args[1])
		week, err2 := strconv.Atoi(args[2])
		if err1 != nil || err2 != nil || week < 1 || week > 53 {
			fmt.Println("Fout: vul een geldig jaar en een weeknummer van 1 t/m 53 in.")
			return
		}
		maakReflectie(db, reader, jaar, week)
	case "rapport":
		maakRapport(db, args)
	case "zoeken":
		if len(args) < 2 {
			fmt.Println("Gebruik: go run . zoeken <zoekwoord>")
			return
		}
		zoekItems(db, strings.Join(args[1:], " "))
	default:
		fmt.Println("Onbekend commando:", args[0])
		fmt.Println("Gebruik: go run . help")
	}
}

func initialiseerDatabase(db *sql.DB) error {
	queries := []string{
		`CREATE TABLE IF NOT EXISTS logboek_items (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			datum TEXT NOT NULL,
			titel TEXT NOT NULL,
			beschrijving TEXT NOT NULL,
			minuten INTEGER NOT NULL CHECK (minuten > 0),
			categorie TEXT NOT NULL DEFAULT ''
		)`,
		`CREATE TABLE IF NOT EXISTS weekreflecties (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			jaar INTEGER NOT NULL,
			week INTEGER NOT NULL,
			antwoord1 TEXT NOT NULL,
			antwoord2 TEXT NOT NULL,
			antwoord3 TEXT NOT NULL,
			antwoord4 TEXT NOT NULL,
			UNIQUE (jaar, week)
		)`,
	}
	for _, query := range queries {
		if _, err := db.Exec(query); err != nil {
			return err
		}
	}

	// Neem oude reflecties mee als de eerdere versie ze in SQLite heeft opgeslagen.
	// De conversie gebeurt in SQLite; de Go-code gebruikt geen JSON-pakket meer.
	var oudeTabel string
	err := db.QueryRow(`SELECT name FROM sqlite_master WHERE type = 'table' AND name = 'reflecties'`).Scan(&oudeTabel)
	if err == nil {
		_, err = db.Exec(`INSERT OR IGNORE INTO weekreflecties (jaar, week, antwoord1, antwoord2, antwoord3, antwoord4)
			SELECT jaar, week,
			COALESCE(json_extract(antwoorden, '$[0]'), ''),
			COALESCE(json_extract(antwoorden, '$[1]'), ''),
			COALESCE(json_extract(antwoorden, '$[2]'), ''),
			COALESCE(json_extract(antwoorden, '$[3]'), '')
			FROM reflecties`)
		if err != nil {
			return fmt.Errorf("oude reflecties overzetten: %w", err)
		}
	} else if err != sql.ErrNoRows {
		return err
	}
	return nil
}

func toonHelp() {
	fmt.Println("=== Go Logboek - Help ===")
	fmt.Println("Gebruik: go run . <commando>")
	fmt.Println()
	fmt.Println("  nieuw                         Nieuw logboekitem toevoegen")
	fmt.Println("  bekijken                      Alle logboekitems bekijken")
	fmt.Println("  dag <datum>                   Items van een dag bekijken")
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
}

func nieuwItem(db *sql.DB, reader *bufio.Reader) {
	titel := vraag(reader, "Titel: ")
	if titel == "" {
		fmt.Println("Fout: titel mag niet leeg zijn.")
		return
	}
	beschrijving := vraag(reader, "Beschrijving: ")
	if beschrijving == "" {
		fmt.Println("Fout: beschrijving mag niet leeg zijn.")
		return
	}
	minutenTekst := vraag(reader, "Aantal minuten: ")
	minuten, err := strconv.Atoi(minutenTekst)
	if err != nil || minuten <= 0 {
		fmt.Println("Fout: vul een geldig aantal minuten groter dan nul in.")
		return
	}
	categorie := vraag(reader, "Categorie (optioneel): ")
	datum := time.Now().Format("02-01-2006")

	resultaat, err := db.Exec(`INSERT INTO logboek_items (datum, titel, beschrijving, minuten, categorie) VALUES (?, ?, ?, ?, ?)`, datum, titel, beschrijving, minuten, categorie)
	if err != nil {
		fmt.Println("Fout bij opslaan in SQLite:", err)
		return
	}
	id, _ := resultaat.LastInsertId()
	fmt.Println("Nieuw item toegevoegd aan SQLite:")
	fmt.Printf("ID: %d\nDatum: %s\nTitel: %s\nBeschrijving: %s\nMinuten: %d\nCategorie: %s\n", id, datum, titel, beschrijving, minuten, categorie)
}

func vraag(reader *bufio.Reader, tekst string) string {
	fmt.Print(tekst)
	antwoord, err := reader.ReadString('\n')
	if err != nil && len(antwoord) == 0 {
		return ""
	}
	return strings.TrimSpace(antwoord)
}

func haalItems(db *sql.DB) ([]LogboekItem, error) {
	rows, err := db.Query(`SELECT id, datum, titel, beschrijving, minuten, categorie FROM logboek_items ORDER BY id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]LogboekItem, 0)
	for rows.Next() {
		var item LogboekItem
		if err := rows.Scan(&item.ID, &item.Datum, &item.Titel, &item.Beschrijving, &item.Minuten, &item.Categorie); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func bekijkItems(db *sql.DB) {
	items, err := haalItems(db)
	if err != nil {
		fmt.Println("Fout bij ophalen van logboekitems:", err)
		return
	}
	if len(items) == 0 {
		fmt.Println("Je logboek is nog leeg.")
		return
	}
	for _, item := range items {
		printItem(item)
	}
}

func printItem(item LogboekItem) {
	fmt.Printf("\nID: %d\nDatum: %s\nTitel: %s\nBeschrijving: %s\nMinuten: %d\nCategorie: %s\n", item.ID, item.Datum, item.Titel, item.Beschrijving, item.Minuten, item.Categorie)
}

func toonDag(db *sql.DB, datum string) {
	if _, err := time.Parse("02-01-2006", datum); err != nil {
		fmt.Println("Fout: gebruik de datumindeling DD-MM-YYYY.")
		return
	}
	rows, err := db.Query(`SELECT id, datum, titel, beschrijving, minuten, categorie FROM logboek_items WHERE datum = ? ORDER BY id`, datum)
	if err != nil {
		fmt.Println("Fout bij ophalen van items:", err)
		return
	}
	defer rows.Close()
	gevonden := false
	for rows.Next() {
		var item LogboekItem
		if err := rows.Scan(&item.ID, &item.Datum, &item.Titel, &item.Beschrijving, &item.Minuten, &item.Categorie); err != nil {
			fmt.Println("Fout bij lezen van item:", err)
			return
		}
		printItem(item)
		gevonden = true
	}
	if err := rows.Err(); err != nil {
		fmt.Println("Fout bij lezen van items:", err)
		return
	}
	if !gevonden {
		fmt.Println("Geen items gevonden voor deze datum.")
	}
}

func toonPeriode(db *sql.DB, startTekst, eindeTekst string) {
	start, err := time.Parse("02-01-2006", startTekst)
	if err != nil {
		fmt.Println("Fout: de begindatum is ongeldig.")
		return
	}
	einde, err := time.Parse("02-01-2006", eindeTekst)
	if err != nil {
		fmt.Println("Fout: de einddatum is ongeldig.")
		return
	}
	if einde.Before(start) {
		fmt.Println("Fout: de einddatum mag niet voor de begindatum liggen.")
		return
	}
	items, err := haalItems(db)
	if err != nil {
		fmt.Println("Fout bij ophalen van items:", err)
		return
	}
	gevonden := false
	for _, item := range items {
		datum, parseErr := time.Parse("02-01-2006", item.Datum)
		if parseErr == nil && !datum.Before(start) && !datum.After(einde) {
			printItem(item)
			gevonden = true
		}
	}
	if !gevonden {
		fmt.Println("Geen items gevonden voor deze periode.")
	}
}

func verwijderItem(db *sql.DB, reader *bufio.Reader) {
	rows, err := db.Query(
		`SELECT id, datum, titel FROM logboek_items ORDER BY id`,
	)
	if err != nil {
		fmt.Println("Fout bij ophalen van items:", err)
		return
	}
	defer rows.Close()

	type verwijderKeuze struct {
		ID    int
		Datum string
		Titel string
	}

	items := make([]verwijderKeuze, 0)

	for rows.Next() {
		var item verwijderKeuze

		if err := rows.Scan(
			&item.ID,
			&item.Datum,
			&item.Titel,
		); err != nil {
			fmt.Println("Fout bij lezen van items:", err)
			return
		}

		items = append(items, item)
	}

	if err := rows.Err(); err != nil {
		fmt.Println("Fout bij ophalen van items:", err)
		return
	}

	if len(items) == 0 {
		fmt.Println("Er zijn geen items om te verwijderen.")
		return
	}

	// Toon de echte database-ID bij ieder item.
	for _, item := range items {
		fmt.Printf(
			"ID %d: %s (%s)\n",
			item.ID,
			item.Titel,
			item.Datum,
		)
	}

	// Vraag om het echte database-ID.
	keuzeTekst := vraag(reader, "Welk database-ID wil je verwijderen? ")

	keuze, err := strconv.Atoi(keuzeTekst)
	if err != nil {
		fmt.Println("Ongeldig ID. Vul een getal uit de lijst in.")
		return
	}

	// Controleer of het gekozen ID bestaat.
	bestaat := false

	for _, item := range items {
		if item.ID == keuze {
			bestaat = true
			break
		}
	}

	if !bestaat {
		fmt.Println("Geen logboekitem gevonden met ID", keuze)
		return
	}

	// Verwijder uitsluitend het gekozen database-ID.
	resultaat, err := db.Exec(
		`DELETE FROM logboek_items WHERE id = ?`,
		keuze,
	)
	if err != nil {
		fmt.Println("Fout bij verwijderen:", err)
		return
	}

	verwijderd, err := resultaat.RowsAffected()
	if err != nil {
		fmt.Println("Item verwijderd, maar het resultaat kon niet worden gecontroleerd.")
		return
	}

	if verwijderd == 0 {
		fmt.Println("Geen item verwijderd. Controleer het ID en probeer opnieuw.")
		return
	}

	fmt.Printf(
		"Logboekitem met ID %d verwijderd uit SQLite.\n",
		keuze,
	)
}

func zoekItems(db *sql.DB, zoekwoord string) {
	rows, err := db.Query(`SELECT id, datum, titel, beschrijving, minuten, categorie FROM logboek_items WHERE LOWER(titel || ' ' || beschrijving || ' ' || categorie) LIKE LOWER(?) ORDER BY id`, "%"+zoekwoord+"%")
	if err != nil {
		fmt.Println("Fout bij zoeken:", err)
		return
	}
	defer rows.Close()
	gevonden := false
	for rows.Next() {
		var item LogboekItem
		if err := rows.Scan(&item.ID, &item.Datum, &item.Titel, &item.Beschrijving, &item.Minuten, &item.Categorie); err != nil {
			fmt.Println("Fout bij lezen van resultaat:", err)
			return
		}
		printItem(item)
		gevonden = true
	}
	if err := rows.Err(); err != nil {
		fmt.Println("Fout bij zoeken:", err)
		return
	}
	if !gevonden {
		fmt.Println("Geen items gevonden.")
	}
}

func itemsVanWeek(db *sql.DB, jaar, week int) ([]LogboekItem, error) {
	items, err := haalItems(db)
	if err != nil {
		return nil, err
	}
	resultaat := make([]LogboekItem, 0)
	for _, item := range items {
		if hoortBijWeek(item.Datum, jaar, week) {
			resultaat = append(resultaat, item)
		}
	}
	return resultaat, nil
}

func maakReflectie(db *sql.DB, reader *bufio.Reader, jaar, week int) {
	items, err := itemsVanWeek(db, jaar, week)
	if err != nil {
		fmt.Println("Fout bij ophalen van weekitems:", err)
		return
	}
	if len(items) == 0 {
		fmt.Println("Er zijn geen logboekitems voor deze week. Reflectie kan niet starten.")
		return
	}

	dagMinuten := make(map[string]int)
	categorieAantal := make(map[string]int)
	totaalMinuten := 0
	for _, item := range items {
		dagMinuten[item.Datum] += item.Minuten
		totaalMinuten += item.Minuten
		if item.Categorie != "" {
			categorieAantal[item.Categorie]++
		}
	}
	druksteDag := ""
	meesteMinuten := 0
	for datum, minuten := range dagMinuten {
		if minuten > meesteMinuten {
			druksteDag, meesteMinuten = datum, minuten
		}
	}
	meestVoorkomendeCategorie := "geen categorie ingevuld"
	hoogsteAantal := 0
	for categorie, aantal := range categorieAantal {
		if aantal > hoogsteAantal {
			meestVoorkomendeCategorie, hoogsteAantal = categorie, aantal
		}
	}

	fmt.Printf("Week %d van %d: %d activiteiten, %d minuten totaal.\n", week, jaar, len(items), totaalMinuten)
	fmt.Printf("Drukste dag: %s (%d minuten).\n", druksteDag, meesteMinuten)
	fmt.Printf("Meest voorkomende categorie: %s.\n", meestVoorkomendeCategorie)
	fmt.Println("\nActiviteiten deze week:")
	for i, item := range items {
		fmt.Printf("%d. %s — %s (%d minuten)\n", i+1, item.Titel, item.Datum, item.Minuten)
	}

	antwoord1 := vraag(reader, "1. Wat heb je deze week geleerd? ")
	if antwoord1 == "" {
		fmt.Println("Antwoord mag niet leeg zijn.")
		return
	}
	antwoord2 := vraag(reader, "2. Wat ging goed en wat kon beter? ")
	if antwoord2 == "" {
		fmt.Println("Antwoord mag niet leeg zijn.")
		return
	}
	antwoord3 := vraag(reader, fmt.Sprintf("3. Je drukste dag was %s. Wat heb je die dag geleerd? ", druksteDag))
	if antwoord3 == "" {
		fmt.Println("Antwoord mag niet leeg zijn.")
		return
	}
	keuzeTekst := vraag(reader, "Kies een activiteitnummer om verder te bespreken: ")
	keuze, err := strconv.Atoi(keuzeTekst)
	if err != nil || keuze < 1 || keuze > len(items) {
		fmt.Println("Ongeldig activiteitnummer.")
		return
	}
	antwoord4 := vraag(reader, fmt.Sprintf("4. Wat heb je geleerd van '%s'? ", items[keuze-1].Titel))
	if antwoord4 == "" {
		fmt.Println("Antwoord mag niet leeg zijn.")
		return
	}

	_, err = db.Exec(`INSERT INTO weekreflecties (jaar, week, antwoord1, antwoord2, antwoord3, antwoord4)
		VALUES (?, ?, ?, ?, ?, ?)
		ON CONFLICT(jaar, week) DO UPDATE SET antwoord1=excluded.antwoord1, antwoord2=excluded.antwoord2, antwoord3=excluded.antwoord3, antwoord4=excluded.antwoord4`, jaar, week, antwoord1, antwoord2, antwoord3, antwoord4)
	if err != nil {
		fmt.Println("Fout bij opslaan van reflectie:", err)
		return
	}
	fmt.Println("Weekreflectie opgeslagen in SQLite.")
}

func maakRapport(db *sql.DB, args []string) {
	if len(args) < 3 {
		fmt.Println("Gebruik: go run . rapport dag <datum>, rapport week <jaar> <week> of rapport maand <jaar> <maand>")
		return
	}
	switch args[1] {
	case "dag":
		datum := args[2]
		if _, err := time.Parse("02-01-2006", datum); err != nil {
			fmt.Println("Fout: gebruik DD-MM-YYYY.")
			return
		}
		items, err := haalItems(db)
		if err != nil {
			fmt.Println("Fout bij ophalen van items:", err)
			return
		}
		var b strings.Builder
		totaal, aantal := 0, 0
		fmt.Fprintf(&b, "Dagrapport: %s\n\n", datum)
		for _, item := range items {
			if item.Datum == datum {
				aantal++
				totaal += item.Minuten
				writeItem(&b, item)
			}
		}
		if aantal == 0 {
			fmt.Println("Er zijn geen logboekitems voor deze datum.")
			return
		}
		fmt.Fprintf(&b, "Aantal activiteiten: %d\nTotaal: %d minuten\n", aantal, totaal)
		bestandsnaam := "dagrapport-" + datum + ".txt"
		schrijfRapport(bestandsnaam, b.String())
	case "week":
		if len(args) < 4 {
			fmt.Println("Gebruik: go run . rapport week <jaar> <week>")
			return
		}
		jaar, e1 := strconv.Atoi(args[2])
		week, e2 := strconv.Atoi(args[3])
		if e1 != nil || e2 != nil || week < 1 || week > 53 {
			fmt.Println("Fout: ongeldig jaar of weeknummer.")
			return
		}
		items, err := itemsVanWeek(db, jaar, week)
		if err != nil {
			fmt.Println("Fout bij ophalen van items:", err)
			return
		}
		if len(items) == 0 {
			fmt.Println("Er zijn geen logboekitems voor deze week.")
			return
		}
		sort.Slice(items, func(i, j int) bool { return datumVoor(itemDatum(items[i]), itemDatum(items[j])) })
		var b strings.Builder
		totaal := 0
		fmt.Fprintf(&b, "Weekrapport - week %d van %d\n\n", week, jaar)
		for _, item := range items {
			writeItem(&b, item)
			totaal += item.Minuten
		}
		fmt.Fprintf(&b, "Totaal aantal minuten deze week: %d\n\n", totaal)
		reflectie, err := haalReflectie(db, jaar, week)
		if err != nil {
			fmt.Println("Fout bij ophalen van reflectie:", err)
			return
		}
		if reflectie != nil {
			fmt.Fprintf(&b, "=== Weekreflectie ===\nWat heb je deze week geleerd?\n%s\n\nWat ging goed en wat kon beter?\n%s\n\nWat heb je geleerd van je drukste dag?\n%s\n\nWat heb je geleerd van de gekozen activiteit?\n%s\n", reflectie.Antwoord1, reflectie.Antwoord2, reflectie.Antwoord3, reflectie.Antwoord4)
		}
		bestandsnaam := fmt.Sprintf("weekrapport-%d-%02d.txt", jaar, week)
		schrijfRapport(bestandsnaam, b.String())
	case "maand":
		if len(args) < 4 {
			fmt.Println("Gebruik: go run . rapport maand <jaar> <maand>")
			return
		}
		jaar, e1 := strconv.Atoi(args[2])
		maand, e2 := strconv.Atoi(args[3])
		if e1 != nil || e2 != nil || maand < 1 || maand > 12 {
			fmt.Println("Fout: ongeldig jaar of maandnummer.")
			return
		}
		items, err := haalItems(db)
		if err != nil {
			fmt.Println("Fout bij ophalen van items:", err)
			return
		}
		var maandItems []LogboekItem
		for _, item := range items {
			datum, parseErr := time.Parse("02-01-2006", item.Datum)
			if parseErr == nil && datum.Year() == jaar && int(datum.Month()) == maand {
				maandItems = append(maandItems, item)
			}
		}
		if len(maandItems) == 0 {
			fmt.Println("Er zijn geen logboekitems voor deze maand.")
			return
		}
		sort.Slice(maandItems, func(i, j int) bool { return datumVoor(itemDatum(maandItems[i]), itemDatum(maandItems[j])) })
		var b strings.Builder
		totaal := 0
		fmt.Fprintf(&b, "Maandrapport - %02d-%d\n\n", maand, jaar)
		for _, item := range maandItems {
			writeItem(&b, item)
			totaal += item.Minuten
		}
		fmt.Fprintf(&b, "Totaal aantal activiteiten deze maand: %d\nTotaal bestede minuten deze maand: %d\n", len(maandItems), totaal)
		bestandsnaam := fmt.Sprintf("maandrapport-%d-%02d.txt", jaar, maand)
		schrijfRapport(bestandsnaam, b.String())
	default:
		fmt.Println("Onbekend rapporttype. Kies dag, week of maand.")
	}
}

func writeItem(b *strings.Builder, item LogboekItem) {
	fmt.Fprintf(b, "Datum: %s\nTitel: %s\nBeschrijving: %s\nMinuten: %d\nCategorie: %s\n\n", item.Datum, item.Titel, item.Beschrijving, item.Minuten, item.Categorie)
}

func schrijfRapport(bestandsnaam, inhoud string) {
	if err := os.WriteFile(bestandsnaam, []byte(inhoud), 0644); err != nil {
		fmt.Println("Fout bij opslaan van rapport:", err)
		return
	}
	fmt.Println("Rapport opgeslagen in:", bestandsnaam)
}

func haalReflectie(db *sql.DB, jaar, week int) (*Reflectie, error) {
	var r Reflectie
	err := db.QueryRow(`SELECT jaar, week, antwoord1, antwoord2, antwoord3, antwoord4 FROM weekreflecties WHERE jaar = ? AND week = ?`, jaar, week).Scan(&r.Jaar, &r.Week, &r.Antwoord1, &r.Antwoord2, &r.Antwoord3, &r.Antwoord4)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &r, nil
}

func hoortBijWeek(datumTekst string, jaar, week int) bool {
	datum, err := time.Parse("02-01-2006", datumTekst)
	if err != nil {
		return false
	}
	weekJaar, weekNummer := datum.ISOWeek()
	return weekJaar == jaar && weekNummer == week
}

func itemDatum(item LogboekItem) time.Time {
	datum, _ := time.Parse("02-01-2006", item.Datum)
	return datum
}

func datumVoor(a, b time.Time) bool { return a.Before(b) }

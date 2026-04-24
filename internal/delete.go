package serve

import (
	"fmt"
	"html/template"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/asano69/hashcards/internal/collection"
	"github.com/asano69/hashcards/internal/db"
	"github.com/asano69/hashcards/internal/parser"
	"github.com/asano69/hashcards/internal/types"
)

// deleteHandler serves the /delete page.
type deleteHandler struct {
	col       *collection.Collection
	db        *db.Database
	staticDir string
}

// deleteCardItem is one entry in the card list on the delete page.
// For cloze families, one item represents the entire C: block.
type deleteCardItem struct {
	Hash  string // representative card hash used as the form value
	Label string // display text shown to the user
}

// deletePageData is the template data for the delete page.
type deletePageData struct {
	Decks        []string
	SelectedDeck string
	Cards        []deleteCardItem
	Message      string
}

func (h *deleteHandler) handleGet(w http.ResponseWriter, r *http.Request) {
	deck := r.URL.Query().Get("deck")
	msg := r.URL.Query().Get("msg")
	decks := collectionDeckNames(h.col)

	var cards []deleteCardItem
	if deck != "" {
		cards = h.buildCardList(deck)
	}

	h.renderPage(w, deletePageData{
		Decks:        decks,
		SelectedDeck: deck,
		Cards:        cards,
		Message:      msg,
	})
}

func (h *deleteHandler) handlePost(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "bad form data", http.StatusBadRequest)
		return
	}

	deckName := r.FormValue("deck")
	selectedHashes := r.Form["hashes"]

	if len(selectedHashes) == 0 {
		http.Redirect(w, r, "/delete?deck="+url.QueryEscape(deckName), http.StatusSeeOther)
		return
	}

	// Expand selected hashes to include all siblings in the same source block.
	toDelete := h.resolveCards(selectedHashes)

	// Group cards by source file.
	fileGroups := make(map[string][]types.Card)
	for _, card := range toDelete {
		fp := card.FilePath()
		fileGroups[fp] = append(fileGroups[fp], card)
	}

	var errorMsgs []string
	for filePath, cards := range fileGroups {
		if err := deleteFromFile(filePath, cards); err != nil {
			errorMsgs = append(errorMsgs, fmt.Sprintf("%s: %v", filepath.Base(filePath), err))
		}
	}

	// Delete from the database.
	deletedCount := 0
	for _, card := range toDelete {
		if err := h.db.DeleteCard(card.Hash()); err == nil {
			deletedCount++
		}
	}

	// Remove deleted cards from the in-memory collection so the page shows
	// updated data immediately. This mutation is intentionally unguarded; it
	// is acceptable for a single-user personal tool.
	deletedSet := make(map[types.CardHash]struct{})
	for _, card := range toDelete {
		deletedSet[card.Hash()] = struct{}{}
	}
	remaining := make([]types.Card, 0, len(h.col.Cards))
	for _, card := range h.col.Cards {
		if _, del := deletedSet[card.Hash()]; !del {
			remaining = append(remaining, card)
		}
	}
	h.col.Cards = remaining

	msg := fmt.Sprintf("Deleted %d card(s).", deletedCount)
	if len(errorMsgs) > 0 {
		msg += " File errors: " + strings.Join(errorMsgs, "; ")
	}
	target := "/delete?deck=" + url.QueryEscape(deckName) + "&msg=" + url.QueryEscape(msg)
	http.Redirect(w, r, target, http.StatusSeeOther)
}

// resolveCards expands the selected card hashes to include all siblings that
// share the same (filePath, lineStart) source block (handles cloze families).
func (h *deleteHandler) resolveCards(hexHashes []string) []types.Card {
	selected := make(map[types.CardHash]struct{})
	for _, hex := range hexHashes {
		ch, err := types.ParseCardHash(hex)
		if err == nil {
			selected[ch] = struct{}{}
		}
	}

	type blockKey struct {
		filePath  string
		lineStart int
	}
	blocks := make(map[blockKey]struct{})
	for _, card := range h.col.Cards {
		if _, ok := selected[card.Hash()]; ok {
			blocks[blockKey{card.FilePath(), card.LineStart()}] = struct{}{}
		}
	}

	seen := make(map[types.CardHash]struct{})
	var result []types.Card
	for _, card := range h.col.Cards {
		key := blockKey{card.FilePath(), card.LineStart()}
		if _, ok := blocks[key]; ok {
			if _, already := seen[card.Hash()]; !already {
				result = append(result, card)
				seen[card.Hash()] = struct{}{}
			}
		}
	}
	return result
}

// buildCardList returns display items for all cards in the given deck.
// Cloze siblings from the same C: block are merged into a single item.
func (h *deleteHandler) buildCardList(deckName string) []deleteCardItem {
	type familyGroup struct {
		representative types.Card
		members        []types.Card
	}

	groups := make(map[string]*familyGroup)
	var order []string

	for _, card := range h.col.Cards {
		if card.DeckName() != deckName {
			continue
		}
		var key string
		if fh := card.FamilyHash(); fh != nil {
			key = "f:" + fh.Hex()
		} else {
			key = card.Hash().Hex()
		}
		if _, ok := groups[key]; !ok {
			groups[key] = &familyGroup{representative: card}
			order = append(order, key)
		}
		groups[key].members = append(groups[key].members, card)
	}

	items := make([]deleteCardItem, 0, len(order))
	for _, key := range order {
		g := groups[key]
		card := g.representative
		cc := card.Content()

		var label string
		switch cc.Kind() {
		case types.CardTypeBasic:
			q := truncateCardText(cc.Question, 80)
			a := truncateCardText(cc.Answer, 80)
			label = "Q: " + q + " / A: " + a
		case types.CardTypeCloze:
			label = "C: " + truncateCardText(reconstructClozeText(g.members), 120)
		}

		items = append(items, deleteCardItem{
			Hash:  card.Hash().Hex(),
			Label: label,
		})
	}
	return items
}

func (h *deleteHandler) renderPage(w http.ResponseWriter, data deletePageData) {
	tmpl, err := template.ParseFiles(
		filepath.Join(h.staticDir, "templates", "base.html"),
		filepath.Join(h.staticDir, "templates", "delete.html"),
	)
	if err != nil {
		http.Error(w, "template error: "+err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := tmpl.ExecuteTemplate(w, "base", data); err != nil {
		fmt.Printf("delete render error: %v\n", err)
	}
}

// collectionDeckNames returns sorted unique deck names found in the collection.
func collectionDeckNames(col *collection.Collection) []string {
	seen := make(map[string]struct{})
	var names []string
	for _, card := range col.Cards {
		name := card.DeckName()
		if _, ok := seen[name]; !ok {
			seen[name] = struct{}{}
			names = append(names, name)
		}
	}
	sort.Strings(names)
	return names
}

// truncateCardText replaces newlines with spaces and truncates to maxLen runes.
func truncateCardText(s string, maxLen int) string {
	s = strings.ReplaceAll(s, "\n", " ")
	s = strings.TrimSpace(s)
	runes := []rune(s)
	if len(runes) <= maxLen {
		return s
	}
	return string(runes[:maxLen]) + "..."
}

// reconstructClozeText rebuilds the C: source text with [brackets] around
// each deletion span, given all members of a cloze family.
func reconstructClozeText(cards []types.Card) string {
	if len(cards) == 0 {
		return ""
	}
	text := cards[0].Content().Text

	type span struct{ start, end int }
	dedup := make(map[[2]int]bool)
	var spans []span
	for _, c := range cards {
		cc := c.Content()
		k := [2]int{cc.Start, cc.End}
		if !dedup[k] {
			dedup[k] = true
			spans = append(spans, span{cc.Start, cc.End})
		}
	}
	sort.Slice(spans, func(i, j int) bool { return spans[i].start < spans[j].start })

	var b strings.Builder
	prev := 0
	for _, s := range spans {
		b.WriteString(text[prev:s.start])
		b.WriteByte('[')
		end := s.end + 1
		if end > len(text) {
			end = len(text)
		}
		b.WriteString(text[s.start:end])
		b.WriteByte(']')
		prev = end
	}
	b.WriteString(text[prev:])
	return b.String()
}

// deleteFromFile removes the given cards from a Markdown deck file by
// deleting their line ranges. Cards with the same lineStart (cloze siblings)
// are treated as a single block. Frontmatter is preserved.
func deleteFromFile(filePath string, cardsToDelete []types.Card) error {
	allCards, err := parser.ParseFile(filePath)
	if err != nil {
		return fmt.Errorf("parse file: %w", err)
	}

	deleteSet := make(map[types.CardHash]struct{})
	for _, card := range cardsToDelete {
		deleteSet[card.Hash()] = struct{}{}
	}

	raw, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("read file: %w", err)
	}

	// Normalise line endings and split.
	text := strings.ReplaceAll(string(raw), "\r\n", "\n")
	text = strings.TrimRight(text, "\n")
	lines := strings.Split(text, "\n")

	// Card line numbers from the parser are relative to the content after
	// any frontmatter. Compute the offset to translate to full-file indices.
	fmOffset := frontmatterLineCount(lines)

	// Sort all cards by lineStart and build a deduplicated list of card groups
	// (multiple cloze siblings share the same lineStart/lineEnd).
	sort.Slice(allCards, func(i, j int) bool {
		return allCards[i].LineStart() < allCards[j].LineStart()
	})

	type cardGroup struct{ lineStart, lineEnd int }
	var groups []cardGroup
	seenLS := make(map[int]bool)
	for _, c := range allCards {
		if !seenLS[c.LineStart()] {
			seenLS[c.LineStart()] = true
			groups = append(groups, cardGroup{c.LineStart(), c.LineEnd()})
		}
	}

	// Mark which groups contain cards from the delete set.
	deleteGroups := make(map[int]bool) // keyed by lineStart
	for _, c := range allCards {
		if _, del := deleteSet[c.Hash()]; del {
			deleteGroups[c.LineStart()] = true
		}
	}

	// Mark individual file lines for deletion (0-indexed).
	toDelete := make([]bool, len(lines))
	for i, g := range groups {
		if !deleteGroups[g.lineStart] {
			continue
		}
		start0 := g.lineStart - 1 + fmOffset
		end0 := g.lineEnd - 1 + fmOffset
		// When this group's lineEnd is the same as the next group's lineStart,
		// that line belongs to the next card and must not be deleted.
		if i+1 < len(groups) && g.lineEnd == groups[i+1].lineStart {
			end0--
		}
		for l := start0; l <= end0 && l < len(toDelete); l++ {
			toDelete[l] = true
		}
	}

	var kept []string
	for i, line := range lines {
		if !toDelete[i] {
			kept = append(kept, line)
		}
	}
	kept = cleanupFileLines(kept)

	out := strings.Join(kept, "\n")
	if len(out) > 0 {
		out += "\n"
	}
	return os.WriteFile(filePath, []byte(out), 0644)
}

// frontmatterLineCount returns the number of lines consumed by the TOML
// frontmatter block (including both "---" delimiters), or 0 if absent.
// This mirrors the logic in the parser package's parseFrontmatter.
func frontmatterLineCount(lines []string) int {
	if len(lines) == 0 || strings.TrimSpace(lines[0]) != "---" {
		return 0
	}
	for i := 1; i < len(lines); i++ {
		if strings.TrimSpace(lines[i]) == "---" {
			return i + 1
		}
	}
	return 0 // no closing delimiter found → treat whole file as content
}

// cleanupFileLines removes leading/trailing blank lines and collapses
// consecutive blank lines into a single blank line.
func cleanupFileLines(lines []string) []string {
	for len(lines) > 0 && strings.TrimSpace(lines[0]) == "" {
		lines = lines[1:]
	}
	for len(lines) > 0 && strings.TrimSpace(lines[len(lines)-1]) == "" {
		lines = lines[:len(lines)-1]
	}
	result := make([]string, 0, len(lines))
	prevBlank := false
	for _, line := range lines {
		isBlank := strings.TrimSpace(line) == ""
		if isBlank && prevBlank {
			continue
		}
		result = append(result, line)
		prevBlank = isBlank
	}
	return result
}

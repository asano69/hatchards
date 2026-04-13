package parser

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/asano69/hashcards/internal/types"
)

// writeDeck creates a temporary .md file with the given deck name stem and
// returns its path. The file is cleaned up automatically via t.Cleanup.
func writeDeck(t *testing.T, stem, content string) string {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, stem+".md")
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("writeDeck: %v", err)
	}
	return path
}

// parseCards is a convenience wrapper around ParseFile that fails the test on
// error and returns the parsed cards.
func parseCards(t *testing.T, content string) []types.Card {
	t.Helper()
	path := writeDeck(t, "test_deck", content)
	cards, err := ParseFile(path)
	if err != nil {
		t.Fatalf("ParseFile: %v", err)
	}
	return cards
}

// assertBasic checks that cards[idx] is a basic card with the given question and answer.
func assertBasic(t *testing.T, cards []types.Card, idx int, question, answer string) {
	t.Helper()
	if idx >= len(cards) {
		t.Fatalf("index %d out of range (len=%d)", idx, len(cards))
	}
	cc := cards[idx].Content()
	if cc.Kind() != types.CardTypeBasic {
		t.Errorf("cards[%d]: expected basic card, got cloze", idx)
		return
	}
	if cc.Question != question {
		t.Errorf("cards[%d].Question = %q, want %q", idx, cc.Question, question)
	}
	if cc.Answer != answer {
		t.Errorf("cards[%d].Answer = %q, want %q", idx, cc.Answer, answer)
	}
}

// assertCloze checks that cards[idx] is a cloze card with the given text, start, end.
func assertCloze(t *testing.T, cards []types.Card, idx int, text string, start, end int) {
	t.Helper()
	if idx >= len(cards) {
		t.Fatalf("index %d out of range (len=%d)", idx, len(cards))
	}
	cc := cards[idx].Content()
	if cc.Kind() != types.CardTypeCloze {
		t.Errorf("cards[%d]: expected cloze card, got basic", idx)
		return
	}
	if cc.Text != text {
		t.Errorf("cards[%d].Text = %q, want %q", idx, cc.Text, text)
	}
	if cc.Start != start {
		t.Errorf("cards[%d].Start = %d, want %d", idx, cc.Start, start)
	}
	if cc.End != end {
		t.Errorf("cards[%d].End = %d, want %d", idx, cc.End, end)
	}
}

// ---- Basic card tests ----

// TestEmptyString matches Rust's test_empty_string.
func TestEmptyString(t *testing.T) {
	cards := parseCards(t, "")
	if len(cards) != 0 {
		t.Errorf("expected 0 cards, got %d", len(cards))
	}
}

// TestWhitespaceString matches Rust's test_whitespace_string.
func TestWhitespaceString(t *testing.T) {
	cards := parseCards(t, "\n\n\n")
	if len(cards) != 0 {
		t.Errorf("expected 0 cards, got %d", len(cards))
	}
}

// TestBasicCard matches Rust's test_basic_card.
func TestBasicCard(t *testing.T) {
	cards := parseCards(t, "Q: What is Rust?\nA: A systems programming language.")
	if len(cards) != 1 {
		t.Fatalf("expected 1 card, got %d", len(cards))
	}
	assertBasic(t, cards, 0, "What is Rust?", "A systems programming language.")
}

// TestMultilineQA matches Rust's test_multiline_qa.
func TestMultilineQA(t *testing.T) {
	cards := parseCards(t, "Q: foo\nbaz\nbaz\nA: FOO\nBAR\nBAZ")
	if len(cards) != 1 {
		t.Fatalf("expected 1 card, got %d", len(cards))
	}
	assertBasic(t, cards, 0, "foo\nbaz\nbaz", "FOO\nBAR\nBAZ")
}

// TestTwoQuestions matches Rust's test_two_questions.
func TestTwoQuestions(t *testing.T) {
	cards := parseCards(t, "Q: foo\nA: bar\n\nQ: baz\nA: quux\n\n")
	if len(cards) != 2 {
		t.Fatalf("expected 2 cards, got %d", len(cards))
	}
	assertBasic(t, cards, 0, "foo", "bar")
	assertBasic(t, cards, 1, "baz", "quux")
}

// ---- Cloze card tests ----

// TestClozeFollowedByQuestion matches Rust's test_cloze_followed_by_question.
func TestClozeFollowedByQuestion(t *testing.T) {
	cards := parseCards(t, "C: [foo]\nQ: Question\nA: Answer")
	if len(cards) != 2 {
		t.Fatalf("expected 2 cards, got %d", len(cards))
	}
	assertCloze(t, cards, 0, "foo", 0, 2)
	assertBasic(t, cards, 1, "Question", "Answer")
}

// TestClozeSingle matches Rust's test_cloze_single.
func TestClozeSingle(t *testing.T) {
	cards := parseCards(t, "C: Foo [bar] baz.")
	if len(cards) != 1 {
		t.Fatalf("expected 1 card, got %d", len(cards))
	}
	assertCloze(t, cards, 0, "Foo bar baz.", 4, 6)
}

// TestClozeMultiple matches Rust's test_cloze_multiple.
func TestClozeMultiple(t *testing.T) {
	cards := parseCards(t, "C: Foo [bar] baz [quux].")
	if len(cards) != 2 {
		t.Fatalf("expected 2 cards, got %d", len(cards))
	}
	assertCloze(t, cards, 0, "Foo bar baz quux.", 4, 6)
	assertCloze(t, cards, 1, "Foo bar baz quux.", 12, 15)
}

// TestClozeWithImage matches Rust's test_cloze_with_image.
func TestClozeWithImage(t *testing.T) {
	cards := parseCards(t, "C: Foo [bar] ![](image.jpg) [quux].")
	if len(cards) != 2 {
		t.Fatalf("expected 2 cards, got %d", len(cards))
	}
	assertCloze(t, cards, 0, "Foo bar ![](image.jpg) quux.", 4, 6)
	assertCloze(t, cards, 1, "Foo bar ![](image.jpg) quux.", 23, 26)
}

// TestClozeWithEscapedSquareBracket matches Rust's test_cloze_with_escaped_square_bracket.
func TestClozeWithEscapedSquareBracket(t *testing.T) {
	cards := parseCards(t, "C: Key: [`\\[`]")
	if len(cards) != 1 {
		t.Fatalf("expected 1 card, got %d", len(cards))
	}
	assertCloze(t, cards, 0, "Key: `[`", 5, 7)
}

// TestClozeWithMultipleEscapedBrackets matches Rust's test_cloze_with_multiple_escaped_square_brackets.
func TestClozeWithMultipleEscapedBrackets(t *testing.T) {
	cards := parseCards(t, "C: \\[markdown\\] [`\\[cloze\\]`]")
	if len(cards) != 1 {
		t.Fatalf("expected 1 card, got %d", len(cards))
	}
	assertCloze(t, cards, 0, "[markdown] `[cloze]`", 11, 19)
}

// TestMultiLineCloze matches Rust's test_multi_line_cloze.
func TestMultiLineCloze(t *testing.T) {
	cards := parseCards(t, "C: [foo]\n[bar]\nbaz.")
	if len(cards) != 2 {
		t.Fatalf("expected 2 cards, got %d", len(cards))
	}
	assertCloze(t, cards, 0, "foo\nbar\nbaz.", 0, 2)
	assertCloze(t, cards, 1, "foo\nbar\nbaz.", 4, 6)
}

// TestTwoClozes matches Rust's test_two_clozes.
func TestTwoClozes(t *testing.T) {
	cards := parseCards(t, "C: [foo]\nC: [bar]")
	if len(cards) != 2 {
		t.Fatalf("expected 2 cards, got %d", len(cards))
	}
	assertCloze(t, cards, 0, "foo", 0, 2)
	assertCloze(t, cards, 1, "bar", 0, 2)
}

// TestClozeWithInitialBlankLine matches Rust's test_cloze_with_initial_blank_line.
func TestClozeWithInitialBlankLine(t *testing.T) {
	input := "C:\nBuild something people want in Lisp.\n\n\u2014 [Paul Graham], [_Hackers and Painters_]\n\n"
	cards := parseCards(t, input)
	if len(cards) != 2 {
		t.Fatalf("expected 2 cards, got %d", len(cards))
	}
	wantText := "Build something people want in Lisp.\n\n\u2014 Paul Graham, _Hackers and Painters_"
	assertCloze(t, cards, 0, wantText, 42, 52)
	assertCloze(t, cards, 1, wantText, 55, 76)
}

// ---- Deduplication tests ----

// TestIdenticalBasicCards matches Rust's test_identical_basic_cards.
func TestIdenticalBasicCards(t *testing.T) {
	cards := parseCards(t, "Q: foo\nA: bar\n\nQ: foo\nA: bar\n\n")
	if len(cards) != 1 {
		t.Errorf("expected 1 card after dedup, got %d", len(cards))
	}
}

// TestIdenticalClozeCards matches Rust's test_identical_cloze_cards.
func TestIdenticalClozeCards(t *testing.T) {
	cards := parseCards(t, "C: foo [bar]\n\nC: foo [bar]")
	if len(cards) != 1 {
		t.Errorf("expected 1 card after dedup, got %d", len(cards))
	}
}

// TestIdenticalCardsAcrossFiles matches Rust's test_identical_cards_across_files.
func TestIdenticalCardsAcrossFiles(t *testing.T) {
	dir := t.TempDir()
	content := "Q: foo\nA: bar"
	for _, name := range []string{"file1.md", "file2.md"} {
		if err := os.WriteFile(filepath.Join(dir, name), []byte(content), 0644); err != nil {
			t.Fatal(err)
		}
	}
	// parseDeck collects all .md files in a directory.
	// Since there's no exported parseDeck in Go, we call ParseFile for each file
	// and check that the hashes are equal (simulating dedup behaviour).
	cards1, err := ParseFile(filepath.Join(dir, "file1.md"))
	if err != nil {
		t.Fatal(err)
	}
	cards2, err := ParseFile(filepath.Join(dir, "file2.md"))
	if err != nil {
		t.Fatal(err)
	}
	if len(cards1) != 1 || len(cards2) != 1 {
		t.Fatalf("expected 1 card from each file, got %d and %d", len(cards1), len(cards2))
	}
	if !cards1[0].Hash().Equal(cards2[0].Hash()) {
		t.Error("same card content in two files should produce the same hash")
	}
}

// ---- Separator tests ----

// TestSeparatorBetweenBasicCards matches Rust's test_separator_between_basic_cards.
func TestSeparatorBetweenBasicCards(t *testing.T) {
	cards := parseCards(t, "Q: foo\nA: bar\n---\nQ: baz\nA: quux")
	if len(cards) != 2 {
		t.Fatalf("expected 2 cards, got %d", len(cards))
	}
	assertBasic(t, cards, 0, "foo", "bar")
	assertBasic(t, cards, 1, "baz", "quux")
}

// TestSeparatorAfterClozeCard matches Rust's test_separator_after_cloze_card.
func TestSeparatorAfterClozeCard(t *testing.T) {
	cards := parseCards(t, "C: [foo]\n---\nQ: Question\nA: Answer")
	if len(cards) != 2 {
		t.Fatalf("expected 2 cards, got %d", len(cards))
	}
	assertCloze(t, cards, 0, "foo", 0, 2)
	assertBasic(t, cards, 1, "Question", "Answer")
}

// TestSeparatorBetweenClozeCards matches Rust's test_separator_between_cloze_cards.
func TestSeparatorBetweenClozeCards(t *testing.T) {
	cards := parseCards(t, "C: [foo]\n---\nC: [bar]")
	if len(cards) != 2 {
		t.Fatalf("expected 2 cards, got %d", len(cards))
	}
	assertCloze(t, cards, 0, "foo", 0, 2)
	assertCloze(t, cards, 1, "bar", 0, 2)
}

// TestMultipleSeparators matches Rust's test_multiple_separators.
func TestMultipleSeparators(t *testing.T) {
	cards := parseCards(t, "Q: foo\nA: bar\n---\n---\nQ: baz\nA: quux")
	if len(cards) != 2 {
		t.Fatalf("expected 2 cards, got %d", len(cards))
	}
	assertBasic(t, cards, 0, "foo", "bar")
	assertBasic(t, cards, 1, "baz", "quux")
}

// TestSeparatorAtEnd matches Rust's test_separator_at_end.
func TestSeparatorAtEnd(t *testing.T) {
	cards := parseCards(t, "Q: foo\nA: bar\n---")
	if len(cards) != 1 {
		t.Fatalf("expected 1 card, got %d", len(cards))
	}
	assertBasic(t, cards, 0, "foo", "bar")
}

// ---- Frontmatter tests ----

// TestFrontmatterWithName matches Rust's test_extract_frontmatter_with_name.
func TestFrontmatterWithName(t *testing.T) {
	input := "---\nname = \"My Deck\"\n---\n\nQ: What is Rust?\nA: A systems programming language."
	path := writeDeck(t, "chapter1", input)
	cards, err := ParseFile(path)
	if err != nil {
		t.Fatalf("ParseFile: %v", err)
	}
	if len(cards) != 1 {
		t.Fatalf("expected 1 card, got %d", len(cards))
	}
	if cards[0].DeckName() != "My Deck" {
		t.Errorf("DeckName = %q, want %q", cards[0].DeckName(), "My Deck")
	}
}

// TestFrontmatterAbsent verifies the deck name defaults to the filename stem.
func TestFrontmatterAbsent(t *testing.T) {
	path := writeDeck(t, "MyDeck", "Q: foo\nA: bar")
	cards, err := ParseFile(path)
	if err != nil {
		t.Fatalf("ParseFile: %v", err)
	}
	if len(cards) != 1 {
		t.Fatalf("expected 1 card, got %d", len(cards))
	}
	if cards[0].DeckName() != "MyDeck" {
		t.Errorf("DeckName = %q, want %q", cards[0].DeckName(), "MyDeck")
	}
}

// ---- Special character tests ----

// TestClozeDeletionWithExclamationSign matches Rust's test_cloze_deletion_with_exclamation_sign.
func TestClozeDeletionWithExclamationSign(t *testing.T) {
	cards := parseCards(t, "C: The notation [$n!$] means 'n factorial'.")
	if len(cards) != 1 {
		t.Fatalf("expected 1 card, got %d", len(cards))
	}
	cc := cards[0].Content()
	if cc.Kind() != types.CardTypeCloze {
		t.Fatal("expected cloze card")
	}
	wantText := "The notation $n!$ means 'n factorial'."
	if cc.Text != wantText {
		t.Errorf("Text = %q, want %q", cc.Text, wantText)
	}
}

// TestClozeDeletionWithMath matches Rust's test_cloze_deletion_with_math.
func TestClozeDeletionWithMath(t *testing.T) {
	cards := parseCards(t, "C: The string `\\alpha` renders as [$\\alpha$].")
	if len(cards) != 1 {
		t.Fatalf("expected 1 card, got %d", len(cards))
	}
	cc := cards[0].Content()
	wantText := "The string `\\alpha` renders as $\\alpha$."
	if cc.Text != wantText {
		t.Errorf("Text = %q, want %q", cc.Text, wantText)
	}
}

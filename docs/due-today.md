# 未学習、期限きれのカードは、今日がスケジュールのカードよりも優先されるのか

いいえ、未学習カードや期限切れカードが今日スケジュールされたカードよりも優先されることはありません。

## 詳細

### DueTodayのロジック

`DueToday`関数は、以下の条件のカードをすべて「期限」として扱います： 

- `due_date`が`null`のカード（未学習カード）
- `due_date <= today`のカード（期限切れまたは今日が期限のカード）

これらはすべて同じセットに含まれ、区別されません。

### シャッフルとフィルタリング

セッション作成時の処理は以下の通りです：

1. `col.DueToday()`ですべての期限カードを取得
2. `rng.Shuffle()`でランダムにシャッフル
3. `FilterDue()`でカード制限と新規カード制限を適用

重要な点は、シャッフルがフィルタリングの前に行われるため、カードの種類（未学習、期限切れ、今日が期限）による優先順位は存在しません。 

### 新規カード制限

唯一の例外は、`newCardLimit`設定です。これによりセッション内の新規カードの数を制限できますが、これは優先順位ではなく単なる数量制限です。

## Notes

コードベースには、カードの種類に基づく優先順位付けロジックは実装されていません。すべての期限カードは等しく扱われ、ランダムにシャッフルされます。

Wiki pages you might want to explore:
- [Glossary (asano69/hashcards)](/wiki/asano69/hashcards#10)

### Citations

**File:** internal/db/db.go (L253-283)
```go
func (db *Database) DueToday(today types.Date) (map[types.CardHash]struct{}, error) {
	rows, err := db.q("select card_hash, due_date from cards;", nil).Rows()
	if err != nil {
		return nil, errs.Newf("query due today: %v", err)
	}
	defer rows.Close()
	due := map[types.CardHash]struct{}{}
	for rows.Next() {
		var hex string
		var dueStr sql.NullString
		if err := rows.Scan(&hex, &dueStr); err != nil {
			return nil, errs.Newf("scan due today row: %v", err)
		}
		h, err := types.ParseCardHash(hex)
		if err != nil {
			return nil, err
		}
		if !dueStr.Valid || dueStr.String == "" {
			due[h] = struct{}{}
			continue
		}
		dd, err := types.ParseDate(dueStr.String)
		if err != nil {
			return nil, err
		}
		if dd.LessOrEqual(today) {
			due[h] = struct{}{}
		}
	}
	return due, rows.Err()
}
```

**File:** internal/cmd/drill/handlers/handlers.go (L77-91)
```go
	if newCardLimit != nil {
		limit := *newCardLimit
		var result []collection.DueCard
		newCount := 0
		for _, dc := range due {
			if dc.Performance.IsNew() {
				if newCount >= limit {
					continue
				}
				newCount++
			}
			result = append(result, dc)
		}
		due = result
	}
```

**File:** internal/cmd/drill/handlers/handlers.go (L127-178)
```go
func (m *Manager) AddSession(
	deckKey string,
	col *collection.Collection,
	database *db.Database,
	answerControls string,
	cardLimit *int,
	newCardLimit *int,
	deckFilter *string,
	burySiblings bool,
	fsrsCfg fsrs.FSRSConfig,
) error {
	due, err := col.DueToday(types.Today())
	if err != nil {
		return err
	}
	due = FilterDue(due, cardLimit, newCardLimit, deckFilter)
	if burySiblings {
		due = BurySiblings(due)
	}

	// Shuffle before filtering so new-card selection is random rather than
	// always picking the same cards in hash order.
	r := rng.FromSeed(uint64(time.Now().UnixNano()))
	due = rng.Shuffle(due, r)
	due = FilterDue(due, cardLimit, newCardLimit, deckFilter)
	if burySiblings {
		due = BurySiblings(due)
	}

	h := &handler{
		mu:             &sync.Mutex{},
		sess:           state.New(due, r, fsrsCfg),
		cache:          drillcache.Build(due, fileMountBase),
		db:             database,
		col:            col,
		macros:         loadMacros(filepath.Join(col.Root, "macros.tex")),
		answerControls: answerControls,
		cardLimit:      cardLimit,
		newCardLimit:   newCardLimit,
		deckFilter:     deckFilter,
		burySiblingsOn: burySiblings,
		fsrsCfg:        fsrsCfg,
	}
	m.sessions[deckKey] = h

	deckLabel := "all decks"
	if deckFilter != nil {
		deckLabel = fmt.Sprintf("deck=%q", *deckFilter)
	}
	fmt.Printf("[session] key=%q %s cards=%d\n", deckKey, deckLabel, len(due))
	return nil
}
```

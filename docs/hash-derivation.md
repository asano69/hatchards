ハッシュの導出アルゴリズムとシェルからの計算方法について説明します。

## ハッシュ導出アルゴリズム

hashcardsでは、カードのハッシュは**BLAKE3**アルゴリズムを使用してカードのコンテンツから計算されます 。ハッシュは32バイトの配列として定義されています

カードはコンテンツアドレス可能（content-addressable）であり、カードのテキストが変更されるとハッシュも変更され、レビュー履歴はリセットされます 。

カードのハッシュ導出はinternal/types/card.goのHash()メソッドで定義されています。

## Basic

Basicカードの場合、以下の3つのバイト列を単純に連結してBLAKE3でハッシュ化しています。

```
goh.Update([]byte("Basic"))
h.Update([]byte(c.Question))
h.Update([]byte(c.Answer))
```

つまり blake3("Basic" + question + answer) です（UTF-8バイト列として連結、区切り文字なし）


## Cloze

### family_hash

- `family_hash`はデータベースには保存されず、**使う場所でその都度メモリ上で再計算**されるだけの値です。
- また、使われている場所は`BurySiblings`だけです。`internal/cmd/drill/handlers/handlers.go`の`BurySiblings`関数の中でのみ使われています。

```go
func BurySiblings(due []collection.DueCard) []collection.DueCard {
	seen := make(map[types.CardHash]struct{})
	result := make([]collection.DueCard, 0, len(due))
	for _, dc := range due {
		fh := dc.Card.FamilyHash()
		if fh != nil {
			if _, ok := seen[*fh]; ok {
				continue
			}
			seen[*fh] = struct{}{}
		}
		result = append(result, dc)
	}
	return result
}
```

これはセッション開始時（`AddSession`/`resetSession`）に呼ばれるたびに:

1. `dc.Card.FamilyHash()`（= `CardContent.FamilyHash()`）でclean textから毎回その場でBLAKE3計算
2. `seen`というメモリ上のmapに一時的に入れて重複チェック
3. 関数を抜けたら`seen`ごと破棄される

`family_hash`は「DBに保存して後から引く値」ではなく、「毎回のセッション構築時にカードのテキストから再計算する導出値（derived value）」という設計です。
カードのテキストが変わればいつでも再計算すれば同じ値が出るので、保存しておく必要がない——というシンプルさ重視の設計ですね（"Cards are content addressed"の思想とも一致しています）。

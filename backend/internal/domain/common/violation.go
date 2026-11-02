package common

import "strconv"

// Violation は入力の 1 項目に対する 1 つの違反（HTTP では errors[{field, rule}] の 1 要素）。
type Violation struct {
	// Field は違反した項目の名前（リクエストの JSON 名）。
	Field string
	// Rule は違反した規則の id（validate のキーワードか、相関ルールの lowerCamel の id）。
	Rule string
}

// Violations は 1 回の入力で見つかった違反の集合。独立して判定できる違反をすべて集めてから返すための error。
// ErrInvalid を包む代わりに Is で ErrInvalid に一致させるので、既存の errors.Is(err, ErrInvalid) は従来どおり真になる。
type Violations []Violation

func (v Violations) Error() string {
	return "invalid: " + strconv.Itoa(len(v)) + " violation(s)"
}

// Is は ErrInvalid に対して真を返す（400 への変換と既存の判定を変えないため）。
func (Violations) Is(target error) bool {
	return target == ErrInvalid
}

// Add は違反を足す。同じ field・rule の重複は 1 つにまとめる（同じ項目を複数の経路で検証しても二重に返さないため）。
func (v *Violations) Add(field, rule string) {
	for _, x := range *v {
		if x.Field == field && x.Rule == rule {
			return
		}
	}
	*v = append(*v, Violation{Field: field, Rule: rule})
}

// Err は違反が無ければ nil、あれば自身を error として返す（呼び出し側の `if err := v.Err(); err != nil` のため）。
func (v Violations) Err() error {
	if len(v) == 0 {
		return nil
	}
	return v
}

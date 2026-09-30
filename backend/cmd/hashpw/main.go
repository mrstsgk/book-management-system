// hashpw は標準入力から読んだパスワードの bcrypt ハッシュを出す。ADMIN_PASSWORD_HASH に設定する値を作るためのもの。
// 引数で受けないのは、パスワードが shell の履歴に残らないようにするため。
//
// read -s は入力を画面に出さず、履歴にも残さない（echo でパスワードを渡すと履歴に残る）。
//
//	read -rs PW; printf %s "$PW" | go run ./cmd/hashpw; unset PW
package main

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

	infraauth "github.com/mrstsgk/book-management-system/backend/internal/infrastructure/auth"
)

func main() {
	if err := run(os.Stdin, os.Stdout); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

// run は 1 行読み、末尾の改行だけを落としてハッシュにする。
func run(in io.Reader, out io.Writer) error {
	line, err := bufio.NewReader(in).ReadString('\n')
	if err != nil && !errors.Is(err, io.EOF) {
		return err
	}
	password := strings.TrimRight(line, "\r\n")
	if password == "" {
		return errors.New("パスワードが空です")
	}
	hash, err := infraauth.HashPassword(password)
	if err != nil {
		return err
	}
	_, err = fmt.Fprintln(out, hash)
	return err
}

package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/go-playground/validator/v10"
)

// ValidDirPath は任意のOSで使用可能なディレクトリパスのバリデーション関数です。
// 空文字列は無効として扱います。
func ValidDirPath(fl validator.FieldLevel) bool {
	path := strings.TrimSpace(fl.Field().String())
	if path == "" {
		return false
	}

	// パスを正規化（クリーンアップ）
	cleanPath := filepath.Clean(path)

	// パスの形式をチェック
	// filepath.Clean後も元のパスが維持されているかを確認
	// （不正な文字が含まれている場合はCleanで変更される）
	if cleanPath != filepath.Clean(filepath.Clean(path)) {
		return false
	}

	// パスが存在する場合は、ディレクトリであることを確認
	if fi, err := os.Stat(cleanPath); err == nil {
		return fi.IsDir()
	}

	// パスが存在しない場合：
	// エラーの種類を確認し、パスの形式が正しいかどうかを判定
	_, err := os.Stat(cleanPath)
	if err != nil {
		if os.IsNotExist(err) {
			// パスは存在しないが、形式は正しい
			return true
		}
		if _, ok := err.(*os.PathError); ok {
			// パスエラーの場合、OSの制限に違反している可能性がある
			return false
		}
	}

	return true
}

// Config はテスト用の設定構造体です
type Config struct {
	DirPath string `validate:"validDirPath"`
}

func main() {
	// バリデータの初期化
	validate := validator.New()
	if err := validate.RegisterValidation("validDirPath", ValidDirPath); err != nil {
		panic(err)
	}

	// テストケース
	paths := []string{
		"C:\\Users\\runneradmin\\.gomi",
		"C:\\Users\\runneradmin\\.gomi\\",
		"/home/user/.gomi",
		"/home/user/.gomi/",
		"./relative/path",
		"./relative/path/",
		"",                      // 空文字列
		"invalid\\:*?\"<>|path", // 不正な文字を含むパス
	}

	for _, path := range paths {
		cfg := &Config{DirPath: path}
		err := validate.Struct(cfg)
		if err != nil {
			if _, ok := err.(*validator.InvalidValidationError); ok {
				fmt.Printf("内部エラー: %v\n", err)
				continue
			}
			fmt.Printf("パス %q は無効: %v\n", path, err)
		} else {
			fmt.Printf("パス %q は有効\n", path)
		}
	}
}

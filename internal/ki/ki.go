package ki

import (
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/masakurapa/gooki/internal/opt"
)

// Make はディレクトリ内容のツリー構造を生成します
func Make(originalPath string, option opt.Option) (Ki, error) {
	absPath, err := filepath.Abs(originalPath)
	if err != nil {
		return nil, err
	}

	ha, err := makeHappa(absPath, option)
	if err != nil {
		return nil, err
	}

	return &ki{
		eda:          makeEda(ha, "."),
		absPath:      absPath,
		originalPath: originalPath,
	}, nil
}

// Happaを作る
func makeHappa(baseAbs string, option opt.Option) ([]Happa, error) {
	ha := make([]Happa, 0)
	// Walkに絶対パスを渡すのでクロージャのpathも絶対パスになる
	err := filepath.Walk(baseAbs, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		// 開始ディレクトリが入ってしまうので除外する
		if path == baseAbs {
			return nil
		}

		h := newHappa(baseAbs, path, info)

		if !option.AllFile && h.IsHiddenFile() {
			return nil
		}
		if option.DirectoryOnly && !h.IsDir() {
			return nil
		}

		ha = append(ha, h)
		return nil
	})

	if err != nil {
		return nil, err
	}
	return ha, nil
}

// Happaの初期化を行います
func newHappa(baseAbsPath, fileAbsPath string, info os.FileInfo) Happa {
	path := strings.TrimPrefix(fileAbsPath, baseAbsPath+"/")
	return &happa{
		absPath:      fileAbsPath,
		relPath:      path,
		dir:          filepath.Dir(path),
		name:         info.Name(),
		isDir:        info.IsDir(),
		isHiddenFile: strings.HasPrefix(fileAbsPath, ".") || strings.HasPrefix(info.Name(), "."),
		isSymlink:    info.Mode()&os.ModeSymlink == os.ModeSymlink,
	}
}

func makeEda(ha []Happa, base string) []Eda {
	ed := make([]Eda, 0, len(ha))

	for _, h := range ha {
		// skip if the file is not directly under the base path.
		if base == h.RelPath() || base != h.Dir() {
			continue
		}

		e := eda{ha: h}
		if h.IsDir() {
			e.eda = makeEda(ha, h.RelPath())
		}
		ed = append(ed, &e)
	}

	return ed
}

type ki struct {
	// ディレクトリツリーの起点となるパスの絶対パス
	absPath string
	// ツリー生成時に渡されるパスを保持
	originalPath string
	// ファイルまたはディレクトリの集合
	eda []Eda
}

func (k *ki) Eda() []Eda {
	return k.eda
}

func (k *ki) WriteTree(out io.Writer, option opt.Option) error {
	w := &treeWriter{
		writer: writer{
			out:    out,
			option: option,
		},
		basePath: k.originalPath,
	}
	return w.Write(k.Eda())
}

// eda はファイルまたはディレクトリ情報を表します
type eda struct {
	eda []Eda
	ha  Happa
}

func (e *eda) Child() []Eda {
	return e.eda
}

func (e *eda) Happa() Happa {
	return e.ha
}

func (e *eda) HasChild() bool {
	return len(e.eda) > 0
}

type happa struct {
	absPath      string
	relPath      string
	dir          string
	name         string
	isDir        bool
	isHiddenFile bool
	isSymlink    bool
}

func (h *happa) AbsPath() string {
	return h.absPath
}

func (h *happa) RelPath() string {
	return h.relPath
}

func (h *happa) Dir() string {
	return h.dir
}

func (h *happa) Name() string {
	return h.name
}

func (h *happa) IsDir() bool {
	return h.isDir
}

func (h *happa) IsHiddenFile() bool {
	return h.isHiddenFile
}

func (h *happa) IsSymlink() bool {
	return h.isSymlink
}

func (h *happa) RealName() string {
	realPath, err := os.Readlink(h.AbsPath())
	if err != nil {
		//TODO: error handling
		return ""
	}
	return filepath.Base(realPath)
}
package file

import (
	"encoding/json"
	"io"
	"os"
	"path"
)

func DirNE(dir string) {
	fi, err := os.Stat(dir)
	if fi == nil || err != nil {
		os.MkdirAll(dir, 0755)
	}

}

func Read(dir, file string) ([]byte, error) {
	DirNE(dir)
	f, err := os.OpenFile(
		path.Join(dir, file),
		os.O_RDONLY,
		0o700)
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = f.Close()
	}()

	return io.ReadAll(f)
}

func ReadJSON[T any](dir, file string) (*T, error) {
	data, err := Read(dir, file)
	if err != nil {
		return nil, err
	}

	var t T
	if err = json.Unmarshal(data, &t); err != nil {
		return nil, err
	}

	return &t, nil
}

func Write(dir, file string, pl []byte) error {
	DirNE(dir)
	f, err := os.OpenFile(
		path.Join(dir, file),
		os.O_WRONLY|os.O_CREATE|os.O_TRUNC,
		0o700)
	if err != nil {
		return err
	}
	defer func() {
		_ = f.Close()
	}()

	_, err = f.Write(pl)

	return err
}

func WriteJSON(dir, file string, v any) error {
	pl, err := json.Marshal(v)
	if err != nil {
		return err
	}
	return Write(dir, file, pl)
}

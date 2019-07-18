package file2

import (
	"io"
	"io/ioutil"
	"os"
)

func Exists(filename string) bool {
	_, err := os.Stat(filename)
	return err == nil || !os.IsNotExist(err)
}

func Size(filename string) (int64, error) {
	info, err := os.Stat(filename)

	if err != nil {
		return 0, err
	}

	return info.Size(), nil
}

func Write(filename string, content []byte, append bool) error {
	if append {
		fd, err := os.OpenFile(filename, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)

		if err != nil {
			return err
		}

		defer fd.Close()

		n, err := fd.Write(content)

		if err != nil {
			return err
		}

		if n < len(content) {
			return io.ErrShortWrite
		}

		return err
	} else {
		return ioutil.WriteFile(filename, content, 0666)
	}
}

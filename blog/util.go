package blog

import (
	"io"
	"io/ioutil"
	"os"
	"path"
)

func walkFiles(dir string, visit func(file os.FileInfo) error) error {
	files, err := ioutil.ReadDir(dir)
	if err != nil {
		return err
	}

	for _, file := range files {
		if !file.IsDir() {
			if err = visit(file); err != nil {
				return err
			}
		}
	}

	return nil
}

func mkdir(dir string) error {
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		return os.Mkdir(dir, 0777)
	}

	return nil
}

func copyDir(dst string, src string) error {
	files, err := ioutil.ReadDir(src)
	if err != nil {
		return err
	}

	for _, file := range files {
		if file.IsDir() {
			dstDir := path.Join(dst, file.Name())
			srcDir := path.Join(src, file.Name())

			if err = os.Mkdir(path.Join(dst, file.Name()), file.Mode()); err != nil {
				return err
			}

			if err = copyDir(dstDir, srcDir); err != nil {
				return err
			}
		} else {
			dstFile, err := os.OpenFile(path.Join(dst, file.Name()), os.O_WRONLY|os.O_CREATE, file.Mode())
			if err != nil {

				return err
			}

			srcFile, err := os.Open(path.Join(src, file.Name()))
			if err != nil {
				dstFile.Close()
				return err
			}

			if _, err = io.Copy(dstFile, srcFile); err != nil {
				dstFile.Close()
				srcFile.Close()
				return err
			}

			dstFile.Close()
			srcFile.Close()
		}
	}

	return nil
}

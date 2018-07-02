package blog

import (
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
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
			dstDir := filepath.Join(dst, file.Name())
			srcDir := filepath.Join(src, file.Name())

			if err = os.Mkdir(filepath.Join(dst, file.Name()), file.Mode()); err != nil {
				return err
			}

			if err = copyDir(dstDir, srcDir); err != nil {
				return err
			}
		} else {
			dstFile, err := os.OpenFile(filepath.Join(dst, file.Name()), os.O_WRONLY|os.O_CREATE, file.Mode())
			if err != nil {

				return err
			}

			srcFile, err := os.Open(filepath.Join(src, file.Name()))
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

func clearDir(dir string) error {
	file, err := os.Open(dir)
	if err != nil {
		return err
	}
	defer file.Close()

	names, err := file.Readdirnames(-1)
	if err != nil {
		return err
	}

	for _, name := range names {
		err = os.RemoveAll(filepath.Join(dir, name))
		if err != nil {
			return err
		}
	}

	return nil
}

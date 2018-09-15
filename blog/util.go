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

func mergeFiles(w io.Writer, files []string) error {
	for _, src := range files {
		srcFile, err := os.Open(src)
		if err != nil {
			return err
		}

		if _, err = io.Copy(w, srcFile); err != nil {
			srcFile.Close()
			return err
		}

		srcFile.Close()
	}

	return nil
}

func copyFileOrDir(dst string, src string) error {
	info, err := os.Stat(src)
	if err != nil {
		return err
	}

	if info.IsDir() {
		if err = os.MkdirAll(dst, info.Mode()); err != nil {
			return err
		}
		err = copyDir(dst, src)
	} else {
		if err = os.MkdirAll(filepath.Dir(dst), 0777); err != nil {
			return err
		}
		err = copyFile(dst, src, info.Mode())
	}

	return err
}

func copyFile(dst string, src string, mode os.FileMode) error {
	dstFile, err := os.OpenFile(dst, os.O_WRONLY|os.O_CREATE, mode)
	if err != nil {
		return err
	}
	defer dstFile.Close()

	srcFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer srcFile.Close()

	_, err = io.Copy(dstFile, srcFile)
	return err
}

func copyDir(dst string, src string) error {
	files, err := ioutil.ReadDir(src)
	if err != nil {
		return err
	}

	for _, file := range files {
		dst := filepath.Join(dst, file.Name())
		src := filepath.Join(src, file.Name())

		if file.IsDir() {
			if err = os.Mkdir(dst, file.Mode()); err != nil {
				return err
			}

			if err = copyDir(dst, src); err != nil {
				return err
			}
		} else {
			if err = copyFile(dst, src, file.Mode()); err != nil {
				return err
			}
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

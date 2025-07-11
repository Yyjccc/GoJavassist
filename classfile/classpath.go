package classfile

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

type ClassPath struct {
	entries []Entry
}

//func Parse(opts *vm.Options) *ClassPath {
//	cp := &ClassPath{}
//	cp.parseBootAndExtClassPath(opts.AbsJavaHome)
//	cp.parseUserClassPath(opts.ClassPath)
//	return cp
//}

func (cp *ClassPath) parseBootAndExtClassPath(absJavaHome string) {
	// jre/lib/*
	jreLibPath := filepath.Join(absJavaHome, "lib", "*")
	cp.entries = append(cp.entries, spreadWildcardEntry(jreLibPath)...)

	// jre/lib/ext/*
	jreExtPath := filepath.Join(absJavaHome, "lib", "ext", "*")
	cp.entries = append(cp.entries, spreadWildcardEntry(jreExtPath)...)
}

func (cp *ClassPath) parseUserClassPath(cpOption string) {
	if cpOption == "" {
		cpOption = "."
	}
	cp.entries = append(cp.entries, parsePath(cpOption)...)
}

// className: fully/qualified/ClassName
//func (cp *ClassPath) ReadClass(className string) (Entry, []byte) {
//	className = className + ".class"
//	for _, entry := range cp.entries {
//		if data, err := entry.readClass(className); err == nil {
//			return entry, data
//		}
//	}
//	return nil, nil
//}

func IsBootClassPath(entry Entry, absJreLib string) bool {
	if entry == nil {
		// todo
		return true
	}

	return strings.HasPrefix(entry.String(), absJreLib)
}

type Entry interface {
	// className: fully/qualified/ClassName.class
	String() string
}

func parsePath(path string) []Entry {
	switch {
	case strings.IndexByte(path, os.PathListSeparator) >= 0:
		return splitPath(path)
	case strings.HasSuffix(path, "*"):
		return spreadWildcardEntry(path)
	default:
		return []Entry{NewDirEntry(path)}
	}
}

func splitPath(pathList string) []Entry {
	list := make([]Entry, 0, 4)

	for _, path := range strings.Split(pathList, string(os.PathListSeparator)) {
		list = append(list, parsePath(path)...)
	}

	return list
}

func spreadWildcardEntry(path string) []Entry {
	baseDir := path[:len(path)-1] // remove *
	files, err := ioutil.ReadDir(baseDir)
	if err != nil {
		panic(err) // TODO
	}

	list := make([]Entry, 0, 4)
	for _, file := range files {

		filename := filepath.Join(baseDir, file.Name())
		list = append(list, NewZipEntry(filename))

	}

	return list
}

type DirEntry struct {
	absPath string
}

func NewDirEntry(path string) *DirEntry {
	return &DirEntry{absPath: path}
}

func (entry *DirEntry) String() string {
	return entry.absPath
}

func (entry *ZipEntry) String() string {
	return entry.absPath
}

type ZipEntry struct {
	absPath string
	//rc      *zip.ReadCloser
}

func NewZipEntry(path string) *ZipEntry {
	return &ZipEntry{absPath: path}
}

package reflect

import (
	"GoJavassist/classfile"
	"archive/zip"
	"bytes"
	"os"
	"path/filepath"

	"io"
	"runtime"
	"strings"
)

type Classpath struct {
	read       bool
	path       string
	Jars       []*Jar
	classFiles []string
}

func NewClasspath(path string) (*Classpath, error) {
	classpath := &Classpath{
		path:       path,
		Jars:       make([]*Jar, 0),
		classFiles: make([]string, 0),
	}
	info, err := os.Stat(path)
	if err != nil {
		return nil, err
	}
	// 如果是文件，直接判断后缀
	if !info.IsDir() {
		if strings.HasSuffix(path, ".jar") {
			classpath.Jars = append(classpath.Jars, NewJar(path))
		}
		if strings.HasSuffix(path, ".class") {
			classpath.classFiles = append(classpath.classFiles, path)
		}
		return classpath, nil
	}
	// 如果是目录，则递归查找
	err = filepath.WalkDir(path, func(p string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() {
			if strings.HasSuffix(p, ".jar") {
				classpath.Jars = append(classpath.Jars, NewJar(p))
			}
			if strings.HasSuffix(p, ".class") {
				classpath.classFiles = append(classpath.classFiles, p)
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return classpath, nil
}

type Jar struct {
	loaded      bool
	path        string
	cleared     bool
	rawDataPool map[string][]byte
	entries     []classfile.ClassFile
}

func NewJar(path string) *Jar {
	return &Jar{
		path:        path,
		loaded:      false,
		entries:     []classfile.ClassFile{},
		rawDataPool: make(map[string][]byte),
	}
}

func (j *Jar) Read() []classfile.ClassFile {
	j.loaded = true
	reader, err := zip.OpenReader(j.path)
	if err != nil {
		return make([]classfile.ClassFile, 0)
	}
	defer reader.Close()

	for _, file := range reader.File {
		// 只处理 .class 文件
		if strings.HasSuffix(file.Name, ".class") {
			rc, err := file.Open()
			if err != nil {
				return make([]classfile.ClassFile, 0)
			}
			defer rc.Close()

			// 读取 class 文件内容
			buf := new(bytes.Buffer)
			_, err = io.Copy(buf, rc)
			if err != nil {
				return j.entries
			}
			data := buf.Bytes()
			cf, err := classfile.Parse(data)
			if err != nil {
				return j.entries
			}
			j.entries = append(j.entries, *cf)
			j.rawDataPool[cf.GetThisClassName()] = data
		}
	}
	return j.entries
}

func (j *Jar) GetRawData(name string) []byte {
	if j == nil {
		return nil
	}
	return j.rawDataPool[name]
}

// Clear 清空数据区，减少内存消耗
func (j *Jar) Clear() {
	j.rawDataPool = make(map[string][]byte)
	j.entries = make([]classfile.ClassFile, 0)
	j.loaded = true
	runtime.GC()
}

// ReadJar 解析 JAR 文件并读取所有 class 文件
func ReadJar(jarPath string) (*Jar, error) {
	reader, err := zip.OpenReader(jarPath)
	if err != nil {
		return nil, err
	}
	defer reader.Close()
	jar := &Jar{
		path:    jarPath,
		entries: make([]classfile.ClassFile, 0),
	}

	for _, file := range reader.File {
		// 只处理 .class 文件
		if strings.HasSuffix(file.Name, ".class") {
			rc, err := file.Open()
			if err != nil {
				return nil, err
			}
			defer rc.Close()

			// 读取 class 文件内容
			buf := new(bytes.Buffer)
			_, err = io.Copy(buf, rc)
			if err != nil {
				return nil, err
			}
			cf, err := classfile.Parse(buf.Bytes())
			if err != nil {
				return nil, err
			}
			jar.entries = append(jar.entries, *cf)
		}
	}
	return jar, nil
}

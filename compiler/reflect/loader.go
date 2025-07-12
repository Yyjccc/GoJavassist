package reflect

import (
	"github.com/Yyjccc/GoJavassist/classfile"
	"github.com/Yyjccc/GoJavassist/compiler/lib"
	"io/ioutil"
	"strings"
)

var bootLoader *ClassLoader
var DefaultPool *ClassPool

func init() {
	bootLoader = NewClassLoader(nil)
	libClassPath, err := NewClasspathFromFS(lib.RtJar, "rt.jar")
	if err == nil {
		bootLoader.AddClassPath(libClassPath)
		bootLoader.LoadClassPaths()
	}
	DefaultPool = NewClassPool(nil)
}

type ClassLoader struct {
	LoadedClasses map[string]classfile.ClassFile
	ClassPath     []*Classpath
	RawDataPool   map[string][]byte
	Parent        *ClassLoader
}

func (l *ClassLoader) AddClassPath(Classpath *Classpath) *ClassLoader {
	l.ClassPath = append(l.ClassPath, Classpath)
	return l
}

func NewClassLoader(parent *ClassLoader) *ClassLoader {
	return &ClassLoader{
		LoadedClasses: make(map[string]classfile.ClassFile),
		Parent:        parent,
		ClassPath:     make([]*Classpath, 0),
		RawDataPool:   make(map[string][]byte),
	}
}

func (l *ClassLoader) LoadClassPaths() {
	for _, path := range l.ClassPath {
		if !path.read {
			for _, jar := range path.Jars {
				var classes []classfile.ClassFile
				if path.isEmbedded {
					classes = jar.ReadFromFS(path.fs)
				} else {
					classes = jar.Read()
				}
				for _, class := range classes {
					name := class.GetThisClassName()
					l.LoadedClasses[name] = class
					l.RawDataPool[name] = jar.GetRawData(name)
				}
				jar.Clear()
			}
			for _, file := range path.classFiles {
				data, err := ioutil.ReadFile(file)
				if err != nil {
					continue
				}
				l.RawDataPool[file] = data
				cf, err := classfile.Parse(data)
				if err != nil {
					continue
				}
				l.LoadedClasses[cf.GetThisClassName()] = *cf
			}
			path.read = true
		}
	}
}

// LoadClass fullName is jvm full class name
func (l *ClassLoader) LoadClass(fullName string) *classfile.ClassFile {
	if class, ok := l.LoadedClasses[fullName]; ok {
		return &class
	}
	if l.Parent != nil {
		return l.Parent.LoadClass(fullName)
	}
	return nil
}

func (l *ClassLoader) GetClasses() []*classfile.ClassFile {
	var classes []*classfile.ClassFile
	for _, class := range l.LoadedClasses {
		classes = append(classes, &class)
	}
	return classes
}

func ClassForName(javaClassName string) *classfile.ClassFile {
	return bootLoader.LoadClass(strings.ReplaceAll(javaClassName, ".", "/"))
}

func GetBootLoader() *ClassLoader {
	return bootLoader
}

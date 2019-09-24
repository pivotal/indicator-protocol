// Code generated for package asset by go-bindata DO NOT EDIT. (@generated)
// sources:
// schemas.yml
package asset

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"time"
)

func bindataRead(data []byte, name string) ([]byte, error) {
	gz, err := gzip.NewReader(bytes.NewBuffer(data))
	if err != nil {
		return nil, fmt.Errorf("Read %q: %v", name, err)
	}

	var buf bytes.Buffer
	_, err = io.Copy(&buf, gz)
	clErr := gz.Close()

	if err != nil {
		return nil, fmt.Errorf("Read %q: %v", name, err)
	}
	if clErr != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

type asset struct {
	bytes []byte
	info  os.FileInfo
}

type bindataFileInfo struct {
	name    string
	size    int64
	mode    os.FileMode
	modTime time.Time
}

// Name return file name
func (fi bindataFileInfo) Name() string {
	return fi.name
}

// Size return file size
func (fi bindataFileInfo) Size() int64 {
	return fi.size
}

// Mode return file mode
func (fi bindataFileInfo) Mode() os.FileMode {
	return fi.mode
}

// Mode return file modify time
func (fi bindataFileInfo) ModTime() time.Time {
	return fi.modTime
}

// IsDir return file whether a directory
func (fi bindataFileInfo) IsDir() bool {
	return fi.mode&os.ModeDir != 0
}

// Sys return file is sys mode
func (fi bindataFileInfo) Sys() interface{} {
	return nil
}

var _schemasYml = []byte("\x1f\x8b\x08\x00\x00\x00\x00\x00\x00\xff\xac\x55\x51\x8f\x83\x36\x0c\x7e\xe7\x57\x58\xea\xa4\x4a\x13\xb4\xbb\xc7\xf1\x36\x69\x2f\x93\xf6\x76\xdb\x3d\xec\x54\x5d\x43\x70\x21\x6b\x88\xd3\xc4\xf4\xc6\x7e\xfd\x14\xa0\x1c\x2d\xb4\xf4\xa4\x21\x55\x6a\x8c\x6d\xfc\x7d\xfe\xec\xac\xe0\x4f\x8f\xb0\x2f\x28\xc9\x94\xc9\x05\x0b\x48\x08\xec\xb1\xd8\x0a\xef\x91\xb7\x5e\x96\x58\x89\x4d\x41\x90\xd8\x63\x01\xad\x11\x3a\xa3\xdf\x34\x95\xde\xc3\xc1\x51\x05\x5c\x22\x38\xb4\x04\x8e\x88\x81\x09\xb2\xda\xe4\x1a\xa3\x15\x70\xa9\x3c\xb4\x79\x95\x61\x02\x01\x05\xc1\x41\x69\x04\x4f\xc0\xa5\x60\x50\x0c\x52\x18\xc8\x10\x94\x91\xba\xce\x31\x07\x65\x00\xff\x41\x59\xb3\xc8\x34\xfa\x4d\x94\x24\x49\xf4\x9b\xc9\x95\x14\x4c\xee\x57\x92\x75\x85\x86\xd3\x08\x80\x1b\x8b\x29\x50\xf6\x37\x4a\x8e\x00\x1c\x9e\x6a\xe5\x30\x0f\xaf\x12\x10\x56\xbd\xa1\xf3\x8a\x4c\x7b\x3c\x2a\x93\xb7\x7f\xbc\x45\x19\x01\x58\x47\x16\x1d\x2b\xf4\xc1\x1d\x46\xee\xdd\xf9\x92\xdd\xb3\x53\xa6\xe8\x4d\x68\xea\x2a\x85\x77\x75\x29\xc6\x3a\x62\x92\xa4\x37\x8a\xb6\xe7\x97\x5d\xeb\x15\x3e\xb4\x9c\x62\x82\xa7\x0b\xae\x90\x45\x60\xeb\x3a\xc1\x80\x30\x3c\xb7\x85\x87\x47\x8b\x0c\xf5\xe8\x7c\x1d\x08\xab\x40\xb1\x21\x06\x49\x86\x85\x32\xb0\xf7\x8c\x76\x0f\x47\x6c\xda\x90\x40\xc9\x25\xf8\x07\x87\x87\x14\xd6\xab\xed\xa4\xc2\x57\x8b\x72\x3d\x6d\xc4\x6b\x1f\xfc\xb0\x19\xd6\x51\x5e\xb7\xf6\xdb\xf2\xfb\x37\x0f\x00\x8f\x33\x85\x27\x01\x23\x2a\x1c\x0e\xe7\xa1\xc9\xf7\xd8\x09\xee\x53\x6e\xae\xba\xd2\x72\xaf\xcc\xef\x68\x0a\x2e\x53\x78\x19\xcc\xe7\x6b\x4d\x7c\x27\x7c\x50\x89\xbf\x06\x27\x9c\x13\x4d\x6f\x51\x8c\xd5\xa8\xd2\x29\xf9\x1d\xe9\x5d\x8f\x1b\xaa\x1f\xf1\x34\x07\x9d\x3e\x0d\xba\xc5\xe2\x59\xb1\x5e\x66\x28\x47\x2f\x9d\xb2\xfc\x0c\x1d\x1e\x65\xf0\x9b\x91\xe4\x18\xfd\x0c\x03\x77\xb0\xdd\x47\x78\x07\xc1\xc3\x56\x3d\x40\xb3\x18\x37\xed\xea\x12\xc2\xbb\x38\xa7\xdf\x82\x15\x18\x82\x4f\xd1\x84\x1d\x7a\x16\x5a\xe5\x82\x31\x6c\x57\x8f\xd1\x95\x24\x16\x07\xae\x1f\x91\x76\xf2\xaa\x93\x9e\x19\xbc\xf1\x54\xcc\xe0\xb5\x82\x19\x9d\x49\x61\xfd\x2e\x92\x7f\x7f\x49\xfe\xfa\x48\x77\xfd\xbf\x9f\x92\x9f\x3f\xd2\xdd\x8f\xeb\xcb\xfc\x56\x27\xfd\x20\xd1\x64\x2e\x5a\x9f\xc5\x05\x79\xb4\x2a\x06\xaf\x55\x0c\xc4\x25\xba\x6e\x3d\x72\xe9\xd0\x97\xa4\xf3\xe7\x47\x6a\x56\x4c\xb7\x2b\x25\x10\xa5\xf1\x8c\x7a\x74\x0e\x6c\x05\xba\x47\xa6\xb3\xd0\x35\x0e\xe7\x79\x35\xb6\x69\xe6\xf4\x3c\x51\xd3\xe5\x03\x4f\x39\x0f\xbc\x68\x8e\x41\x33\xc6\x50\x70\xf8\x61\x0c\x78\x8a\xc1\xe0\x69\x37\xf2\x6e\x0b\x9d\xcb\x6b\xea\x2a\xc3\x0e\x92\xd0\xe8\xbe\xbb\x4e\x0e\x74\x77\x99\xb4\xd2\x65\x90\xb5\x73\x68\x58\x37\x83\x7c\x73\x10\x1e\xf2\xda\x09\xfe\xda\xd2\x00\xe1\x02\xfa\x1f\x52\x59\x87\x1e\x0d\x0b\x9e\xdc\xdb\x8b\x58\x64\x29\x1c\xff\x31\xd2\xe2\x03\xf6\x7b\xee\x43\xd1\x31\x64\xc2\xc5\xe0\x59\x70\xed\x63\x38\xd5\xc4\xe2\x8b\xfa\xbe\xe6\xb7\x5b\xfe\xbb\xbc\x19\x91\x46\xf1\x45\xc2\x21\xe8\x10\x8d\x6c\xa6\xae\xca\x30\x16\xe8\x16\x6f\xf8\xa7\xd7\xe9\x08\x53\xde\xdf\xdc\xf7\x59\x8b\xfe\x0b\x00\x00\xff\xff\x81\x0d\x8e\x94\x17\x0a\x00\x00")

func schemasYmlBytes() ([]byte, error) {
	return bindataRead(
		_schemasYml,
		"schemas.yml",
	)
}

func schemasYml() (*asset, error) {
	bytes, err := schemasYmlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "schemas.yml", size: 2583, mode: os.FileMode(420), modTime: time.Unix(1569349547, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

// Asset loads and returns the asset for the given name.
// It returns an error if the asset could not be found or
// could not be loaded.
func Asset(name string) ([]byte, error) {
	cannonicalName := strings.Replace(name, "\\", "/", -1)
	if f, ok := _bindata[cannonicalName]; ok {
		a, err := f()
		if err != nil {
			return nil, fmt.Errorf("Asset %s can't read by error: %v", name, err)
		}
		return a.bytes, nil
	}
	return nil, fmt.Errorf("Asset %s not found", name)
}

// MustAsset is like Asset but panics when Asset would return an error.
// It simplifies safe initialization of global variables.
func MustAsset(name string) []byte {
	a, err := Asset(name)
	if err != nil {
		panic("asset: Asset(" + name + "): " + err.Error())
	}

	return a
}

// AssetInfo loads and returns the asset info for the given name.
// It returns an error if the asset could not be found or
// could not be loaded.
func AssetInfo(name string) (os.FileInfo, error) {
	cannonicalName := strings.Replace(name, "\\", "/", -1)
	if f, ok := _bindata[cannonicalName]; ok {
		a, err := f()
		if err != nil {
			return nil, fmt.Errorf("AssetInfo %s can't read by error: %v", name, err)
		}
		return a.info, nil
	}
	return nil, fmt.Errorf("AssetInfo %s not found", name)
}

// AssetNames returns the names of the assets.
func AssetNames() []string {
	names := make([]string, 0, len(_bindata))
	for name := range _bindata {
		names = append(names, name)
	}
	return names
}

// _bindata is a table, holding each asset generator, mapped to its name.
var _bindata = map[string]func() (*asset, error){
	"schemas.yml": schemasYml,
}

// AssetDir returns the file names below a certain
// directory embedded in the file by go-bindata.
// For example if you run go-bindata on data/... and data contains the
// following hierarchy:
//     data/
//       foo.txt
//       img/
//         a.png
//         b.png
// then AssetDir("data") would return []string{"foo.txt", "img"}
// AssetDir("data/img") would return []string{"a.png", "b.png"}
// AssetDir("foo.txt") and AssetDir("notexist") would return an error
// AssetDir("") will return []string{"data"}.
func AssetDir(name string) ([]string, error) {
	node := _bintree
	if len(name) != 0 {
		cannonicalName := strings.Replace(name, "\\", "/", -1)
		pathList := strings.Split(cannonicalName, "/")
		for _, p := range pathList {
			node = node.Children[p]
			if node == nil {
				return nil, fmt.Errorf("Asset %s not found", name)
			}
		}
	}
	if node.Func != nil {
		return nil, fmt.Errorf("Asset %s not found", name)
	}
	rv := make([]string, 0, len(node.Children))
	for childName := range node.Children {
		rv = append(rv, childName)
	}
	return rv, nil
}

type bintree struct {
	Func     func() (*asset, error)
	Children map[string]*bintree
}

var _bintree = &bintree{nil, map[string]*bintree{
	"schemas.yml": &bintree{schemasYml, map[string]*bintree{}},
}}

// RestoreAsset restores an asset under the given directory
func RestoreAsset(dir, name string) error {
	data, err := Asset(name)
	if err != nil {
		return err
	}
	info, err := AssetInfo(name)
	if err != nil {
		return err
	}
	err = os.MkdirAll(_filePath(dir, filepath.Dir(name)), os.FileMode(0755))
	if err != nil {
		return err
	}
	err = ioutil.WriteFile(_filePath(dir, name), data, info.Mode())
	if err != nil {
		return err
	}
	err = os.Chtimes(_filePath(dir, name), info.ModTime(), info.ModTime())
	if err != nil {
		return err
	}
	return nil
}

// RestoreAssets restores an asset under the given directory recursively
func RestoreAssets(dir, name string) error {
	children, err := AssetDir(name)
	// File
	if err != nil {
		return RestoreAsset(dir, name)
	}
	// Dir
	for _, child := range children {
		err = RestoreAssets(dir, filepath.Join(name, child))
		if err != nil {
			return err
		}
	}
	return nil
}

func _filePath(dir, name string) string {
	cannonicalName := strings.Replace(name, "\\", "/", -1)
	return filepath.Join(append([]string{dir}, strings.Split(cannonicalName, "/")...)...)
}

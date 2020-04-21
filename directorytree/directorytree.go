package directorytree

import (
//	"errors"
	"fmt"
	"os"
	"path"
	"strings"

	"github.com/imtlab/pkg/loggers"
)

/*
	Package directorytree builds a hierarchical tree of directory names from a list of directory paths having a common root.
	It also provides a recursive function for creating each directory as needed under another root.
*/

type DirectoryTree struct {
	Children map[string]*DirectoryTree
}

func (pRoot *DirectoryTree) AddPath(dirPath string) {
	if "/" == dirPath {
		return
	}

	names := strings.Split(dirPath, "/")
	pParent := pRoot	//	start at root

	for _, folderName := range names {
		var pChild *DirectoryTree
		if nil == pParent.Children {
			pParent.Children = make(map[string]*DirectoryTree)

			pChild = new(DirectoryTree)
			pParent.Children[folderName] = pChild
		} else {
			//	does a child node already exist representing this folderName?
			var ok bool
			if pChild, ok = pParent.Children[folderName]; !ok {
				pChild = new(DirectoryTree)
				pParent.Children[folderName] = pChild
			}
		}
		pParent = pChild	//	prepare for next level down
	}
}

func (pRoot *DirectoryTree) Build(xDirPath []string) {
	for _, dirPath := range xDirPath {
		pRoot.AddPath(dirPath)
	}
}

/*	Findings re: FileMode passed to os.MkdirAll() and os.Mkdir()
-	On macOS, despite specifying os.ModePerm (0777) all directories were created with permissions 0755.
	When I executed the equivalent "mkdir -p -m 0777 $path" on the FlexSync server,
	only the final directory in $path were created with permissions 0777,
	while all parent directories were were created with permissions 0755.

	So, I'll have to blow this off and create directories one at a time using recurseMkdirWhereNotExists().
-	Nope, on macOS I got the same result - all directories created with permissions 0755.
	I now suspect the passed permissions are masked by the permissions of the root directory which is 0755.

	I'll set its permissions to 0777, and retry recurseMkdirAllAtLeaves().
-	Nope, on macOS I still got the same result - all directories created with 0755.

	One last try with recurseMkdirWhereNotExists() just to leave no stone unturned.
-	Same result.

-	Ah-ha!  Here's the key phrase from https://golang.org/pkg/os/#Mkdir explaining it all:
		Mkdir creates a new directory with the specified name and permission bits (before umask)
	My macOS umask is set to 0022, which explains why 0777 is masked to 0755.
	Q:	What is the umask on the FlexSync server?
	A:	MDC 1 & 2:	0002
		FlexSync:	0022

	So probably my only recourse is to call os.Chmod() after each directory is created in recurseMkdirWhereNotExists().
	(There is so equivalent to "chmod -R".)  Chmod does not apply umask.
-	Success!
*/
/*
func RecurseMkdirAllAtLeaves(pParent *DirectoryTree, parentPath string) (err error) {
	if nil == pParent.Children {
		var finfo os.FileInfo
		if finfo, err = os.Stat(parentPath); nil == err {
			if !finfo.IsDir() {
				err = fmt.Errorf("\"%v\" is not a directory", finfo)
				loggers.Error.Println(err)
				return
			}
		} else {
			if os.IsNotExist(err) {
				//	create the directory
				if err = os.MkdirAll(parentPath, os.ModePerm); nil != err {	//	ModePerm FileMode = 0777 // Unix permission bits
					loggers.Error.Println(err)
					return
				}
			}
		}
	} else {
		for folderName, pChild := range pParent.Children {
			err = RecurseMkdirAllAtLeaves(pChild, path.Join(parentPath, folderName))	//	resurse
		}
	}
	return
}
*/
func RecurseMkdirWhereNotExists(pParent *DirectoryTree, parentPath string) (err error) {
	//	check existence of parentPath (and make sure it's a directory). create it if it ain't there
	var finfo os.FileInfo
	if finfo, err = os.Stat(parentPath); nil == err {
		if !finfo.IsDir() {
			err = fmt.Errorf("\"%v\" is not a directory", finfo)
		}
	} else if os.IsNotExist(err) {
		//	create the directory
		if err = os.Mkdir(parentPath, os.ModePerm); nil == err {	//	ModePerm FileMode = 0777 // Unix permission bits
			//	because Mkdir doesn't give us the mode we specified due to umask...
			err = os.Chmod(parentPath, os.ModePerm)
		} else {
			/*	watch out for race condition: it's possible (and has happened) that another worker created
				this directory between the time os.Stat() said it didn't exist and os.Mkdir() was executed.
			*/
//			if "file exists" == err.(*os.PathError).Err.Error() {	//	type assertion
			if os.IsExist(err) {
				err = nil	//	don't consider this an error
			}
		}
	}

	if nil != err {
		loggers.Error.Println(err)
		return
	}

	if nil != pParent.Children {
		for folderName, pChild := range pParent.Children {
			if err = RecurseMkdirWhereNotExists(pChild, path.Join(parentPath, folderName)); nil != err {	//	resurse
				break
			}
		}
	}
	return
}


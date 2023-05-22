package m

import (
	"fmt"

	"github.com/andreyvit/buddyd/mvp/flake"
)

type FolderID = flake.ID

type Folder struct {
	ID         FolderID   `msgpack:"-"`
	AccountID  AccountID  `msgpack:"a"`
	Name       string     `msgpack:"n"`
	Slug       string     `msgpack:"s"`
	ParentID   FolderID   `msgpack:"p"`
	ChildenIDs []FolderID `msgpack:"c"`
}

func (fldr *Folder) IsRoot() bool {
	return fldr.ParentID == 0
}

func (fldr *Folder) FirstLetter() string {
	if fldr.Name == "" {
		return ""
	} else {
		return fldr.Name[:1] // TODO: unicode
	}
}

func (fldr *Folder) SemanticPath() string {
	if fldr.IsRoot() {
		return "lib/folders/root"
	} else {
		return fmt.Sprintf("lib/folders/%v", fldr.ID)
	}
}

type AccountLibrary struct {
	RootFolderID  FolderID
	Folders       map[FolderID]*Folder
	FoldersBySlug map[string]*Folder
}

func (lib *AccountLibrary) RootFolder() *Folder {
	if lib.RootFolderID == 0 {
		return nil
	}
	return lib.Folders[lib.RootFolderID]
}

func (lib *AccountLibrary) Folder(id FolderID) *Folder {
	if id == 0 {
		return nil
	}
	return lib.Folders[id]
}

func (lib *AccountLibrary) FolderBySlug(slug string) *Folder {
	return lib.FoldersBySlug[slug]
}

type FolderWithItemsVM struct {
	*Folder
	Subfolders []*Folder
	Items      []*Item
}

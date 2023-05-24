package main

import (
	"github.com/andreyvit/buddyd/internal/flogger"
	m "github.com/andreyvit/buddyd/model"
	"github.com/andreyvit/edb"
	"golang.org/x/exp/slices"
)

func loadAccountEmbeddings(rc *RC, reload bool) {
	if rc.Embeddings != nil && !reload {
		return
	}
	embs := new(m.AccountEmbeddings)
	embs.Embeddings = edb.All(edb.ExactIndexScan[m.ContentEmbedding](rc, EmbeddingsByAccountType, m.ContentEmbeddingAccountTypeKey{
		AccountID: rc.AccountID(),
		Type:      m.CurrentEmbeddingType,
	}))
	rc.Embeddings = embs
	flogger.Log(rc, "Loaded %d embeddings", len(rc.Embeddings.Embeddings))
}

func deleteContentByItem(rc *RC, itemID m.ItemID) {
	edb.DeleteAll(rc.DBTx().IndexScan(ContentByIRO, edb.ExactScan(m.ContentIROKey{ItemID: itemID}).Prefix(1)))
	edb.DeleteAll(rc.DBTx().IndexScan(EmbeddingsByItem, edb.ExactScan(itemID)))
}

func deleteItem(rc *RC, itemID m.ItemID) {
	deleteContentByItem(rc, itemID)
	rc.DBTx().DeleteByKey(Items, itemID)
}

func ensureFolderBySlug(rc *RC, slug, name string, parentFolderID m.FolderID) *m.Folder {
	fldr := rc.Library.FolderBySlug(slug)
	if fldr == nil {
		fldr = &m.Folder{
			ID:        rc.App().NewID(),
			AccountID: rc.AccountID(),
			Name:      name,
			Slug:      slug,
			ParentID:  parentFolderID,
		}
		saveFolder(rc, fldr)
	}
	return fldr
}

func saveFolder(rc *RC, fldr *m.Folder) {
	edb.Put(rc, fldr)
	updateFolderParent(rc, fldr)
	if rc.Library.Folder(fldr.ID) == nil {
		rc.Library.AddFolder(fldr)
	}
}

func updateFolderParent(rc *RC, fldr *m.Folder) {
	parent := rc.Library.Folder(fldr.ParentID)
	if !slices.Contains(parent.ChildenIDs, fldr.ID) {
		parent.ChildenIDs = append(parent.ChildenIDs, fldr.ID)
		edb.Put(rc, parent)
	}
}
